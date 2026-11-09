package gui

import (
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// Empty is the slot where a column or a lane holding nothing says so: a dashed line at a card's corners, and the words.
type Empty struct{ Said string }

// Draw draws the slot in at and says what it says in its middle.
func (e Empty) Draw(k Kit, at Box) Box {
	k.put("empty", at)
	Dashed{Radius: k.skin.Radius.Card}.Draw(k.inside("empty"), at)
	sayCentred(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: at.X, Y: at.Y + (at.H-lineHigh)/2, W: at.W, H: lineHigh}, e.Said, small)
	return at
}

// Dashed is a dashed line round a box, its corners rounded at the box's own radius.
type Dashed struct{ Radius int }

// Draw draws the dashed line round at.
func (d Dashed) Draw(k Kit, at Box) Box {
	k.put("dashed", at)
	gtx, ink := k.gtx, k.skin.Colour.Line.Edge
	dash, gap, line := k.skin.Size.Dash, k.skin.Space.DashGap, k.skin.Size.Rule
	r := min(d.Radius, at.W/2, at.H/2)
	from := 0
	if r > 0 {
		from = r + gap
		wide, tall := r, r
		if at.W-2*r <= gap {
			wide = at.W - r
		}
		if at.H-2*r <= gap {
			tall = at.H - r
		}
		right, foot := at.X+at.W-r, at.Y+at.H-r
		for _, corner := range []Box{{X: at.X, Y: at.Y, W: wide, H: tall}, {X: right, Y: at.Y, W: r, H: tall}, {X: at.X, Y: foot, W: wide, H: r}, {X: right, Y: foot, W: r, H: r}} {
			cut := clip.Rect(rect(corner)).Push(gtx.Ops)
			paint.FillShape(gtx.Ops, ink, clip.Stroke{Path: clip.UniformRRect(rect(at), r).Path(gtx.Ops), Width: float32(line)}.Op())
			cut.Pop()
		}
	}
	for x := at.X + from; x < at.X+at.W-r; x += dash + gap {
		long := min(dash, at.X+at.W-r-x)
		fill(gtx, Box{X: x, Y: at.Y, W: long, H: line}, ink)
		fill(gtx, Box{X: x, Y: at.Y + at.H - line, W: long, H: line}, ink)
	}
	for y := at.Y + from; y < at.Y+at.H-r; y += dash + gap {
		long := min(dash, at.Y+at.H-r-y)
		fill(gtx, Box{X: at.X, Y: y, W: line, H: long}, ink)
		fill(gtx, Box{X: at.X + at.W - line, Y: y, W: line, H: long}, ink)
	}
	return at
}
