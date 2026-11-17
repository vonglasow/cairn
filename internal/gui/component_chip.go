package gui

import (
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"image/color"
	"strings"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"
)

// LabelChip is a label's name in its colour on a fifth of that colour, with a cross that takes it off when it can be.
type LabelChip struct {
	Label, Colour string
	Removable     bool
}

// Draw draws the chip in at.
func (c LabelChip) Draw(k Kit, at geom.Box) geom.Box {
	k.put("label", at)
	ground, ink := k.skin.LabelPill(c.Colour)
	rounded(k.gtx, at, k.skin.Radius.Full(at), ground)
	said := at
	if c.Removable {
		said.W -= crossWide
		Cross{Ink: ink}.Draw(k.inside("label"), geom.Box{X: at.X + at.W - pillPad/2 - crossWide, Y: at.Y + (at.H-crossWide)/2, W: crossWide, H: crossWide})
	}
	sayCentred(k.gtx, k.th, ink, pillLine(k.th, said), c.Label, small)
	return at
}

// Tone is what a pill says by its colours: nothing more than its word, a standing, a setting given here or inherited, a link.
type Tone int

// The tones a pill is drawn in.
const (
	ToneNeutral Tone = iota
	ToneStanding
	ToneHere
	ToneInherited
	ToneLink
)

// Pill is a word in a rounded ground: a state, a count, a standing, where a setting comes from, a link.
type Pill struct {
	Said      string
	Tone      Tone
	Standing  string
	OnControl bool
}

// Draw draws the pill in at, its word centred by the ink of a line of letters.
func (p Pill) Draw(k Kit, at geom.Box) geom.Box {
	k.put("pill", at)
	ground, ink, grounded := p.inks(k.skin)
	if grounded {
		rounded(k.gtx, at, k.skin.Radius.Full(at), ground)
	} else {
		Dashed{Radius: k.skin.Radius.Full(at)}.Draw(k.inside("pill"), at)
	}
	face := font.Font{}
	if p.Tone == ToneLink {
		face = k.skin.Type.Mono
	}
	sayIn(k.gtx, k.th, ink, pillLine(k.th, at), p.Said, small, face, text.Middle)
	return at
}

func (p Pill) inks(s design.Skin) (ground, ink color.NRGBA, grounded bool) {
	c := s.Colour
	switch p.Tone {
	case ToneStanding:
		if inked, ok := c.Standing[p.Standing]; ok {
			return inked.Back, inked.Ink, true
		}
	case ToneHere:
		return c.Standing["active"].Back, c.Standing["active"].Ink, true
	case ToneInherited:
		return ground, c.Ink.Faint, false
	case ToneLink:
		return c.Surface.Tint, c.Accent.Base, true
	}
	if p.OnControl {
		return c.Surface.Raised, c.Ink.Faint, true
	}
	return c.Surface.Control, c.Ink.Faint, true
}

func pillLine(th *material.Theme, pill geom.Box) geom.Box {
	top, bottom := inkOf(th, font.Font{Typeface: th.Face}, pillReference)
	return geom.Box{X: pill.X, Y: pill.Y + (pill.H-top-bottom)/2, W: pill.W, H: lineHigh}
}

// KeyChip is a key, or a sequence of keys apart by spaces, each on a cap of its own: one look wherever a key is named.
type KeyChip struct {
	Key string
}

// Draw draws the caps at the left of row, in its middle, and gives back where they stand.
func (c KeyChip) Draw(k Kit, row geom.Box) geom.Box {
	y := row.Y + (row.H-chipHigh)/2
	if c.Key == "" {
		return geom.Box{X: row.X, Y: y, H: chipHigh}
	}
	x := row.X
	for _, key := range capsOf(c.Key) {
		capped(k, geom.Box{X: x, Y: y, W: capWide(key), H: chipHigh}, key)
		x += capWide(key) + capGap
	}
	return geom.Box{X: row.X, Y: y, W: x - capGap - row.X, H: chipHigh}
}

func capped(k Kit, at geom.Box, key string) {
	k.put("key", at)
	r := k.skin.Size.Rule
	rounded(k.gtx, at, k.skin.Radius.Chip, k.skin.Colour.Line.Raised)
	rounded(k.gtx, geom.Box{X: at.X + r, Y: at.Y + r, W: at.W - 2*r, H: at.H - 3*r}, k.skin.Radius.Chip-r, k.skin.Colour.Surface.Control)
	if key == ShiftMark {
		shiftDrawn(k, at)
		return
	}
	sayOnChip(k, k.skin.Colour.Ink.Text, at, key)
}

// ShiftMark is how the shift key is written; it is drawn on a cap of its own, as an outlined arrow, since the faces cairn writes with draw it as a thin arrow lost among the arrows.
const ShiftMark = "⇧"

func capsOf(key string) []string {
	var caps []string
	for _, field := range strings.Fields(key) {
		for field != "" {
			before, after, found := strings.Cut(field, ShiftMark)
			if before != "" {
				caps = append(caps, before)
			}
			if found {
				caps = append(caps, ShiftMark)
			}
			field = after
		}
	}
	return caps
}

func shiftDrawn(k Kit, at geom.Box) {
	face := float32(at.H - 3*k.skin.Size.Rule)
	size := face * 0.6
	x := float32(at.X) + (float32(at.W)-size)/2
	y := float32(at.Y) + (face-size)/2 + float32(k.skin.Size.Rule)
	var shape clip.Path
	shape.Begin(k.gtx.Ops)
	for i, p := range [][2]float32{{0.5, 0}, {0.95, 0.48}, {0.7, 0.48}, {0.7, 1}, {0.3, 1}, {0.3, 0.48}, {0.05, 0.48}} {
		pt := f32.Pt(x+p[0]*size, y+p[1]*size)
		if i == 0 {
			shape.MoveTo(pt)
			continue
		}
		shape.LineTo(pt)
	}
	shape.Close()
	paint.FillShape(k.gtx.Ops, k.skin.Colour.Ink.Text, clip.Stroke{Path: shape.End(), Width: float32(k.skin.Size.Stroke)}.Op())
}

const chipHigh = 18

func sayOnChip(k Kit, ink color.NRGBA, chip geom.Box, key string) {
	top, bottom := inkOf(k.th, k.skin.Type.MonoBold, key)
	at := chip.Y + chip.H/2 - (top+bottom)/2
	sayIn(k.gtx, k.th, ink, geom.Box{X: chip.X, Y: at, W: chip.W, H: lineHigh}, key, small, k.skin.Type.MonoBold, text.Middle)
}

// Dot is a colour standing beside its name, or a ring where there is no colour to show.
type Dot struct {
	Colour     string
	Ring, Late bool
}

// Draw draws the dot in at.
func (d Dot) Draw(k Kit, at geom.Box) geom.Box {
	k.put("dot", at)
	if d.Ring {
		paint.FillShape(k.gtx.Ops, k.skin.Colour.Ink.Faint, clip.Stroke{Path: clip.UniformRRect(rect(at), k.skin.Radius.Full(at)).Path(k.gtx.Ops), Width: k.skin.Size.Stroke}.Op())
		return at
	}
	ink := k.skin.Colour.Labels[d.Colour]
	if d.Late {
		ink = k.skin.Colour.State.Late
	}
	rounded(k.gtx, at, k.skin.Radius.Full(at), ink)
	return at
}

const dotSide = 8
