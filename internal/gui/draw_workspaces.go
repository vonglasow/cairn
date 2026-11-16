package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"

	"gioui.org/font"
	"gioui.org/text"
)

func drawPlace(k Kit, l geom.Layout, w Window) {
	if l.Place.W == 0 {
		return
	}
	pill := WorkspacePill{Name: w.PlaceShown(), Loose: w.Loose(), Open: w.Screen == board.OnWorkspaces}
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
		line = geom.Box{X: line.X + at, Y: line.Y, W: line.W - at, H: line.H}
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
		Rule{}.Draw(in, geom.Box{X: box.X, Y: acts[0].At.Y - 1, W: box.W})
	}
	for _, act := range acts {
		row := Row{Said: act.Said}
		row.Draw(in, act.At)
		if act.Key != "" {
			spelled := w.keyboard.Spelled(act.Key)
			inside := Inside(act.At)
			row.Chip(in, KeyChip{Key: spelled}, geom.Box{X: inside.X + inside.W - KeyChipWide(spelled), Y: act.At.Y, H: act.At.H})
		}
	}
}

func drawSwitcherLine(k Kit, m Measure, one SwitcherLine) {
	line := Row{Picked: one.On}
	line.Draw(k, one.At)
	row := Inside(one.At)
	dot := geom.Box{X: row.X, Y: row.Y + (row.H-dotSide)/2, W: dotSide, H: dotSide}
	ink := k.skin.Colour.Ink.Text
	if one.Gone {
		ink = k.skin.Colour.Ink.Faint
	}
	right := row.X + row.W
	if one.Key != "" {
		chip := line.Chip(k, KeyChip{Key: one.Key}, geom.Box{X: right - KeyChipWide(one.Key), Y: one.At.Y, H: one.At.H})
		right = chip.X - between
	}
	tag := ends(m, one.Tag, small)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, geom.Box{X: right - tag, Y: row.Y, W: tag, H: row.H}, one.Tag, small, font.Font{}, text.End)
	right -= tag + padding
	x := dot.X + dot.W + between
	Dot{Colour: one.Colour, Ring: one.Gone || one.Recent}.Draw(k, dot)
	sayStyled(k.gtx, k.th, ink, geom.Box{X: x, Y: row.Y, W: max(0, right-x), H: row.H}, one.Name, reads, font.Medium, 1, text.Start)
}

func drawNewSpace(k Kit, w Window) {
	Veil{}.Draw(k, geom.Box{W: w.width, H: w.height})
	k = k.inside("veil")
	p := w.NewSpaceLaid()
	Panel{}.Draw(k, p.Box)
	in := k.inside("panel")
	ink := in.skin.Colour.Ink
	Heading{Said: NewSpaceHead}.Draw(in, p.Head)
	sayIn(in.gtx, in.th, ink.Faint, p.Says, NewSpaceSays, small, font.Font{}, text.Start)
	for _, said := range []struct {
		at  geom.Box
		say string
	}{{p.NameSaid, NewSpaceName}, {p.FolderSaid, NewSpaceFolder}, {p.ColourSaid, NewSpaceColour}} {
		sayStyled(in.gtx, in.th, ink.Text, said.at, said.say, small, font.Medium, 1, text.Start)
	}
	m := w.measured()
	for i, at := range []geom.Box{p.Name, p.Folder} {
		field := Field{Lit: w.NewSpaceOn() == i}.Draw(in, at)
		line := geom.Box{X: field.X + pickerPad, Y: field.Y + (field.H-lineHigh)/2, W: field.W - 2*pickerPad, H: lineHigh}
		switch {
		case w.NewSpaceOn() == i && (i == 0 || w.NewSpaceTyped(1).Reads() != ""):
			writing(in.inside("field"), m, line, w.NewSpaceTyped(i), true, small, 1, Follow)
		case i == 1:
			sayIn(in.gtx, in.th, ink.Low, line, w.NewSpaceFolderShown(), small, in.skin.Type.Mono, text.Start)
		default:
			sayIn(in.gtx, in.th, ink.Text, line, w.NewSpaceTyped(0).Reads(), small, in.skin.Type.Mono, text.Start)
		}
	}
	Button{Said: NewSpaceChoose, Small: true}.Draw(in, p.Choose)
	sayIn(in.gtx, in.th, ink.Faint, p.Hint, NewSpaceHint, small, font.Font{}, text.Start)
	picked := w.NewSpaceColour()
	for _, swatch := range p.Swatches {
		Swatch{Colour: swatch.Colour, Picked: swatch.Colour == picked}.Draw(in, swatch.At)
	}
	Check{Taken: w.NewSpaceAway()}.Draw(in, CheckAt(p.Switch))
	sayIn(in.gtx, in.th, ink.Text, geom.Box{X: p.Switch.X + checkSide + between, Y: p.Switch.Y, W: p.Switch.W - checkSide - between, H: p.Switch.H}, NewSpaceSwitch, small, font.Font{}, text.Start)
	if said := w.NewSpaceRefused(); said != "" {
		sayStyled(in.gtx, in.th, in.skin.Colour.State.Wrong, p.Said, said, small, font.Normal, 2, text.Start)
	}
	Button{Said: NewSpaceCancel}.Draw(in, p.Cancel)
	Button{Said: NewSpaceCreate, Look: buttonLit}.Draw(in, p.Create)
}
