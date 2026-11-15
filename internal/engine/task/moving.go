package task

import (
	"cairn/internal/engine/item"
	"errors"
	"fmt"
)

// ErrMovedBehindItself is returned when a task is asked to be moved behind itself.
var ErrMovedBehindItself = errors.New("a task cannot be moved behind itself")

// ErrOtherColumn is returned when a task is asked to be moved behind one in another column.
var ErrOtherColumn = errors.New("a task is moved inside its own column")

// MayMoveBehind says whether this task may take the place behind target.
func (t Task) MayMoveBehind(target Task) error {
	if t.number == target.number {
		return fmt.Errorf("task %d: %w", t.number, ErrMovedBehindItself)
	}
	if t.state != target.state {
		return fmt.Errorf("task %d is in %s and task %d is in %s: %w",
			t.number, t.state, target.number, target.state, ErrOtherColumn)
	}
	return nil
}

// ErrNowhereToMove is returned when a move names no place to go.
var ErrNowhereToMove = errors.New("a task is moved to the first place, behind another, or to the last")

// ErrMoreThanOnePlace is returned when a move names several places at once.
var ErrMoreThanOnePlace = errors.New("a task is moved to one place, not several")

// ErrDoneHoldsNoFocus is returned when a task already done is put in focus: there is nothing left to work on.
var ErrDoneHoldsNoFocus = errors.New("a task that is done cannot be in focus")

// Place is where a move puts a task among the ones it sits with.
type Place int

const (
	_ Place = iota
	// First is the head of the tasks it sits with.
	First
	// Behind is directly after another task.
	Behind
	// Last is after all the tasks it sits with.
	Last
)

// PlaceNamed is how the trace names the place a move put a task.
func PlaceNamed(at Place, target int) string {
	switch at {
	case First:
		return "first"
	case Behind:
		return fmt.Sprintf("behind task %d", target)
	default:
		return "last"
	}
}

// WaitingForSomebody is the columns a task focused on is brought into current from.
var WaitingForSomebody = []item.State{item.Backlog, item.Todo, item.Paused}

// Crossed is the column a move took a task out of, and the one it took it into.
type Crossed struct {
	Task int    `json:"task" jsonschema:"the number of the task that changed column"`
	From string `json:"from" jsonschema:"the column it was in"`
	To   string `json:"to" jsonschema:"the column it is in now"`
}
