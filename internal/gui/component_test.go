package gui

import (
	"image"
	"image/color"
	"slices"
	"testing"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"cairn/internal/engine"
)

func drawnAlone(t *testing.T, wide, high int, draw func(k Kit)) (*image.RGBA, []Drawing) {
	t.Helper()
	window := offScreen(t, wide, high)
	defer window.Release()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Constraints: layout.Exact(image.Pt(wide, high)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	var drawn Drawn
	k := NewKit(gtx, Theme(), Dark(), &drawn)
	fill(gtx, Box{W: wide, H: high}, k.skin.Colour.Surface.Back)
	draw(k)
	if err := window.Frame(gtx.Ops); err != nil {
		t.Fatal(err)
	}
	shot := image.NewRGBA(image.Rect(0, 0, wide, high))
	if err := window.Screenshot(shot); err != nil {
		t.Fatal(err)
	}
	return shot, drawn.All()
}

func pixel(shot *image.RGBA, x, y int) color.NRGBA {
	c := shot.RGBAAt(x, y)
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}
}

func parts(drawn []Drawing) []string {
	var named []string
	for _, d := range drawn {
		named = append(named, d.Part)
	}
	return named
}

func TestAGroundIsRoundedInTheToneOfWhatItCarries(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 100, H: 60}
	for _, c := range []struct {
		ground Ground
		want   color.NRGBA
	}{{Ground{}, s.Colour.Surface.Ground}, {Ground{Raised: true}, s.Colour.Surface.Raised}} {
		shot, drawn := drawnAlone(t, 140, 90, func(k Kit) { c.ground.Draw(k, box) })
		if got := pixel(shot, 60, 40); got != c.want {
			t.Errorf("%+v stands on %v, not %v", c.ground, got, c.want)
		}
		if pixel(shot, box.X, box.Y) == c.want {
			t.Errorf("%+v is square at its corner", c.ground)
		}
		if want := []Drawing{{Part: "ground", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
			t.Errorf("%+v recorded %+v, want %+v", c.ground, drawn, want)
		}
	}
}

func TestAPanelIsRaisedAtAPanelsCornersAndHeadedInside(t *testing.T) {
	s := Dark()
	box := Box{X: 20, Y: 20, W: 220, H: 120}
	shot, drawn := drawnAlone(t, 260, 180, func(k Kit) { Panel{Head: "Backups"}.Draw(k.on(LayerPanel), box) })
	if got := pixel(shot, box.X+box.W/2, box.Y+box.H-10); got != s.Colour.Surface.Raised {
		t.Errorf("the panel stands on %v, not on a raised ground", got)
	}
	if pixel(shot, box.X+3, box.Y+3) == s.Colour.Surface.Raised {
		t.Error("the panel is rounded less than a panel's corners")
	}
	if lettered(shot, FormHeading(box), s.Colour.Ink.Text) == 0 {
		t.Error("the panel's heading is not written in its heading")
	}
	want := []Drawing{{Part: "panel", At: box, Layer: LayerPanel}, {Part: "panel/heading", At: FormHeading(box), Layer: LayerPanel}}
	if !slices.Equal(drawn, want) {
		t.Errorf("the panel recorded %+v, want %+v", drawn, want)
	}
	if _, bare := drawnAlone(t, 260, 180, func(k Kit) { Panel{}.Draw(k, box) }); !slices.Equal(parts(bare), []string{"panel"}) {
		t.Errorf("a panel with no head recorded %v", parts(bare))
	}
}

func TestAPopoverIsRaisedAtAGroundsCornersRounderThanACardLessThanAPanel(t *testing.T) {
	s := Dark()
	box := Box{X: 20, Y: 20, W: 120, H: 80}
	shot, drawn := drawnAlone(t, 180, 140, func(k Kit) { Popover{}.Draw(k.on(LayerPopover), box) })
	if got := pixel(shot, box.X+3, box.Y+3); got != s.Colour.Surface.Raised {
		t.Errorf("the popover is rounded more than a ground's corners: %v three points in", got)
	}
	if pixel(shot, box.X+1, box.Y+1) == s.Colour.Surface.Raised {
		t.Error("the popover is rounded no more than a card")
	}
	if want := []Drawing{{Part: "popover", At: box, Layer: LayerPopover}}; !slices.Equal(drawn, want) {
		t.Errorf("the popover recorded %+v, want %+v", drawn, want)
	}
}

func TestAVeilDimsWhatIsUnderItTheLightOneLess(t *testing.T) {
	s := Dark()
	if (Veil{}).ink(s) != s.Colour.Surface.Veil || (Veil{Light: true}).ink(s) != s.Colour.Surface.VeilLight {
		t.Error("a veil does not take its tokens")
	}
	box := Box{W: 40, H: 40}
	under := func(k Kit) { fill(k.gtx, box, s.Colour.Ink.Text) }
	dark, drawn := drawnAlone(t, 40, 40, func(k Kit) { under(k); Veil{}.Draw(k.on(LayerPanel), box) })
	light, _ := drawnAlone(t, 40, 40, func(k Kit) { under(k); Veil{Light: true}.Draw(k, box) })
	if pixel(dark, 20, 20).R >= pixel(light, 20, 20).R || pixel(light, 20, 20).R >= s.Colour.Ink.Text.R {
		t.Errorf("a veil leaves %v and a light one %v over %v", pixel(dark, 20, 20), pixel(light, 20, 20), s.Colour.Ink.Text)
	}
	if want := []Drawing{{Part: "veil", At: box, Layer: LayerPanel}}; !slices.Equal(drawn, want) {
		t.Errorf("the veil recorded %+v, want %+v", drawn, want)
	}
}

func TestARuleTakesTheLineOfWhatItIsDrawnOn(t *testing.T) {
	s := Dark()
	along := Box{X: 10, Y: 20, W: 80}
	for _, c := range []struct {
		layer Layer
		want  color.NRGBA
	}{{LayerBoard, s.Colour.Line.Rule}, {LayerPanel, s.Colour.Line.Raised}, {LayerPopover, s.Colour.Line.Raised}} {
		shot, drawn := drawnAlone(t, 100, 40, func(k Kit) { Rule{}.Draw(k.on(c.layer), along) })
		if got := pixel(shot, 50, 20); got != c.want {
			t.Errorf("a rule on layer %v is drawn in %v, not %v", c.layer, got, c.want)
		}
		if got := pixel(shot, 50, 21); got == c.want {
			t.Errorf("a rule on layer %v is thicker than a rule", c.layer)
		}
		if want := []Drawing{{Part: "rule", At: Box{X: 10, Y: 20, W: 80, H: s.Size.Rule}, Layer: c.layer}}; !slices.Equal(drawn, want) {
			t.Errorf("the rule recorded %+v, want %+v", drawn, want)
		}
	}
}

func TestAHeadingIsWrittenInTheReadingInk(t *testing.T) {
	box := Box{X: 10, Y: 10, W: 200, H: titleHi}
	shot, drawn := drawnAlone(t, 240, 50, func(k Kit) { Heading{Said: "Backups"}.Draw(k, box) })
	if lettered(shot, box, Dark().Colour.Ink.Text) == 0 {
		t.Error("the heading is not written")
	}
	if want := []Drawing{{Part: "heading", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the heading recorded %+v, want %+v", drawn, want)
	}
}

func TestAStripeIsSquareAlongAnEdgeAndRoundedStandingFree(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 5, H: 40}
	high := s.Colour.Priority[string("high")]
	attached, drawn := drawnAlone(t, 40, 60, func(k Kit) { Stripe{Priority: "high"}.Draw(k, box) })
	if pixel(attached, box.X, box.Y) != high || pixel(attached, box.X+box.W-1, box.Y) == high {
		t.Error("a stripe along an edge is not square on its edge and rounded away from it")
	}
	if want := []Drawing{{Part: "stripe", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the stripe recorded %+v, want %+v", drawn, want)
	}
	free, _ := drawnAlone(t, 40, 60, func(k Kit) { Stripe{Priority: "high", Free: true}.Draw(k, box) })
	if pixel(free, box.X, box.Y) == high || pixel(free, box.X+2, box.Y+20) != high {
		t.Error("a stripe standing free is not rounded at every corner")
	}
	none, _ := drawnAlone(t, 40, 60, func(k Kit) { Stripe{}.Draw(k, box) })
	if pixel(none, box.X+2, box.Y+20) != s.Colour.Line.Edge {
		t.Error("a stripe of no priority is not drawn in an edge's line")
	}
}

func cardDrawn(t *testing.T, c Card) (*image.RGBA, []Drawing, Parts) {
	t.Helper()
	if c.Colour == nil {
		c.Colour = func(string) string { return "mauve" }
	}
	slot := Box{X: 10, Y: 10, W: 280, H: CardHeight}
	shot, drawn := drawnAlone(t, 300, 120, func(k Kit) { c.Draw(k, slot) })
	return shot, drawn, PartsOf(slot)
}

func TestACardStandsOnTheGroundOfItsState(t *testing.T) {
	s := Dark()
	face := engine.Card{Number: 7, Title: "Write the scenario", State: "todo", Priority: "high"}
	focused := face
	focused.InFocus = true
	for _, c := range []struct {
		name string
		card Card
		want color.NRGBA
	}{
		{"at rest", Card{Face: face}, s.Colour.Surface.Card},
		{"under the pointer", Card{Face: face, Hovered: true}, s.Colour.Surface.Hover},
		{"in focus", Card{Face: focused, Hovered: true}, s.Colour.Surface.Tint},
		{"chosen", Card{Face: face, Chosen: true}, s.Colour.Surface.Card},
		{"dragged", Card{Face: face, Dragged: true}, s.Colour.Surface.Ghost},
		{"a reference", Card{Face: face, Reference: true}, s.Colour.Surface.Back},
	} {
		shot, drawn, p := cardDrawn(t, c.card)
		if got := pixel(shot, p.Panel.X+p.Panel.W-4, p.Panel.Y+44); got != c.want {
			t.Errorf("a card %s stands on %v, not %v", c.name, got, c.want)
		}
		if len(drawn) == 0 || drawn[0].Part != "card" || (!c.card.Dragged && drawn[0].At != p.Panel) {
			t.Errorf("a card %s recorded %+v first", c.name, drawn)
		}
	}
	shot, _, p := cardDrawn(t, Card{Face: face, Chosen: true})
	if pixel(shot, p.Panel.X+p.Panel.W-1, p.Panel.Y+44) != s.Colour.Line.Selected {
		t.Error("a chosen card is not ringed in the line of what is chosen")
	}
	shot, _, p = cardDrawn(t, Card{Face: face, Reference: true})
	if painted(shot, Box{X: p.Panel.X + s.Radius.Card, Y: p.Panel.Y, W: p.Panel.W - 2*s.Radius.Card, H: 1}, s.Colour.Line.Edge) == 0 {
		t.Error("a reference is not edged by a dashed line")
	}
	if lettered(shot, p.Title, s.Colour.Ink.Faint) == 0 || lettered(shot, p.Title, s.Colour.Ink.Text) != 0 {
		t.Error("a reference's title is not written faint")
	}
	dimmed, _, p := cardDrawn(t, Card{Face: face, Dimmed: true})
	if got := pixel(dimmed, p.Panel.X+p.Panel.W-4, p.Panel.Y+44); got == s.Colour.Surface.Card || got == s.Colour.Surface.Back {
		t.Errorf("a dimmed card stands on %v, whole or not at all", got)
	}
}

func TestACardIsStripedInItsPriorityAndAnArchivedOneInAnEdge(t *testing.T) {
	s := Dark()
	face := engine.Card{Number: 7, Title: "Write the scenario", State: "todo", Priority: "critical"}
	shot, drawn, p := cardDrawn(t, Card{Face: face})
	if got := pixel(shot, p.Edge.X+1, p.Edge.Y+p.Edge.H/2); got != s.Colour.Priority["critical"] {
		t.Errorf("a critical card is striped in %v", got)
	}
	if !slices.Contains(drawn, Drawing{Part: "card/stripe", At: p.Edge, Layer: LayerBoard}) {
		t.Errorf("the card recorded no stripe inside it: %v", parts(drawn))
	}
	face.Archived = true
	shot, _, p = cardDrawn(t, Card{Face: face})
	if got := pixel(shot, p.Edge.X+1, p.Edge.Y+p.Edge.H/2); got != s.Colour.Line.Edge {
		t.Errorf("an archived card is striped in %v", got)
	}
}

func TestACardCountsTheLabelsItLeavesOutInAPillWithNoRing(t *testing.T) {
	s := Dark()
	face := engine.Card{Number: 7, Title: "Write the scenario", State: "todo", Labels: []string{"an-altogether-overlong-label", "another-just-as-long-one", "gui"}}
	shot, drawn, p := cardDrawn(t, Card{Face: face})
	laid := CardRowOf(face, p.Number, Measured(Theme()), time.Time{}, engine.Pages{})
	if laid.Count == "" {
		t.Fatal("the card leaves no label out")
	}
	top := pixel(shot, laid.More.X+laid.More.W/2, laid.More.Y)
	if top != s.Colour.Surface.Control {
		t.Errorf("the count's pill is edged in %v, not drawn as a plain pill", top)
	}
	if !slices.Contains(parts(drawn), "card/pill") || !slices.Contains(parts(drawn), "card/label") {
		t.Errorf("the card recorded %v, with no label and no count inside it", parts(drawn))
	}
}

func TestAnObjectivesCardDrawsItsStandingAsAPillOfThatStanding(t *testing.T) {
	s := Dark()
	face := goalFace(engine.Goal{Number: 3, Title: "Ship v0.11", State: "on-hold"})
	shot, drawn, _ := cardDrawn(t, Card{Face: face, Standing: "on-hold"})
	pill := StandingPill(face, Box{X: 10, Y: 10, W: 280, H: CardHeight}, Measured(Theme()))
	if got := pixel(shot, pill.X+pill.W/2, pill.Y+1); got != s.Colour.Standing["on-hold"].Back {
		t.Errorf("the standing stands on %v", got)
	}
	if slices.Contains(parts(drawn), "card/label") || !slices.Contains(parts(drawn), "card/pill") {
		t.Errorf("the standing is recorded as %v, not as a pill", parts(drawn))
	}
}

func TestALabelChipIsALabelsColourOverACard(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 80, H: pillHigh}
	ground, ink := s.LabelPill("mauve")
	shot, drawn := drawnAlone(t, 120, 40, func(k Kit) { LabelChip{Label: "feature", Colour: "mauve"}.Draw(k, box) })
	if got := pixel(shot, box.X+box.W/2, box.Y+1); got != ground {
		t.Errorf("a label chip stands on %v, not %v", got, ground)
	}
	if lettered(shot, box, ink) == 0 {
		t.Error("a label chip is not written in its colour")
	}
	if want := []Drawing{{Part: "label", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the chip recorded %+v, want %+v", drawn, want)
	}
	plain, _ := drawnAlone(t, 120, 40, func(k Kit) { LabelChip{Label: "someday"}.Draw(k, box) })
	if got := pixel(plain, box.X+box.W/2, box.Y+1); got != s.Colour.Surface.Control {
		t.Errorf("a label of no colour stands on %v", got)
	}
}

func TestARemovableLabelChipCarriesAStrokedCrossInsideIt(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 90, H: pillHigh}
	ground, ink := s.LabelPill("mauve")
	shot, drawn := drawnAlone(t, 120, 40, func(k Kit) { LabelChip{Label: "feature", Colour: "mauve", Removable: true}.Draw(k, box) })
	if len(drawn) != 2 || drawn[1].Part != "label/cross" {
		t.Fatalf("a removable chip recorded %+v", drawn)
	}
	cross := drawn[1].At
	if cross.X < box.X || cross.X+cross.W > box.X+box.W || cross.Y < box.Y || cross.Y+cross.H > box.Y+box.H {
		t.Errorf("the cross %+v stands outside the chip %+v", cross, box)
	}
	if lettered(shot, cross, ink) == 0 {
		t.Error("the cross is not drawn in the label's ink")
	}
	if got := pixel(shot, cross.X, cross.Y); got != ground {
		t.Errorf("the cross stands on %v, not on the chip's own ground", got)
	}
}

func TestAPillTakesTheGroundAndInkOfItsTone(t *testing.T) {
	s := Dark()
	for _, c := range []struct {
		pill        Pill
		ground, ink color.NRGBA
		grounded    bool
	}{
		{Pill{Said: "todo"}, s.Colour.Surface.Control, s.Colour.Ink.Faint, true},
		{Pill{Said: "deploy", OnControl: true}, s.Colour.Surface.Raised, s.Colour.Ink.Faint, true},
		{Pill{Said: "on-hold", Tone: ToneStanding, Standing: "on-hold"}, s.Colour.Standing["on-hold"].Back, s.Colour.Standing["on-hold"].Ink, true},
		{Pill{Said: "nowhere", Tone: ToneStanding, Standing: "nowhere"}, s.Colour.Surface.Control, s.Colour.Ink.Faint, true},
		{Pill{Said: "here", Tone: ToneHere}, s.Colour.Standing["active"].Back, s.Colour.Standing["active"].Ink, true},
		{Pill{Said: "↻3", Tone: ToneLink}, s.Colour.Surface.Tint, s.Colour.Accent.Base, true},
		{Pill{Said: "global", Tone: ToneInherited}, color.NRGBA{}, s.Colour.Ink.Faint, false},
	} {
		ground, ink, grounded := c.pill.inks(s)
		if grounded != c.grounded || ink != c.ink || (grounded && ground != c.ground) {
			t.Errorf("%+v takes %v / %v (grounded %v)", c.pill, ground, ink, grounded)
		}
	}
	box := Box{X: 10, Y: 10, W: 60, H: pillHigh}
	shot, drawn := drawnAlone(t, 100, 40, func(k Kit) { Pill{Said: "+2"}.Draw(k, box) })
	if got := pixel(shot, box.X+box.W/2, box.Y); got != s.Colour.Surface.Control {
		t.Errorf("a pill's edge is %v, not its ground", got)
	}
	if want := []Drawing{{Part: "pill", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the pill recorded %+v, want %+v", drawn, want)
	}
	dashed, drawn := drawnAlone(t, 100, 40, func(k Kit) { Pill{Said: "global", Tone: ToneInherited}.Draw(k, box) })
	if pixel(dashed, box.X+box.W/2, box.Y+box.H/2+3) != s.Colour.Surface.Back || painted(dashed, Box{X: box.X + 9, Y: box.Y, W: box.W - 18, H: 1}, s.Colour.Line.Edge) == 0 {
		t.Error("an inherited pill is not a dashed line around nothing")
	}
	if !slices.Equal(parts(drawn), []string{"pill", "pill/dashed"}) {
		t.Errorf("an inherited pill recorded %v", parts(drawn))
	}
}

func TestEveryPillIsCentredByTheInkOfALineOfLetters(t *testing.T) {
	box := Box{X: 10, Y: 10, W: 80, H: pillHigh}
	var at []Box
	noteSaid = func(_ layout.Context, _ *material.Theme, b Box, _ string, _ int, _ font.Font, _ int, _ text.Alignment) {
		at = append(at, b)
	}
	defer func() { noteSaid = nil }()
	drawnAlone(t, 120, 40, func(k Kit) {
		Pill{Said: "active", Tone: ToneStanding, Standing: "active"}.Draw(k, box)
		Pill{Said: "+2"}.Draw(k, box)
		LabelChip{Label: "gui", Colour: "mauve"}.Draw(k, box)
	})
	want := pillLine(Theme(), box)
	for i, b := range at {
		if b.Y != want.Y || b.H != want.H {
			t.Errorf("pill %d writes at %+v, not on the line %+v", i, b, want)
		}
	}
	if len(at) != 3 {
		t.Errorf("%d pills wrote", len(at))
	}
}

func TestAKeyChipIsAlwaysGroundedLoudOrQuiet(t *testing.T) {
	s := Dark()
	row := Box{X: 10, Y: 10, H: 28}
	want := Box{X: 10, Y: 15, W: KeyChipWide("esc"), H: chipHigh}
	for _, c := range []struct {
		chip        KeyChip
		ground, ink color.NRGBA
	}{
		{KeyChip{Key: "esc"}, s.Colour.Surface.Control, s.Colour.Ink.Text},
		{KeyChip{Key: "esc", Quiet: true}, s.Colour.Surface.Control, s.Colour.Ink.Faint},
		{KeyChip{Key: "esc", OnControl: true}, s.Colour.Surface.ChipOnControl, s.Colour.Ink.Text},
		{KeyChip{Key: "esc", Quiet: true, OnControl: true}, s.Colour.Surface.ChipOnControl, s.Colour.Ink.Faint},
	} {
		var got Box
		shot, drawn := drawnAlone(t, 80, 50, func(k Kit) { got = c.chip.Draw(k, row) })
		if got != want {
			t.Errorf("%+v is drawn at %+v, not %+v", c.chip, got, want)
		}
		if pixel(shot, want.X+want.W/2, want.Y+1) != c.ground {
			t.Errorf("%+v stands on %v, not %v", c.chip, pixel(shot, want.X+want.W/2, want.Y+1), c.ground)
		}
		if lettered(shot, want, c.ink) == 0 {
			t.Errorf("%+v is not written in %v", c.chip, c.ink)
		}
		if w := []Drawing{{Part: "key", At: want, Layer: LayerBoard}}; !slices.Equal(drawn, w) {
			t.Errorf("%+v recorded %+v", c.chip, drawn)
		}
	}
	var none Box
	_, drawn := drawnAlone(t, 80, 50, func(k Kit) { none = KeyChip{}.Draw(k, row) })
	if none != (Box{X: 10, Y: 15, H: chipHigh}) || len(drawn) != 0 {
		t.Errorf("no key gives %+v and records %+v", none, drawn)
	}
}

func TestAButtonsLooksAreTokens(t *testing.T) {
	c := Dark().Colour
	for look, want := range map[buttonLook][3]color.NRGBA{
		buttonPlain:     {c.Surface.Control, c.Ink.Text, c.Surface.ChipOnControl},
		buttonLit:       {c.Surface.ControlLit, c.Ink.Text, c.Surface.ChipOnControlLit},
		buttonOn:        {c.Accent.Base, c.Ink.OnAccent, c.Accent.Chip},
		buttonOnLit:     {c.Accent.Lit, c.Ink.OnAccent, c.Accent.LitChip},
		buttonFaint:     {c.Surface.Control, c.Ink.Faint, c.Surface.ChipOnControl},
		buttonDanger:    {c.State.Danger, c.State.Wrong, c.State.DangerChip},
		buttonDangerLit: {c.State.DangerLit, c.State.Wrong, c.State.DangerLitChip},
	} {
		if back, ink, chip := Dark().button(look); [3]color.NRGBA{back, ink, chip} != want {
			t.Errorf("look %d takes %v %v %v", look, back, ink, chip)
		}
	}
}

func TestAButtonIsDrawnInItsLookWithItsKeyOnAChipInside(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 120, H: 30}
	shot, drawn := drawnAlone(t, 150, 50, func(k Kit) {
		Button{Said: "Save", Key: "⌘⏎", Look: buttonOn, Keyed: true}.Draw(k.on(LayerPanel), box)
	})
	if got := pixel(shot, box.X+s.Radius.Control, box.Y+2); got != s.Colour.Accent.Base {
		t.Errorf("a button on stands on %v", got)
	}
	if !slices.Equal(parts(drawn), []string{"button", "button/key"}) || drawn[0].At != box {
		t.Fatalf("a keyed button recorded %+v", drawn)
	}
	key := drawn[1].At
	if got := pixel(shot, key.X+key.W/2, key.Y+1); got != s.Colour.Accent.Chip {
		t.Errorf("the key of a button on stands on %v", got)
	}
	plain, drawn := drawnAlone(t, 150, 50, func(k Kit) { Button{Said: "Month", Look: buttonPlain}.Draw(k, box) })
	if !slices.Equal(parts(drawn), []string{"button"}) || pixel(plain, box.X+s.Radius.Control, box.Y+2) != s.Colour.Surface.Control {
		t.Errorf("a plain button recorded %v", parts(drawn))
	}
	if lettered(plain, ButtonSaid(box), s.Colour.Ink.Text) == 0 {
		t.Error("a plain button does not write its word")
	}
}

func TestADotIsAColourOrARingBesideAName(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: dotSide, H: dotSide}
	if dotSide != 8 {
		t.Errorf("a dot is %d wide", dotSide)
	}
	shot, drawn := drawnAlone(t, 30, 30, func(k Kit) { Dot{Colour: "mauve"}.Draw(k, box) })
	if got := pixel(shot, box.X+box.W/2, box.Y+box.H/2); got != s.Colour.Labels["mauve"] {
		t.Errorf("a dot is %v", got)
	}
	if want := []Drawing{{Part: "dot", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the dot recorded %+v, want %+v", drawn, want)
	}
	ring, _ := drawnAlone(t, 30, 30, func(k Kit) { Dot{Ring: true}.Draw(k, box) })
	if pixel(ring, box.X+box.W/2, box.Y+box.H/2).R > s.Colour.Surface.Back.R+8 || inked(ring, box, s.Colour.Surface.Back) == 0 {
		t.Error("a ring is not a faint line around nothing")
	}
}

func TestAnEmptySlotIsDashedAtACardsCornersAndSaysSo(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 200, H: emptyHigh}
	shot, drawn := drawnAlone(t, 230, 70, func(k Kit) { Empty{Said: "Nothing here"}.Draw(k, box) })
	if pixel(shot, box.X, box.Y) == s.Colour.Line.Edge {
		t.Error("the empty slot is square")
	}
	along := Box{X: box.X + s.Radius.Card + s.Space.DashGap, Y: box.Y, W: s.Size.Dash, H: 1}
	if painted(shot, along, s.Colour.Line.Edge) != s.Size.Dash {
		t.Error("the empty slot's line does not start with a dash past its corner")
	}
	if lettered(shot, box, s.Colour.Ink.Low) == 0 {
		t.Error("the empty slot says nothing")
	}
	if !slices.Equal(parts(drawn), []string{"empty", "empty/dashed"}) {
		t.Errorf("the empty slot recorded %v", parts(drawn))
	}
	square, _ := drawnAlone(t, 230, 70, func(k Kit) { Dashed{}.Draw(k, box) })
	if pixel(square, box.X, box.Y) != s.Colour.Line.Edge {
		t.Error("a dashed line of no radius is not square")
	}
}

func TestAKeyOnAPickedRowStandsOutFromTheRow(t *testing.T) {
	w := Window{screen: OnMenu, width: 1200, height: 800, menuState: menuState{menu: []string{"Edit", "Delete"}, menuAt: Point{X: 300, Y: 200}}}
	shot := drawn(t, w)
	row := MenuRowAt(w.MenuBox(), w.MenuOn())
	for _, d := range partsDrawn(w) {
		if d.Part != "row/key" || d.At.Y < row.Y || d.At.Y+d.At.H > row.Y+row.H {
			continue
		}
		chip := pixel(shot, d.At.X+d.At.W/2, d.At.Y+1)
		if got := Contrast(chip, Dark().Colour.Surface.Control); got < 1.2 {
			t.Errorf("the key on the picked row stands at %.2f from the row, and needs 1.2", got)
		}
		return
	}
	t.Fatal("the picked row draws no key")
}
