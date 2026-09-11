package engine

// The task operations: add, list, show, edit and delete a task, described
// once so both surfaces build the same command from the same facts.

import (
	"errors"
	"fmt"
	"time"

	"cairn/internal/domain"
	"cairn/internal/store"
	"cairn/internal/task"
)

// Name is how an operation is named as a command or a tool: "task_add".
func (o Operation) Name() string { return o.Thing + "_" + o.Verb }

// Line is how an operation reads on the command line: "task add".
func (o Operation) Line() string { return o.Thing + " " + o.Verb }

// All is everything cairn can be asked to do, in a stable order.
func All() []Operation {
	return []Operation{
		TaskAdd, TaskList, TaskShow, TaskEdit, TaskDelete, TaskSearch,
		TaskMove, TaskFocus, TaskUnfocus,
		ObjectiveAdd, ObjectiveList, ObjectiveShow, ObjectiveEdit, ObjectiveDelete, ObjectiveSearch,
		BackupAdd, BackupList, BackupShow, BackupDelete, BackupRestore,
		MigrationList, MigrationShow, MigrationApply, MigrationRollback,
		TraceShow, TraceList,
		ConfigShow,
	}
}

// TaskAdd records a new task.
var TaskAdd = Operation{
	Thing:   "task",
	Verb:    "add",
	Short:   "Add a task",
	Does:    "Add a task, carrying whatever of its attributes were given.",
	Level:   Write,
	Needs:   AStore,
	Answers: Task{},
	Takes: append([]Argument{
		{Name: "title", Kind: Text, Positional: true, Required: true, Means: "what the task is called"},
	}, taskAttributes...),
	Run: func(on On, in Input) (Answer, error) {
		title := ""
		if given := in.Change("title"); given != nil {
			title = *given
		}
		changes, err := taskChanges(in)
		if err != nil {
			return Answer{}, err
		}
		added, err := on.Engine.Add(title, changes)
		if err != nil {
			return Answer{}, withTypedTitle(err, title)
		}
		// Nothing hangs under a task that has just been created, so the
		// progress is asked of nobody rather than of the store.
		return Answer{Task: &added, Progress: domain.Progress{}}, nil
	},
}

// TaskList names every task, or only the ones carrying what was named.
var TaskList = Operation{
	Thing:   "task",
	Verb:    "list",
	Short:   "List tasks, all of them or only some",
	Does:    "List every task, or only the ones carrying everything named. A task has to carry all of it, not one of it.",
	Level:   Read,
	Needs:   AStore,
	Answers: Tasks{},
	Takes:   taskNarrowing,
	Run:     listing,
}

// taskNarrowing are the facts a listing of tasks can be kept down to. The four
// that come from closed sets or from something numbered: what a person typed
// themselves is what `task search` looks in.
var taskNarrowing = narrowedBy(taskAttributes, map[string]string{
	"state":     "keep only the tasks sitting in this column",
	"priority":  "keep only the tasks carrying this priority",
	"label":     "keep only the tasks carrying this label; named twice, only those carrying both",
	"objective": "keep only the tasks serving this objective",
})

// TaskShow says what a task carries, what hangs under it and how far along it
// is.
var TaskShow = Operation{
	Thing:   "task",
	Verb:    "show",
	Short:   "Show what a task carries",
	Does:    "Show what a task carries, what hangs under it and how far along it is.",
	Level:   Read,
	Needs:   AStore,
	Answers: Task{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		shown, err := on.Engine.Show(n)
		if err != nil {
			return Answer{}, err
		}
		under, err := on.Engine.Under(n)
		if err != nil {
			return Answer{}, err
		}
		p, err := on.Engine.TaskProgress(n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Task: &shown, Under: under, Progress: p}, nil
	},
}

// TaskEdit changes what a task carries.
var TaskEdit = Operation{
	Thing:   "task",
	Verb:    "edit",
	Short:   "Change what a task carries",
	Does:    "Change what a task carries.",
	Level:   Write,
	Needs:   AStore,
	Answers: Task{},
	Takes: append([]Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
		{Name: "title", Kind: Text, Means: "what the task is called"},
	}, taskAttributes...),
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		changes, err := taskChanges(in)
		if err != nil {
			return Answer{}, err
		}
		changes.Title = in.Change("title")
		edited, err := on.Engine.Edit(n, changes)
		if err != nil {
			return Answer{}, editReason(err)
		}
		p, err := on.Engine.TaskProgress(n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Task: &edited, Progress: p}, nil
	},
}

// TaskDelete removes a task.
var TaskDelete = Operation{
	Thing:   "task",
	Verb:    "delete",
	Short:   "Delete a task",
	Does:    "Delete a task.",
	Level:   Full,
	Needs:   AStore,
	Answers: Deleted{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		if err := on.Engine.Delete(n); err != nil {
			return Answer{}, err
		}
		return Answer{Deleted: &n}, nil
	},
}

// TaskSearch names every task carrying everything named, wherever among its
// facts each one happens to sit.
var TaskSearch = Operation{
	Thing:   "task",
	Verb:    "search",
	Short:   "Look for a task by anything it carries",
	Does:    "List every task carrying everything named, wherever among its facts each one sits.",
	Level:   Read,
	Needs:   AStore,
	Answers: Tasks{},
	Takes: []Argument{
		{Name: "named", Kind: List, Positional: true, Required: true, Means: "what to look for, read the way it is shown"},
	},
	Run: func(on On, in Input) (Answer, error) {
		named := in.Changed("named")
		tasks, err := on.Engine.List()
		if err != nil {
			return Answer{}, err
		}
		found := []task.Task{}
		for _, t := range tasks {
			if named != nil && task.Matching(t, *named) {
				found = append(found, t)
			}
		}
		return Answer{Matched: found}, nil
	},
}

// taskAttributes are the twelve attributes a task carries, minus title, which is
// taken as a positional argument where it can be given at all.
var taskAttributes = []Argument{
	{Name: "desc", Kind: Text, Means: "what the task is about"},
	{Name: "state", Kind: Text, OneOf: oneOf(domain.States), Means: "the column it sits in"},
	{Name: "priority", Kind: Text, OneOf: oneOf(domain.Priorities), Means: "the priority it carries"},
	{Name: "label", Kind: List, Means: "a label it carries"},
	{Name: "blocker", Kind: List, Of: "a task", Means: "the number of a task in its way; +3 adds one, -3 takes one away"},
	{Name: "because", Kind: Text, Means: "what is in its way and is not a task cairn holds"},
	{Name: "deadline", Kind: Day, Means: "the day it is due"},
	{Name: "estimate", Kind: Duration, Means: "how long it should take"},
	{Name: "spent", Kind: Duration, Means: "the time already spent on it; +30m adds, -15m takes off"},
	{Name: "parent", Kind: Number, Of: "a task", Means: "the number of the task it hangs under"},
	{Name: "objective", Kind: List, Of: "an objective", Means: "the number of an objective it serves; +4 adds one, -1 takes one away"},
}

// oneOf turns a closed set of typed values into the strings Argument.OneOf
// needs, in one place rather than once per attribute that has one.
func oneOf[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = string(v)
	}
	return out
}

// withTypedTitle puts the title back in the reason when the title is what was
// wrong, so whoever asked sees what cairn read — including nothing at all.
func withTypedTitle(err error, typed string) error {
	if errors.Is(err, domain.ErrTitleRequired) {
		return fmt.Errorf("%w, got %q", err, typed)
	}
	return err
}

// editReason says what an edit can change, when it was asked to change nothing.
func editReason(err error) error {
	if errors.Is(err, ErrNothingToChange) {
		return fmt.Errorf("%w: name one of --title, --desc, --state, --priority, --label, --deadline, --estimate", err)
	}
	return err
}

// taskChanges turns what was named in in into the changes Add and
// Edit take.
func taskChanges(in Input) (Changes, error) {
	var c Changes
	c.Description = in.Change("desc")
	c.State = in.Change("state")
	c.Priority = in.Change("priority")
	c.Labels = in.Changed("label")
	c.Blocker = in.Changed("blocker")
	c.Because = in.Change("because")
	c.Deadline = in.Change("deadline")
	c.Estimate = in.Change("estimate")
	c.Spent = in.Change("spent")
	if in.Named("parent") {
		// Empty is how a task is set free, and Whole reads it as zero — the
		// same task.NoParent a number nobody could tell apart from "hang
		// under task zero" would otherwise have to spell out here.
		under, _ := in.Whole("parent")
		c.Parent = &under
	}
	// A set now, so it is read the way the labels and the blockers are rather
	// than as a single number: naming nothing sets the task loose, which the
	// grammar already handles without a case of its own.
	c.Objective = in.Changed("objective")
	return c, nil
}

// TaskMove changes where a task sits in the order.
var TaskMove = Operation{
	Thing:   "task",
	Verb:    "move",
	Short:   "Move a task in the order",
	Does:    "Move a task to another place among the ones it sits with. It never changes the column.",
	Level:   Write,
	Needs:   AStore,
	Answers: Tasks{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
		{Name: "first", Kind: Bool, Means: "put it at the head of the ones it sits with"},
		{Name: "after", Kind: Number, Of: "a task", Means: "put it directly behind this task"},
		{Name: "last", Kind: Bool, Means: "put it behind all the ones it sits with"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		at, target, err := placeFrom(in)
		if err != nil {
			return Answer{}, err
		}
		if err := on.Engine.Move(n, at, target); err != nil {
			return Answer{}, err
		}
		// A move changes the order and nothing else, so the order is what it
		// answers with: the card of the task moved would not show what moved.
		//
		// Answered as a listing nobody narrowed, and told so with an empty
		// Input rather than handed the move's own: what a move was given says
		// where to put a task, and reading it as facts to keep by would narrow
		// the answer by whatever a later argument happened to be called.
		return listing(on, Input{})
	},
}

// TaskFocus says which task is being worked on, or puts one there.
var TaskFocus = Operation{
	Thing:   "task",
	Verb:    "focus",
	Short:   "Put a task in focus, or say which one is",
	Does:    "Put a task in focus, or say which one is. One task is in focus at a time.",
	Level:   Write,
	Needs:   AStore,
	Answers: Focus{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a task", Means: "the number of the task being worked on"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, named := in.Whole("number")
		if !named {
			return whatIsInFocus(on)
		}
		t, err := on.Engine.PutInFocus(n)
		if err != nil {
			return Answer{}, err
		}
		p, err := on.Engine.TaskProgress(n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Task: &t, Progress: p, InFocus: &n}, nil
	},
}

// TaskUnfocus leaves nothing in focus.
var TaskUnfocus = Operation{
	Thing:   "task",
	Verb:    "unfocus",
	Short:   "Let go of the task in focus",
	Does:    "Leave nothing in focus. Letting go of nothing is not a failure.",
	Level:   Write,
	Needs:   AStore,
	Answers: Focus{},
	Run: func(on On, _ Input) (Answer, error) {
		if err := on.Engine.LetGo(); err != nil {
			return Answer{}, err
		}
		nothing := store.NoFocus
		return Answer{InFocus: &nothing}, nil
	},
}

// placeFrom reads where a move was told to put a task. Naming nothing and
// naming several are both refused: one is a command that asked for nothing,
// the other a command that asked for two things, and picking either quietly
// would move a task somewhere nobody said.
func placeFrom(in Input) (Place, int, error) {
	named := make([]Place, 0, 3)
	target := 0
	if in.Yes("first") {
		named = append(named, First)
	}
	if in.Named("after") {
		named = append(named, Behind)
		target, _ = in.Whole("after")
	}
	if in.Yes("last") {
		named = append(named, Last)
	}
	if len(named) == 0 {
		return 0, 0, ErrNowhereToMove
	}
	if len(named) > 1 {
		return 0, 0, ErrMoreThanOnePlace
	}
	return named[0], target, nil
}

// listing is what every listing answers: every task in the one order there is,
// and which of them is in focus, since a listing marks it.
func listing(on On, in Input) (Answer, error) {
	kept, err := keeping(in, taskNarrowing)
	if err != nil {
		return Answer{}, err
	}
	tasks, err := on.Engine.Narrowed(kept)
	if err != nil {
		return Answer{}, err
	}
	inFocus, err := on.Engine.Focus()
	if err != nil {
		return Answer{}, err
	}
	held := heldAmong(tasks, inFocus)
	free, err := on.Engine.NothingBlocks()
	if err != nil {
		return Answer{}, err
	}
	// A narrowed listing is answered flat: what matched is what is shown, and
	// where each task sits is not what was asked. Which field carries them is
	// how Of and RenderText tell the two apart.
	if kept.Narrows() {
		return Answer{Matched: tasks, InFocus: held, NothingBlocks: free}, nil
	}
	return Answer{Tasks: tasks, InFocus: held, NothingBlocks: free}, nil
}

// whatIsInFocus answers about the task being worked on, and says so plainly
// when there is none: nothing in focus is an answer, not an absence.
func whatIsInFocus(on On) (Answer, error) {
	t, number, err := on.Engine.InFocus()
	if err != nil {
		return Answer{}, err
	}
	if number == store.NoFocus {
		return Answer{InFocus: &number}, nil
	}
	p, err := on.Engine.TaskProgress(number)
	if err != nil {
		return Answer{}, err
	}
	return Answer{Task: &t, Progress: p, InFocus: &number}, nil
}

// TraceShow says what has been done to a task.
var TraceShow = Operation{
	Thing:   "trace",
	Verb:    "show",
	Short:   "Show what has been done to a task",
	Does:    "Show everything done to a task, oldest first, with what each change went from and to.",
	Level:   Read,
	Needs:   AStore,
	Answers: Acts{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		acts, err := on.Engine.History(n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Acts: acts}, nil
	},
}

// TraceList says what was done over a period, and what that comes to.
var TraceList = Operation{
	Thing:   "trace",
	Verb:    "list",
	Short:   "List what was done over a period",
	Does:    "List what was done from a day onward, today when none is given, and what it comes to.",
	Level:   Read,
	Needs:   AStore,
	Answers: Acts{},
	Takes: []Argument{
		{Name: "since", Kind: Day, Means: "the first day to look at; today when left out"},
	},
	Run: func(on On, in Input) (Answer, error) {
		from, err := sinceWhen(in)
		if err != nil {
			return Answer{}, err
		}
		acts, err := on.Engine.Since(from)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Since: acts}, nil
	},
}

// sinceWhen reads the day a period starts at: the one given, or the start of
// today when none was. Today rather than everything, because a listing of every
// act cairn ever recorded is not a question anybody asked.
func sinceWhen(in Input) (time.Time, error) {
	given := in.Given["since"]
	if given == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	}
	return domain.ParseDeadline(given)
}
