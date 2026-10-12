package engine

// The task operations: add, list, show, edit and delete a task, described
// once so both surfaces build the same command from the same facts.

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/task"
)

// TaskNew records a new task.
var TaskNew = Operation{
	Thing: "task",
	Verb:  "new",
	Short: "Make a new task",
	Does:  "Make a new task, carrying whatever of its attributes were given.",
	Example: `cairn task new "Write the spec" --state todo --priority high --label spec --deadline 2026-09-01
cairn task new "Write the tests" --parent 1 --objective 2
cairn task new "Write the spec" --desc - < notes.md`,
	Level:   Write,
	Needs:   AStore,
	Again:   DoesItAgain,
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
	Thing: "task",
	Verb:  "list",
	Short: "List tasks, all of them or only some",
	Does:  "List every task, or only the ones carrying everything named. A task has to carry all of it, not one of it.",
	Example: `cairn task list
cairn task list --state todo --priority high
cairn task list --label team --archived`,
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
		Means: "keep only the archived tasks, instead of the ones still on the board",
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
	Example: "cairn task show 1",
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
		since, err := on.Engine.stoodSince(n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Task: &shown, Under: under, Progress: p, Context: on.contextOf(n), StoodSince: since}, nil
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
	Thing: "task",
	Verb:  "archive",
	Short: "Archive a task, keeping everything it carries",
	Does:  "Archive a task. Everything it carries is kept, and it can be unarchived; deleting is the other thing.",
	Example: `cairn task archive 1
cairn task archive --state done`,
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Swept{},
	Takes: append([]Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a task", Means: "the number of the task"},
		alsoOf("a task"),
	}, taskNarrowing...),
	Run: func(on On, in Input) (Answer, error) {
		if n, named := in.Whole("number"); named && n != 0 {
			numbers, err := together(in)
			if err != nil {
				return Answer{}, err
			}
			for _, n := range numbers {
				if err := on.Engine.archive(n); err != nil {
					return Answer{}, err
				}
			}
			shown, err := shownTask(on, n)
			count := len(numbers)
			shown.Archived = &count
			return shown, err
		}
		if len(in.Listed["also"]) > 0 {
			return Answer{}, ErrAlsoWithoutNumber
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
	Thing:      "task",
	Verb:       "label",
	Short:      "Put a label on every task a narrowing reaches",
	Does:       "Put a label on every task the narrowing reaches, keeping the ones each already carries.",
	Example:    "cairn task label sprint-12 --state done",
	Level:      Write,
	Needs:      AStore,
	Again:      ChangesNothingMore,
	Overwrites: true,
	Answers:    Labelled{},
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

// ErrAlsoWithoutNumber is returned when archiving names more tasks beside a
// first, and no first: a narrowing already takes many, and "also" said beside
// one is a question about which.
var ErrAlsoWithoutNumber = errors.New("--also names tasks beside a number, and no number was given")

// TaskUnarchive brings one back onto the board.
var TaskUnarchive = Operation{
	Thing:   "task",
	Verb:    "unarchive",
	Short:   "Unarchive a task, back on the board",
	Does:    "Unarchive a task: it is back on the board, as it was.",
	Example: "cairn task unarchive 1",
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
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
	Thing: "task",
	Verb:  "edit",
	Short: "Change what a task carries",
	Does:  "Change what a task carries.",
	Example: `cairn task edit 1 --state current --estimate 2h
cairn task edit 1 --label +urgent --label -spec
cairn task edit 5 --state blocked --blocker 3 --because "the client's answer"`,
	Level:      Write,
	Needs:      AStore,
	Again:      ChangesNothingMore,
	Overwrites: true,
	Answers:    Task{},
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
	Thing: "task",
	Verb:  "delete",
	Short: "Delete a task",
	Does:  "Delete a task.",
	Example: `cairn task delete 1 --yes
cairn task delete 1 --also 4 --also 7 --yes`,
	Level:   Full,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Deleted{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
		alsoOf("a task"),
	},
	Run: func(on On, in Input) (Answer, error) {
		numbers, err := together(in)
		if err != nil {
			return Answer{}, err
		}
		var waiting []int
		for _, n := range numbers {
			// Read before it goes: afterwards the link is gone with it.
			was, err := on.Engine.waitingOn(n)
			if err != nil {
				return Answer{}, err
			}
			if err := on.Engine.delete(n); err != nil {
				return Answer{}, err
			}
			waiting = append(waiting, was...)
		}
		// A task deleted in the same act is not one left waiting.
		waiting = slices.DeleteFunc(waiting, func(n int) bool { return slices.Contains(numbers, n) })
		return Answer{Deleted: &numbers[0], Also: numbers[1:], Unblocked: waiting}, nil
	},
}

// TaskSearch names every task carrying everything named, wherever among its
// facts each one happens to sit.
var TaskSearch = Operation{
	Thing: "task",
	Verb:  "search",
	Short: "Look for a task by anything it carries",
	Does:  "List every task carrying everything named, wherever among its facts each one sits.",
	Example: `cairn task search spec
cairn task search spec team --state todo`,
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
	{Name: "state", Kind: Text, OneOf: oneOf(domain.States), Outside: domain.ErrNotAState, Means: "the column it sits in"},
	{Name: "priority", Kind: Text, OneOf: oneOf(domain.Priorities), Outside: domain.ErrNotAPriority, Means: "the priority it carries"},
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
	// Unnarrowed, the listing holds every part of every task it shows — a
	// part is never taken off the board without its task — so it is all the
	// counting needs. Narrowed, it does not, and every task is read.
	along, _, err := on.Engine.alongEvery(tasks, !kept.Narrows())
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
