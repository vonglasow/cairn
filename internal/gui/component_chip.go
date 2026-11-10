package gui

import (
	"image/color"

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
func (c LabelChip) Draw(k Kit, at Box) Box {
	k.put("label", at)
	ground, ink := k.skin.LabelPill(c.Colour)
	rounded(k.gtx, at, k.skin.Radius.Full(at), ground)
	said := at
	if c.Removable {
		said.W -= crossWide
		Cross{Ink: ink}.Draw(k.inside("label"), Box{X: at.X + at.W - pillPad/2 - crossWide, Y: at.Y + (at.H-crossWide)/2, W: crossWide, H: crossWide})
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
func (p Pill) Draw(k Kit, at Box) Box {
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

func (p Pill) inks(s Skin) (ground, ink color.NRGBA, grounded bool) {
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

func pillLine(th *material.Theme, pill Box) Box {
	top, bottom := inkOf(th, font.Font{Typeface: th.Face}, pillReference)
	return Box{X: pill.X, Y: pill.Y + (pill.H-top-bottom)/2, W: pill.W, H: lineHigh}
}

// KeyChip is a key on a chip of its own, loud where it is offered and quiet where it only reminds, always grounded.
type KeyChip struct {
	Key              string
	Quiet, OnControl bool
}

// Draw draws the chip at the left of row, in its middle, and gives back where it stands.
func (c KeyChip) Draw(k Kit, row Box) Box {
	chip := Box{X: row.X, Y: row.Y + (row.H-chipHigh)/2, W: KeyChipWide(c.Key), H: chipHigh}
	if c.Key == "" {
		return Box{X: row.X, Y: chip.Y, H: chipHigh}
	}
	ground, ink := c.inks(k.skin)
	return c.grounded(k, chip, ground, ink)
}

func (c KeyChip) inks(s Skin) (ground, ink color.NRGBA) {
	ground, ink = s.Colour.Surface.Control, s.Colour.Ink.Text
	if c.OnControl {
		ground = s.Colour.Surface.ChipOnControl
	}
	if c.Quiet {
		ink = s.Colour.Ink.Faint
	}
	return ground, ink
}

func (c KeyChip) grounded(k Kit, chip Box, ground, ink color.NRGBA) Box {
	k.put("key", chip)
	rounded(k.gtx, chip, k.skin.Radius.Chip, ground)
	sayOnChip(k, ink, chip, c.Key)
	return chip
}

const chipHigh = 18

func sayOnChip(k Kit, ink color.NRGBA, chip Box, key string) {
	top, bottom := inkOf(k.th, k.skin.Type.MonoBold, key)
	at := chip.Y + chip.H/2 - (top+bottom)/2
	sayIn(k.gtx, k.th, ink, Box{X: chip.X, Y: at, W: chip.W, H: lineHigh}, key, small, k.skin.Type.MonoBold, text.Middle)
}

// Dot is a colour standing beside its name, or a ring where there is no colour to show.
type Dot struct {
	Colour     string
	Ring, Late bool
}

// Draw draws the dot in at.
func (d Dot) Draw(k Kit, at Box) Box {
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
