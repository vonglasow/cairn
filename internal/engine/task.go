package engine

import (
	"cairn/internal/engine/item"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/engine/task"
)

// TaskNew records a new task.
var TaskNew = Operation{
	Thing: "task",
	Verb:  "new",
	Short: "Make a new task",
	Does:  "Make a new task, carrying whatever of its attributes were given.",
	Example: `cairn task new "Write the spec" --state todo --priority high --label spec --deadline 2026-09-01
cairn task new "Ship it" --objective 2
cairn task new "Write the spec" --desc - < notes.md`,
	Level:   write,
	Needs:   aStore,
	Again:   doesItAgain,
	Answers: Task{},
	Takes: append(append(append([]Argument{
		{Name: "title", Kind: Text, Positional: true, Required: true, Means: "what the task is called"},
	}, taskAttributes...), stepsArgument), everyArguments...),
	Run: func(on workedOn, in Input) (Answer, error) {
		title := ""
		if given := in.change("title"); given != nil {
			title = *given
		}
		changes, err := taskChanges(in)
		if err != nil {
			return Answer{}, err
		}
		begun, err := on.Engine.beginning(in, &changes)
		if err != nil {
			return Answer{}, err
		}
		added, err := on.Engine.add(title, changes)
		if err != nil {
			return Answer{}, task.WithTypedTitle(err, title)
		}
		n := added.Carried().Number
		if begun != nil {
			if err := on.Engine.beginSeries(added, begun.Rule, begun.From, begun.Day); err != nil {
				return Answer{}, err
			}
		}
		p, err := on.Engine.taskProgress(n)
		if err != nil {
			return Answer{}, err
		}
		warnings, err := on.Engine.unmet(n, "")
		if err != nil {
			return Answer{}, err
		}
		occurring, err := on.Engine.occurring([]int{n})
		return Answer{Task: &added, Progress: p, Warnings: warnings, Occurring: occurring}, err
	},
}

// TaskList names every task, or only the ones carrying what was named.
var TaskList = Operation{
	Thing: "task",
	Verb:  "list",
	Short: "List tasks, all of them or only some",
	Does:  "List every task, or only the ones carrying everything named. A task has to carry all of it, not one of it. Given numbers, only those tasks, archived or not, the answer saying under unknown which number is no task. Each row carries every field, or only the fields named. Given a limit or an offset, only that page of them, in the board's order, the answer saying under total how many there are in all. Given count, no row: only how many sit in each column or carry each label.",
	Example: `cairn task list
cairn task list --state todo --priority high
cairn task list --label team --archived
cairn task list --number 25 --number 18
cairn task list --limit 50 --offset 100
cairn task list --count state`,
	Level:   Read,
	Needs:   aStore,
	Answers: Tasks{},
	Takes:   append(append(slices.Clone(taskNarrowing), blockerArgument, viewArgument, numbersArgument, fieldsArgument, countArgument, inSeriesArgument), pageArguments...),
	Run:     listing,
}

var taskNarrowing = append(narrowedBy(taskAttributes, map[string]string{
	"state":     "keep only the tasks sitting in this column",
	"priority":  "keep only the tasks carrying this priority",
	"label":     "keep only the tasks carrying this label; named twice, only those carrying both",
	"objective": "keep only the tasks serving this objective or one of its sub-objectives",
}),
	Argument{
		Name:  "deadline-before",
		Kind:  day,
		Means: "keep only the tasks whose deadline is on or before this day",
	},
	Argument{
		Name:  "unblocked",
		Kind:  Bool,
		Means: "keep only the tasks nothing stops: no open task in their way, and no reason said",
	},
	Argument{
		Name:  "focus",
		Kind:  Bool,
		Means: "keep only the tasks in focus",
	},
	Argument{
		Name:  "no-focus",
		Kind:  Bool,
		Means: "keep only the tasks not in focus",
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
	Argument{
		Name:  "itself",
		Kind:  Bool,
		Means: "keep to the tasks serving the objective named itself, leaving out those serving only its sub-objectives",
	},
)

// TaskShow says what a task carries, its steps, how far along it is, and where its outside context lives.
var TaskShow = Operation{
	Thing:   "task",
	Verb:    "show",
	Short:   "Show what a task carries",
	Does:    "Show what a task carries, its steps, how far along it is, and where what was written about it elsewhere lives. That directory is where a dated note of the work goes — one file per entry, added beside the others — rather than into the description, which an edit rewrites whole.",
	Example: "cairn task show 1",
	Level:   Read,
	Needs:   aStore,
	Answers: Task{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, _ := in.whole("number")
		shown, err := on.Engine.show(n)
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
		occurring, err := on.Engine.occurring([]int{n})
		if err != nil {
			return Answer{}, err
		}
		chosen, err := on.Engine.choosing([]task.Task{shown})
		return Answer{Task: &shown, Progress: p, Context: on.contextOf(n), StoodSince: since, Occurring: occurring, Choosing: chosen}, err
	},
}

func (on workedOn) contextOf(n int) string {
	if on.Notes == "" {
		return ""
	}
	return filepath.Join(on.Notes, strconv.Itoa(n))
}

// TaskArchive takes a task off the board.
var TaskArchive = Operation{
	Thing: "task",
	Verb:  "archive",
	Short: "Archive a task, keeping everything it carries",
	Does:  "Archive a task. Everything it carries is kept, and it can be unarchived; deleting is the other thing.",
	Example: `cairn task archive 1
cairn task archive --state done`,
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: Swept{},
	Takes: append([]Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a task", Means: "the number of the task"},
		alsoOf("a task"),
	}, taskNarrowing...),
	Sweeps: sweepsByNarrowing,
	Run: func(on workedOn, in Input) (Answer, error) {
		if n, named := in.whole("number"); named && n != 0 {
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
			return Answer{}, task.ErrAlsoWithoutNumber
		}
		kept, err := keeping(in, taskNarrowing, on.Engine)
		if err != nil {
			return Answer{}, err
		}
		if !kept.Narrows() {
			return Answer{}, task.ErrNothingToArchive
		}
		swept, err := on.Engine.archiveEvery(kept)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Archived: &swept}, nil
	},
}

// TaskLabel puts one label on every task a narrowing reaches.
var TaskLabel = Operation{
	Thing:      "task",
	Verb:       "label",
	Short:      "Put a label on every task a narrowing reaches",
	Does:       "Put a label on every task the narrowing reaches, keeping the ones each already carries.",
	Example:    "cairn task label sprint-12 --state done",
	Level:      write,
	Needs:      aStore,
	Again:      ChangesNothingMore,
	Overwrites: true,
	Answers:    Labelled{},
	Takes: append([]Argument{
		{Name: "put", Kind: Text, Positional: true, Required: true, Means: "the label to put on them"},
	}, taskNarrowing...),
	Run: func(on workedOn, in Input) (Answer, error) {
		label := ""
		if given := in.change("put"); given != nil {
			label = *given
		}
		kept, err := keeping(in, taskNarrowing, on.Engine)
		if err != nil {
			return Answer{}, err
		}
		if !kept.Narrows() {
			return Answer{}, task.ErrNothingToLabel
		}
		put, err := on.Engine.labelEvery(kept, label)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Labelled: &put}, nil
	},
}

// TaskUnarchive brings one back onto the board.
var TaskUnarchive = Operation{
	Thing:   "task",
	Verb:    "unarchive",
	Short:   "Unarchive a task, back on the board",
	Does:    "Unarchive a task: it is back on the board, as it was.",
	Example: "cairn task unarchive 1",
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: Task{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, _ := in.whole("number")
		if err := on.Engine.unarchive(n); err != nil {
			return Answer{}, err
		}
		return shownTask(on, n)
	},
}

func shownTask(on workedOn, number int) (Answer, error) {
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
	Does:  "Change what a task carries. Put in done while one of its steps is still open, it is refused unless confirmed; confirmed, its steps are ticked with it.",
	Example: `cairn task edit 1 --state current --estimate 2h
cairn task edit 1 --label +urgent --label -spec
cairn task edit 5 --state blocked --blocker 3 --because "the client's answer"`,
	Finishes:   true,
	Level:      write,
	Needs:      aStore,
	Again:      ChangesNothingMore,
	Overwrites: true,
	Answers:    Task{},
	Takes: append(append(append([]Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
		{Name: "title", Kind: Text, Means: "what the task is called"},
	}, taskAttributes...), append([]Argument{stepsArgument}, replaceArguments...)...), append(slices.Clone(everyArguments), seriesTooArgument, alsoOf("a task"))...),
	Run: func(on workedOn, in Input) (Answer, error) {
		numbers, err := together(in)
		if err != nil {
			return Answer{}, err
		}
		n := numbers[0]
		changes, err := taskChanges(in)
		if err != nil {
			return Answer{}, err
		}
		changes.Replace, changes.By = in.change("replace"), in.change("by")
		changes.Title = in.change("title")
		changes.Finishing = in.Confirmed
		every, err := on.Engine.repeating(n, in)
		if err != nil {
			return Answer{}, err
		}
		was, _ := on.Engine.show(n)
		if changes.named() || !every {
			for _, m := range numbers {
				if _, err := on.Engine.edit(m, changes); err != nil {
					return Answer{}, editReason(err)
				}
			}
		}
		if every {
			if err := on.Engine.everyGiven(n, in); err != nil {
				return Answer{}, err
			}
		}
		if in.yes("series") {
			if err := on.Engine.seriesFollowsTask(n, in); err != nil {
				return Answer{}, err
			}
		}
		edited, err := on.Engine.show(n)
		if err != nil {
			return Answer{}, err
		}
		p, err := on.Engine.taskProgress(n)
		if err != nil {
			return Answer{}, err
		}
		warnings, err := on.Engine.unmet(n, was.Carried().State)
		if err != nil {
			return Answer{}, err
		}
		occurring, err := on.Engine.occurring([]int{n})
		return Answer{Task: &edited, Progress: p, Moved: movedOf(was, edited), Warnings: warnings, Occurring: occurring}, err
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
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: Deleted{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
		alsoOf("a task"),
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		numbers, err := together(in)
		if err != nil {
			return Answer{}, err
		}
		var waiting []int
		for _, n := range numbers {
			was, err := on.Engine.waitingOn(n)
			if err != nil {
				return Answer{}, err
			}
			if err := on.Engine.delete(n); err != nil {
				return Answer{}, err
			}
			waiting = append(waiting, was...)
		}
		waiting = slices.DeleteFunc(waiting, func(n int) bool { return slices.Contains(numbers, n) })
		return Answer{Deleted: &numbers[0], Also: numbers[1:], Unblocked: waiting}, nil
	},
}

func somethingToLookFor(named *[]string) error {
	if named == nil {
		return nil
	}
	for _, n := range *named {
		if strings.TrimSpace(n) != "" {
			return nil
		}
	}
	return Refusing("nothing-to-look-for", "search: there is nothing to look for")
}

// TaskSearch names every task carrying everything named, wherever among its facts each one happens to sit.
var TaskSearch = Operation{
	Thing: "task",
	Verb:  "search",
	Short: "Look for a task by anything it carries",
	Does:  "List every task carrying everything named, wherever among its facts each one sits — or, with any, every task carrying at least one of it, the answer then saying under matched which tasks each thing named matched. Each row carries every field, or only the fields named. Given a limit or an offset, only that page of what it kept, in the board's order, the answer saying under total how many it kept in all, and matched speaking only of the tasks on the page.",
	Example: `cairn task search spec
cairn task search spec team --state todo
cairn task search T010 T013 --any
cairn task search spec --limit 50 --offset 100`,
	Level:   Read,
	Needs:   aStore,
	Answers: Tasks{},
	Takes: append([]Argument{
		{Name: "named", Kind: List, Positional: true, Required: true, Means: "what to look for, read the way it is shown"},
		{Name: "any", Kind: Bool, Means: "keep a task carrying any one of what is named rather than all of it, and say under matched which tasks each one matched"},
	}, append(append(slices.Clone(taskNarrowing), fieldsArgument), pageArguments...)...),
	Run: func(on workedOn, in Input) (Answer, error) {
		named := in.changed("named")
		if err := somethingToLookFor(named); err != nil {
			return Answer{}, err
		}
		kept, err := keeping(in, taskNarrowing, on.Engine)
		if err != nil {
			return Answer{}, err
		}
		fields, err := fieldsOf(in)
		if err != nil {
			return Answer{}, err
		}
		cut, err := askedPage(in)
		if err != nil {
			return Answer{}, err
		}
		tasks, err := on.Engine.narrowed(kept)
		if err != nil {
			return Answer{}, err
		}
		var found []task.Task
		var terms map[string][]int
		if in.yes("any") {
			found, terms = task.MatchingAny(tasks, *named)
		} else {
			found = []task.Task{}
			for _, t := range tasks {
				if named != nil && task.Matching(t, *named) {
					found = append(found, t)
				}
			}
		}
		answer := Answer{Fields: fields, Offset: cut.offset}
		if cut.asked {
			total := len(found)
			answer.Total = &total
			found = cut.of(found)
			terms = task.OnPage(terms, found)
		}
		answer.Matched, answer.Terms = found, terms
		answer.Choosing, err = on.Engine.choosing(found)
		return answer, err
	},
}

// Reaches is the engine's search over what a board already holds: what a task or an objective said a search looks in (task.Said, objective.Said), and what somebody typed, word by word.
func Reaches(said []string, typed string) bool {
	return item.Reaches(said, strings.Fields(typed))
}

func sweepsByNarrowing(in Input) bool {
	if n, named := in.whole("number"); named && n != 0 {
		return false
	}
	for _, by := range taskNarrowing {
		if in.Named(by.Name) {
			return true
		}
	}
	return false
}

var taskAttributes = []Argument{
	{Name: "desc", Kind: Text, Means: "what the task is; what was done or decided goes in its notes"},
	{Name: "state", Kind: Text, OneOf: oneOf(item.States), Outside: item.ErrNotAState, Means: "the column it sits in"},
	{Name: "priority", Kind: Text, OneOf: oneOf(item.Priorities), Outside: item.ErrNotAPriority, Means: "the priority it carries"},
	{Name: "label", Kind: List, Means: "a label it carries; when changing, x or +x adds one, -x takes one off, =x replaces them all"},
	{Name: "blocker", Kind: List, Of: "a task", Means: "the number of a task in its way; when changing, 3 or +3 adds one, -3 takes one away, =3 replaces them all"},
	{Name: "link", Kind: List, Of: "a task", Means: "the number of a task it is linked to; when changing, 3 or +3 adds one, -3 takes one away, =3 replaces them all"},
	{Name: "because", Kind: Text, Means: "what is in its way and is not a task cairn holds"},
	{Name: "deadline", Kind: day, Means: "the day it is wanted by"},
	{Name: "estimate", Kind: duration, Means: "how long it should take"},
	{Name: "spent", Kind: duration, Means: "the time spent on it, the time in focus included; +30m adds time today, -15m takes some back"},
	{Name: "objective", Kind: List, Of: "an objective", Means: "the number of an objective it serves; when changing, 4 or +4 adds one, -1 takes one away, =4 replaces them all"},
}

func oneOf[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = string(v)
	}
	return out
}

func editReason(err error) error {
	if errors.Is(err, errNothingToChange) {
		return fmt.Errorf("%w: name one of --title, --desc, --state, --priority, --label, --deadline, --estimate", err)
	}
	return err
}

var replaceArguments = []Argument{
	{Name: "replace", Kind: Text, Means: "a piece of the description to change, there exactly once; --by says what goes in its place"},
	{Name: "by", Kind: Text, Means: "what goes in place of the piece --replace names; empty takes it out"},
}

func taskChanges(in Input) (Changes, error) {
	var c Changes
	c.Description = in.change("desc")
	c.State = in.change("state")
	c.Priority = in.change("priority")
	c.Labels = in.changed("label")
	c.Blocker = in.changed("blocker")
	c.Link = in.changed("link")
	c.Because = in.change("because")
	c.Deadline = in.change("deadline")
	c.Estimate = in.change("estimate")
	c.Spent = in.change("spent")
	c.Objective = in.changed("objective")
	if said := in.change("steps"); said != nil {
		steps, err := stepsIn(*said)
		if err != nil {
			return Changes{}, err
		}
		c.Steps = &steps
	}
	return c, nil
}

func listing(on workedOn, in Input) (Answer, error) {
	kept, err := keeping(in, taskNarrowing, on.Engine)
	if err != nil {
		return Answer{}, err
	}
	fields, err := fieldsOf(in)
	if err != nil {
		return Answer{}, err
	}
	numbers, err := askedNumbers(in)
	if err != nil {
		return Answer{}, err
	}
	blocker, byBlocker, err := askedBlocker(in)
	if err != nil {
		return Answer{}, err
	}
	cut, err := askedPage(in)
	if err != nil {
		return Answer{}, err
	}
	counting, err := askedCount(in, fields, cut)
	if err != nil {
		return Answer{}, err
	}
	if numbers != nil && !in.yes("archived") && !in.yes("with-archived") {
		kept.WithArchived()
	}
	seen, err := on.Engine.inView(in, kept)
	if err != nil {
		return Answer{}, err
	}
	occurrences, err := on.Engine.occurrencesIn(in)
	if err != nil {
		return Answer{}, err
	}
	tasks, err := on.Engine.narrowed(kept)
	if err != nil {
		return Answer{}, err
	}
	if seen != nil {
		tasks = seen.Kept(tasks)
	}
	tasks = occurrences.Kept(tasks)
	if byBlocker {
		tasks = task.InTheWayOf(tasks, blocker)
	}
	var unknown []int
	if numbers != nil {
		if tasks, unknown, err = on.Engine.numbered(tasks, numbers); err != nil {
			return Answer{}, err
		}
	}
	total := len(tasks)
	var counts []count
	if counting != "" {
		counts, tasks = countedBy(tasks, counting), []task.Task{}
	}
	tasks = cut.of(tasks)
	inFocus, err := on.Engine.focus()
	if err != nil {
		return Answer{}, err
	}
	held := heldAmong(tasks, inFocus)
	free, err := on.Engine.tasks.NothingBlocks()
	if err != nil {
		return Answer{}, err
	}
	narrows := kept.Narrows() || numbers != nil || seen != nil || occurrences.Of != nil || byBlocker
	along, _, err := on.Engine.alongEvery(tasks, !narrows && !cut.asked)
	if err != nil {
		return Answer{}, err
	}
	contexts := map[int]string{}
	for _, t := range tasks {
		n := t.Carried().Number
		contexts[n] = on.contextOf(n)
	}
	shown := make([]int, 0, len(tasks))
	for _, t := range tasks {
		shown = append(shown, t.Carried().Number)
	}
	occurring, err := on.Engine.occurring(shown)
	if err != nil {
		return Answer{}, err
	}
	chosen, err := on.Engine.choosing(tasks)
	if err != nil {
		return Answer{}, err
	}
	answer := Answer{InFocus: &held, NothingBlocks: free, Along: along, Contexts: contexts, Fields: fields, Unknown: unknown, Counts: counts, Offset: cut.offset, Occurring: occurring, Choosing: chosen}
	if counting != "" || cut.asked {
		answer.Total = &total
	}
	if seen != nil {
		answer.InView, answer.Lost, answer.KeepsNothing = seen.Name, seen.Lost, seen.Nothing
		if counting == "" {
			if answer.Outside, err = on.Engine.outside(tasks); err != nil {
				return Answer{}, err
			}
		}
	}
	if narrows {
		answer.Matched = tasks
		return answer, nil
	}
	answer.Tasks = tasks
	return answer, nil
}
