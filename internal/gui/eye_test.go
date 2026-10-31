package gui

import (
	"cairn/internal/engine"
	"testing"
)

func TestACardInFocusEndsItsTitleWithACairn(t *testing.T) {
	title := Box{X: 100, Y: 200, W: 240, H: 2 * lineHigh}
	mark, ok := FocusMark(engine.Card{InFocus: true}, title, 2, 60)
	if !ok {
		t.Fatal("a card in focus has no cairn")
	}
	if mark.X != title.X+60+markSpace || mark.Y < title.Y+lineHigh || mark.Y+mark.H > title.Y+2*lineHigh || mark.W != markSide {
		t.Errorf("the cairn is at %+v, after a second line 60 wide in %+v", mark, title)
	}
	if _, ok := FocusMark(engine.Card{}, title, 2, 60); ok {
		t.Error("a card nobody is on has a cairn")
	}
}

func TestTheCairnIsDrawn(t *testing.T) {
	on := engine.Card{Number: 1, Title: "Write the spec", State: "todo", InFocus: true}
	off := engine.Card{Number: 2, Title: "Ship it", State: "todo"}
	w := holding(newFake(), on, off)
	shot := drawn(t, w)
	l := w.Layout()
	if got := painted(shot, PartsOf(l.CardAt(1, 0, 0)).Title, Dark().Chosen); got == 0 {
		t.Error("the card in focus draws no cairn")
	}
	if got := painted(shot, PartsOf(l.CardAt(1, 1, 0)).Title, Dark().Chosen); got != 0 {
		t.Errorf("the card nobody is on draws %d pixels of a cairn", got)
	}
}

func TestACardIsRoundedAndStripedByItsPriority(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo", Priority: "high"})
	shot := drawn(t, w)
	panel := PartsOf(w.Layout().CardAt(1, 0, 0)).Panel
	high := Dark().Priority["high"]
	if painted(shot, Box{X: panel.X, Y: panel.Y, W: 1, H: 1}, Dark().Card) != 0 {
		t.Error("the card's corner is not rounded")
	}
	if painted(shot, Box{X: panel.X, Y: panel.Y + panel.H/2, W: edgeWide, H: 1}, high) != edgeWide {
		t.Errorf("the stripe is not %d wide in the priority's colour", edgeWide)
	}
	if painted(shot, Box{X: panel.X + edgeWide, Y: panel.Y + panel.H/2, W: 1, H: 1}, high) != 0 {
		t.Error("the stripe runs wider than its width")
	}
	if painted(shot, Box{X: panel.X + panel.W/2, Y: panel.Y + 2, W: 1, H: 1}, Dark().Card) != 1 {
		t.Error("the card does not stand on its own ground")
	}
}

func TestACardInFocusStandsOnTheTint(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo", InFocus: true})
	panel := PartsOf(w.Layout().CardAt(1, 0, 0)).Panel
	if painted(drawn(t, w), Box{X: panel.X + panel.W/2, Y: panel.Y + 2, W: 1, H: 1}, Dark().Tint) != 1 {
		t.Error("the card in focus does not stand on the tint")
	}
}

func TestTheCountOfLabelsOpensABubbleOfThemAll(t *testing.T) {
	c := engine.Card{Number: 1, Title: "Write the spec", State: "todo", Labels: []string{"an-altogether-overlong-label", "another-just-as-long-one", "gui"}}
	w := holding(newFake(), c)
	parts := PartsOf(w.Layout().CardAt(1, 0, 0))
	more := CardRowOf(c, parts.Number, w.measured(), w.now).More
	w = w.Hovering(Point{X: more.X + more.W/2, Y: more.Y + more.H/2})
	box, pills, ok := w.LabelsBubbleBox()
	if !ok || len(pills) != 3 {
		t.Fatalf("resting on the count opens %v pills in %+v", len(pills), box)
	}
	for _, pill := range pills {
		if pill.X < box.X || pill.X+pill.W > box.X+box.W || pill.Y+pill.H > box.Y+box.H {
			t.Errorf("a pill %+v runs out of its bubble %+v", pill, box)
		}
	}
	if painted(drawn(t, w), Box{X: box.X + box.W/2, Y: box.Y + box.H - 3, W: 1, H: 1}, Dark().Panel) != 1 {
		t.Error("the bubble is not drawn")
	}
}

func TestOnlyACardWithStepsDrawsItsBar(t *testing.T) {
	with := engine.Card{Number: 1, Title: "Write the spec", State: "todo", SubDone: 1, SubTotal: 3}
	without := engine.Card{Number: 2, Title: "Ship it", State: "todo"}
	w := holding(newFake(), with, without)
	shot := drawn(t, w)
	l := w.Layout()
	if painted(shot, PartsOf(l.CardAt(1, 0, 0)).Steps, Dark().Done) == 0 {
		t.Error("a card with a step done draws no bar")
	}
	if got := painted(shot, PartsOf(l.CardAt(1, 1, 0)).Steps, Dark().Pill); got != 0 {
		t.Errorf("a card with no step draws %d pixels of a track", got)
	}
}
