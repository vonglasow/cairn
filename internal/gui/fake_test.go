package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/app"
	"cairn/internal/engine/workspace"
	"cairn/internal/gui/geom"
	"errors"
	"time"

	"gioui.org/io/pointer"

	"cairn/internal/engine"
	"cairn/internal/engine/act"
	"cairn/internal/engine/agent"
	"cairn/internal/engine/task"
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

	acts []act.Act
	warn []string

	asked []asked
	reads int

	shown  *workspace.Shown
	listed *workspace.Workspaces
	config *workspace.Config

	agents    *agent.Agents
	terminals []string
	refuseGo  error
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
	if op.Thing == "agent" && op.Verb == "go" && f.refuseGo != nil {
		return engine.Answer{}, f.refuseGo
	}
	if op.Level != engine.Read && len(f.warn) > 0 {
		return engine.Answer{Warnings: f.warn}, nil
	}
	switch op.Thing + " " + op.Verb {
	case " upgrade":
		return engine.Answer{Upgraded: &app.Upgraded{From: "v0.12.0", To: "v0.12.0", Binary: "/usr/local/bin/cairn"}}, nil
	case "task list":
		f.reads++
		return engine.Answer{Tasks: f.tasks, InFocus: &f.focus}, nil
	case "objective list":
		return engine.Answer{}, nil
	case "trace show":
		return engine.Answer{Acts: f.acts}, nil
	case "workspace show":
		return engine.Answer{WorkspaceShown: f.shown}, nil
	case "workspace list":
		return engine.Answer{Workspaces: f.listed}, nil
	case "agent list":
		return engine.Answer{Agents: f.agents}, nil
	case "agent terminals":
		return engine.Answer{Terminals: &agent.Terminals{Terminals: f.terminals}}, nil
	case "config show":
		if f.config != nil {
			return engine.Answer{Config: f.config}, nil
		}
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

func holding(f *fake, cards ...board.Card) Window {
	w := New(f, 1200, 800)
	w.Snap = test.Searchable(board.Snapshot{Cards: cards, Focus: f.focus, Version: f.version})
	return w
}

func card(number int, title, state string) board.Card {
	return board.Card{Number: number, Title: title, State: state}
}

func middleOf(w Window, column, at int) geom.Point {
	box := w.Layout().CardAt(column, at, w.scroll[column])
	return geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
}

func (w Window) clicked(at geom.Point) Window {
	w, _ = tapped(w, Gesture{}, at)
	return w
}

var moment = time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)

func tapped(w Window, g Gesture, at geom.Point) (Window, Gesture) {
	w, g = Pointing(w, g, pointer.Press, at, 0, moment)
	return Pointing(w, g, pointer.Release, at, 0, moment)
}

func (w Window) tripleClicked(at geom.Point) Window {
	w, g := tapped(w, Gesture{}, at)
	for i := 1; i <= 2; i++ {
		w, g = Pointing(w, g, pointer.Press, at, 0, moment.Add(time.Duration(i)*time.Millisecond))
		w, g = Pointing(w, g, pointer.Release, at, 0, moment.Add(time.Duration(i)*time.Millisecond))
	}
	return w
}

func (w Window) doubleClicked(at geom.Point) Window {
	w, g := tapped(w, Gesture{}, at)
	w, g = Pointing(w, g, pointer.Press, at, 0, moment.Add(time.Millisecond))
	w, _ = Pointing(w, g, pointer.Release, at, 0, moment.Add(time.Millisecond))
	return w
}

func formName(box geom.Box, fields []board.Field, at int) geom.Box {
	name, _, _ := FormPlace(box, fields, at)
	return name
}

func formValue(box geom.Box, fields []board.Field, at int) geom.Box {
	_, value, _ := FormPlace(box, fields, at)
	return value
}

func formRows(box geom.Box, fields []board.Field, at int) int {
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

type heldDrafts struct{ values map[string]string }

func (h *heldDrafts) Keep(values map[string]string) error { h.values = values; return nil }
func (h *heldDrafts) Held() (map[string]string, error)    { return h.values, nil }
func (h *heldDrafts) Drop() error                         { h.values = nil; return nil }
