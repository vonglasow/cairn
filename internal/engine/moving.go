package engine

import (
	"errors"
	"fmt"
	"slices"

	"cairn/internal/domain"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// The reasons a move is refused, re-exported so a surface can recognise them: a surface asks the engine and may not import the task, so the names it answers with have to live here too.
var (
	ErrMovedBehindItself = task.ErrMovedBehindItself
	ErrOtherColumn       = task.ErrOtherColumn
)

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

type places interface {
	MoveFirst(int) (bool, error)
	MoveLast(int) (bool, error)
	MoveAfter(int, int) (bool, error)
	Focus() ([]int, error)
	AddFocus(int) error
	DropFocus(int) error
	ClearFocus() error
}

func (e *Engine) move(number int, at Place, target int) error {
	moved, err := e.moveTo(number, at, target)
	if err != nil || !moved {
		return err
	}
	return e.record(number, trace.Moved, "", "", placeNamed(at, target))
}

func (e *Engine) moveInto(number int, state string, finishing bool) error {
	was, err := e.tasks.Find(number)
	if err != nil {
		return err
	}
	if factsOf(was).State == domain.State(state) {
		return nil
	}
	_, err = e.edit(number, Changes{State: &state, Finishing: finishing})
	return editReason(err)
}

func (e *Engine) moveTo(number int, at Place, target int) (bool, error) {
	moving, err := e.tasks.Find(number)
	if err != nil {
		return false, err
	}
	if at == First {
		return e.places.MoveFirst(number)
	}
	if at != Behind {
		return e.places.MoveLast(number)
	}
	sits, err := e.tasks.Find(target)
	if err != nil {
		return false, err
	}
	if err := moving.MayMoveBehind(sits); err != nil {
		return false, err
	}
	return e.places.MoveAfter(number, target)
}

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

func (e *Engine) focus() ([]int, error) {
	return e.places.Focus()
}

func (e *Engine) putInFocus(number int) (task.Task, error) {
	t, err := e.tasks.Find(number)
	if err != nil {
		return task.Task{}, err
	}
	if factsOf(t).State == domain.Done {
		return task.Task{}, ErrDoneHoldsNoFocus
	}
	held, err := e.places.Focus()
	if err != nil {
		return task.Task{}, err
	}
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

// TaskMove changes where a task sits in the order.
var TaskMove = Operation{
	Thing: "task",
	Verb:  "move",
	Short: "Move a task in the order",
	Does:  "Move a task to another place among the ones it sits with, or into another column and a place in it, in one act; the tasks named under also follow it there, one behind the other, in the same act. Moved into done while one of its steps is still open, it is refused unless confirmed; confirmed, its steps are ticked with it.",
	Example: `cairn task move 3 --first
cairn task move 3 --after 1
cairn task move 3 --state todo --after 1`,
	Finishes: true,
	Level:    Write,
	Needs:    AStore,
	Again:    ChangesNothingMore,
	Answers:  Tasks{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
		{Name: "first", Kind: Bool, Means: "put it at the head of the ones it sits with"},
		{Name: "after", Kind: Number, Of: "a task", Means: "put it directly behind this task"},
		{Name: "last", Kind: Bool, Means: "put it behind all the ones it sits with"},
		{Name: "state", Kind: Text, OneOf: oneOf(domain.States), Outside: domain.ErrNotAState, Means: "the column to move it into first, in the same act"},
		alsoOf("a task"),
	},
	Run: func(on On, in Input) (Answer, error) {
		numbers, err := together(in)
		if err != nil {
			return Answer{}, err
		}
		n := numbers[0]
		into := in.Change("state")
		if into != nil {
			if err := on.Engine.moveInto(n, *into, in.Confirmed); err != nil {
				return Answer{}, err
			}
		}
		at, target, err := placeFrom(in)
		if errors.Is(err, ErrNowhereToMove) && into != nil {
			return listing(on, Input{})
		}
		if err != nil {
			return Answer{}, err
		}
		if err := on.Engine.move(n, at, target); err != nil {
			return Answer{}, err
		}
		if err := on.Engine.alongBehind(n, numbers[1:], in.Confirmed); err != nil {
			return Answer{}, err
		}
		return listing(on, Input{})
	},
}

func (e *Engine) alongBehind(first int, others []int, finishing bool) error {
	led, err := e.tasks.Find(first)
	if err != nil {
		return err
	}
	state, behind := string(factsOf(led).State), first
	for _, number := range others {
		if err := e.moveInto(number, state, finishing); err != nil {
			return err
		}
		if err := e.move(number, Behind, behind); err != nil {
			return err
		}
		behind = number
	}
	return nil
}

// TaskFocus says which task is being worked on, or puts one there.
var TaskFocus = Operation{
	Thing: "task",
	Verb:  "focus",
	Short: "Put a task in focus, or say which ones are",
	Does:  "Put a task in focus alongside whatever else is being worked on, or name every task in focus. Several can be at once.",
	Example: `cairn task focus 3
cairn task focus`,
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Focus{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a task", Means: "the task to focus, or nothing to name every task in focus"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, named := in.Whole("number")
		if !named {
			return whatIsInFocus(on)
		}
		t, err := on.Engine.putInFocus(n)
		if err != nil {
			return Answer{}, err
		}
		p, err := on.Engine.taskProgress(n)
		if err != nil {
			return Answer{}, err
		}
		held, err := on.Engine.focus()
		if err != nil {
			return Answer{}, err
		}
		return Answer{Task: &t, Progress: p, InFocus: &held}, nil
	},
}

// TaskUnfocus leaves nothing in focus.
var TaskUnfocus = Operation{
	Thing: "task",
	Verb:  "unfocus",
	Short: "Unfocus a task in focus, or all of them",
	Does:  "Unfocus one task and keep the rest in focus, or every task in focus when no number is given. Unfocusing nothing is not a failure.",
	Example: `cairn task unfocus 3
cairn task unfocus`,
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Focus{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a task", Means: "the task to unfocus, or nothing to unfocus them all"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, named := in.Whole("number")
		if named {
			if err := on.Engine.letGoOf(n); err != nil {
				return Answer{}, err
			}
		} else if err := on.Engine.letGo(); err != nil {
			return Answer{}, err
		}
		return whatIsInFocus(on)
	},
}

func placeFrom(in Input) (Place, int, error) {
	named := make([]Place, 0, 3)
	target := 0
	if in.Yes("first") {
		named = append(named, First)
	}
	if in.Given["after"] != "" {
		named = append(named, Behind)
		target, _ = in.Whole("after")
	}
	if in.Yes("last") {
		named = append(named, Last)
	}
	if len(named) == 0 {
		return 0, 0, ErrNowhereToMove
	}
	if len(named) > 1 {
		return 0, 0, ErrMoreThanOnePlace
	}
	return named[0], target, nil
}

func whatIsInFocus(on On) (Answer, error) {
	held, numbers, err := on.Engine.inFocus()
	if err != nil {
		return Answer{}, err
	}
	return Answer{Held: held, InFocus: &numbers}, nil
}
