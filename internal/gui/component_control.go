package gui

import (
	"gioui.org/font"
	"gioui.org/text"
)

// Tabs are the window's tabs, each a name and what it holds, the open one on a control's ground.
type Tabs struct {
	Said, Counts []string
	At           []Box
	On           int
}

// Draw draws the tabs, each in its own box, the band they stand in being at.
func (t Tabs) Draw(k Kit, at Box) Box {
	k.put("tabs", at)
	in := k.inside("tabs")
	m := k.measured()
	for i, tab := range t.At {
		in.put("tab", tab)
		ink := k.skin.Colour.Ink.Faint
		if i == t.On {
			rounded(k.gtx, tab, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
			ink = k.skin.Colour.Ink.Text
		}
		room := TabName(tab)
		sayStyled(k.gtx, k.th, ink, room, t.Said[i], reads, font.Medium, 1, text.Start)
		counted := ends(m, t.Said[i], reads) + countGap
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: room.X + counted, Y: room.Y, W: max(0, room.W-counted), H: room.H}, t.Counts[i], small, k.skin.Type.Mono, text.Start)
	}
	return at
}

const countGap = 6

// Segmented is one choice among a few, side by side on one track, the chosen one lit inside it; a segment may count what it holds.
type Segmented struct {
	Said, Counts []string
	At           []Box
	On           int
	Small        bool
}

// Draw draws the track over at and every segment in its own box.
func (s Segmented) Draw(k Kit, at Box) Box {
	k.put("segmented", at)
	hair := k.skin.Space.Hair
	rounded(k.gtx, at, k.skin.Radius.Control+hair, k.skin.Colour.Surface.Control)
	in := k.inside("segmented")
	size := reads
	if s.Small {
		size = small
	}
	m := k.measured()
	for i, segment := range s.At {
		in.put("segment", segment)
		ink := k.skin.Colour.Ink.Faint
		if i == s.On {
			rounded(k.gtx, Box{X: segment.X + hair, Y: segment.Y + hair, W: segment.W - 2*hair, H: segment.H - 2*hair}, k.skin.Radius.Control, k.skin.Colour.Surface.ControlLit)
			ink = k.skin.Colour.Ink.Text
		}
		line := ButtonSaid(segment)
		if i >= len(s.Counts) || s.Counts[i] == "" {
			sayStyled(k.gtx, k.th, ink, line, s.Said[i], size, font.Medium, 1, text.Middle)
			continue
		}
		said, counted := ends(m, s.Said[i], size), monoWide(s.Counts[i], small)
		x := segment.X + (segment.W-said-countGap-counted)/2
		sayStyled(k.gtx, k.th, ink, Box{X: x, Y: line.Y, W: said + countGap, H: line.H}, s.Said[i], size, font.Medium, 1, text.Start)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: x + said + countGap, Y: line.Y, W: counted, H: line.H}, s.Counts[i], small, k.skin.Type.Mono, text.Start)
	}
	return at
}

// Field is the frame one types into or reads from, lit with a halo while it is written into.
type Field struct{ Lit bool }

// Draw draws the frame at at, and its halo round it when lit.
func (f Field) Draw(k Kit, at Box) Box {
	k.put("field", at)
	edge := k.skin.Colour.Line.Rule
	if f.Lit {
		edge = k.skin.Colour.Accent.Base
		halo := k.skin.Space.Halo
		rounded(k.gtx, Box{X: at.X - halo, Y: at.Y - halo, W: at.W + 2*halo, H: at.H + 2*halo}, k.skin.Radius.Control+halo, k.skin.Colour.Accent.Halo)
	}
	framed(k, at, k.skin.Radius.Control, edge, k.skin.Colour.Surface.Field)
	return at
}

// Check is the box before a step or a value: empty, done and receding, or taken and in force.
type Check struct{ Done, Taken bool }

// Draw draws the box in at.
func (c Check) Draw(k Kit, at Box) Box {
	k.put("check", at)
	radius := k.skin.Radius.Check
	switch {
	case c.Taken:
		rounded(k.gtx, at, radius, k.skin.Colour.Accent.Base)
		say(k.gtx, k.th, k.skin.Colour.Ink.OnAccent, TickAt(at), TickedMark, small)
	case c.Done:
		rounded(k.gtx, at, radius, k.skin.Colour.Ink.Faint)
		say(k.gtx, k.th, k.skin.Colour.Surface.Raised, TickAt(at), TickedMark, small)
	default:
		ringed(k, at, radius, k.skin.Colour.Line.Edge)
	}
	return at
}

// TickedMark says an objective is in, or a step is done.
const TickedMark = "✓"

// CheckAt is where a line's box stands: at its left, in its middle.
func CheckAt(line Box) Box {
	return Box{X: line.X, Y: line.Y + (line.H-checkSide)/2, W: checkSide, H: checkSide}
}

// Progress is how much of something is done: a track, what is done on it, and the count at its right when it says one.
type Progress struct {
	Done, Total int
	Said        string
}

// Draw draws the track along the middle of at.
func (p Progress) Draw(k Kit, at Box) Box {
	k.put("progress", at)
	high := k.skin.Size.Track
	room, counted := at.W, 0
	if p.Said != "" {
		counted = monoWide(p.Said, small)
		room = max(0, at.W-counted-pillGap*2)
	}
	track := Box{X: at.X, Y: at.Y + (at.H-high)/2, W: room, H: high}
	rounded(k.gtx, track, high, k.skin.Colour.Surface.Control)
	if p.Done > 0 && p.Total > 0 {
		rounded(k.gtx, Box{X: track.X, Y: track.Y, W: max(high, track.W*p.Done/p.Total), H: high}, high, k.skin.Colour.State.Done)
	}
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: at.X + at.W - counted, Y: at.Y + (at.H-lineHigh)/2, W: counted, H: lineHigh}, p.Said, small, k.skin.Type.Mono, text.End)
	return at
}

// Spent is how long a task took on a day, a bar of the accent as long as its share of the longest.
type Spent struct{}

// Draw fills at with the accent.
func (Spent) Draw(k Kit, at Box) Box {
	fill(k.gtx, k.put("spent", at), k.skin.Colour.Accent.Base)
	return at
}
