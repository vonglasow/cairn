package gui

import (
	"cairn/internal/engine"
	"testing"
)

// The band and its three lanes, and the edges of every rectangle. A point one
// pixel outside a box is outside it, and a board that answered otherwise would
// pick up a card nobody pointed at.

func TestABoxHoldsItsOwnEdgesAndNotTheNextOnes(t *testing.T) {
	box := Box{X: 10, Y: 20, W: 30, H: 40}
	for _, c := range []struct {
		what string
		p    Point
		in   bool
	}{
		{"its own corner", Point{X: 10, Y: 20}, true},
		{"one left of it", Point{X: 9, Y: 20}, false},
		{"one above it", Point{X: 10, Y: 19}, false},
		{"the last point inside", Point{X: 39, Y: 59}, true},
		{"one past the right", Point{X: 40, Y: 59}, false},
		{"one past the bottom", Point{X: 39, Y: 60}, false},
		{"the middle", Point{X: 25, Y: 40}, true},
	} {
		if got := box.In(c.p); got != c.in {
			t.Errorf("%s: In(%v) = %v", c.what, c.p, got)
		}
	}
}

func TestACardInALaneSitsUnderItsLane(t *testing.T) {
	l := LayoutFor(1200, 800)
	for lane := range l.Lane {
		for at := range 2 {
			box := l.LaneCardAt(lane, at, 0)
			if box.H != CardHeight {
				t.Errorf("lane %d card %d is %d tall", lane, at, box.H)
			}
			if box.X != l.Lane[lane].X || box.W != l.Lane[lane].W {
				t.Errorf("lane %d card %d is not the width of its lane", lane, at)
			}
			if want := l.Lane[lane].Y + at*CardHeight; box.Y != want {
				t.Errorf("lane %d card %d starts at %d, want %d", lane, at, box.Y, want)
			}
		}
		if got := l.LaneCardAt(lane, 0, CardHeight).Y; got != l.Lane[lane].Y-CardHeight {
			t.Errorf("scrolling a lane by one card put the first at %d", got)
		}
	}
}

func TestWhichLaneAPointFallsIn(t *testing.T) {
	l := LayoutFor(1200, 800)
	for want := range l.Lane {
		box := l.Lane[want]
		got, found := l.LaneUnder(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		if !found || got != want {
			t.Errorf("the middle of lane %d answered (%d, %v)", want, got, found)
		}
	}
	above := Point{X: l.Lane[0].X, Y: l.Lane[0].Y - 1}
	if _, found := l.LaneUnder(above); found {
		t.Error("a point above the lanes answered one")
	}
}

// A card in the band is clicked the same way as one in a column, and the
// window has to look in the lanes to find it.
func TestClickingACardInALaneChoosesIt(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "Client's answer", State: "blocked"},
		engine.Card{Number: 2, Title: "Handed over", State: "watching"},
		engine.Card{Number: 3, Title: "Not now", State: "paused"})

	for lane, number := range map[int]int{0: 1, 1: 2, 2: 3} {
		box := w.Layout().LaneCardAt(lane, 0, 0)
		got := w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		if got.Chosen() != number {
			t.Errorf("clicking lane %d chose %d, want %d", lane, got.Chosen(), number)
		}
	}
}

// A lane with nothing in it answers nothing, rather than the card that would
// have been there.
func TestClickingAnEmptyLaneChoosesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Client's answer", State: "blocked"})

	box := w.Layout().LaneCardAt(1, 0, 0)
	if got := w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2}); got.Chosen() != 0 {
		t.Errorf("an empty lane chose %d", got.Chosen())
	}
	below := w.Layout().LaneCardAt(0, 4, 0)
	if got := w.clicked(Point{X: below.X + below.W/2, Y: below.Y + below.H/2}); got.Chosen() != 0 {
		t.Errorf("past the last card of a lane chose %d", got.Chosen())
	}
}

// Dragging out of the band puts the task back in a column, which is the way
// work starts moving again.
func TestDraggingOutOfALaneReturnsATaskToAColumn(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Client's answer", State: "blocked"})

	from := w.Layout().LaneCardAt(0, 0, 0)
	onto := w.Layout().CardAt(2, 0, 0)
	w = w.PressAt(Point{X: from.X + from.W/2, Y: from.Y + from.H/2})
	w = w.MoveTo(Point{X: onto.X + onto.W/2, Y: onto.Y + onto.H/2})
	w.LetGo(Point{X: onto.X + onto.W/2, Y: onto.Y + onto.H/2})

	in, asked := f.gave("task", "edit")
	if !asked {
		t.Fatal("nothing was asked of the store")
	}
	if in.Given["state"] != "current" {
		t.Errorf("asked for %q, want current", in.Given["state"])
	}
}

// A drop in no column and no lane — the margin, the bar, the gap between two
// columns — changes nothing rather than guessing.
func TestADropOnNothingChangesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})

	from := middleOf(w, 0, 0)
	w = w.PressAt(from)
	w = w.MoveTo(Point{X: 0, Y: 0})
	w = w.LetGo(Point{X: 0, Y: 0})

	if len(f.asked) != 0 {
		t.Errorf("a drop on nothing asked the store %d things", len(f.asked))
	}
	if w.Dragging() {
		t.Error("the card is still in hand")
	}
}

// Letting go having taken hold of nothing is nothing at all.
func TestLettingGoOfNothingDoesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})

	w = w.PressAt(Point{X: 0, Y: 0})
	w = w.MoveTo(middleOf(w, 1, 0))
	w = w.LetGo(middleOf(w, 1, 0))

	if len(f.asked) != 0 || w.Chosen() != 0 {
		t.Error("taking hold of nothing and letting go somewhere did something")
	}
}

// Dropping past the last card of a column puts the task behind all of them,
// which is what the bottom of a column means.
func TestDroppingBelowEverythingPutsItLast(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "First", State: "todo"},
		engine.Card{Number: 2, Title: "Second", State: "backlog"})

	onto := w.Layout().CardAt(1, 3, 0)
	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(Point{X: onto.X + onto.W/2, Y: onto.Y + onto.H/2})
	w.LetGo(Point{X: onto.X + onto.W/2, Y: onto.Y + onto.H/2})

	in, asked := f.gave("task", "move")
	if !asked {
		t.Fatal("nothing was asked to move")
	}
	if in.Given["last"] != "true" {
		t.Errorf("asked to move %+v, want it last", in.Given)
	}
}

// Scrolling never goes above the first card nor past the last, whatever it is
// asked for.
func TestScrollingSettlesBetweenItsTwoEnds(t *testing.T) {
	f := newFake()
	var many []engine.Card
	for i := 1; i <= 20; i++ {
		many = append(many, engine.Card{Number: i, Title: "Task", State: "todo"})
	}
	w := holding(f, many...)

	l := w.Layout()
	most := len(many)*CardHeight - l.Column[1].H
	// A point in the column's body rather than on a card: once a column is
	// scrolled to its end, its first card is off the top of it.
	over := Point{X: l.Column[1].X + l.Column[1].W/2, Y: l.Column[1].Y + l.Column[1].H/2}

	w = w.Scroll(over, CardHeight*100)
	if w.scroll[1] != most {
		t.Errorf("scrolled to %d, and the end is %d", w.scroll[1], most)
	}
	w = w.Scroll(over, -CardHeight*100)
	if w.scroll[1] != 0 {
		t.Errorf("scrolled back to %d", w.scroll[1])
	}
}

// A lane of the band is scrolled the same way, between the same two ends, and
// leaves the columns above it where they were.
func TestALaneSettlesBetweenItsTwoEnds(t *testing.T) {
	f := newFake()
	var many []engine.Card
	for i := 1; i <= 7; i++ {
		many = append(many, engine.Card{Number: i, Title: "Task", State: "watching"})
	}
	w := holding(f, many...)

	l := w.Layout()
	most := len(many)*CardHeight - l.Lane[1].H
	over := Point{X: l.Lane[1].X + l.Lane[1].W/2, Y: l.Lane[1].Y + l.Lane[1].H/2}

	w = w.Scroll(over, CardHeight*100)
	if w.lanes[1] != most {
		t.Errorf("the lane scrolled to %d, and its end is %d", w.lanes[1], most)
	}
	if w.scroll != [Columns]int{} {
		t.Errorf("scrolling a lane moved the columns: %v", w.scroll)
	}
	w = w.Scroll(over, -CardHeight*100)
	if w.lanes[1] != 0 {
		t.Errorf("the lane scrolled back to %d", w.lanes[1])
	}
}

// A point in no column at all scrolls nothing.
func TestScrollingOverNothingScrollsNothing(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	if got := w.Scroll(Point{X: 0, Y: 0}, CardHeight); got.scroll != w.scroll {
		t.Error("a point in no column scrolled one")
	}
}

// Just past the last card of a column is the column and no card: a board that
// answered the last card there would choose one nobody pointed at.
func TestJustPastTheLastCardChoosesNothing(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "First", State: "todo"},
		engine.Card{Number: 2, Title: "Second", State: "todo"})

	last := w.Layout().CardAt(1, 1, 0)
	if got := w.clicked(Point{X: last.X + 10, Y: last.Y + 10}); got.Chosen() != 2 {
		t.Fatalf("the last card chose %d", got.Chosen())
	}
	past := w.Layout().CardAt(1, 2, 0)
	if got := w.clicked(Point{X: past.X + 10, Y: past.Y + 10}); got.Chosen() != 0 {
		t.Errorf("just past the last card chose %d", got.Chosen())
	}
}

// Exactly one slot past the last card is the boundary the "behind this one"
// reckoning stops at. One short of it names a task; on it, there is no task to
// name and the drop means last.
func TestTheSlotJustPastTheLastCardMeansLast(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "First", State: "todo"},
		engine.Card{Number: 2, Title: "Second", State: "todo"},
		engine.Card{Number: 3, Title: "Newcomer", State: "backlog"})

	drop := func(slot int) map[string]string {
		f.asked = nil
		onto := w.Layout().CardAt(1, slot, 0)
		at := Point{X: onto.X + onto.W/2, Y: onto.Y + onto.H/2}
		w.PressAt(middleOf(w, 0, 0)).MoveTo(at).LetGo(at)
		in, _ := f.gave("task", "move")
		return in.Given
	}

	if got := drop(2); got["after"] != "2" {
		t.Errorf("dropped on the last card, asked %+v, want behind task 2", got)
	}
	if got := drop(3); got["last"] != "true" {
		t.Errorf("dropped one past the last card, asked %+v, want last", got)
	}
}
