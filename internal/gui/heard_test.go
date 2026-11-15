package gui

import (
	"cairn/internal/board"
	"cairn/internal/debug"
	"cairn/internal/gui/geom"
	"image"
	"os"
	"strings"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func frames(t *testing.T, w Window, events ...event.Event) (Window, time.Time, geom.Point) {
	t.Helper()
	return framesAt(t, w, time.Time{}, geom.Point{}, time.Now(), events...)
}

func framesTraced(t *testing.T, w Window, traced *debug.Log, events ...event.Event) Window {
	t.Helper()
	w, _, _ = framesInto(t, w, time.Time{}, geom.Point{}, time.Now(), traced, events...)
	return w
}

func framesAt(
	t *testing.T, w Window,
	pressed time.Time, pressedAt geom.Point, now time.Time,
	events ...event.Event,
) (Window, time.Time, geom.Point) {
	t.Helper()
	return framesInto(t, w, pressed, pressedAt, now, nil, events...)
}

func framesInto(
	t *testing.T, w Window,
	pressed time.Time, pressedAt geom.Point, now time.Time, traced *debug.Log,
	events ...event.Event,
) (Window, time.Time, geom.Point) {
	t.Helper()
	var router input.Router
	var ops op.Ops

	gtx := func() layout.Context { return frameOf(&ops, &router, w, now) }

	first := gtx()
	w, pressed, pressedAt = heard(first, w, pressed, pressedAt, traced)
	router.Frame(first.Ops)
	router.Queue(events...)

	second := gtx()
	w, pressed, pressedAt = heard(second, w, pressed, pressedAt, traced)
	router.Frame(second.Ops)
	return w, pressed, pressedAt
}

func frameOf(ops *op.Ops, router *input.Router, w Window, now time.Time) layout.Context {
	ops.Reset()
	return layout.Context{
		Ops:         ops,
		Constraints: layout.Exact(image.Pt(w.width, w.height)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         now,
		Source:      router.Source(),
	}
}

func at(w Window, column, index int) f32.Point {
	box := w.Layout().CardAt(column, index, 0)
	return f32.Pt(float32(box.X+box.W/2), float32(box.Y+box.H/2))
}

func TestAFrameOfEventsTypesIntoTheForm(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title")

	w, _, _ = frames(t, w, key.EditEvent{Range: key.Range{Start: 14, End: 14}, Text: " properly"})
	if got := valueOf(w.Form, "title"); got != "Write the spec properly" {
		t.Fatalf("the title reads %q", got)
	}

	w, _, _ = frames(t, w, key.Event{Name: key.NameTab, State: key.Press})
	if got := w.Form.Fields[w.Form.On()].Name; got != "description" {
		t.Fatalf("tab through the loop landed on %q", got)
	}

	w, _, _ = frames(t, w, key.Event{Name: key.NameTab, Modifiers: key.ModShift, State: key.Press})
	if got := w.Form.Fields[w.Form.On()].Name; got != "title" {
		t.Fatalf("shift+tab through the loop landed on %q", got)
	}

	w, _, _ = frames(t, w, key.Event{Name: key.NameDeleteBackward, State: key.Press})
	if got := valueOf(w.Form, "title"); got != "Write the spec properl" {
		t.Fatalf("backspace through the loop left %q", got)
	}

	w, _, _ = frames(t, w, key.Event{
		Name: key.NameReturn, Modifiers: key.ModCommand, State: key.Press})
	if w.Screen != board.OnTheBoard {
		t.Fatalf("⌘⏎ through the loop left the window on %v", w.Screen)
	}
	if f.what("task", "edit") != 1 {
		t.Errorf("it asked for %d edits", f.what("task", "edit"))
	}
}

func TestTheLoopWritesDownEachEventInDebugMode(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	traced, err := debug.Open(t.TempDir(), "gui", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	framesTraced(t, w, traced,
		pointer.Event{Kind: pointer.Move, Position: at(w, 1, 0)},
		pointer.Event{Kind: pointer.Press, Position: at(w, 1, 0), Buttons: pointer.ButtonPrimary},
		key.Event{Name: "?", State: key.Press},
	)
	_ = traced.Close()
	read, err := os.ReadFile(traced.Path())
	if err != nil {
		t.Fatal(err)
	}
	said := string(read)
	if strings.Contains(said, "pointer Move") {
		t.Errorf("a hover was written down: %q", said)
	}
	pressed := strings.Index(said, "pointer Press ")
	typed := strings.Index(said, `key Press name="?"`)
	if pressed < 0 || !strings.Contains(said, "held=1") || typed < pressed {
		t.Errorf("the loop wrote %q, want the press, then the key", said)
	}
}
