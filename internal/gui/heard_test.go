package gui

import (
	"bytes"
	"cairn/internal/engine"
	"image"
	"io"
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

func frames(t *testing.T, w Window, events ...event.Event) (Window, time.Time, Point) {
	t.Helper()
	return framesAt(t, w, time.Time{}, Point{}, time.Now(), events...)
}

func framesTraced(t *testing.T, w Window, traced io.Writer, events ...event.Event) Window {
	t.Helper()
	w, _, _ = framesInto(t, w, time.Time{}, Point{}, time.Now(), traced, events...)
	return w
}

func framesAt(
	t *testing.T, w Window,
	pressed time.Time, pressedAt Point, now time.Time,
	events ...event.Event,
) (Window, time.Time, Point) {
	t.Helper()
	return framesInto(t, w, pressed, pressedAt, now, nil, events...)
}

func framesInto(
	t *testing.T, w Window,
	pressed time.Time, pressedAt Point, now time.Time, traced io.Writer,
	events ...event.Event,
) (Window, time.Time, Point) {
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

func composing(w Window, made, dead, letter string) Window {
	var router input.Router
	var ops op.Ops
	frame := func() {
		gtx := frameOf(&ops, &router, w, time.Now())
		w, _, _ = heard(gtx, w, time.Time{}, Point{}, nil)
		router.Frame(gtx.Ops)
	}
	frame()
	before := router.EditorState()
	at := before.Selection.Range
	router.Queue(key.EditEvent{Range: at, Text: dead})
	frame()
	after := router.EditorState()
	said := []rune(before.Snippet.Text)
	from, to := at.Start-before.Snippet.Start, at.End-before.Snippet.Start
	if from < 0 || to > len(said) || from > to {
		from, to = len(said), len(said)
	}
	told := string(said[:from]) + dead + string(said[to:])
	end := at.Start + len([]rune(dead))
	if after.Snippet.Text == told && after.Selection.Range == (key.Range{Start: end, End: end}) {
		router.Queue(key.EditEvent{Range: key.Range{Start: at.Start, End: end}, Text: made})
	} else {
		router.Queue(key.EditEvent{Range: after.Selection.Range, Text: letter})
	}
	frame()
	return w
}

func at(w Window, column, index int) f32.Point {
	box := w.Layout().CardAt(column, index, 0)
	return f32.Pt(float32(box.X+box.W/2), float32(box.Y+box.H/2))
}

func TestAFrameOfEventsChoosesACard(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	where := at(w, 1, 0)

	got, _, _ := frames(t, w,
		pointer.Event{Kind: pointer.Press, Position: where, Buttons: pointer.ButtonPrimary},
		pointer.Event{Kind: pointer.Release, Position: where, Buttons: pointer.ButtonPrimary},
	)
	if got.Chosen() != 1 {
		t.Fatalf("a press and a release chose %d", got.Chosen())
	}
}

func TestAFrameOfEventsDragsACardToAnotherColumn(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	from, onto := at(w, 0, 0), at(w, 1, 0)

	frames(t, w,
		pointer.Event{Kind: pointer.Press, Position: from, Buttons: pointer.ButtonPrimary},
		pointer.Event{Kind: pointer.Move, Position: onto, Buttons: pointer.ButtonPrimary},
		pointer.Event{Kind: pointer.Release, Position: onto, Buttons: pointer.ButtonPrimary},
	)
	if f.what("task", "move") != 1 {
		t.Fatalf("a drag across columns asked for %d moves", f.what("task", "move"))
	}
}

func TestAFrameOfEventsOpensTheDetailOnTwoPresses(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	where := at(w, 1, 0)
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)

	got, _, _ := framesAt(t, w, now, Point{X: int(where.X), Y: int(where.Y)}, now.Add(time.Millisecond),
		pointer.Event{Kind: pointer.Press, Position: where, Buttons: pointer.ButtonPrimary},
	)
	if got.Screen() != OnForm {
		t.Fatalf("a second press left the window on %v", got.Screen())
	}
}

func TestAFrameOfEventsRemembersAPressForNextTime(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	where := at(w, 1, 0)
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)

	_, pressed, pressedAt := framesAt(t, w, time.Time{}, Point{}, now,
		pointer.Event{Kind: pointer.Press, Position: where, Buttons: pointer.ButtonPrimary},
	)
	if pressed != now {
		t.Errorf("the press was remembered at %v, want %v", pressed, now)
	}
	if pressedAt.X != int(where.X) {
		t.Errorf("the press was remembered at %v", pressedAt)
	}
}

func TestAFrameOfEventsScrollsAColumn(t *testing.T) {
	f := newFake()
	var many []engine.Card
	for i := 1; i <= 12; i++ {
		many = append(many, engine.Card{Number: i, Title: "Task", State: "todo"})
	}
	w := holding(f, many...)
	where := at(w, 1, 0)

	got, _, _ := frames(t, w, pointer.Event{
		Kind:     pointer.Scroll,
		Position: where,
		Scroll:   f32.Pt(0, float32(CardHeight)),
	})
	if got.scroll[1] == 0 {
		t.Fatal("a scroll over a column moved nothing")
	}
}

func TestAFrameOfEventsTakesAKey(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})

	got, _, _ := frames(t, w,
		key.Event{Name: "N", Modifiers: key.ModCommand, State: key.Press},
	)
	if got.Screen() != OnForm {
		t.Fatalf("⌘N through the loop left the window on %v", got.Screen())
	}
}

func TestAFrameOfEventsIgnoresAKeyGoingBackUp(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})

	got, _, _ := frames(t, w,
		key.Event{Name: "N", Modifiers: key.ModCommand, State: key.Release},
	)
	if got.Screen() != OnTheBoard {
		t.Fatalf("a key going back up left the window on %v", got.Screen())
	}
}

func TestAFrameOfNothingChangesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})

	got, pressed, _ := frames(t, w)
	if got.Chosen() != 0 || !pressed.IsZero() || len(f.asked) != 0 {
		t.Error("a frame with no events did something")
	}
}

func TestAFrameOfEventsChoosesThenEdits(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	card := at(w, 1, 0)

	w, _, _ = frames(t, w,
		pointer.Event{Kind: pointer.Press, Position: card, Buttons: pointer.ButtonPrimary},
		pointer.Event{Kind: pointer.Release, Position: card, Buttons: pointer.ButtonPrimary},
	)
	if w.Chosen() != 1 {
		t.Fatalf("clicking the card chose %d", w.Chosen())
	}

	offer := -1
	for i, name := range w.Offers() {
		if name == "Edit" {
			offer = i
		}
	}
	if offer < 0 {
		t.Fatalf("the bar offers %v", w.Offers())
	}
	box := OfferAt(w.Layout(), offer)
	on := f32.Pt(float32(box.X+box.W/2), float32(box.Y+box.H/2))

	w, _, _ = frames(t, w,
		pointer.Event{Kind: pointer.Press, Position: on, Buttons: pointer.ButtonPrimary},
		pointer.Event{Kind: pointer.Release, Position: on, Buttons: pointer.ButtonPrimary},
	)
	if w.Screen() != OnForm {
		t.Fatalf("pressing Edit in the bar left the window on %v", w.Screen())
	}
}

func TestAFrameOfEventsTypesIntoTheForm(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title")

	w, _, _ = frames(t, w, key.EditEvent{Range: key.Range{Start: 14, End: 14}, Text: " properly"})
	if got := w.Form().ValueOf("title"); got != "Write the spec properly" {
		t.Fatalf("the title reads %q", got)
	}

	w, _, _ = frames(t, w, key.Event{Name: key.NameTab, State: key.Press})
	if got := w.Form().Fields()[w.Form().On()].Name; got != "description" {
		t.Fatalf("tab through the loop landed on %q", got)
	}

	w, _, _ = frames(t, w, key.Event{Name: key.NameTab, Modifiers: key.ModShift, State: key.Press})
	if got := w.Form().Fields()[w.Form().On()].Name; got != "title" {
		t.Fatalf("shift+tab through the loop landed on %q", got)
	}

	w, _, _ = frames(t, w, key.Event{Name: key.NameDeleteBackward, State: key.Press})
	if got := w.Form().ValueOf("title"); got != "Write the spec properl" {
		t.Fatalf("backspace through the loop left %q", got)
	}

	w, _, _ = frames(t, w, key.Event{
		Name: key.NameReturn, Modifiers: key.ModCommand, State: key.Press})
	if w.Screen() != OnTheBoard {
		t.Fatalf("⌘⏎ through the loop left the window on %v", w.Screen())
	}
	if f.what("task", "edit") != 1 {
		t.Errorf("it asked for %d edits", f.what("task", "edit"))
	}
}

func TestAFrameOfEventsSavesByPressingSaveOnTheForm(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title").Types(" properly")

	button := PanelOfferAt(w.formPanel(), 0)
	on := f32.Pt(float32(button.X+button.W/2), float32(button.Y+button.H/2))

	w, _, _ = frames(t, w,
		pointer.Event{Kind: pointer.Press, Position: on, Buttons: pointer.ButtonPrimary},
		pointer.Event{Kind: pointer.Release, Position: on, Buttons: pointer.ButtonPrimary},
	)
	if w.Screen() != OnTheBoard {
		t.Fatalf("pressing Save on the form left the window on %v", w.Screen())
	}
	in, asked := f.gave("task", "edit")
	if !asked {
		t.Fatal("pressing Save asked the store for nothing")
	}
	if in.Given["title"] != "Write the spec properly" {
		t.Errorf("it saved the title as %q", in.Given["title"])
	}
}

func TestTheLoopWritesDownEachEventWhenAsked(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	var traced bytes.Buffer
	framesTraced(t, w, &traced,
		pointer.Event{Kind: pointer.Press, Position: at(w, 1, 0), Buttons: pointer.ButtonPrimary},
		key.Event{Name: "?", State: key.Press},
	)
	said := traced.String()
	pressed := strings.Index(said, "pointer Press ")
	typed := strings.Index(said, `key Press name="?"`)
	if pressed < 0 || !strings.Contains(said, "held=1") || typed < pressed {
		t.Errorf("the loop wrote %q, want the press, then the key", said)
	}
}

func TestTheWindowTellsTheSystemTheTextBeingTyped(t *testing.T) {
	w, _, _ := frames(t, holding(newFake()).Key("cmd+n"))
	w = w.Types("Fenê")
	var router input.Router
	var ops op.Ops
	gtx := frameOf(&ops, &router, w, time.Now())
	_, _, _ = heard(gtx, w, time.Time{}, Point{}, nil)
	router.Frame(gtx.Ops)
	told := router.EditorState()
	if told.Snippet.Text != "Fenê" || told.Snippet.Start != 0 {
		t.Errorf("the system was told the text is %+v", told.Snippet)
	}
	if told.Selection.Range != (key.Range{Start: 4, End: 4}) {
		t.Errorf("the system was told the caret is at %+v", told.Selection.Range)
	}
	w = w.Key("shift+left").Key("shift+left")
	gtx = frameOf(&ops, &router, w, time.Now())
	_, _, _ = heard(gtx, w, time.Time{}, Point{}, nil)
	router.Frame(gtx.Ops)
	if held := router.EditorState().Selection.Range; held != (key.Range{Start: 2, End: 4}) {
		t.Errorf("the system was told the run held is %+v", held)
	}
}

func TestAnEditReplacesTheRunTheSystemNamed(t *testing.T) {
	w, _, _ := frames(t, holding(newFake()).Key("cmd+n"), key.EditEvent{Text: "Fen^"})
	w, _, _ = frames(t, w, key.EditEvent{Range: key.Range{Start: 3, End: 4}, Text: "ê"})
	if got := w.Form().Writing().Reads(); got != "Fenê" {
		t.Errorf("the title reads %q", got)
	}
}
