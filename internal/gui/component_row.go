package gui

import (
	"gioui.org/font"
	"gioui.org/text"
)

// Row is one line of a list: its words in the reading ink, or in the wrong's when it destroys, on a control's ground when picked and a tint's when taken.
type Row struct {
	Said                  string
	Picked, Taken, Danger bool
	Hover                 bool
}

// Draw draws the row's ground over at and writes its words inside it.
func (r Row) Draw(k Kit, at Box) Box {
	k.put("row", at)
	if r.Hover {
		rounded(k.gtx, at, k.skin.Radius.Row, k.skin.Colour.Surface.Hover)
	}
	if r.Taken {
		rounded(k.gtx, at, k.skin.Radius.Row, k.skin.Colour.Surface.Tint)
	}
	if r.Picked {
		rounded(k.gtx, at, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
	}
	ink := k.skin.Colour.Ink.Text
	if r.Danger {
		ink = k.skin.Colour.State.Wrong
	}
	say(k.gtx, k.th, ink, Inside(at), r.Said, reads)
	return at
}

// Chip draws a key on the row, on the chip a picked row's ground calls for.
func (r Row) Chip(k Kit, c KeyChip, at Box) Box {
	c.OnControl = r.Picked
	return c.Draw(k.inside("row"), at)
}

// Check draws the box before the row's value, taken or not.
func (r Row) Check(k Kit, c Check, at Box) Box {
	return c.Draw(k.inside("row"), at)
}

// Connector draws the line joining the row to the one it sits under, in the room left of its box.
func (r Row) Connector(k Kit, c Connector, at Box) Box {
	return c.Draw(k.inside("row"), at)
}

// Pill draws a pill on the row, on the ground a picked row's ground calls for.
func (r Row) Pill(k Kit, p Pill, at Box) Box {
	p.OnControl = r.Picked
	return p.Draw(k.inside("row"), at)
}

// Head is the small line naming what follows — a column, a group, a part of a form — with its count at the right when it has one.
type Head struct{ Said, Count string }

// Draw writes the head in at.
func (h Head) Draw(k Kit, at Box) Box {
	k.put("head", at)
	sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Faint, at, h.Said, small, font.Medium, 1, text.Start)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, at, h.Count, small, k.skin.Type.Mono, text.End)
	return at
}

// Tile is one line of a list inside a field, a step or a task serving an objective, ghosted while it is carried.
type Tile struct{ Picked bool }

// Draw lays the tile's ground over at.
func (t Tile) Draw(k Kit, at Box) Box {
	k.put("tile", at)
	ground := k.skin.Colour.Surface.Raised
	if t.Picked {
		ground = k.skin.Colour.Surface.Ghost
	}
	rounded(k.gtx, at, k.skin.Radius.Row, ground)
	return at
}

// Struck is the line through the words of what is done.
type Struck struct{}

// Draw draws the line along at.
func (Struck) Draw(k Kit, at Box) Box {
	fill(k.gtx, k.put("struck", at), k.skin.Colour.Ink.Faint)
	return at
}

// Landing is the line where what is carried would land.
type Landing struct{}

// Draw draws the line along at.
func (Landing) Draw(k Kit, at Box) Box {
	fill(k.gtx, k.put("landing", at), k.skin.Colour.Accent.Base)
	return at
}

// Bar is the command bar's ground: raised, edged with a rule, at a ground's corners.
type Bar struct{}

// Draw lays the bar over at.
func (Bar) Draw(k Kit, at Box) Box {
	k.put("bar", at)
	framed(k, at, k.skin.Radius.Ground, k.skin.Colour.Line.Rule, k.skin.Colour.Surface.Raised)
	return at
}

// Hint is one offer of the command bar: its key on a chip and its word, faint under a panel, lit under the pointer.
type Hint struct {
	Key, Word  string
	Faint, Lit bool
}

// Draw draws the hint in at.
func (h Hint) Draw(k Kit, at Box) Box {
	k.put("hint", at)
	if h.Lit {
		rounded(k.gtx, Box{X: at.X, Y: at.Y + 3, W: at.W - hintSpace, H: at.H - 6}, k.skin.Radius.Row, k.skin.Colour.Surface.Hover)
	}
	chip := KeyChip{Key: h.Key}.Draw(k.inside("hint"), Box{X: at.X + hintSpace/2, Y: at.Y, H: at.H})
	ink := k.skin.Colour.Ink.Text
	if h.Faint {
		ink = k.skin.Colour.Ink.Faint
	}
	from := chip.X + chip.W + hintSpace
	sayIn(k.gtx, k.th, ink, Box{X: from, Y: at.Y + (at.H-lineHigh)/2, W: at.X + at.W - from, H: lineHigh}, h.Word, reads, font.Font{}, text.Start)
	return at
}
