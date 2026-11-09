package gui

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"
)

func drawPlace(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	if l.Place.W == 0 {
		return
	}
	rounded(gtx, l.Place, tabRadius, skin.Pill)
	dot, name, caret := PlaceParts(l.Place)
	if w.Loose() {
		drawRing(gtx, dot, skin.Faint)
		sayIn(gtx, th, skin.Ink, name, w.PlaceShown(), small, font.Font{Typeface: MonoFace}, text.Start)
	} else {
		rounded(gtx, dot, dot.W/2, skin.Labels[w.Place().Colour])
		sayStyled(gtx, th, skin.Ink, name, w.PlaceShown(), reads, font.Medium, 1, text.Start)
	}
	drawChevron(gtx, caret, skin.Faint)
}

func drawRing(gtx layout.Context, b Box, ink color.NRGBA) {
	paint.FillShape(gtx.Ops, ink, clip.Stroke{Path: clip.UniformRRect(rect(b), b.W/2).Path(gtx.Ops), Width: 1.5}.Op())
}

func drawSwitcher(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box := w.SwitcherBox()
	raised(gtx, box, panelRadius, skin.Panel)
	clipped := clip.Rect(rect(box)).Push(gtx.Ops)
	defer clipped.Pop()
	m := w.measured()
	field := w.SwitcherField()
	rounded(gtx, field, buttonRadius, skin.Chosen)
	rounded(gtx, Box{X: field.X + 1, Y: field.Y + 1, W: field.W - 2, H: field.H - 2}, buttonRadius-1, skin.Field)
	line := Inside(field)
	if asks := w.SwitcherAsks(); asks != "" {
		said := asks + "  "
		sayIn(gtx, th, skin.Faint, line, said, small, font.Font{}, text.Start)
		at := ends(m, said, small)
		line = Box{X: line.X + at, Y: line.Y, W: line.W - at, H: line.H}
	}
	if w.SwitcherTyped().Reads() == "" && w.SwitcherAsks() == "" {
		say(gtx, th, skin.Faint, line, SwitcherSays, reads)
	} else {
		writing(gtx, th, skin, m, line, w.SwitcherTyped(), true, reads, 1, Follow)
	}
	heads := []string{WorkspacesHead, RecentHead}
	for i, head := range w.SwitcherHeads() {
		sayIn(gtx, th, skin.Low, Inside(head), heads[i], small, font.Font{Typeface: MonoFace, Weight: font.SemiBold}, text.Start)
	}
	for _, one := range w.SwitcherLines() {
		drawSwitcherLine(gtx, th, skin, m, one)
	}
	acts := SwitcherActs(w)
	if len(acts) > 0 {
		fill(gtx, Box{X: box.X, Y: acts[0].At.Y - 1, W: box.W, H: 1}, skin.GroundEdge)
	}
	for _, act := range acts {
		row := Inside(act.At)
		say(gtx, th, skin.Ink, row, act.Said, reads)
		if act.Key != "" {
			drawKeyChip(gtx, th, skin, Box{X: row.X + row.W - KeyChipWide(w.keyboard.Spelled(act.Key)), Y: act.At.Y, H: act.At.H}, w.keyboard.Spelled(act.Key))
		}
	}
}

func drawSwitcherLine(gtx layout.Context, th *material.Theme, skin Skin, m Measure, one SwitcherLine) {
	if one.On {
		rounded(gtx, one.At, rowRadius, skin.Pill)
	}
	row := Inside(one.At)
	dot := Box{X: row.X, Y: row.Y + (row.H-placeDot)/2, W: placeDot, H: placeDot}
	ink := skin.Ink
	if one.Gone {
		ink = skin.Faint
	}
	right := row.X + row.W
	if one.Key != "" {
		chip := drawKeyChip(gtx, th, skin, Box{X: right - KeyChipWide(one.Key), Y: one.At.Y, H: one.At.H}, one.Key)
		right = chip.X - between
	}
	tag := ends(m, one.Tag, small)
	sayIn(gtx, th, skin.Faint, Box{X: right - tag, Y: row.Y, W: tag, H: row.H}, one.Tag, small, font.Font{}, text.End)
	right -= tag + padding
	x := dot.X + dot.W + between
	if one.Recent {
		drawRing(gtx, dot, skin.Faint)
		sayIn(gtx, th, ink, Box{X: x, Y: row.Y, W: max(0, right-x), H: row.H}, one.Name, small, font.Font{Typeface: MonoFace}, text.Start)
		return
	}
	if one.Gone {
		drawRing(gtx, dot, skin.Faint)
	} else {
		rounded(gtx, dot, dot.W/2, skin.Labels[one.Colour])
	}
	named := min(ends(m, one.Name, reads)+placeSlack, max(0, right-x))
	sayStyled(gtx, th, ink, Box{X: x, Y: row.Y, W: named, H: row.H}, one.Name, reads, font.Medium, 1, text.Start)
	folder := x + named + padding
	sayIn(gtx, th, skin.Faint, Box{X: folder, Y: row.Y, W: max(0, right-folder), H: row.H}, one.Folder, small, font.Font{Typeface: MonoFace}, text.Start)
}

func drawTimeSegments(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	segments, ok := w.TimeSegments()
	if !ok {
		return
	}
	first, last := segments[0].At, segments[len(segments)-1].At
	rounded(gtx, Box{X: first.X, Y: first.Y, W: last.X + last.W - first.X, H: first.H}, buttonRadius, skin.Pill)
	for _, segment := range segments {
		ink := skin.Faint
		if segment.On {
			rounded(gtx, Box{X: segment.At.X + segmentInset, Y: segment.At.Y + segmentInset, W: segment.At.W - 2*segmentInset, H: segment.At.H - 2*segmentInset}, buttonRadius, lighter(skin.Pill))
			ink = skin.Ink
		}
		sayStyled(gtx, th, ink, Box{X: segment.At.X, Y: segment.At.Y + (segment.At.H-lineHigh)/2, W: segment.At.W, H: lineHigh}, segment.Said, small, font.Medium, 1, text.Middle)
	}
}
