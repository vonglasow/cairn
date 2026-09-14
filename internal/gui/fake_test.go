package gui

import (
	"errors"

	"cairn/internal/engine"
	"cairn/internal/reading"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// A stand-in for whatever holds the work. The window asks it questions and
// tells it to change things; a test answers, and writes down what it was
// asked. No database, so a hundred of these run in the time one real store
// takes to open.

type fake struct {
	tasks []task.Task
	focus int

	version int64
	// fail is a store that cannot be read at all; failOp is one that reads
	// perfectly well and refuses what it was asked to change. They are two
	// different faults, and a window that muddled them would say the wrong one.
	fail   error
	failOp error
	// refuseMove refuses the second half of a drop and nothing else, which is
	// the one thing a real store cannot be asked for on demand.
	refuseMove error

	// acts is the history a detail is given when it asks for one.
	acts []trace.Act

	// asked is every operation the window asked for, in order, with what it
	// gave.
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

// what says how many times the window asked for a given operation.
func (f *fake) what(thing, verb string) int {
	count := 0
	for _, a := range f.asked {
		if a.op.Thing == thing && a.op.Verb == verb {
			count++
		}
	}
	return count
}

// gave answers what the window handed to the last call of an operation.
func (f *fake) gave(thing, verb string) (engine.Input, bool) {
	for i := len(f.asked) - 1; i >= 0; i-- {
		if f.asked[i].op.Thing == thing && f.asked[i].op.Verb == verb {
			return f.asked[i].in, true
		}
	}
	return engine.Input{}, false
}

// holding builds a window already showing these cards, so a test starts where
// a person starts rather than at an empty screen.
func holding(f *fake, cards ...reading.Card) Window {
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Cards: cards, Focus: f.focus, Version: f.version}
	return w
}

// card is one, in a column, titled.
func card(number int, title, state string) reading.Card {
	return reading.Card{Number: number, Title: title, State: state}
}

// middleOf is the point a person would click to hit the n-th card of a column.
func middleOf(w Window, column, at int) Point {
	box := w.Layout().CardAt(column, at, w.scroll[column])
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
}

// inLane is the point a person would drop onto to name one of the three kinds.
func inLane(w Window, lane int) Point {
	box := w.Layout().Lane[lane]
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
}
