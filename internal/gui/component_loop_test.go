package gui

import (
	"slices"
	"testing"
)

func TestTheLoopMarkIsAnOpenRingDrawnInTheInkGiven(t *testing.T) {
	s := Dark()
	box := Box{X: 4, Y: 4, W: 13, H: 13}
	shot, drawn := drawnAlone(t, 21, 21, func(k Kit) { LoopMark{Ink: s.Colour.Ink.Faint}.Draw(k, box) })
	if lettered(shot, box, s.Colour.Ink.Faint) == 0 {
		t.Error("the loop is not drawn in the ink given")
	}
	if lettered(shot, Box{X: box.X + box.W/2 - 1, Y: box.Y + box.H/2 - 1, W: 2, H: 2}, s.Colour.Ink.Faint) != 0 {
		t.Error("the loop is filled where it should be stroked")
	}
	if want := []Drawing{{Part: "loop", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the loop recorded %+v", drawn)
	}
}

func TestASeriesNumberIsTheLoopThenItsFiguresInTheInkGiven(t *testing.T) {
	s := Dark()
	box := Box{X: 4, Y: 4, W: SeriesNumberWide(12, small), H: lineHigh}
	shot, drawn := drawnAlone(t, 80, 30, func(k Kit) { SeriesNumber{Number: 12, Ink: s.Colour.Accent.Base}.Draw(k, box) })
	if got := parts(drawn); !slices.Equal(got, []string{"series", "series/loop"}) {
		t.Errorf("a series' number recorded %v", got)
	}
	if drawn[1].At.X != box.X || drawn[1].At.W != markSide-2 {
		t.Errorf("the loop stands at %+v, not first in %+v", drawn[1].At, box)
	}
	if lettered(shot, Box{X: drawn[1].At.X + drawn[1].At.W, Y: box.Y, W: box.W - drawn[1].At.W, H: box.H}, s.Colour.Accent.Base) == 0 {
		t.Error("the figures are not written after the loop, in the ink given")
	}
}

func TestAHeadingNumberingASeriesWritesTheLoopBeforeItsNumber(t *testing.T) {
	box := Box{X: 10, Y: 10, W: 300, H: 26}
	_, drawn := drawnAlone(t, 330, 50, func(k Kit) { Heading{Said: "Stand-up", Number: "3", Loop: true}.Draw(k, box) })
	if got := parts(drawn); !slices.Equal(got, []string{"heading", "heading/series", "heading/series/loop"}) {
		t.Errorf("a heading numbering a series recorded %v", got)
	}
}

func TestAListItemOfASeriesSaysItsRuleItsDayAndItsState(t *testing.T) {
	s := Dark()
	slot := Box{X: 10, Y: 10, W: 400, H: ItemHigh}
	shot, drawn := drawnAlone(t, 440, 80, func(k Kit) {
		ListItem{Number: "3", Loop: true, Title: "Stand-up", Says: "every weekday 10:30", Aside: "today", Badge: "active", BadgeStanding: "active"}.Draw(k, slot)
	})
	if got := parts(drawn); !slices.Equal(got, []string{"item", "item/stripe", "item/pill", "item/series", "item/series/loop"}) {
		t.Errorf("a series' item recorded %v", got)
	}
	_, _, second := ItemLine(ItemOf(slot, 0))
	if lettered(shot, second, s.Colour.Ink.Faint) == 0 {
		t.Error("the rule is not written on the second line, in the faint ink")
	}
}
