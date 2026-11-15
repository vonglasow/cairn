package engine

import (
	"cairn/internal/engine/item"
	"errors"
	"fmt"
	"slices"

	"cairn/internal/engine/act"
	"cairn/internal/engine/task"
)

func (e *Engine) move(number int, at task.Place, target int) error {
	moved, err := e.moveTo(number, at, target)
	if err != nil || !moved {
		return err
	}
	return e.record(number, act.Moved, "", "", task.PlaceNamed(at, target))
}

func (e *Engine) moveInto(number int, state string, finishing bool) error {
	was, err := e.tasks.Find(number)
	if err != nil {
		return err
	}
	if was.Carried().State == item.State(state) {
		return nil
	}
	_, err = e.edit(number, Changes{State: &state, Finishing: finishing})
	return editReason(err)
}

func (e *Engine) moveTo(number int, at task.Place, target int) (bool, error) {
	moving, err := e.tasks.Find(number)
	if err != nil {
		return false, err
	}
	if at == task.First {
		return e.places.MoveFirst(number)
	}
	if at != task.Behind {
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

func (e *Engine) focus() ([]int, error) {
	return e.places.Focus()
}

func (e *Engine) putInFocus(number int) (task.Task, error) {
	t, err := e.tasks.Find(number)
	if err != nil {
		return task.Task{}, err
	}
	if t.Carried().State == item.Done {
		return task.Task{}, task.ErrDoneHoldsNoFocus
	}
	held, err := e.places.Focus()
	if err != nil {
		return task.Task{}, err
	}
	if slices.Contains(held, number) {
		return t, e.renewed(number)
	}
	if err := e.places.AddFocus(number); err != nil {
		return task.Task{}, err
	}
	became := ""
	if !e.timeCounted() {
		became = act.WithoutATimer
	}
	if err := e.record(number, act.FocusTaken, "", "", became); err != nil {
		return task.Task{}, err
	}
	return e.broughtIntoCurrent(t)
}

func (e *Engine) broughtIntoCurrent(t task.Task) (task.Task, error) {
	number := t.Carried().Number
	if !slices.Contains(task.WaitingForSomebody, t.Carried().State) {
		return t, nil
	}
	if err := e.moveInto(number, string(item.Current), false); err != nil {
		return task.Task{}, err
	}
	if err := e.move(number, task.First, 0); err != nil {
		return task.Task{}, err
	}
	return e.tasks.Find(number)
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
	return e.record(number, act.FocusReleased, "", "", "")
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
		if err := e.record(number, act.FocusReleased, "", "", ""); err != nil {
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
	Does:  "Move a task to another place among the ones it sits with, or into another column and a place in it, in one act — behind a task in another column, it goes into that column; the tasks named under also follow it there, one behind the other, in the same act. Moved into done while one of its steps is still open, it is refused unless confirmed; confirmed, its steps are ticked with it.",
	Example: `cairn task move 3 --first
cairn task move 3 --after 1
cairn task move 3 --state todo --after 1`,
	Finishes: true,
	Level:    write,
	Needs:    aStore,
	Again:    ChangesNothingMore,
	Answers:  Tasks{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
		{Name: "first", Kind: Bool, Means: "put it at the head of the ones it sits with"},
		{Name: "after", Kind: Number, Of: "a task", Means: "put it directly behind this task, in this task's column"},
		{Name: "last", Kind: Bool, Means: "put it behind all the ones it sits with"},
		{Name: "state", Kind: Text, OneOf: oneOf(item.States), Outside: item.ErrNotAState, Means: "the column to move it into first, in the same act"},
		alsoOf("a task"),
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		numbers, err := together(in)
		if err != nil {
			return Answer{}, err
		}
		n := numbers[0]
		at, target, placed := placeFrom(in)
		into := in.change("state")
		if placed != nil && (into == nil || !errors.Is(placed, task.ErrNowhereToMove)) {
			return Answer{}, placed
		}
		was, err := on.Engine.tasks.Find(n)
		if err != nil {
			return Answer{}, err
		}
		states, err := on.Engine.statesOf(numbers)
		if err != nil {
			return Answer{}, err
		}
		if into == nil && at == task.Behind {
			if into, err = on.Engine.columnOf(target); err != nil {
				return Answer{}, err
			}
		}
		if into != nil {
			if err := on.Engine.moveInto(n, *into, in.Confirmed); err != nil {
				return Answer{}, err
			}
		}
		if placed == nil {
			if err := on.Engine.move(n, at, target); err != nil {
				return Answer{}, err
			}
		}
		if err := on.Engine.alongBehind(n, numbers[1:], in.Confirmed); err != nil {
			return Answer{}, err
		}
		answer, err := listing(on, Input{})
		if err != nil {
			return Answer{}, err
		}
		if answer.Crossed, err = on.Engine.crossed(was); err != nil {
			return Answer{}, err
		}
		answer.Warnings, err = on.Engine.unmetEach(states)
		return answer, err
	},
}

func (e *Engine) crossed(was task.Task) (*task.Crossed, error) {
	now, err := e.tasks.Find(was.Carried().Number)
	if err != nil || now.Carried().State == was.Carried().State {
		return nil, err
	}
	return &task.Crossed{Task: was.Carried().Number, From: string(was.Carried().State), To: string(now.Carried().State)}, nil
}

func (e *Engine) columnOf(number int) (*string, error) {
	t, err := e.tasks.Find(number)
	if err != nil {
		return nil, err
	}
	column := string(t.Carried().State)
	return &column, nil
}

func (e *Engine) alongBehind(first int, others []int, finishing bool) error {
	led, err := e.tasks.Find(first)
	if err != nil {
		return err
	}
	state, behind := string(led.Carried().State), first
	for _, number := range others {
		if err := e.moveInto(number, state, finishing); err != nil {
			return err
		}
		if err := e.move(number, task.Behind, behind); err != nil {
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
	Does:  "Put a task in focus alongside whatever else is being worked on, or name every task in focus. Several can be at once. A task waiting in backlog, todo or paused goes into current, at its head, as it takes the focus; one blocked or watching stays where it is, and one done is refused. Each task in focus says how long its session has been up; focusing one already in focus renews its session once it is past the length the configuration gives.",
	Example: `cairn task focus 3
cairn task focus`,
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: Focus{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a task", Means: "the task to focus, or nothing to name every task in focus"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, named := in.whole("number")
		if !named {
			return whatIsInFocus(on)
		}
		states, err := on.Engine.statesOf([]int{n})
		if err != nil {
			return Answer{}, err
		}
		t, err := on.Engine.putInFocus(n)
		if err != nil {
			return Answer{}, err
		}
		warnings, err := on.Engine.unmetEach(states)
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
		sessions, err := on.Engine.sessions()
		if err != nil {
			return Answer{}, err
		}
		return Answer{Task: &t, Progress: p, InFocus: &held, Sessions: sessions, Warnings: warnings}, nil
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
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: Focus{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a task", Means: "the task to unfocus, or nothing to unfocus them all"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, named := in.whole("number")
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

func placeFrom(in Input) (task.Place, int, error) {
	named := make([]task.Place, 0, 3)
	target := 0
	if in.yes("first") {
		named = append(named, task.First)
	}
	if in.Given["after"] != "" {
		named = append(named, task.Behind)
		target, _ = in.whole("after")
	}
	if in.yes("last") {
		named = append(named, task.Last)
	}
	if len(named) == 0 {
		return 0, 0, task.ErrNowhereToMove
	}
	if len(named) > 1 {
		return 0, 0, task.ErrMoreThanOnePlace
	}
	return named[0], target, nil
}

func whatIsInFocus(on workedOn) (Answer, error) {
	held, numbers, err := on.Engine.inFocus()
	if err != nil {
		return Answer{}, err
	}
	sessions, err := on.Engine.sessions()
	if err != nil {
		return Answer{}, err
	}
	return Answer{Held: held, InFocus: &numbers, Sessions: sessions}, nil
}
