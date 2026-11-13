package gui

import (
	"fmt"
	"image"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Shot is the window drawn off screen in the skin given, as its pixels: what a picture of it shows.
func Shot(w Window, th *material.Theme, skin Skin) (*image.RGBA, error) {
	window, err := headless.NewWindow(w.width, w.height)
	if err != nil {
		return nil, fmt.Errorf("no headless window here: %w", err)
	}
	defer window.Release()
	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Exact(image.Pt(w.width, w.height)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}
	Draw(NewKit(gtx, th, skin, &Drawn{}), w)
	if err := window.Frame(gtx.Ops); err != nil {
		return nil, fmt.Errorf("drawing: %w", err)
	}
	shot := image.NewRGBA(image.Rect(0, 0, w.width, w.height))
	if err := window.Screenshot(shot); err != nil {
		return nil, fmt.Errorf("reading the pixels back: %w", err)
	}
	return shot, nil
}
