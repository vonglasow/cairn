package gui

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"

	"cairn/internal/linuxapp"
	cairnlogo "cairn/internal/logo"
)

var x11Sizes = []int{16, 32, 48, 64, 128}

// X11Icon is the logo as X11's _NET_WM_ICON holds it: at each size a bar asks for, its width, its height, then every pixel as ARGB, row by row.
func X11Icon() []uint32 {
	var held []uint32
	for _, size := range x11Sizes {
		raw, err := cairnlogo.At(size)
		if err != nil {
			panic(fmt.Sprintf("the logo at %d pixels is not embedded: %v", size, err))
		}
		drawn, err := png.Decode(bytes.NewReader(raw))
		if err != nil {
			panic(fmt.Sprintf("the logo at %d pixels does not decode: %v", size, err))
		}
		b := drawn.Bounds()
		held = append(held, uint32(b.Dx()), uint32(b.Dy()))
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := color.NRGBAModel.Convert(drawn.At(x, y)).(color.NRGBA)
				held = append(held, uint32(c.A)<<24|uint32(c.R)<<16|uint32(c.G)<<8|uint32(c.B))
			}
		}
	}
	return held
}

// Settling settles the window in on a Linux desktop under data, started with starts, and asks about an entry somebody else wrote that hides the logo, unless something else is being asked already.
func (w Window) Settling(data, starts string) Window {
	_ = linuxapp.SettleIn(data, starts)
	if hiding := linuxapp.Hiding(data); hiding.Asks() && w.screen != OnAsking {
		w.hiding, w.starts = hiding, starts
		w.screen, w.question = OnAsking, aboutHiding
	}
	return w
}

func (w Window) answeringHiding(replace bool) Window {
	w.screen, w.question = OnTheBoard, aboutDeleting
	err := w.hiding.Kept()
	if replace {
		err = w.hiding.Replaced(w.starts)
	}
	if err != nil {
		return w.failing(err)
	}
	return w
}
