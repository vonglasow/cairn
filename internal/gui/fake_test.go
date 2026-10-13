package gui

import (
	"errors"
	"time"

	"gioui.org/io/pointer"

	"cairn/internal/engine"
	"cairn/internal/task"
	"cairn/internal/trace"
	"cairn/test"
)

type fake struct {
	tasks []task.Task
	focus []int

	version      int64
	fail         error
	failOp       error
	refuseMove   error
	refuseDelete error

	acts []trace.Act

	asked []asked
	reads int
}

type asked struct {
	op engine.Operation
	in engine.Input
}

var (
	errNope   = errors.New("the store is not answering")
	errRefuse = errors.New("that cannot be done")
)

func newFake() *fake { return &fake{} }

func (f *fake) Ask(op engine.Operation, in engine.Input) (engine.Answer, error) {
	f.asked = append(f.asked, asked{op: op, in: in})
	if f.fail != nil {
		return engine.Answer{}, f.fail
	}
	if op.Level != engine.Read && f.failOp != nil {
		return engine.Answer{}, f.failOp
	}
	if op.Thing == "task" && op.Verb == "move" && f.refuseMove != nil {
		return engine.Answer{}, f.refuseMove
	}
	if op.Thing == "task" && op.Verb == "delete" && f.refuseDelete != nil {
		return engine.Answer{}, f.refuseDelete
	}
	switch op.Thing + " " + op.Verb {
	case "task list":
		f.reads++
		return engine.Answer{Tasks: f.tasks, InFocus: &f.focus}, nil
	case "objective list":
		return engine.Answer{}, nil
	case "trace show":
		return engine.Answer{Acts: f.acts}, nil
	}
	return engine.Answer{}, nil
}

func (f *fake) Version() (int64, error) { return f.version, f.fail }

func (f *fake) what(thing, verb string) int {
	count := 0
	for _, a := range f.asked {
		if a.op.Thing == thing && a.op.Verb == verb {
			count++
		}
	}
	return count
}

func (f *fake) gave(thing, verb string) (engine.Input, bool) {
	for i := len(f.asked) - 1; i >= 0; i-- {
		if f.asked[i].op.Thing == thing && f.asked[i].op.Verb == verb {
			return f.asked[i].in, true
		}
	}
	return engine.Input{}, false
}

func holding(f *fake, cards ...engine.Card) Window {
	w := New(f, 1200, 800)
	w.snap = test.Searchable(engine.Snapshot{Cards: cards, Focus: f.focus, Version: f.version})
	return w
}

func card(number int, title, state string) engine.Card {
	return engine.Card{Number: number, Title: title, State: state}
}

func middleOf(w Window, column, at int) Point {
	box := w.Layout().CardAt(column, at, w.scroll[column])
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
}

func (w Window) clicked(at Point) Window {
	w, _ = tapped(w, Gesture{}, at)
	return w
}

var moment = time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)

func tapped(w Window, g Gesture, at Point) (Window, Gesture) {
	w, g = Pointing(w, g, pointer.Press, at, 0, moment)
	return Pointing(w, g, pointer.Release, at, 0, moment)
}

func (w Window) doubleDragged(from, to Point) Window {
	w, g := tapped(w, Gesture{}, from)
	w, g = Pointing(w, g, pointer.Press, from, 0, moment.Add(time.Millisecond))
	w, g = Pointing(w, g, pointer.Drag, to, 0, moment.Add(2*time.Millisecond))
	w, _ = Pointing(w, g, pointer.Release, to, 0, moment.Add(2*time.Millisecond))
	return w
}

func (w Window) tripleClicked(at Point) Window {
	w, g := tapped(w, Gesture{}, at)
	for i := 1; i <= 2; i++ {
		w, g = Pointing(w, g, pointer.Press, at, 0, moment.Add(time.Duration(i)*time.Millisecond))
		w, g = Pointing(w, g, pointer.Release, at, 0, moment.Add(time.Duration(i)*time.Millisecond))
	}
	return w
}

func (w Window) doubleClicked(at Point) Window {
	w, g := tapped(w, Gesture{}, at)
	w, g = Pointing(w, g, pointer.Press, at, 0, moment.Add(time.Millisecond))
	w, _ = Pointing(w, g, pointer.Release, at, 0, moment.Add(time.Millisecond))
	return w
}

func formName(box Box, fields []engine.Field, at int) Box {
	name, _, _ := FormPlace(box, fields, at)
	return name
}

func formValue(box Box, fields []engine.Field, at int) Box {
	_, value, _ := FormPlace(box, fields, at)
	return value
}

func formRows(box Box, fields []engine.Field, at int) int {
	_, _, rows := FormPlace(box, fields, at)
	return rows
}

func (f *fake) wrote() []asked {
	var changed []asked
	for _, a := range f.asked {
		if a.op.Level != engine.Read {
			changed = append(changed, a)
		}
	}
	return changed
}
