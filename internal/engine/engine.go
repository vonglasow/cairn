package engine

import (
	"errors"
	"fmt"

	"cairn/internal/domain"
	"cairn/internal/store"
)

// ErrNothingToChange is returned when an edit names nothing to change. Doing
// nothing quietly would look like success to whoever asked.
var ErrNothingToChange = errors.New("nothing was named to change")

// ErrStore is returned when the store itself cannot be used — as opposed to a
// request the store understood and refused. A surface tells the two apart to
// answer with the right exit code, so a script knows whether to retry or to
// fix its arguments.
var ErrStore = errors.New("the store cannot be read")

// ErrLabelSignAtCreation is returned when a task is created with a label
// carrying a + or a -. Those say how an existing set changes, and a task being
// created has no set yet — accepting them would mean `--label -spec` quietly
// doing nothing at all.
var ErrLabelSignAtCreation = errors.New("a label is added or taken away only when changing a task")

// ErrNoSuchTask is re-exported so a surface can recognise it: a surface asks
// the engine and may not import the store, so the engine is where the name it
// answers with has to live.
var ErrNoSuchTask = store.ErrNoSuchTask

// tasks is what the engine needs from whatever keeps its tasks. Declared here,
// by the consumer, so the rules stay testable and a second backing can arrive
// without the engine noticing.
type tasks interface {
	Keep(domain.Task) (domain.Task, error)
	Update(domain.Task) error
	Delete(int) error
	Find(int) (domain.Task, error)
	All() ([]domain.Task, error)
	Close() error
}

// Engine carries out the operations. Every change of state goes through it.
type Engine struct {
	tasks tasks
}

// Changes are the attributes an operation was given. A nil field was not
// named at all; a field pointing at an empty value was named and emptied —
// which is how an attribute is taken away.
type Changes struct {
	Title       *string
	Description *string
	State       *string
	Priority    *string
	Labels      *[]string
	Deadline    *string
	Estimate    *string
}

// named says whether anything at all was given to change.
func (c Changes) named() bool {
	return c.Title != nil || c.Description != nil || c.State != nil ||
		c.Priority != nil || c.Labels != nil || c.Deadline != nil || c.Estimate != nil
}

// Open opens the store at path and builds the engine over it.
func Open(path string) (*Engine, error) {
	kept, err := store.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %w", path, ErrStore, err)
	}
	return &Engine{tasks: kept}, nil
}

// Close releases the store.
func (e *Engine) Close() error {
	return e.tasks.Close()
}

// Add records a new task carrying what it was given, and nothing else.
func (e *Engine) Add(title string, c Changes) (domain.Task, error) {
	task, err := domain.NewTask(title)
	if err != nil {
		return domain.Task{}, err
	}
	if task, err = apply(task, c, creating); err != nil {
		return domain.Task{}, err
	}
	return e.tasks.Keep(task)
}

// List hands back every task, in the order they were added.
func (e *Engine) List() ([]domain.Task, error) {
	return e.tasks.All()
}

// Show hands back the task carrying number.
func (e *Engine) Show(number int) (domain.Task, error) {
	return e.tasks.Find(number)
}

// Delete removes the task carrying number.
func (e *Engine) Delete(number int) error {
	return e.tasks.Delete(number)
}

// Edit changes what was named on a task and leaves the rest alone.
func (e *Engine) Edit(number int, c Changes) (domain.Task, error) {
	if !c.named() {
		return domain.Task{}, ErrNothingToChange
	}
	task, err := e.tasks.Find(number)
	if err != nil {
		return domain.Task{}, err
	}
	if task, err = apply(task, c, changing); err != nil {
		return domain.Task{}, err
	}
	if err := e.tasks.Update(task); err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

// Whether a task is being created or changed. Labels are the only attribute
// that cares: a sign says how an existing set moves, and there is no set to
// move at creation.
const (
	creating = true
	changing = false
)

// apply lays the changes over a task, refusing any value it cannot make sense
// of. A field that was not named is left exactly as it was.
func apply(task domain.Task, c Changes, atCreation bool) (domain.Task, error) {
	if c.Title != nil {
		retitled, err := domain.NewTask(*c.Title)
		if err != nil {
			return domain.Task{}, err
		}
		task.Title = retitled.Title
	}
	if c.Description != nil {
		task.Description = *c.Description
	}
	if c.State != nil {
		state, err := domain.ParseState(*c.State)
		if err != nil {
			return domain.Task{}, err
		}
		task.State = state
	}
	if c.Priority != nil {
		priority, err := domain.ParsePriority(*c.Priority)
		if err != nil {
			return domain.Task{}, err
		}
		task.Priority = priority
	}
	if c.Labels != nil {
		change, err := domain.ParseLabelChange(*c.Labels)
		if err != nil {
			return domain.Task{}, err
		}
		if atCreation && change.Adjusts() {
			return domain.Task{}, ErrLabelSignAtCreation
		}
		labels, err := change.Apply(task.Labels)
		if err != nil {
			return domain.Task{}, err
		}
		task.Labels = labels
	}
	if c.Deadline != nil {
		deadline, err := domain.ParseDeadline(*c.Deadline)
		if err != nil {
			return domain.Task{}, err
		}
		task.Deadline = deadline
	}
	if c.Estimate != nil {
		estimate, err := domain.ParseEstimate(*c.Estimate)
		if err != nil {
			return domain.Task{}, err
		}
		task.Estimate = estimate
	}
	return task, nil
}
