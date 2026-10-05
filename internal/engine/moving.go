package engine

// Moving a task, and which task is being worked on. Both are about the one
// order there is rather than about what a task carries, which is why they are
// here together and not among the five that change a task's attributes.

import (
	"errors"
	"fmt"
	"slices"

	"cairn/internal/task"
	"cairn/internal/trace"
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
	Focus() ([]int, error)
	AddFocus(int) error
	DropFocus(int) error
	ClearFocus() error
}

// move puts a task in a place among the ones it sits with. It never changes
// the column: that is an edit, and one act with two meanings is an act nobody
// can undo confidently.
func (e *Engine) move(number int, at Place, target int) error {
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
		if err := e.places.MoveAfter(number, target); err != nil {
			return err
		}
	case First:
		if err := e.places.MoveFirst(number); err != nil {
			return err
		}
	default:
		if err := e.places.MoveLast(number); err != nil {
			return err
		}
	}
	return e.record(number, trace.Moved, "", "", placeNamed(at, target))
}

// placeNamed is how a place reads in a history. A move records where it went
// and not what that did to the places of its sisters: nineteen lines saying a
// number went up by one would bury the one line saying what a person did.
func placeNamed(at Place, target int) string {
	switch at {
	case First:
		return "first"
	case Behind:
		return fmt.Sprintf("behind task %d", target)
	default:
		return "last"
	}
}

// focus says which tasks are being worked on, oldest first, and nothing when
// none are. A listing asks for the numbers rather than the tasks: it marks
// lines, and reading every task to mark a few would be a second query per
// listing.
func (e *Engine) focus() ([]int, error) {
	return e.places.Focus()
}

// putInFocus takes up a task alongside whatever else is being worked on, and
// hands it back so whoever asked sees what they are on.
//
// Alongside and not instead. One task at a time is how a person works, and how
// this read for a long while; but an agent session is started per task, and
// several run at once — a focus that held one made every other one invisible
// to whoever asked what was being worked on.
func (e *Engine) putInFocus(number int) (task.Task, error) {
	t, err := e.tasks.Find(number)
	if err != nil {
		return task.Task{}, err
	}
	held, err := e.places.Focus()
	if err != nil {
		return task.Task{}, err
	}
	// Already being worked on: nothing happens, and nothing is written down.
	// Letting it through would open a second stretch on a task whose first has
	// not closed, and every duration after it would count the same work twice.
	if slices.Contains(held, number) {
		return t, nil
	}
	if err := e.places.AddFocus(number); err != nil {
		return task.Task{}, err
	}
	if err := e.record(number, trace.FocusTaken, "", "", ""); err != nil {
		return task.Task{}, err
	}
	return t, nil
}

// letGoOf leaves one task and keeps the rest. Letting go of one that was never
// taken up is not a failure — but it is not an act either, and nothing is
// written down: a stretch closing that never opened would fault every duration
// around it.
func (e *Engine) letGoOf(number int) error {
	held, err := e.places.Focus()
	if err != nil {
		return err
	}
	if err := e.places.DropFocus(number); err != nil {
		return err
	}
	if !slices.Contains(held, number) {
		return nil
	}
	return e.record(number, trace.FocusReleased, "", "", "")
}

// letGo leaves nothing in focus at all, and writes down the release of every
// task that was being worked on. Every one of them: a release that named only
// the first would leave the others' stretches open for ever.
func (e *Engine) letGo() error {
	held, err := e.places.Focus()
	if err != nil {
		return err
	}
	if err := e.places.ClearFocus(); err != nil {
		return err
	}
	for _, number := range held {
		if err := e.record(number, trace.FocusReleased, "", "", ""); err != nil {
			return err
		}
	}
	return nil
}

// inFocus hands back the tasks being worked on and their numbers, and nothing
// at all when none are. The numbers come back alongside so a caller answering
// about the focus does not have to ask twice for them.
func (e *Engine) inFocus() ([]task.Task, []int, error) {
	numbers, err := e.focus()
	if err != nil {
		return nil, nil, err
	}
	held := make([]task.Task, 0, len(numbers))
	for _, number := range numbers {
		t, err := e.tasks.Find(number)
		if err != nil {
			return nil, nil, fmt.Errorf("reading a task in focus: %w", err)
		}
		held = append(held, t)
	}
	return held, numbers, nil
}
