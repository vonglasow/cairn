package gui

import (
	"gioui.org/font"
	"gioui.org/text"
)

func drawPlace(k Kit, l Layout, w Window) {
	if l.Place.W == 0 {
		return
	}
	pill := WorkspacePill{Name: w.PlaceShown(), Loose: w.Loose(), Open: w.Screen() == OnWorkspaces}
	if !pill.Loose {
		pill.Colour = w.Place().Colour
	}
	pill.Draw(k, l.Place)
}

func drawSwitcher(k Kit, w Window) {
	box := w.SwitcherBox()
	Popover{}.Draw(k, box)
	in, clipped := k.inside("popover").within(box)
	defer clipped.Pop()
	m := w.measured()
	field := Field{Lit: true}.Draw(in, w.SwitcherField())
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
		writing(in, m, line, w.SwitcherTyped(), true, reads, 1, Follow)
	}
	heads := []string{WorkspacesHead, RecentHead}
	for i, head := range w.SwitcherHeads() {
		Head{Said: heads[i]}.Draw(in, Inside(head))
	}
	for _, one := range w.SwitcherLines() {
		drawSwitcherLine(in, m, one)
	}
	acts := SwitcherActs(w)
	if len(acts) > 0 {
		Rule{}.Draw(in, Box{X: box.X, Y: acts[0].At.Y - 1, W: box.W})
	}
	for _, act := range acts {
		row := Row{Said: act.Said}
		row.Draw(in, act.At)
		if act.Key != "" {
			spelled := w.keyboard.Spelled(act.Key)
			inside := Inside(act.At)
			row.Chip(in, KeyChip{Key: spelled}, Box{X: inside.X + inside.W - KeyChipWide(spelled), Y: act.At.Y, H: act.At.H})
		}
	}
}

func drawSwitcherLine(k Kit, m Measure, one SwitcherLine) {
	line := Row{Picked: one.On}
	line.Draw(k, one.At)
	row := Inside(one.At)
	dot := Box{X: row.X, Y: row.Y + (row.H-dotSide)/2, W: dotSide, H: dotSide}
	ink := k.skin.Colour.Ink.Text
	if one.Gone {
		ink = k.skin.Colour.Ink.Faint
	}
	right := row.X + row.W
	if one.Key != "" {
		chip := line.Chip(k, KeyChip{Key: one.Key}, Box{X: right - KeyChipWide(one.Key), Y: one.At.Y, H: one.At.H})
		right = chip.X - between
	}
	tag := ends(m, one.Tag, small)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: right - tag, Y: row.Y, W: tag, H: row.H}, one.Tag, small, font.Font{}, text.End)
	right -= tag + padding
	x := dot.X + dot.W + between
	if one.Recent {
		Dot{Ring: true}.Draw(k, dot)
		sayIn(k.gtx, k.th, ink, Box{X: x, Y: row.Y, W: max(0, right-x), H: row.H}, one.Name, small, k.skin.Type.Mono, text.Start)
		return
	}
	Dot{Colour: one.Colour, Ring: one.Gone}.Draw(k, dot)
	named := min(ends(m, one.Name, reads)+placeSlack, max(0, right-x))
	sayStyled(k.gtx, k.th, ink, Box{X: x, Y: row.Y, W: named, H: row.H}, one.Name, reads, font.Medium, 1, text.Start)
	folder := x + named + padding
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: folder, Y: row.Y, W: max(0, right-folder), H: row.H}, one.Folder, small, k.skin.Type.Mono, text.Start)
}
