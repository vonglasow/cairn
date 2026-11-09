package gui

import (
	"image"
	"image/color"

	"gioui.org/layout"
)

func sayHolding(k Kit, ink color.NRGBA, box Box, said Written, size int) {
	paintHeld(k, box, said, size)
	say(k.gtx, k.th, ink, box, said.Value, size)
}

func paintHeld(k Kit, box Box, said Written, size int) {
	if !said.Holding() {
		return
	}
	kept := []rune(said.Value)
	from, to := said.Run()
	from, to = min(from, len(kept)), min(to, len(kept))
	loose := k.gtx
	loose.Constraints = layout.Constraints{Max: image.Pt(box.W, box.H)}
	fill(k.gtx, HeldRun(box, wide(loose, k.th, string(kept[:from]), size).Size.X,
		wide(loose, k.th, string(kept[from:to]), size).Size.X), k.skin.Colour.Surface.Ghost)
}

func writing(k Kit, m Measure, b Box, said Written, carets bool, size, rows, scrolled int) {
	_, clipped := k.within(b)
	defer clipped.Pop()

	at := said.At
	lines := wrapIn(m, said.Reads(), b, size)
	line, column := lines.At(at)
	from := StartsAt(lines, rows, line, scrolled)

	begins, ends := said.Run()
	loose := k.gtx
	loose.Constraints = layout.Constraints{Max: image.Pt(b.W, lineHigh)}
	for i := from; i < Lasting(len(lines.Said), from, rows); i++ {
		if a, z, on := HeldOn(lines, begins, ends, i); said.Holding() && on {
			kept := []rune(lines.Said[i])
			at := LineAt(b, i, from)
			left := wide(loose, k.th, string(kept[:a]), size).Size.X
			fill(k.gtx, Box{X: at.X + left, Y: at.Y, H: at.H,
				W: wide(loose, k.th, string(kept[a:z]), size).Size.X}, k.skin.Colour.Surface.Ghost)
		}
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, LineAt(b, i, from), lines.Said[i], size)
	}
	if !carets {
		return
	}
	kept := []rune(lines.Said[line])
	before := wide(loose, k.th, string(kept[:min(len(kept), column)]), size)

	fill(k.gtx, CaretAt(b, line, from, before.Size.X), k.skin.Colour.Accent.Base)
}
