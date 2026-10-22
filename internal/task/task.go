// Package task holds the task: what it carries, the rules that keep it valid, and the membrane anything uses to read or build one.
package task

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"cairn/internal/domain"
)

// NoObjective is the number nothing serves.
const NoObjective = 0

// Task is the unit of work.
type Task struct {
	number      int
	objectives  []int
	title       string
	description string
	state       domain.State
	priority    domain.Priority
	labels      []string
	blockers    []int
	linked      []int
	because     string
	deadline    time.Time
	estimate    time.Duration
	timeSpent   time.Duration
	archived    bool
	steps       []Step
}

// Carried is everything a task carries, as plain facts.
type Carried struct {
	Number               int
	Title, Description   string
	State                domain.State
	Priority             domain.Priority
	Labels               []string
	Blockers, Objectives []int
	Because              string
	Deadline             time.Time
	Estimate, Spent      time.Duration
	Archived             bool
	Steps                []Step
	Linked               []int
}

// Carried hands over what the task carries.
func (t Task) Carried() Carried {
	return Carried{
		Number: t.number, Title: t.title, Description: t.description,
		State: t.state, Priority: t.priority, Labels: slices.Clone(t.labels),
		Blockers: slices.Clone(t.blockers), Objectives: slices.Clone(t.objectives),
		Because: t.because, Deadline: t.deadline, Estimate: t.estimate, Spent: t.timeSpent,
		Archived: t.archived, Steps: slices.Clone(t.steps),
		Linked: slices.Clone(t.linked),
	}
}

// Made is the task that carries what is given.
func Made(c Carried) Task {
	return Task{
		number: c.Number, title: c.Title, description: c.Description,
		state: c.State, priority: c.Priority, labels: c.Labels, blockers: c.Blockers,
		objectives: c.Objectives, because: c.Because, deadline: c.Deadline,
		estimate: c.Estimate, timeSpent: c.Spent, archived: c.Archived, steps: c.Steps,
		linked: c.Linked,
	}
}

// ErrAlreadyArchived is returned when a task off the board is archived again.
var ErrAlreadyArchived = errors.New("the task is already archived")

// ErrNotArchived is returned when a task on the board is brought back onto it.
var ErrNotArchived = errors.New("the task is on the board")

// MayArchive says whether this task may be taken off the board.
func (t Task) MayArchive() error {
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

// NewTask builds a task from a title, refusing one that carries no text.
func NewTask(raw string) (Task, error) {
	title, err := domain.Title("a task", raw)
	if err != nil {
		return Task{}, err
	}
	return Task{title: title, state: domain.Backlog, priority: domain.Normal}, nil
}
