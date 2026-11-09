package gui

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
)

func drawPlace(k Kit, l Layout, w Window) {
	if l.Place.W == 0 {
		return
	}
	rounded(k.gtx, l.Place, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
	dot, name, caret := PlaceParts(l.Place)
	if w.Loose() {
		drawRing(k, dot, k.skin.Colour.Ink.Faint)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, name, w.PlaceShown(), small, k.skin.Type.Mono, text.Start)
	} else {
		rounded(k.gtx, dot, k.skin.Radius.Full(dot), k.skin.Colour.Labels[w.Place().Colour])
		sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, name, w.PlaceShown(), reads, font.Medium, 1, text.Start)
	}
	drawChevron(k, caret, k.skin.Colour.Ink.Faint)
}

func drawRing(k Kit, b Box, ink color.NRGBA) {
	paint.FillShape(k.gtx.Ops, ink, clip.Stroke{Path: clip.UniformRRect(rect(b), k.skin.Radius.Full(b)).Path(k.gtx.Ops), Width: k.skin.Size.Stroke}.Op())
}

func drawSwitcher(k Kit, w Window) {
	box := w.SwitcherBox()
	raised(k, box, k.skin.Radius.Panel, k.skin.Colour.Surface.Raised)
	clipped := clip.Rect(rect(box)).Push(k.gtx.Ops)
	defer clipped.Pop()
	m := w.measured()
	field := w.SwitcherField()
	framed(k, field, k.skin.Radius.Control, k.skin.Colour.Accent.Base, k.skin.Colour.Surface.Field)
	line := Inside(field)
	if asks := w.SwitcherAsks(); asks != "" {
		said := asks + "  "
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, said, small, font.Font{}, text.Start)
		at := ends(m, said, small)
		line = Box{X: line.X + at, Y: line.Y, W: line.W - at, H: line.H}
	}
	if w.SwitcherTyped().Reads() == "" && w.SwitcherAsks() == "" {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, SwitcherSays, reads)
	} else {
		writing(k, m, line, w.SwitcherTyped(), true, reads, 1, Follow)
	}
	heads := []string{WorkspacesHead, RecentHead}
	for i, head := range w.SwitcherHeads() {
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Inside(head), heads[i], small, k.skin.Type.MonoBold, text.Start)
	}
	for _, one := range w.SwitcherLines() {
		drawSwitcherLine(k, m, one)
	}
	acts := SwitcherActs(w)
	if len(acts) > 0 {
		fill(k.gtx, Box{X: box.X, Y: acts[0].At.Y - 1, W: box.W, H: 1}, k.skin.Colour.Line.Raised)
	}
	for _, act := range acts {
		row := Inside(act.At)
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, row, act.Said, reads)
		if act.Key != "" {
			drawKeyChip(k, Box{X: row.X + row.W - KeyChipWide(w.keyboard.Spelled(act.Key)), Y: act.At.Y, H: act.At.H}, w.keyboard.Spelled(act.Key))
		}
	}
}

func drawSwitcherLine(k Kit, m Measure, one SwitcherLine) {
	if one.On {
		rounded(k.gtx, one.At, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
	}
	row := Inside(one.At)
	dot := Box{X: row.X, Y: row.Y + (row.H-placeDot)/2, W: placeDot, H: placeDot}
	ink := k.skin.Colour.Ink.Text
	if one.Gone {
		ink = k.skin.Colour.Ink.Faint
	}
	right := row.X + row.W
	if one.Key != "" {
		chip := drawKeyChip(k, Box{X: right - KeyChipWide(one.Key), Y: one.At.Y, H: one.At.H}, one.Key)
		right = chip.X - between
	}
	tag := ends(m, one.Tag, small)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: right - tag, Y: row.Y, W: tag, H: row.H}, one.Tag, small, font.Font{}, text.End)
	right -= tag + padding
	x := dot.X + dot.W + between
	if one.Recent {
		drawRing(k, dot, k.skin.Colour.Ink.Faint)
		sayIn(k.gtx, k.th, ink, Box{X: x, Y: row.Y, W: max(0, right-x), H: row.H}, one.Name, small, k.skin.Type.Mono, text.Start)
		return
	}
	if one.Gone {
		drawRing(k, dot, k.skin.Colour.Ink.Faint)
	} else {
		rounded(k.gtx, dot, k.skin.Radius.Full(dot), k.skin.Colour.Labels[one.Colour])
	}
	named := min(ends(m, one.Name, reads)+placeSlack, max(0, right-x))
	sayStyled(k.gtx, k.th, ink, Box{X: x, Y: row.Y, W: named, H: row.H}, one.Name, reads, font.Medium, 1, text.Start)
	folder := x + named + padding
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: folder, Y: row.Y, W: max(0, right-folder), H: row.H}, one.Folder, small, k.skin.Type.Mono, text.Start)
}

func drawTimeSegments(k Kit, w Window) {
	segments, ok := w.TimeSegments()
	if !ok {
		return
	}
	first, last := segments[0].At, segments[len(segments)-1].At
	rounded(k.gtx, Box{X: first.X, Y: first.Y, W: last.X + last.W - first.X, H: first.H}, k.skin.Radius.Control, k.skin.Colour.Surface.Control)
	for _, segment := range segments {
		ink := k.skin.Colour.Ink.Faint
		if segment.On {
			rounded(k.gtx, Box{X: segment.At.X + segmentInset, Y: segment.At.Y + segmentInset, W: segment.At.W - 2*segmentInset, H: segment.At.H - 2*segmentInset}, k.skin.Radius.Control, k.skin.Colour.Surface.ControlLit)
			ink = k.skin.Colour.Ink.Text
		}
		sayStyled(k.gtx, k.th, ink, Box{X: segment.At.X, Y: segment.At.Y + (segment.At.H-lineHigh)/2, W: segment.At.W, H: lineHigh}, segment.Said, small, font.Medium, 1, text.Middle)
	}
}
