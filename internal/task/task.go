// Package task holds the task: what it carries, the rules that keep it valid, and the membrane anything uses to read or build one.
package task

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
)

// NoParent is the number a task carries when it hangs under nothing.
const NoParent = 0

// NoObjective is the number nothing serves.
const NoObjective = 0

// ErrParentIsSubTask is returned when a task is asked to hang under one that is itself a sub-task.
var ErrParentIsSubTask = errors.New("a sub-task cannot carry sub-tasks")

// ErrParentIsSelf is returned when a task is asked to hang under itself.
var ErrParentIsSelf = errors.New("a task cannot hang under itself")

// ErrSubTaskServesNoObjective is returned when a task is asked to carry both a parent and an objective at once.
var ErrSubTaskServesNoObjective = errors.New("a sub-task serves no objective")

// Task is the unit of work.
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
	archived    bool
}

// Told is the membrane of a task: one method per fact it has to tell, absence included.
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

// Tell says every fact this task carries to the given membrane, absence included, in the order a card lays them out.
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
	domain.TellOrNot(t.description, to.Described, to.NotDescribed)
	domain.TellAll(t.labels, to.Labelled, to.NotLabelled, to.LabelsTold)
	domain.TellAll(t.blockers, to.BlockedBy, to.NotBlocked, to.BlockersTold)
	domain.TellOrNot(t.because, to.Because, to.NotBecause)
	domain.TellOrNot(t.deadline, to.Due, to.NotDue)
	domain.TellOrNot(t.estimate, to.Estimated, to.NotEstimated)
	domain.TellOrNot(t.timeSpent, to.Spent, to.NothingSpent)
	domain.TellOrNot(t.parent, to.Under, to.UnderNothing)
	domain.TellAll(t.objectives, to.Serves, to.ServesNothing, to.ObjectivesTold)
}

// MayCarrySubTasks says whether a task may have another hang under it.
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
var ErrPartGoesWithItsWork = errors.New("a part is archived with the work it belongs to, and unarchived with it")

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

// MayUnarchive says whether a task can be brought back onto the board.
func (t Task) MayUnarchive() error {
	if !t.archived {
		return fmt.Errorf("task %d: %w", t.number, ErrNotArchived)
	}
	return nil
}

// ErrSubTaskHasNoColumn is returned when a sub-task is put in a column that says where work stands.
var ErrSubTaskHasNoColumn = errors.New("a sub-task is done or it is not")

// MayCarryItsColumn says whether this task may stand in the column it carries.
func (t Task) MayCarryItsColumn() error {
	if t.parent == NoParent || t.state == domain.Done || t.state == domain.Todo {
		return nil
	}
	return fmt.Errorf("task %d in %s: %w", t.number, t.state, ErrSubTaskHasNoColumn)
}

// Settled is this task with a column a sub-task may carry: done stays done and anything else becomes todo.
func (t Task) Settled() Task {
	if t.parent == NoParent || t.state == domain.Done {
		return t
	}
	t.state = domain.Todo
	return t
}

// MayServeAnObjective says whether a task may carry an objective alongside whatever parent it carries: a sub-task never does, since it counts towards its parent's objective by way of that parent and is never given one of its own.
func (t Task) MayServeAnObjective() error {
	if t.parent != NoParent && len(t.objectives) > 0 {
		return ErrSubTaskServesNoObjective
	}
	return nil
}

// MayHangUnderItsParent says whether the parent this task already carries is one it may take: never itself.
func (t Task) MayHangUnderItsParent() error {
	if t.parent == t.number {
		return fmt.Errorf("task %d: %w", t.number, ErrParentIsSelf)
	}
	return nil
}

// ErrSubTaskStillOpen is returned when a task is asked to be finished while something hanging under it is not.
var ErrSubTaskStillOpen = errors.New("something hanging under it is not done")

// MayFinish says whether the task carrying number may be moved to done, given what hangs under it.
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
func NewTask(raw string) (Task, error) {
	title, err := domain.Title("a task", raw)
	if err != nil {
		return Task{}, err
	}
	return Task{title: title, state: domain.Backlog, priority: domain.Normal}, nil
}
