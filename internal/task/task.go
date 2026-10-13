// Package task holds the task: what it carries, the rules that keep it valid, and the membrane anything uses to read or build one.
package task

import (
	"errors"
	"fmt"
	"slices"
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

// Carried is everything a task carries, as plain facts.
type Carried struct {
	Number, Parent       int
	Title, Description   string
	State                domain.State
	Priority             domain.Priority
	Labels               []string
	Blockers, Objectives []int
	Because              string
	Deadline             time.Time
	Estimate, Spent      time.Duration
	Archived             bool
}

// Carried hands over what the task carries.
func (t Task) Carried() Carried {
	return Carried{
		Number: t.number, Parent: t.parent, Title: t.title, Description: t.description,
		State: t.state, Priority: t.priority, Labels: slices.Clone(t.labels),
		Blockers: slices.Clone(t.blockers), Objectives: slices.Clone(t.objectives),
		Because: t.because, Deadline: t.deadline, Estimate: t.estimate, Spent: t.timeSpent,
		Archived: t.archived,
	}
}

// Made is the task that carries what is given.
func Made(c Carried) Task {
	return Task{
		number: c.Number, parent: c.Parent, title: c.Title, description: c.Description,
		state: c.State, priority: c.Priority, labels: c.Labels, blockers: c.Blockers,
		objectives: c.Objectives, because: c.Because, deadline: c.Deadline,
		estimate: c.Estimate, timeSpent: c.Spent, archived: c.Archived,
	}
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
