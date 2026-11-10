package gui

import (
	"image"
	"slices"
	"strings"
	"testing"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"cairn/internal/engine"
)

func cardTitled(title string) engine.Card {
	return engine.Card{Number: 1, Title: title, State: "backlog"}
}

func TestAConnectorRunsDownAndElbowsIntoWhatItJoins(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 20, H: 40}
	shot, drawn := drawnAlone(t, 60, 70, func(k Kit) { Connector{Elbow: 15}.Draw(k, box) })
	if got := pixel(shot, box.X, box.Y+35); got != s.Colour.Line.Edge {
		t.Errorf("a connector with more below does not run to its foot: %v", got)
	}
	if got := pixel(shot, box.X+12, box.Y+15); got != s.Colour.Line.Edge {
		t.Errorf("the connector does not elbow towards what it joins: %v", got)
	}
	if want := []Drawing{{Part: "connector", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the connector recorded %+v", drawn)
	}
	last, _ := drawnAlone(t, 60, 70, func(k Kit) { Connector{Last: true, Elbow: 15}.Draw(k, box) })
	if got := pixel(last, box.X, box.Y+35); got == s.Colour.Line.Edge {
		t.Error("the last connector runs on past what it joins")
	}
	if got := pixel(last, box.X, box.Y+5); got != s.Colour.Line.Edge {
		t.Errorf("the last connector does not come down from above: %v", got)
	}
}

func TestAListItemIsIndentedByItsDepthAndJoinedToItsParent(t *testing.T) {
	slot := Box{X: 10, Y: 10, W: 400, H: ItemHigh}
	_, top := drawnAlone(t, 440, 80, func(k Kit) {
		ListItem{Number: "#4", Title: "Home automation", Standing: "active", Counts: "1/2 · 50%", Done: 1, Total: 2}.Draw(k, slot)
	})
	if got := parts(top); !slices.Equal(got, []string{"item", "item/stripe", "item/standing", "item/progress"}) {
		t.Errorf("a level-one item recorded %v", got)
	}
	if top[0].At != ItemOf(slot, 0) {
		t.Errorf("a level-one item stands at %+v, want %+v", top[0].At, ItemOf(slot, 0))
	}
	_, under := drawnAlone(t, 440, 80, func(k Kit) {
		ListItem{Number: "#5", Title: "Enclosure", Depth: 1, Last: true, Chip: "project"}.Draw(k, slot)
	})
	if got := parts(under); !slices.Equal(got, []string{"connector", "item", "item/stripe", "item/label"}) {
		t.Errorf("a sub-objective's item recorded %v", got)
	}
	if item := ItemOf(slot, 1); item.X != slot.X+indent || under[1].At != item {
		t.Errorf("a sub-objective's item stands at %+v, want %d in", under[1].At, indent)
	}
	if connector := under[0].At; connector.X+connector.W != under[1].At.X {
		t.Errorf("the connector %+v does not reach the item %+v", connector, under[1].At)
	}
}

func TestAListItemHasNoGroundAtRestAndIsRingedWhenChosen(t *testing.T) {
	s := Dark()
	slot := Box{X: 10, Y: 10, W: 300, H: ItemHigh}
	inside := func(item Box) (int, int) { return item.X + item.W - 30, item.Y + item.H - 4 }
	rest, _ := drawnAlone(t, 340, 80, func(k Kit) { ListItem{Title: "Rest"}.Draw(k, slot) })
	if x, y := inside(ItemOf(slot, 0)); pixel(rest, x, y) != s.Colour.Surface.Back {
		t.Error("an item at rest stands on a ground")
	}
	hovered, _ := drawnAlone(t, 340, 80, func(k Kit) { ListItem{Title: "Hovered", Hovered: true}.Draw(k, slot) })
	if x, y := inside(ItemOf(slot, 0)); pixel(hovered, x, y) != s.Colour.Surface.Hover {
		t.Error("an item under the pointer is not on the hover's ground")
	}
	chosen, drawn := drawnAlone(t, 340, 80, func(k Kit) { ListItem{Title: "Chosen", Chosen: true}.Draw(k, slot) })
	if x, y := inside(ItemOf(slot, 0)); pixel(chosen, x, y) != s.Colour.Surface.Card {
		t.Error("the item chosen is not on a card's ground")
	}
	if !slices.Contains(parts(drawn), "item/ring") {
		t.Errorf("the item chosen is not ringed: %v", parts(drawn))
	}
}

func TestTheStandingMarksAreDrawnInTheirStandingsInk(t *testing.T) {
	s := Dark()
	box := Box{X: 4, Y: 4, W: 13, H: 13}
	for _, state := range []string{"active", "on-hold", "reached", "dropped"} {
		shot, drawn := drawnAlone(t, 21, 21, func(k Kit) { StandingMark{State: state}.Draw(k, box) })
		if lettered(shot, box, s.Colour.Standing[state].Ink) == 0 {
			t.Errorf("the mark of %q is not in its standing's ink", state)
		}
		if want := []Drawing{{Part: "standing", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
			t.Errorf("the mark of %q recorded %+v", state, drawn)
		}
	}
	shapes := map[string]int{}
	for _, state := range []string{"active", "on-hold", "reached", "dropped"} {
		shot, _ := drawnAlone(t, 21, 21, func(k Kit) { StandingMark{State: state}.Draw(k, box) })
		shapes[state] = ink(shot, box, s.Colour.Surface.Back)
	}
	if shapes["active"] == shapes["on-hold"] || shapes["reached"] == shapes["dropped"] {
		t.Errorf("two standings draw the same mark: %v", shapes)
	}
}

func TestTheServesMarkIsADiamondDrawnInTheInkGiven(t *testing.T) {
	s := Dark()
	box := Box{X: 4, Y: 4, W: 13, H: 13}
	shot, drawn := drawnAlone(t, 21, 21, func(k Kit) { ServesMark{Ink: s.Colour.Ink.Faint}.Draw(k, box) })
	if lettered(shot, box, s.Colour.Ink.Faint) == 0 {
		t.Error("the mark is not drawn in the ink given")
	}
	if got := pixel(shot, box.X+box.W/2, box.Y+box.H/2); got != s.Colour.Surface.Back {
		t.Error("the diamond is filled where it should be stroked")
	}
	if want := []Drawing{{Part: "serves", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the mark recorded %+v", drawn)
	}
}

func TestAHeadingWritesItsNumberBesideItsTitle(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 300, H: 26}
	shot, _ := drawnAlone(t, 330, 50, func(k Kit) { Heading{Said: "Home", Number: "#4"}.Draw(k, box) })
	if lettered(shot, Box{X: box.X + 40, Y: box.Y, W: 80, H: box.H}, s.Colour.Ink.Low) == 0 {
		t.Error("the number is not written beside the title, in the low ink")
	}
}

func TestFactsWriteEachNameBesideItsValueAndTheAsideUnder(t *testing.T) {
	facts := Facts{Lines: []Fact{{Name: "done", Value: "3 of 6 tasks"}, {Name: "spent", Value: "8h25m", Aside: "of 12h estimated"}}}
	if got := facts.High(); got != 3*lineHigh {
		t.Errorf("two facts, one with an aside, are %d high, want %d", got, 3*lineHigh)
	}
	s := Dark()
	box := Box{X: 10, Y: 10, W: 300, H: facts.High()}
	shot, drawn := drawnAlone(t, 330, 80, func(k Kit) { facts.Draw(k, box) })
	if lettered(shot, Box{X: box.X + factName, Y: box.Y, W: 100, H: lineHigh}, s.Colour.Ink.Text) == 0 {
		t.Error("a value is not written in the ink beside its name")
	}
	if lettered(shot, Box{X: box.X + factName, Y: box.Y + 2*lineHigh, W: 100, H: lineHigh}, s.Colour.Ink.Faint) == 0 {
		t.Error("the aside is not written faint under its value")
	}
	if want := []Drawing{{Part: "facts", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the facts recorded %+v", drawn)
	}
}

func TestFiguresWriteTheLateOneInTheLateInk(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 400, H: FiguresHigh}
	shot, drawn := drawnAlone(t, 430, 60, func(k Kit) {
		Figures{Said: []Figure{{Figure: "46", Word: "done"}, {Figure: "2", Word: "missed", Late: true}}}.Draw(k, box)
	})
	if lettered(shot, box, s.Colour.State.Late) == 0 {
		t.Error("the late figure is not in the late ink")
	}
	if lettered(shot, box, s.Colour.Ink.Text) == 0 {
		t.Error("the figure is not in the ink")
	}
	if want := []Drawing{{Part: "figures", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the figures recorded %+v", drawn)
	}
}

func TestABannerRefusesInTheWrongsTonesAndWarnsInOnHolds(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 300, H: 60}
	refused, drawn := drawnAlone(t, 330, 90, func(k Kit) { Banner{Said: "Already counted", Then: "Untick both."}.Draw(k, box) })
	if got := pixel(refused, box.X+box.W-6, box.Y+box.H-6); got != s.Colour.State.Danger {
		t.Errorf("a refusal stands on %v, not the danger's ground", got)
	}
	if lettered(refused, box, s.Colour.State.Wrong) == 0 {
		t.Error("a refusal is not written in the wrong's ink")
	}
	if want := []Drawing{{Part: "banner", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the banner recorded %+v", drawn)
	}
	warned, _ := drawnAlone(t, 330, 90, func(k Kit) { Banner{Said: "Every part is reached", Warn: true}.Draw(k, box) })
	if got := pixel(warned, box.X+box.W-6, box.Y+box.H-6); got != s.Colour.Standing["on-hold"].Back {
		t.Errorf("a warning stands on %v, not on hold's ground", got)
	}
	if got := BannerHigh(Reckoned, "short", "", 300); got >= BannerHigh(Reckoned, "short", "then", 300) {
		t.Errorf("a banner with a second line is no taller: %d", got)
	}
	if one, two := BannerHigh(Reckoned, "short", "", 300), BannerHigh(Reckoned, strings.Repeat("long words ", 12), "", 300); one >= two {
		t.Errorf("a banner whose words wrap is no taller: %d, %d", one, two)
	}
}

func TestACardInTileLayoutSaysItsCountsItsNotesAndItsWarning(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 220, H: TileHigh}
	tile := &GoalTile{Number: "#7", Title: "Features", State: "on-hold", Chip: "project", Done: 1, Total: 3, Counts: "1 of 3 · 33%",
		Notes: []string{"#44 also in Firmware"}, Warn: "waits on Firmware"}
	shot, drawn := drawnAlone(t, 250, 150, func(k Kit) { Card{Goal: tile, Chosen: true}.Draw(k, box) })
	if lettered(shot, box, s.Colour.State.Late) == 0 {
		t.Error("the tile's warning is not in the late ink")
	}
	want := []string{"card", "card/ring", "card/stripe", "card/standing", "card/label", "card/progress"}
	if got := parts(drawn); !slices.Equal(got, want) {
		t.Errorf("a chosen tile recorded %v, want %v", got, want)
	}
}

func TestAReferenceCardSaysWhyItIsThereInTheLateInk(t *testing.T) {
	s := Dark()
	slot := Box{X: 10, Y: 10, W: 240, H: CardHeight}
	shot, _ := drawnAlone(t, 270, 120, func(k Kit) {
		Card{Face: cardTitled("Order the relay"), Reference: true, Why: "outside this view · blocks #2", Colour: func(string) string { return "" }}.Draw(k, slot)
	})
	if lettered(shot, PartsOf(slot).Number, s.Colour.State.Late) == 0 {
		t.Error("a reference does not say why it is there, in the late ink")
	}
}

func TestFiguresGiveEachFigureTheRoomItsWeightTakes(t *testing.T) {
	th, s := Theme(), Dark()
	boxes := map[string]Box{}
	was := noteSaid
	noteSaid = func(_ layout.Context, _ *material.Theme, b Box, said string, _ int, _ font.Font, _ int, _ text.Alignment) {
		boxes[said] = b
	}
	defer func() { noteSaid = was }()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Constraints: layout.Exact(image.Pt(600, 100)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	Figures{Said: []Figure{{Figure: "—", Word: "spent"}, {Figure: "3", Word: "left"}, {Figure: "14h50m", Word: "spent"}}}.Draw(NewKit(gtx, th, s, &Drawn{}), Box{X: 0, Y: 0, W: 600, H: FiguresHigh})
	for _, figure := range []string{"—", "3", "14h50m"} {
		label := material.Label(th, unit.Sp(float32(s.Type.Figure)), figure)
		label.Font.Weight = font.SemiBold
		laid := label.Layout(layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(1<<20, 1<<20)}, Metric: gtx.Metric}).Size
		if got, ok := boxes[figure]; !ok || got.W < laid.X {
			t.Errorf("the figure %q is written in %+v, %d wide whole", figure, got, laid.X)
		}
	}
}
