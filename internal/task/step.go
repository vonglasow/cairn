package task

import (
	"errors"

	"cairn/internal/domain"
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
	title, err := domain.Title("a step", raw)
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
