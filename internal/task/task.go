// Package task holds the task: what it carries, the rules that keep it
// valid, and the membrane anything uses to read or build one. It knows only
// domain, for the attributes a task carries and the title rule it shares
// with an objective.
package task

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"cairn/internal/domain"
)

// NoParent is the number a task carries when it hangs under nothing. Zero is
// safe to mean "none" because the store never hands out zero: numbering starts
// at one.
const NoParent = 0

// NoObjective is the number a task carries when it serves none. Zero is safe
// to mean "none" because the store never hands out zero: numbering starts at
// one, the same way NoParent works.
const NoObjective = 0

// ErrParentIsSubTask is returned when a task is asked to hang under one that
// is itself a sub-task. Only one level: what hangs under a task is never
// something with more hanging under it.
var ErrParentIsSubTask = errors.New("a sub-task cannot carry sub-tasks")

// ErrParentIsSelf is returned when a task is asked to hang under itself.
var ErrParentIsSelf = errors.New("a task cannot hang under itself")

// ErrCarriesSubTasks is returned when a task carrying sub-tasks is asked to
// be deleted. What depends on it would go with it, and a deletion that takes
// more than the number typed is a deletion nobody asked for.
var ErrCarriesSubTasks = errors.New("the task carries sub-tasks")

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
	objective   int
	title       string
	description string
	state       domain.State
	priority    domain.Priority
	labels      []string
	deadline    time.Time
	estimate    time.Duration
	timeSpent   time.Duration
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
	Due(time.Time)
	NotDue()
	Estimated(time.Duration)
	NotEstimated()
	Spent(time.Duration)
	NothingSpent()
	Under(int)
	UnderNothing()
	Serves(int)
	ServesNothing()
}

// Tell says every fact this task carries to the given membrane, absence
// included, in the order a card lays them out.
func (t Task) Tell(to Told) {
	to.Numbered(t.number)
	to.Titled(t.title)
	to.InColumn(t.state)
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
	if t.objective == NoObjective {
		to.ServesNothing()
	} else {
		to.Serves(t.objective)
	}
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

// MayServeAnObjective says whether a task may carry an objective alongside
// whatever parent it carries: a sub-task never does, since it counts towards
// its parent's objective by way of that parent and is never given one of its
// own.
func (t Task) MayServeAnObjective() error {
	if t.parent != NoParent && t.objective != NoObjective {
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

// MayDelete says whether the task carrying number may be deleted, given what
// hangs under it. What depends on it would go with it, and a deletion that
// takes more than the number typed is a deletion nobody asked for.
func MayDelete(number int, hanging []Task) error {
	if len(hanging) > 0 {
		return fmt.Errorf("task %d: %w", number, ErrCarriesSubTasks)
	}
	return nil
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
