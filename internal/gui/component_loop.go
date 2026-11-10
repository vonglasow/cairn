package gui

import (
	"image/color"
	"strconv"

	"gioui.org/f32"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
)

// LoopMark is the mark of what comes back, an open ring with its arrow, drawn rather than written since no face cairn carries has ↻.
type LoopMark struct{ Ink color.NRGBA }

// Draw draws the loop in at.
func (l LoopMark) Draw(k Kit, at Box) Box {
	k.put("loop", at)
	reach := float32(min(at.W, at.H))/2 - 1.5
	cx, cy := float32(at.X)+float32(at.W)/2, float32(at.Y)+float32(at.H)/2
	var ring clip.Path
	ring.Begin(k.gtx.Ops)
	ring.MoveTo(f32.Pt(cx+reach, cy))
	ring.QuadTo(f32.Pt(cx+reach, cy-reach), f32.Pt(cx, cy-reach))
	ring.QuadTo(f32.Pt(cx-reach, cy-reach), f32.Pt(cx-reach, cy))
	ring.QuadTo(f32.Pt(cx-reach, cy+reach), f32.Pt(cx, cy+reach))
	paint.FillShape(k.gtx.Ops, l.Ink, clip.Stroke{Path: ring.End(), Width: k.skin.Size.Stroke}.Op())
	tip := reach * 0.55
	var head clip.Path
	head.Begin(k.gtx.Ops)
	head.MoveTo(f32.Pt(cx+reach-tip, cy-tip*0.2))
	head.LineTo(f32.Pt(cx+reach, cy))
	head.LineTo(f32.Pt(cx+reach+tip*0.4, cy-tip))
	paint.FillShape(k.gtx.Ops, l.Ink, clip.Stroke{Path: head.End(), Width: k.skin.Size.Stroke}.Op())
	return at
}

// SeriesNumber is how a series is named, apart from a task's number: the loop, then its figures.
type SeriesNumber struct {
	Number int
	Ink    color.NRGBA
	Size   int
}

const loopSide = markSide - 2

// SeriesNumberWide is how wide a series' number is written at a size: the loop, a hair, and its figures in Go Mono.
func SeriesNumberWide(n, size int) int { return loopSide + 1 + monoWide(strconv.Itoa(n), size) }

// Draw draws the number in at, the loop at its left.
func (s SeriesNumber) Draw(k Kit, at Box) Box {
	k.put("series", at)
	size := s.Size
	if size == 0 {
		size = small
	}
	LoopMark{Ink: s.Ink}.Draw(k.inside("series"), Box{X: at.X, Y: at.Y + (at.H-loopSide)/2, W: loopSide, H: loopSide})
	sayIn(k.gtx, k.th, s.Ink, Box{X: at.X + loopSide + 1, Y: at.Y, W: max(0, at.W-loopSide-1), H: at.H}, strconv.Itoa(s.Number), size, k.skin.Type.Mono, text.Start)
	return at
}
