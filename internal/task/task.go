// Package task holds the task: what it carries, the rules that keep it
// valid, and the membrane anything uses to read or build one. It knows only
// domain, for the attributes a task carries and the title rule it shares
// with an objective.
package task

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
)

// NoParent is the number a task carries when it hangs under nothing. Zero is
// safe to mean "none" because the store never hands out zero: numbering starts
// at one.
const NoParent = 0

// NoObjective is the number nothing serves. Zero is safe to mean "none"
// because the store never hands out zero: numbering starts at one, the same way
// NoParent works.
//
// A task carries a set of objectives rather than one, so it is no longer the
// value a task holds when it serves nothing — an empty set is. What it still
// answers is "is this a real objective number", which is what a caller naming
// one has to be held to.
const NoObjective = 0

// ErrParentIsSubTask is returned when a task is asked to hang under one that
// is itself a sub-task. Only one level: what hangs under a task is never
// something with more hanging under it.
var ErrParentIsSubTask = errors.New("a sub-task cannot carry sub-tasks")

// ErrParentIsSelf is returned when a task is asked to hang under itself.
var ErrParentIsSelf = errors.New("a task cannot hang under itself")

// ErrSubTaskServesNoObjective is returned when a task is asked to carry both
// a parent and an objective at once. A sub-task lives with the task it hangs
// under, counts towards that task's objective through it, and is never given
// one of its own — an objective it served alone would be one its parent
// might have nothing to do with.
var ErrSubTaskServesNoObjective = errors.New("a sub-task serves no objective")

// Task is the unit of work. It carries a title, and a number to refer to it by
// afterwards. Everything else is optional and has a stated default: a task
// starts in the backlog, at normal priority, carrying nothing else.
type Task struct {
	number      int
	parent      int
	objectives  []int
	title       string
	description string
	state       domain.State
	priority    domain.Priority
	labels      []string
	blockers    []int
	because     string
	deadline    time.Time
	estimate    time.Duration
	timeSpent   time.Duration
	// archived is beside the column, not one of them: a task archived after
	// being finished is still finished.
	archived bool
}

// Told is the membrane of a task: one method per fact it has to tell, absence
// included. Whoever wants to know something about a task implements this and
// is told it; nothing reads a task's insides, and a task knows none of them.
type Told interface {
	Numbered(int)
	Titled(string)
	InColumn(domain.State)
	Carries(domain.Priority)
	Described(string)
	NotDescribed()
	Labelled(string)
	NotLabelled()
	LabelsTold()
	BlockedBy(int)
	NotBlocked()
	BlockersTold()
	Because(string)
	NotBecause()
	Due(time.Time)
	NotDue()
	Estimated(time.Duration)
	NotEstimated()
	Spent(time.Duration)
	NothingSpent()
	Archived()
	NotArchived()
	Under(int)
	UnderNothing()
	Serves(int)
	ServesNothing()
	ObjectivesTold()
}

// Tell says every fact this task carries to the given membrane, absence
// included, in the order a card lays them out.
func (t Task) Tell(to Told) {
	to.Numbered(t.number)
	to.Titled(t.title)
	to.InColumn(t.state)
	if t.archived {
		to.Archived()
	} else {
		to.NotArchived()
	}
	to.Carries(t.priority)
	if t.description == "" {
		to.NotDescribed()
	} else {
		to.Described(t.description)
	}
	if len(t.labels) == 0 {
		to.NotLabelled()
	} else {
		for _, label := range t.labels {
			to.Labelled(label)
		}
	}
	to.LabelsTold()
	if len(t.blockers) == 0 {
		to.NotBlocked()
	} else {
		for _, blocker := range t.blockers {
			to.BlockedBy(blocker)
		}
	}
	to.BlockersTold()
	if t.because == "" {
		to.NotBecause()
	} else {
		to.Because(t.because)
	}
	if t.deadline.IsZero() {
		to.NotDue()
	} else {
		to.Due(t.deadline)
	}
	if t.estimate == 0 {
		to.NotEstimated()
	} else {
		to.Estimated(t.estimate)
	}
	if t.timeSpent == 0 {
		to.NothingSpent()
	} else {
		to.Spent(t.timeSpent)
	}
	if t.parent == NoParent {
		to.UnderNothing()
	} else {
		to.Under(t.parent)
	}
	if len(t.objectives) == 0 {
		to.ServesNothing()
	} else {
		for _, objective := range t.objectives {
			to.Serves(objective)
		}
	}
	to.ObjectivesTold()
}

// MayCarrySubTasks says whether a task may have another hang under it. Only
// one level: a sub-task cannot itself carry sub-tasks, so what hangs under a
// task is never something with more hanging under it.
func (t Task) MayCarrySubTasks() error {
	if t.parent != NoParent {
		return fmt.Errorf("task %d: %w", t.number, ErrParentIsSubTask)
	}
	return nil
}

// ErrAlreadyArchived is returned when a task off the board is archived again.
var ErrAlreadyArchived = errors.New("the task is already archived")

// ErrNotArchived is returned when a task on the board is brought back onto it.
var ErrNotArchived = errors.New("the task is on the board")

// ErrPartGoesWithItsWork is returned when a sub-task is archived on its own.
var ErrPartGoesWithItsWork = errors.New("a part goes off the board with the work it belongs to")

// MayArchive says whether this task may be taken off the board.
func (t Task) MayArchive() error {
	if t.parent != NoParent {
		return fmt.Errorf("task %d: %w", t.parent, ErrPartGoesWithItsWork)
	}
	if t.archived {
		return fmt.Errorf("task %d: %w", t.number, ErrAlreadyArchived)
	}
	return nil
}

// MayUnarchive says whether a task can be brought back onto the board. The
// mirror of MayArchive: one gesture in two directions has to answer a mistake
// the same way in both, or it is a pair nobody can reason about.
func (t Task) MayUnarchive() error {
	if !t.archived {
		return fmt.Errorf("task %d: %w", t.number, ErrNotArchived)
	}
	return nil
}

// ErrSubTaskHasNoColumn is returned when a sub-task is put in a column that
// says where work stands. A sub-task is a part of the work its parent names,
// split off to be sure the whole is finished, and the only thing worth knowing
// about a part is whether it is done.
//
// It had all seven columns and used two: measured on a store of 219 tasks, the
// 49 sub-tasks in it were done forty-eight times and todo once, never current,
// never stopped, never with a deadline or a blocker of their own. A part being
// "blocked" is the parent being blocked; a part being "current" is the parent
// being worked on.
var ErrSubTaskHasNoColumn = errors.New("a sub-task is done or it is not")

// MayCarryItsColumn says whether this task may stand in the column it carries.
// Every column is open to a task hanging under nothing; a sub-task takes done
// or todo and nothing else.
func (t Task) MayCarryItsColumn() error {
	if t.parent == NoParent || t.state == domain.Done || t.state == domain.Todo {
		return nil
	}
	return fmt.Errorf("task %d in %s: %w", t.number, t.state, ErrSubTaskHasNoColumn)
}

// Settled is this task with a column a sub-task may carry: done stays done and
// anything else becomes todo. What a task hung under another was doing on its
// own stops being true the moment it becomes a part of something — except
// being finished, which stays finished.
func (t Task) Settled() Task {
	if t.parent == NoParent || t.state == domain.Done {
		return t
	}
	t.state = domain.Todo
	return t
}

// MayServeAnObjective says whether a task may carry an objective alongside
// whatever parent it carries: a sub-task never does, since it counts towards
// its parent's objective by way of that parent and is never given one of its
// own.
func (t Task) MayServeAnObjective() error {
	if t.parent != NoParent && len(t.objectives) > 0 {
		return ErrSubTaskServesNoObjective
	}
	return nil
}

// MayHangUnderItsParent says whether the parent this task already carries is
// one it may take: never itself.
func (t Task) MayHangUnderItsParent() error {
	if t.parent == t.number {
		return fmt.Errorf("task %d: %w", t.number, ErrParentIsSelf)
	}
	return nil
}

// ErrSubTaskStillOpen is returned when a task is asked to be finished while
// something hanging under it is not. A parent is done because what hangs under
// it is done: if it could be finished with a sub-task still open, then either
// that sub-task was never part of it — in which case it is another task — or
// the parent was closed over work nobody did.
var ErrSubTaskStillOpen = errors.New("something hanging under it is not done")

// MayFinish says whether the task carrying number may be moved to done, given
// what hangs under it.
//
// Refused rather than cascaded. Marking the children done on the parent's
// behalf would finish work nobody finished, and it would do it silently; the
// refusal names which ones are in the way, which is the useful thing to be
// told. It is asked only of done: blocked, watching and paused say something
// about the parent itself, and done is the only state that claims something
// about what hangs under it.
func MayFinish(number int, hanging []Task) error {
	var open []string
	for _, t := range hanging {
		if t.state != domain.Done {
			open = append(open, strconv.Itoa(t.number))
		}
	}
	if len(open) == 0 {
		return nil
	}
	return fmt.Errorf("task %d: task %s: %w", number, strings.Join(open, ", "), ErrSubTaskStillOpen)
}

// NewTask builds a task from a title, refusing one that carries no text.
// What surrounds the title is not part of it; what sits inside it is.
func NewTask(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, fmt.Errorf("a task %w", domain.ErrTitleRequired)
	}
	if err := domain.OneLine(title); err != nil {
		return Task{}, err
	}
	return Task{title: title, state: domain.Backlog, priority: domain.Normal}, nil
}
