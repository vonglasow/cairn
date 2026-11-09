package gui

import (
	"image"

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

func (k Kit) within(at Box) clip.Stack { return clip.Rect(rect(at)).Push(k.gtx.Ops) }

func (k Kit) faded(by float32) paint.OpacityStack { return paint.PushOpacity(k.gtx.Ops, by) }

func (k Kit) shifted(down int) op.TransformStack {
	return op.Offset(image.Pt(0, down)).Push(k.gtx.Ops)
}

func (k Kit) put(part string, at Box) Box {
	k.drawn.Put(k.in+part, at, k.layer)
	return at
}

// Layer is one of the planes the window is drawn in, each over the one before.
type Layer int

// The layers, from the board up.
const (
	LayerBoard Layer = iota
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
