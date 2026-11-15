package task

import (
	"cairn/internal/engine/item"
	"errors"
)

// Step is one of the smaller pieces a task is split into: what it is called, and whether it is done. It is not a task — it has no number, column, focus, time or notes of its own.
type Step struct {
	Title string `json:"title" jsonschema:"what the step is called"`
	Done  bool   `json:"done" jsonschema:"true once the step is done"`
}

// ErrNoSuchStep is returned when a task is asked for a step at a place it has none.
var ErrNoSuchStep = errors.New("the task has no such step")

// NewStep builds a step, refusing one that carries no text.
func NewStep(raw string, done bool) (Step, error) {
	title, err := item.Title("a step", raw)
	if err != nil {
		return Step{}, err
	}
	return Step{Title: title, Done: done}, nil
}

// Ticked is this task with every one of its steps done: what finishing it, once confirmed, does to them.
func (t Task) Ticked() Task {
	steps := make([]Step, len(t.steps))
	for i, s := range t.steps {
		steps[i] = Step{Title: s.Title, Done: true}
	}
	t.steps = steps
	return t
}

// ErrDoneAndNotDone is returned when a step is asked to be ticked and unticked in one go.
var ErrDoneAndNotDone = errors.New("a step is done or not done, not both")

// Steps is how a task's steps cross to a machine: the task they belong to, and each step with its place.
type Steps struct {
	Task  int         `json:"task"`
	Steps []StepShape `json:"steps"`
}

// StepShape is one step as it crosses, with the place it is named by.
type StepShape struct {
	Place int    `json:"place"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// OfSteps is a task's steps as they cross.
func OfSteps(t Task) Steps {
	said := Steps{Task: t.Carried().Number, Steps: []StepShape{}}
	for i, s := range t.Carried().Steps {
		said.Steps = append(said.Steps, StepShape{Place: i + 1, Title: s.Title, Done: s.Done})
	}
	return said
}
