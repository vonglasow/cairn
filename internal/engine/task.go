package engine

// The task operations: add, list, show, edit and delete a task, described
// once so both surfaces build the same command from the same facts.

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
)

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
		added, err := on.Engine.add(title, changes)
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
var taskNarrowing = append(narrowedBy(taskAttributes, map[string]string{
	"state":     "keep only the tasks sitting in this column",
	"priority":  "keep only the tasks carrying this priority",
	"label":     "keep only the tasks carrying this label; named twice, only those carrying both",
	"objective": "keep only the tasks serving this objective",
}),
	// Not one of a task's attributes: being off the board is done to a task,
	// not carried by one.
	Argument{
		Name:  "due-before",
		Kind:  Day,
		Means: "keep only the tasks due on or before this day",
	},
	Argument{
		Name:  "archived",
		Kind:  Bool,
		Means: "keep only the tasks taken off the board, instead of the ones still on it",
	},
	Argument{
		Name:  "with-archived",
		Kind:  Bool,
		Means: "keep both the tasks on the board and the ones taken off it",
	},
)

// TaskShow says what a task carries, what hangs under it, how far along it is,
// and where its outside context lives.
var TaskShow = Operation{
	Thing: "task",
	Verb:  "show",
	Short: "Show what a task carries",
	// The last sentence is for an agent, which reads this as the tool's
	// description: what it learns while working belongs in a dated file in
	// that directory, added beside the rest, and not in the description,
	// which is rewritten whole every time it is changed. See docs/context.md.
	Does:    "Show what a task carries, what hangs under it, how far along it is, and where what was written about it elsewhere lives. That directory is where a dated note of the work goes — one file per entry, added beside the others — rather than into the description, which an edit rewrites whole.",
	Level:   Read,
	Needs:   AStore,
	Answers: Task{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		shown, err := on.Engine.show(n)
		if err != nil {
			return Answer{}, err
		}
		under, err := on.Engine.under(n)
		if err != nil {
			return Answer{}, err
		}
		p, err := on.Engine.taskProgress(n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Task: &shown, Under: under, Progress: p, Context: on.contextOf(n)}, nil
	},
}

// contextOf is where a task's outside context lives: the directory named for
// its number under the notes, or nothing when nobody said where the notes are.
// Composed, never looked at: cairn says where and does nothing about it — see
// docs/context.md.
func (on On) contextOf(n int) string {
	if on.Notes == "" {
		return ""
	}
	return filepath.Join(on.Notes, strconv.Itoa(n))
}

// TaskArchive takes a task off the board. Its own verb rather than a flag on
// edit: it is done to a task, not carried by one.
var TaskArchive = Operation{
	Thing:   "task",
	Verb:    "archive",
	Short:   "Take a task off the board, keeping everything it carries",
	Does:    "Take a task off the board. Everything it carries is kept, and it can be brought back; deleting is the other thing.",
	Level:   Write,
	Needs:   AStore,
	Answers: Swept{},
	Takes: append([]Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a task", Means: "the number of the task"},
	}, taskNarrowing...),
	Run: func(on On, in Input) (Answer, error) {
		if n, named := in.Whole("number"); named && n != 0 {
			if err := on.Engine.archive(n); err != nil {
				return Answer{}, err
			}
			shown, err := shownTask(on, n)
			one := 1
			shown.Archived = &one
			return shown, err
		}
		kept, err := keeping(in, taskNarrowing)
		if err != nil {
			return Answer{}, err
		}
		// Nothing to go on would take the whole board off it, which is not
		// something to find out afterwards.
		if !kept.Narrows() {
			return Answer{}, ErrNothingToArchive
		}
		swept, err := on.Engine.archiveEvery(kept)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Archived: &swept}, nil
	},
}

// TaskLabel puts one label on every task a narrowing reaches. Marking a sprint
// one task at a time is not marking it at all.
var TaskLabel = Operation{
	Thing:   "task",
	Verb:    "label",
	Short:   "Put a label on every task a narrowing reaches",
	Does:    "Put a label on every task the narrowing reaches, keeping the ones each already carries.",
	Level:   Write,
	Needs:   AStore,
	Answers: Labelled{},
	Takes: append([]Argument{
		// Named "put" and not "label": the narrowing below already takes a
		// "label" — the ones a task has to carry to be reached — and an agent
		// names its arguments in one object, where a name is held once.
		{Name: "put", Kind: Text, Positional: true, Required: true, Means: "the label to put on them"},
	}, taskNarrowing...),
	Run: func(on On, in Input) (Answer, error) {
		label := ""
		if given := in.Change("put"); given != nil {
			label = *given
		}
		kept, err := keeping(in, taskNarrowing)
		if err != nil {
			return Answer{}, err
		}
		if !kept.Narrows() {
			return Answer{}, ErrNothingToLabel
		}
		put, err := on.Engine.labelEvery(kept, label)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Labelled: &put}, nil
	},
}

// ErrNothingToLabel is returned when labelling is given nothing to narrow by.
var ErrNothingToLabel = errors.New("labelling needs something to narrow by")

// ErrNothingToArchive is returned when archiving is given neither a number nor
// a narrowing.
var ErrNothingToArchive = errors.New("archiving needs a number or something to narrow by")

// TaskUnarchive brings one back onto the board.
var TaskUnarchive = Operation{
	Thing:   "task",
	Verb:    "unarchive",
	Short:   "Bring a task back onto the board",
	Does:    "Bring a task back onto the board.",
	Level:   Write,
	Needs:   AStore,
	Answers: Task{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		if err := on.Engine.unarchive(n); err != nil {
			return Answer{}, err
		}
		return shownTask(on, n)
	},
}

// shownTask answers with the task as it ends up, so nobody has to ask again.
func shownTask(on On, number int) (Answer, error) {
	found, err := on.Engine.show(number)
	if err != nil {
		return Answer{}, err
	}
	return Answer{Task: &found}, nil
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
		edited, err := on.Engine.edit(n, changes)
		if err != nil {
			return Answer{}, editReason(err)
		}
		p, err := on.Engine.taskProgress(n)
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
		if err := on.Engine.delete(n); err != nil {
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
	// The one place the two vocabularies meet: what is looked for is free text
	// and what narrows it is a closed set. A word worth searching for is a word
	// common enough to answer the whole board without one.
	Takes: append([]Argument{
		{Name: "named", Kind: List, Positional: true, Required: true, Means: "what to look for, read the way it is shown"},
	}, taskNarrowing...),
	Run: func(on On, in Input) (Answer, error) {
		named := in.Changed("named")
		kept, err := keeping(in, taskNarrowing)
		if err != nil {
			return Answer{}, err
		}
		tasks, err := on.Engine.narrowed(kept)
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
		if err := on.Engine.move(n, at, target); err != nil {
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
	Short:   "Put a task in focus, or say which ones are",
	Does:    "Put a task in focus alongside whatever else is being worked on, or name every task in focus. Several can be at once.",
	Level:   Write,
	Needs:   AStore,
	Answers: Focus{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a task", Means: "the task to take up, or nothing to name every task in focus"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, named := in.Whole("number")
		if !named {
			return whatIsInFocus(on)
		}
		t, err := on.Engine.putInFocus(n)
		if err != nil {
			return Answer{}, err
		}
		p, err := on.Engine.taskProgress(n)
		if err != nil {
			return Answer{}, err
		}
		held, err := on.Engine.focus()
		if err != nil {
			return Answer{}, err
		}
		return Answer{Task: &t, Progress: p, InFocus: &held}, nil
	},
}

// TaskUnfocus leaves nothing in focus.
var TaskUnfocus = Operation{
	Thing:   "task",
	Verb:    "unfocus",
	Short:   "Let go of a task in focus, or of them all",
	Does:    "Let go of one task and keep the rest, or of every task in focus when no number is given. Letting go of nothing is not a failure.",
	Level:   Write,
	Needs:   AStore,
	Answers: Focus{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a task", Means: "the task to let go of, or nothing to let go of them all"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, named := in.Whole("number")
		if named {
			if err := on.Engine.letGoOf(n); err != nil {
				return Answer{}, err
			}
		} else if err := on.Engine.letGo(); err != nil {
			return Answer{}, err
		}
		return whatIsInFocus(on)
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
	tasks, err := on.Engine.narrowed(kept)
	if err != nil {
		return Answer{}, err
	}
	inFocus, err := on.Engine.focus()
	if err != nil {
		return Answer{}, err
	}
	held := heldAmong(tasks, inFocus)
	free, err := on.Engine.nothingBlocks()
	if err != nil {
		return Answer{}, err
	}
	along, _, err := on.Engine.alongEvery(tasks, !kept.Narrows() && in.Yes("with-archived"))
	if err != nil {
		return Answer{}, err
	}
	contexts := map[int]string{}
	for _, t := range tasks {
		n := factsOf(t).number
		contexts[n] = on.contextOf(n)
	}
	// A narrowed listing is answered flat: what matched is what is shown, and
	// where each task sits is not what was asked. Which field carries them is
	// how Of and RenderText tell the two apart.
	answer := Answer{InFocus: &held, NothingBlocks: free, Along: along, Contexts: contexts}
	if kept.Narrows() {
		answer.Matched = tasks
		return answer, nil
	}
	answer.Tasks = tasks
	return answer, nil
}

// whatIsInFocus answers about the task being worked on, and says so plainly
// when there is none: nothing in focus is an answer, not an absence.
func whatIsInFocus(on On) (Answer, error) {
	held, numbers, err := on.Engine.inFocus()
	if err != nil {
		return Answer{}, err
	}
	return Answer{Held: held, InFocus: &numbers}, nil
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
		acts, err := on.Engine.historyOf(n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Acts: acts}, nil
	},
}

// TraceList says what was done over a period, across every task.
var TraceList = Operation{
	Thing:   "trace",
	Verb:    "list",
	Short:   "List what was done over a period",
	Does:    "List what was done from a day onward, today when none is given, across every task.",
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
		acts, err := on.Engine.since(from)
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
