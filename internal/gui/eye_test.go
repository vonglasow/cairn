package gui

import (
	"cairn/internal/engine"
	"image/color"
	"testing"
)

func TestACardInFocusStandsItsGlowAtTheRightOfItsFirstLine(t *testing.T) {
	title := Box{X: 100, Y: 200, W: 240, H: 2 * lineHigh}
	mark, ok := FocusMark(engine.Card{InFocus: true}, title)
	if !ok {
		t.Fatal("a card in focus has no glowing dot")
	}
	if mark.X+mark.W != title.X+title.W || mark.Y < title.Y || mark.Y+mark.H > title.Y+lineHigh || mark.W != markSide {
		t.Errorf("the glowing dot is at %+v, not at the right of the first line of %+v", mark, title)
	}
	if _, ok := FocusMark(engine.Card{}, title); ok {
		t.Error("a card nobody is on has a glowing dot")
	}
}

func TestOnlyTheFirstLineOfATitleInFocusMakesRoomForTheGlow(t *testing.T) {
	title := Box{W: 120, H: 2 * lineHigh}
	first, rest := TitleRoom(engine.Card{InFocus: true}, title)
	if first != 120-markSide-markSpace || rest != 120 {
		t.Errorf("a title in focus runs %d then %d wide", first, rest)
	}
	got := TitleLines("Write the abcdefghijklmno", first, rest, eightEach)
	if len(got) != 2 || got[0] != "Write the" || got[1] != "abcdefghijklmno" {
		t.Errorf("the title in focus was laid out %q", got)
	}
	if first, rest := TitleRoom(engine.Card{}, title); first != 120 || rest != 120 {
		t.Errorf("a title nobody is on runs %d then %d wide", first, rest)
	}
}

func TestTheGlowingDotIsDrawn(t *testing.T) {
	on := engine.Card{Number: 1, Title: "Write the spec", State: "todo", InFocus: true}
	off := engine.Card{Number: 2, Title: "Ship it", State: "todo"}
	w := holding(newFake(), on, off)
	shot := drawn(t, w)
	l := w.Layout()
	if got := painted(shot, PartsOf(l.CardAt(1, 0, 0)).Title, Dark().Colour.Accent.Base); got == 0 {
		t.Error("the card in focus draws no glowing dot")
	}
	if got := painted(shot, PartsOf(l.CardAt(1, 1, 0)).Title, Dark().Colour.Accent.Base); got != 0 {
		t.Errorf("the card nobody is on draws %d pixels of a glowing dot", got)
	}
}

func TestACardIsRoundedAndStripedByItsPriority(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo", Priority: "high"})
	shot := drawn(t, w)
	panel := PartsOf(w.Layout().CardAt(1, 0, 0)).Panel
	high := Dark().Colour.Priority["high"]
	if painted(shot, Box{X: panel.X, Y: panel.Y, W: 1, H: 1}, Dark().Colour.Surface.Card) != 0 {
		t.Error("the card's corner is not rounded")
	}
	if painted(shot, Box{X: panel.X, Y: panel.Y + panel.H/2, W: edgeWide, H: 1}, high) != edgeWide {
		t.Errorf("the stripe is not %d wide in the priority's colour", edgeWide)
	}
	if painted(shot, Box{X: panel.X + edgeWide, Y: panel.Y + panel.H/2, W: 1, H: 1}, high) != 0 {
		t.Error("the stripe runs wider than its width")
	}
	if painted(shot, Box{X: panel.X + panel.W/2, Y: panel.Y + 2, W: 1, H: 1}, Dark().Colour.Surface.Card) != 1 {
		t.Error("the card does not stand on its own ground")
	}
}

func TestACardInFocusStandsOnTheTint(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo", InFocus: true})
	panel := PartsOf(w.Layout().CardAt(1, 0, 0)).Panel
	if painted(drawn(t, w), Box{X: panel.X + panel.W/2, Y: panel.Y + 2, W: 1, H: 1}, Dark().Colour.Surface.Tint) != 1 {
		t.Error("the card in focus does not stand on the tint")
	}
}

func TestTheCountOfLabelsOpensABubbleOfThemAll(t *testing.T) {
	c := engine.Card{Number: 1, Title: "Write the spec", State: "todo", Labels: []string{"an-altogether-overlong-label", "another-just-as-long-one", "gui"}}
	w := holding(newFake(), c)
	parts := PartsOf(w.Layout().CardAt(1, 0, 0))
	more := CardRowOf(c, parts.Number, w.measured(), w.now, engine.Pages{}).More
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
	if painted(drawn(t, w), Box{X: box.X + box.W/2, Y: box.Y + box.H - 3, W: 1, H: 1}, Dark().Colour.Surface.Raised) != 1 {
		t.Error("the bubble is not drawn")
	}
}

func TestOnlyACardWithStepsDrawsItsBar(t *testing.T) {
	with := engine.Card{Number: 1, Title: "Write the spec", State: "todo", SubDone: 1, SubTotal: 3}
	without := engine.Card{Number: 2, Title: "Ship it", State: "todo"}
	w := holding(newFake(), with, without)
	shot := drawn(t, w)
	l := w.Layout()
	if painted(shot, PartsOf(l.CardAt(1, 0, 0)).Steps, Dark().Colour.State.Done) == 0 {
		t.Error("a card with a step done draws no bar")
	}
	if got := painted(shot, PartsOf(l.CardAt(1, 1, 0)).Steps, Dark().Colour.Surface.Control); got != 0 {
		t.Errorf("a card with no step draws %d pixels of a track", got)
	}
}

func TestTheMarkOfACardInFocusIsADotInAFaintHaloOfItsOwnInk(t *testing.T) {
	got := GlowOf(Box{X: 10, Y: 20, W: 14, H: 14})
	if got != (Glow{X: 17, Y: 27, Dot: 3.4, Halo: 6}) {
		t.Errorf("the glow in a box of 14 is %+v", got)
	}
	small := GlowOf(Box{X: 0, Y: 0, W: markSide, H: markSide})
	if small.X-small.Halo < 0 || small.X+small.Halo > markSide || small.Y-small.Halo < 0 || small.Y+small.Halo > markSide {
		t.Errorf("the halo %+v spills out of its box of %d", small, markSide)
	}
	if small.Dot >= small.Halo {
		t.Errorf("the dot %+v is no smaller than its halo", small)
	}
	accent, glow := Dark().Colour.Accent.Base, Dark().Colour.Accent.Glow
	if want := (color.NRGBA{R: accent.R, G: accent.G, B: accent.B, A: uint8(int(accent.A) * 3 / 10)}); glow != want {
		t.Errorf("the glow is %+v, not the accent at 30%%: %+v", glow, want)
	}
}

func TestAboutCairnPaintsTheLogo(t *testing.T) {
	w := holding(newFake())
	w.screen, w.question = OnAsking, aboutCairn
	shot := drawn(t, w)
	if got := painted(shot, w.AboutLogo(), color.NRGBA{R: 0x7B, G: 0x93, B: 0xFF, A: 255}); got == 0 {
		t.Error("About cairn paints no logo")
	}
}
