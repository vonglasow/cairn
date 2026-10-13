package gui

import (
	"fmt"
	"testing"

	"cairn/internal/engine"
)

func TestDroppingACardInAnotherColumnChangesItsState(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "backlog"))

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(middleOf(w, 1, 0))
	w.LetGo(middleOf(w, 1, 0))

	if f.what("task", "move") != 1 || f.what("task", "edit") != 0 {
		t.Fatalf("the store was asked %d moves and %d edits", f.what("task", "move"), f.what("task", "edit"))
	}
	in, _ := f.gave("task", "move")
	if in.Given["number"] != "1" || in.Given["state"] != "todo" {
		t.Errorf("asked to move %+v, want task 1 into todo", in.Given)
	}
}

func TestDroppingACardAboveAnotherCarriesBothTheColumnAndThePlace(t *testing.T) {
	f := newFake()
	w := holding(f,
		card(1, "First", "todo"),
		card(2, "Second", "todo"),
		card(3, "Third", "backlog"))

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(middleOf(w, 1, 1))
	w.LetGo(middleOf(w, 1, 1))

	if f.what("task", "edit") != 0 || f.what("task", "move") != 1 {
		t.Fatalf("asked %d edits and %d moves, want one move",
			f.what("task", "edit"), f.what("task", "move"))
	}
	in, _ := f.gave("task", "move")
	if in.Given["number"] != "3" || in.Given["after"] != "1" || in.Given["state"] != "todo" {
		t.Errorf("asked to move %+v, want task 3 into todo behind task 1", in.Given)
	}
}

func TestDroppingACardInItsOwnColumnOnlyMovesIt(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "First", "todo"), card(2, "Second", "todo"))

	w = w.PressAt(middleOf(w, 1, 1))
	w = w.MoveTo(middleOf(w, 1, 0))
	w.LetGo(middleOf(w, 1, 0))

	if f.what("task", "edit") != 0 {
		t.Error("a card dropped in its own column had its column changed")
	}
	if f.what("task", "move") != 1 {
		t.Fatalf("asked %d moves", f.what("task", "move"))
	}
	in, _ := f.gave("task", "move")
	if in.Given["first"] != "true" {
		t.Errorf("asked to move %+v, want it first", in.Given)
	}
}

func TestDroppingACardInALaneSaysWhichKindOfStopped(t *testing.T) {
	for lane, kind := range engine.Stalled {
		t.Run(kind, func(t *testing.T) {
			f := newFake()
			w := holding(f, card(1, "Client's answer", "todo"))

			w = w.PressAt(middleOf(w, 1, 0))
			w = w.MoveTo(inLane(w, lane))
			w.LetGo(inLane(w, lane))

			in, asked := f.gave("task", "move")
			if !asked {
				t.Fatal("nothing was asked of the store")
			}
			if in.Given["state"] != kind {
				t.Errorf("asked for %q, want %q", in.Given["state"], kind)
			}
		})
	}
}

func TestADropTheStoreRefusesSaysSoAndChangesNothing(t *testing.T) {
	f := newFake()
	f.failOp = errRefuse
	w := holding(f, card(1, "Write the spec", "backlog"))

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(middleOf(w, 1, 0))
	w = w.LetGo(middleOf(w, 1, 0))

	if w.Wrong() == nil {
		t.Fatal("a refused drop said nothing")
	}
	if w.Dragging() {
		t.Error("the card is still held after a refusal")
	}
}

func TestACardBeingDraggedFollowsThePointerAndShowsWhereItLands(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "First", State: "todo"},
		engine.Card{Number: 2, Title: "Second", State: "backlog"})

	if _, dragging := w.Ghost(); dragging {
		t.Fatal("a board nobody is dragging on shows a card in mid-air")
	}

	from := middleOf(w, 0, 0)
	over := middleOf(w, 1, 0)
	w = w.PressAt(from)
	w = w.MoveTo(over)

	ghost, dragging := w.Ghost()
	if !dragging {
		t.Fatal("a card being dragged is not shown")
	}
	if ghost.X > over.X || over.X > ghost.X+ghost.W || ghost.Y > over.Y || over.Y > ghost.Y+ghost.H {
		t.Errorf("the card at %+v does not hold the pointer at %+v", ghost, over)
	}

	landing, somewhere := w.Landing()
	if !somewhere {
		t.Fatal("the board does not say where the card would land")
	}
	if landing.X != w.Layout().Column[1].X {
		t.Errorf("it lands at %d, and todo is at %d", landing.X, w.Layout().Column[1].X)
	}
}

func TestOverNothingThereIsNowhereToLand(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "First", State: "todo"})
	w = w.PressAt(middleOf(w, 1, 0))
	w = w.MoveTo(Point{X: 0, Y: 0})
	if _, somewhere := w.Landing(); somewhere {
		t.Error("a card over the margin was told where it would land")
	}
}

func TestWhereACardWouldLandIsSnappedToTheGrid(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "First", State: "todo"},
		engine.Card{Number: 2, Title: "Second", State: "backlog"})

	l := w.Layout()
	for _, off := range []int{2, CardHeight / 2, CardHeight - 2} {
		over := Point{X: l.Column[1].X + 20, Y: l.Column[1].Y + CardHeight + off}
		held := w.PressAt(middleOf(w, 0, 0)).MoveTo(over)
		landing, somewhere := held.Landing()
		if !somewhere {
			t.Fatalf("%d into the second slot, the board says nowhere", off)
		}
		if want := l.Column[1].Y + CardHeight; landing.Y != want {
			t.Errorf("%d into the second slot, the line is at %d, want %d", off, landing.Y, want)
		}
		if landing.X != l.Column[1].X || landing.W != l.Column[1].W {
			t.Errorf("the line runs from %d for %d, and the column from %d for %d",
				landing.X, landing.W, l.Column[1].X, l.Column[1].W)
		}
	}
}

func TestDroppedUnder_asksNothingForAPartAlreadyThere(t *testing.T) {
	w := openOnAParent(t, 3)
	box := BesideBox(Over(w.width, w.height))
	at := Point{X: box.X + 5, Y: box.Y + SubHeight + 2}
	past := Point{X: at.X + Grip + 2, Y: at.Y}
	w = w.PressAt(at).MoveTo(past).LetGo(at)
	for _, one := range w.work.(*fake).asked {
		if one.op.Verb == engine.TaskMove.Verb && one.op.Thing == engine.TaskMove.Thing {
			t.Errorf("a part let go where it already was asked %v", one.in.Given)
		}
	}
}

func openOnAParent(t *testing.T, parts int) Window {
	t.Helper()
	f := newFake()
	cards := []engine.Card{{Number: 1, Title: "Ship it", State: "todo"}}
	for i := range parts {
		cards = append(cards, engine.Card{
			Number: i + 2, Title: fmt.Sprintf("Part %d", i+1), State: "todo", Parent: 1,
		})
	}
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Cards: cards}
	return w.clicked(middleOf(w, 1, 0)).Key("enter")
}

func TestLanding_saysNothingOutsideTheBox(t *testing.T) {
	w := openOnAParent(t, 3)
	box := BesideBox(Over(w.width, w.height))
	from := Point{X: box.X + 5, Y: box.Y + 2}
	out := Point{X: box.X + 5, Y: box.Y + box.H + CardHeight}
	w = w.PressAt(from).MoveTo(out)
	if landing, somewhere := w.Landing(); somewhere {
		t.Errorf("it says a tile would land at %+v, outside its box", landing)
	}
}

func TestDraggingAnObjectiveMovesNoTask(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{
		Cards: []engine.Card{card(1, "Write the spec", "todo")},
		Goals: []engine.Goal{{Number: 1, Title: "Ship cairn", State: "active"}},
	}
	w = w.Key("tab")
	if w.Tab() != Objectives {
		t.Fatalf("the window is on %v", w.Tab())
	}
	g := GoalAt(w.Layout(), 0, 0)
	at := Point{X: g.X + g.W/2, Y: g.Y + g.H/2}
	w = w.PressAt(at).MoveTo(Point{X: at.X + 400, Y: at.Y}).LetGo(Point{X: at.X + 400, Y: at.Y})

	for _, one := range f.asked {
		if one.op.Thing == "task" && (one.op.Verb == "edit" || one.op.Verb == "move") {
			t.Errorf("dragging an objective asked %s %s: %v", one.op.Thing, one.op.Verb, one.in.Given)
		}
	}
}
