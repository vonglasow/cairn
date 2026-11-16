package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"image"
	"image/color"
	"slices"
	"strings"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func drawnAlone(t *testing.T, wide, high int, draw func(k Kit)) (*image.RGBA, []Drawing) {
	t.Helper()
	window := offScreen(t, wide, high)
	defer window.Release()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Constraints: layout.Exact(image.Pt(wide, high)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	var drawn Drawn
	k := NewKit(gtx, drawingTheme(), design.Dark(), &drawn)
	fill(gtx, geom.Box{W: wide, H: high}, k.skin.Colour.Surface.Back)
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

func cardDrawn(t *testing.T, c Card) (*image.RGBA, []Drawing, Parts) {
	t.Helper()
	if c.Colour == nil {
		c.Colour = func(string) string { return "mauve" }
	}
	slot := geom.Box{X: 10, Y: 10, W: 280, H: geom.CardHeight}
	shot, drawn := drawnAlone(t, 300, 120, func(k Kit) { c.Draw(k, slot) })
	return shot, drawn, PartsOf(slot)
}

func TestACardStandsOnTheGroundOfItsState(t *testing.T) {
	s := design.Dark()
	face := board.Card{Number: 7, Title: "Write the scenario", State: "todo", Priority: "high"}
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
	if painted(shot, geom.Box{X: p.Panel.X + s.Radius.Card, Y: p.Panel.Y, W: p.Panel.W - 2*s.Radius.Card, H: 1}, s.Colour.Line.Edge) == 0 {
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

func TestAnObjectivesCardDrawsItsStandingAsAPillOfThatStanding(t *testing.T) {
	s := design.Dark()
	face := goalFace(board.Goal{Number: 3, Title: "Ship v0.11", State: "on-hold"})
	shot, drawn, _ := cardDrawn(t, Card{Face: face, Standing: "on-hold"})
	pill := StandingPill(face, geom.Box{X: 10, Y: 10, W: 280, H: geom.CardHeight}, Measured(drawingTheme()))
	if got := pixel(shot, pill.X+pill.W/2, pill.Y+1); got != s.Colour.Standing["on-hold"].Back {
		t.Errorf("the standing stands on %v", got)
	}
	if slices.Contains(parts(drawn), "card/label") || !slices.Contains(parts(drawn), "card/pill") {
		t.Errorf("the standing is recorded as %v, not as a pill", parts(drawn))
	}
}

func TestAPillTakesTheGroundAndInkOfItsTone(t *testing.T) {
	s := design.Dark()
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
	box := geom.Box{X: 10, Y: 10, W: 60, H: pillHigh}
	shot, drawn := drawnAlone(t, 100, 40, func(k Kit) { Pill{Said: "+2"}.Draw(k, box) })
	if got := pixel(shot, box.X+box.W/2, box.Y); got != s.Colour.Surface.Control {
		t.Errorf("a pill's edge is %v, not its ground", got)
	}
	if want := []Drawing{{Part: "pill", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the pill recorded %+v, want %+v", drawn, want)
	}
	dashed, drawn := drawnAlone(t, 100, 40, func(k Kit) { Pill{Said: "global", Tone: ToneInherited}.Draw(k, box) })
	if pixel(dashed, box.X+box.W/2, box.Y+box.H/2+3) != s.Colour.Surface.Back || painted(dashed, geom.Box{X: box.X + 9, Y: box.Y, W: box.W - 18, H: 1}, s.Colour.Line.Edge) == 0 {
		t.Error("an inherited pill is not a dashed line around nothing")
	}
	if !slices.Equal(parts(drawn), []string{"pill", "pill/dashed"}) {
		t.Errorf("an inherited pill recorded %v", parts(drawn))
	}
}

func TestAKeyCapHasOneLookWhereverItIsDrawn(t *testing.T) {
	s := design.Dark()
	cap := func(at geom.Box) geom.Box { return geom.Box{X: at.X, Y: at.Y, W: KeyChipWide("esc"), H: chipHigh} }
	for _, c := range []struct {
		name string
		draw func(k Kit) geom.Box
	}{
		{"alone", func(k Kit) geom.Box { return KeyChip{Key: "esc"}.Draw(k, geom.Box{X: 10, Y: 10, H: 28}) }},
		{"on a danger button", func(k Kit) geom.Box {
			b := geom.Box{X: 4, Y: 4, W: 150, H: 36}
			Button{Said: "Delete", Key: "esc", Look: buttonDanger, Keyed: true}.Draw(k, b)
			_, key := ButtonParts(b, k.measured(), "Delete", "esc")
			return key
		}},
	} {
		var at geom.Box
		shot, _ := drawnAlone(t, 160, 50, func(k Kit) { at = cap(c.draw(k)) })
		if got := pixel(shot, at.X+2, at.Y+at.H/2); got != s.Colour.Surface.Control {
			t.Errorf("%s: the cap stands on %v, not %v", c.name, got, s.Colour.Surface.Control)
		}
		if got := pixel(shot, at.X+at.W/2, at.Y+at.H-1); got != s.Colour.Line.Raised {
			t.Errorf("%s: the cap's lower edge is %v, not %v", c.name, got, s.Colour.Line.Raised)
		}
		if lettered(shot, at, s.Colour.Ink.Text) == 0 {
			t.Errorf("%s: the cap is not written in %v", c.name, s.Colour.Ink.Text)
		}
	}
	var none geom.Box
	_, drawn := drawnAlone(t, 80, 50, func(k Kit) { none = KeyChip{}.Draw(k, geom.Box{X: 10, Y: 10, H: 28}) })
	if none != (geom.Box{X: 10, Y: 15, H: chipHigh}) || len(drawn) != 0 {
		t.Errorf("no key gives %+v and records %+v", none, drawn)
	}
}

func TestASequenceOfKeysIsOneCapPerKey(t *testing.T) {
	one := KeyChipWide("esc")
	if got := KeyChipWide("esc esc"); got != 2*one+capGap {
		t.Errorf("esc esc is %d wide, not %d", got, 2*one+capGap)
	}
	var got geom.Box
	_, drawn := drawnAlone(t, 120, 50, func(k Kit) { got = KeyChip{Key: "esc esc"}.Draw(k, geom.Box{X: 10, Y: 10, H: 28}) })
	want := []Drawing{
		{Part: "key", At: geom.Box{X: 10, Y: 15, W: one, H: chipHigh}, Layer: LayerBoard},
		{Part: "key", At: geom.Box{X: 10 + one + capGap, Y: 15, W: one, H: chipHigh}, Layer: LayerBoard},
	}
	if !slices.Equal(drawn, want) || got != (geom.Box{X: 10, Y: 15, W: 2*one + capGap, H: chipHigh}) {
		t.Errorf("esc esc drew %+v in %+v", drawn, got)
	}
}

func TestASentenceDrawsTheKeysItNamesAsCaps(t *testing.T) {
	said := "Played on the board · `r` again · `⏎` keeps it"
	_, drawn := drawnAlone(t, 400, 40, func(k Kit) { Sentence{Said: said}.Draw(k, geom.Box{X: 4, Y: 4, W: 390, H: lineHigh}) })
	var keys []geom.Box
	for _, d := range drawn {
		if d.Part == "key" {
			keys = append(keys, d.At)
		}
	}
	if len(keys) != 2 || keys[0].W != KeyChipWide("r") || keys[1].W != KeyChipWide("⏎") || keys[1].X <= keys[0].X+keys[0].W {
		t.Errorf("%q drew its keys at %+v", said, keys)
	}
	if plain := (Sentence{Said: said}).Plain(); strings.Contains(plain, "`") || plain != "Played on the board · r again · ⏎ keeps it" {
		t.Errorf("%q reads %q", said, plain)
	}
}

func TestAButtonsLooksAreTokens(t *testing.T) {
	c := design.Dark().Colour
	for look, want := range map[buttonLook][2]color.NRGBA{
		buttonPlain:     {c.Surface.Control, c.Ink.Text},
		buttonLit:       {c.Surface.ControlLit, c.Ink.Text},
		buttonOn:        {c.Accent.Base, c.Ink.OnAccent},
		buttonOnLit:     {c.Accent.Lit, c.Ink.OnAccent},
		buttonFaint:     {c.Surface.Control, c.Ink.Faint},
		buttonDanger:    {c.State.Danger, c.State.Wrong},
		buttonDangerLit: {c.State.DangerLit, c.State.Wrong},
	} {
		if back, ink := buttonColours(design.Dark(), look); [2]color.NRGBA{back, ink} != want {
			t.Errorf("look %d takes %v %v", look, back, ink)
		}
	}
}
