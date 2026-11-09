package gui

import (
	"image"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
)

// Kit is what every drawing is handed: Gio's context, the theme, the skin, the layer it draws on, and the record of what was drawn.
type Kit struct {
	gtx   layout.Context
	th    *material.Theme
	skin  Skin
	layer Layer
	drawn *Drawn
	in    string
	ruler Measure
	shown *Box
}

// NewKit is a kit drawing on the board in that skin, recording into drawn.
func NewKit(gtx layout.Context, th *material.Theme, skin Skin, drawn *Drawn) Kit {
	return Kit{gtx: gtx, th: th, skin: skin, layer: LayerBoard, drawn: drawn}
}

func (k Kit) on(layer Layer) Kit {
	k.layer = layer
	return k
}

func (k Kit) inside(part string) Kit {
	k.in += part + "/"
	return k
}

func (k Kit) measuring(m Measure) Kit {
	k.ruler = m
	return k
}

func (k Kit) measured() Measure {
	if k.ruler == nil {
		return Measured(k.th)
	}
	return k.ruler
}

func (k Kit) within(at Box) (Kit, clip.Stack) {
	if k.shown != nil {
		at = common(*k.shown, at)
	}
	k.shown = &at
	return k, clip.Rect(rect(at)).Push(k.gtx.Ops)
}

func (k Kit) faded(by float32) paint.OpacityStack { return paint.PushOpacity(k.gtx.Ops, by) }

func (k Kit) shifted(down int) op.TransformStack {
	return op.Offset(image.Pt(0, down)).Push(k.gtx.Ops)
}

func (k Kit) put(part string, at Box) Box {
	seen := at
	if k.shown != nil {
		seen = common(*k.shown, at)
	}
	k.drawn.Put(k.in+part, seen, k.layer)
	return at
}

// Layer is one of the planes the window is drawn in, each over the one before.
type Layer int

// The layers, from the window's own back up.
const (
	LayerBack Layer = iota - 1
	LayerBoard
	LayerBar
	LayerDrag
	LayerPanel
	LayerPopover
)

// Drawing is one part drawn: what it is, its box, and the layer it is on.
type Drawing struct {
	Part  string
	At    Box
	Layer Layer
}

// Drawn is the record of every part a frame drew.
type Drawn struct{ all []Drawing }

// Put records a part drawn in that box on that layer.
func (d *Drawn) Put(part string, at Box, on Layer) {
	d.all = append(d.all, Drawing{Part: part, At: at, Layer: on})
}

// All is every part recorded, in the order it was drawn.
func (d *Drawn) All() []Drawing { return d.all }

// Overlaps is every pair of parts on one layer, under one parent, whose boxes cross. A part's parent is the last part drawn before it under the name it is inside, and is never compared with it; a part drawn over the whole of its parent — a ring, a dashed line — is its frame, and is compared with nothing.
func (d *Drawn) Overlaps() [][2]Drawing {
	parents := make([]int, len(d.all))
	for i, one := range d.all {
		parents[i] = -1
		for j := i - 1; j >= 0; j-- {
			if d.all[j].Part == parentOf(one.Part) {
				parents[i] = j
				break
			}
		}
	}
	framing := func(i int) bool { return parents[i] >= 0 && d.all[parents[i]].At == d.all[i].At }
	var crossed [][2]Drawing
	for i, a := range d.all {
		for j := i + 1; j < len(d.all); j++ {
			b := d.all[j]
			if a.Layer == b.Layer && parents[i] == parents[j] && parentOf(a.Part) == parentOf(b.Part) && !framing(i) && !framing(j) && cross(a.At, b.At) {
				crossed = append(crossed, [2]Drawing{a, b})
			}
		}
	}
	return crossed
}

func parentOf(part string) string {
	at := strings.LastIndex(part, "/")
	if at < 0 {
		return ""
	}
	return part[:at]
}

func cross(a, b Box) bool {
	seen := common(a, b)
	return seen.W > 0 && seen.H > 0
}

func common(a, b Box) Box {
	x, y := max(a.X, b.X), max(a.Y, b.Y)
	return Box{X: x, Y: y, W: max(0, min(a.X+a.W, b.X+b.W)-x), H: max(0, min(a.Y+a.H, b.Y+b.H)-y)}
}
