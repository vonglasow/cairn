package engine

// Moving a task, and which task is being worked on. Both are about the one
// order there is rather than about what a task carries, which is why they are
// here together and not among the five that change a task's attributes.

import (
	"errors"
	"fmt"

	"cairn/internal/store"
	"cairn/internal/task"
)

// The reasons a move is refused, re-exported so a surface can recognise them:
// a surface asks the engine and may not import the task, so the names it
// answers with have to live here too.
var (
	ErrMovedBehindItself  = task.ErrMovedBehindItself
	ErrOtherColumn        = task.ErrOtherColumn
	ErrNotAmongItsSisters = task.ErrNotAmongItsSisters
)

// ErrNowhereToMove is returned when a move names no place to go. A command
// that parsed is not a command that asked for something.
var ErrNowhereToMove = errors.New("a task is moved to the first place, behind another, or to the last")

// ErrMoreThanOnePlace is returned when a move names several places at once.
// Picking one quietly would move the task somewhere nobody asked for.
var ErrMoreThanOnePlace = errors.New("a task is moved to one place, not several")

// Place is where a move puts a task among the ones it sits with.
type Place int

const (
	_ Place = iota // the zero value: never a place.
	// First is the head of the tasks it sits with.
	First
	// Behind is directly after another task.
	Behind
	// Last is after all the tasks it sits with.
	Last
)

// places is what the engine needs from whatever keeps the order and the focus.
// Declared beside tasks, by the consumer, for the same reason.
type places interface {
	MoveFirst(int) error
	MoveLast(int) error
	MoveAfter(int, int) error
	Focus() (int, error)
	SetFocus(int) error
	ClearFocus() error
}

// Move puts a task in a place among the ones it sits with. It never changes
// the column: that is an edit, and one act with two meanings is an act nobody
// can undo confidently.
func (e *Engine) Move(number int, at Place, target int) error {
	moving, err := e.tasks.Find(number)
	if err != nil {
		return err
	}
	switch at {
	case Behind:
		sits, err := e.tasks.Find(target)
		if err != nil {
			return err
		}
		if err := moving.MayMoveBehind(sits); err != nil {
			return err
		}
		return e.places.MoveAfter(number, target)
	case First:
		return e.places.MoveFirst(number)
	default:
		return e.places.MoveLast(number)
	}
}

// Focus says which task is being worked on, or zero when none is. A listing
// asks for the number rather than the task: it marks a line, and reading the
// whole task to mark one would be a second query per listing.
func (e *Engine) Focus() (int, error) {
	return e.places.Focus()
}

// PutInFocus makes a task the one being worked on, releasing whichever was
// before, and hands back the task so whoever asked sees what they are on.
func (e *Engine) PutInFocus(number int) (task.Task, error) {
	t, err := e.tasks.Find(number)
	if err != nil {
		return task.Task{}, err
	}
	if err := e.places.SetFocus(number); err != nil {
		return task.Task{}, err
	}
	return t, nil
}

// LetGo leaves nothing in focus. Letting go of nothing is not a failure, so
// this answers the same either way.
func (e *Engine) LetGo() error {
	return e.places.ClearFocus()
}

// InFocus hands back the task being worked on and its number, or an empty task
// and store.NoFocus when none is. The number comes back alongside so a caller
// answering about the focus does not have to ask twice for it.
func (e *Engine) InFocus() (task.Task, int, error) {
	number, err := e.Focus()
	if err != nil {
		return task.Task{}, store.NoFocus, err
	}
	if number == store.NoFocus {
		return task.Task{}, store.NoFocus, nil
	}
	t, err := e.tasks.Find(number)
	if err != nil {
		return task.Task{}, store.NoFocus, fmt.Errorf("reading the task in focus: %w", err)
	}
	return t, number, nil
}
