package gui

import (
	"gioui.org/font"
	"gioui.org/text"
)

// Ring is a line round what the keyboard is on, or, a target, round what is written into or where a drop will land; always at the element's own radius.
type Ring struct {
	Target bool
	Radius int
}

// Draw draws the ring round at, leaving what it rings as it was.
func (r Ring) Draw(k Kit, at Box) Box {
	k.put("ring", at)
	ink := k.skin.Colour.Line.Selected
	if r.Target {
		ink = k.skin.Colour.Accent.Base
	}
	ringed(k, at, r.Radius, ink)
	return at
}

// TodayMark is how today is told apart wherever days are laid out: a ring in the accent, a fill meaning a day picked.
type TodayMark struct{ Radius int }

// Draw draws the ring round at.
func (t TodayMark) Draw(k Kit, at Box) Box {
	k.put("today", at)
	ringed(k, at, t.Radius, k.skin.Colour.Accent.Base)
	return at
}

// DayCell is a day of a month: picked and filled with the accent, today and ringed, of another month and faint, on a raised ground in the month's grid.
type DayCell struct {
	Said                         string
	Today, Picked, Faded, Raised bool
}

// Draw draws the day in at and writes its number in the middle.
func (d DayCell) Draw(k Kit, at Box) Box {
	k.put("day", at)
	if d.Picked {
		rounded(k.gtx, at, k.skin.Radius.Control, k.skin.Colour.Accent.Base)
		sayStyled(k.gtx, k.th, k.skin.Colour.Ink.OnAccent, ButtonSaid(at), d.Said, reads, font.Medium, 1, text.Middle)
		return at
	}
	if d.Raised {
		rounded(k.gtx, at, k.skin.Radius.Control, k.skin.Colour.Surface.Raised)
	}
	if d.Today {
		TodayMark{Radius: k.skin.Radius.Control}.Draw(k.inside("day"), at)
	}
	ink := k.skin.Colour.Ink.Text
	if d.Faded {
		ink = k.skin.Colour.Ink.Faint
	}
	sayCentred(k.gtx, k.th, ink, ButtonSaid(at), d.Said, reads)
	return at
}

// Swatch is a colour to pick, ringed in the reading ink when it is the one picked.
type Swatch struct {
	Colour string
	Picked bool
}

// Draw draws the swatch in at, its ring round it.
func (s Swatch) Draw(k Kit, at Box) Box {
	k.put("swatch", at)
	if s.Picked {
		hair := k.skin.Space.Hair
		ring := Box{X: at.X - hair, Y: at.Y - hair, W: at.W + 2*hair, H: at.H + 2*hair}
		rounded(k.gtx, ring, k.skin.Radius.Full(ring), k.skin.Colour.Ink.Text)
	}
	rounded(k.gtx, at, k.skin.Radius.Full(at), k.skin.Colour.Labels[s.Colour])
	return at
}

// WorkspacePill is the first thing of the top line: the workspace's dot, its name and a chevron on a control, ringed while its switcher is open; a loose store has a ring for a dot and its folder in mono.
type WorkspacePill struct {
	Name, Colour string
	Loose, Open  bool
}

// Draw draws the pill in at.
func (p WorkspacePill) Draw(k Kit, at Box) Box {
	k.put("workspace", at)
	rounded(k.gtx, at, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
	in := k.inside("workspace")
	dot, name, caret := PlaceParts(at)
	if p.Loose {
		Dot{Ring: true}.Draw(in, dot)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, name, p.Name, small, k.skin.Type.Mono, text.Start)
	} else {
		Dot{Colour: p.Colour}.Draw(in, dot)
		sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, name, p.Name, reads, font.Medium, 1, text.Start)
	}
	Chevron{Ink: k.skin.Colour.Ink.Faint}.Draw(in, caret)
	if p.Open {
		Ring{Target: true, Radius: k.skin.Radius.Row}.Draw(in, at)
	}
	return at
}

// Scrollbar is the track beside what scrolls and the thumb showing which of its lines are in sight.
type Scrollbar struct{ Lines, Rows, At int }

// Draw draws the track along at and the thumb on it, and nothing when every line is in sight.
func (s Scrollbar) Draw(k Kit, at Box) Box {
	if s.Lines <= s.Rows {
		return at
	}
	k.put("scrollbar", at)
	fill(k.gtx, at, k.skin.Colour.Surface.Back)
	fill(k.gtx, Thumb(at, s.Lines, s.Rows, s.At), k.skin.Colour.Line.Edge)
	return at
}

// Backdrop is the window's own back, under everything else and under what is written over the bar.
type Backdrop struct{}

// Draw lays the back over at.
func (Backdrop) Draw(k Kit, at Box) Box {
	fill(k.gtx, k.put("backdrop", at), k.skin.Colour.Surface.Back)
	return at
}

// InfoBox is a box of facts inside a panel or a detail: a card's ground at a card's corners.
type InfoBox struct{}

// Draw lays the box over at.
func (InfoBox) Draw(k Kit, at Box) Box {
	k.put("info", at)
	rounded(k.gtx, at, k.skin.Radius.Card, k.skin.Colour.Surface.Card)
	return at
}
