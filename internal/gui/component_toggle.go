package gui

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/text"

	"cairn/internal/theme"
)

// Toggle is a switch a setting is turned on or off by: a track on the control, its knob at the left; on the accent, its knob at the right.
type Toggle struct{ On bool }

const (
	toggleWide = 34
	toggleHigh = 20
	knobInset  = 3
)

// Draw draws the track in at and its knob inside it.
func (t Toggle) Draw(k Kit, at Box) Box {
	k.put("toggle", at)
	track, knob, x := k.skin.Colour.Surface.ControlLit, k.skin.Colour.Ink.Faint, at.X+knobInset
	if t.On {
		track, knob, x = k.skin.Colour.Accent.Base, k.skin.Colour.Ink.OnAccent, at.X+at.W-knobInset-(at.H-2*knobInset)
	}
	rounded(k.gtx, at, k.skin.Radius.Full(at), track)
	side := at.H - 2*knobInset
	disc := Box{X: x, Y: at.Y + knobInset, W: side, H: side}
	rounded(k.gtx, disc, k.skin.Radius.Full(disc), knob)
	return at
}

// ToggleAt is where a row's switch stands in the box its control is given: at the left, in its middle.
func ToggleAt(control Box) Box {
	return Box{X: control.X, Y: control.Y + (control.H-toggleHigh)/2, W: toggleWide, H: toggleHigh}
}

// Picker is a value chosen from a list that opens elsewhere: on a control, a dot or a strip of a scheme's colours, the value, and a chevron.
type Picker struct {
	Said, Dot string
	Strip     []color.NRGBA
}

const (
	stripCell = 6
	stripHigh = 14
	pickerPad = 9
	chevron   = 10
)

// PickerWide is how wide a picker saying that is, with a dot or a strip before it.
func PickerWide(m Measure, said string, dot, strip bool) int {
	wide := 2*pickerPad + ends(m, said, reads) + gap + chevron
	if dot {
		wide += dotSide + gap
	}
	if strip {
		wide += 6*stripCell + gap
	}
	return wide
}

// Draw draws the picker in at.
func (p Picker) Draw(k Kit, at Box) Box {
	k.put("picker", at)
	rounded(k.gtx, at, k.skin.Radius.Control, k.skin.Colour.Surface.Control)
	in := k.inside("picker")
	x := at.X + pickerPad
	if p.Dot != "" {
		Dot{Colour: p.Dot}.Draw(in, Box{X: x, Y: at.Y + (at.H-dotSide)/2, W: dotSide, H: dotSide})
		x += dotSide + gap
	}
	if len(p.Strip) > 0 {
		Strip{Colours: p.Strip}.Draw(in, Box{X: x, Y: at.Y + (at.H-stripHigh)/2, W: len(p.Strip) * stripCell, H: stripHigh})
		x += len(p.Strip)*stripCell + gap
	}
	said := Box{X: x, Y: at.Y + (at.H-lineHigh)/2, W: max(0, at.X+at.W-pickerPad-chevron-gap-x), H: lineHigh}
	sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, said, p.Said, reads, font.Medium, 1, text.Start)
	Chevron{Ink: k.skin.Colour.Ink.Faint}.Draw(in, Box{X: at.X + at.W - pickerPad - chevron, Y: at.Y + (at.H-chevron)/2, W: chevron, H: chevron})
	return at
}

// Strip is a scheme told at a glance: a cell of each of a few of its colours, side by side.
type Strip struct{ Colours []color.NRGBA }

// Draw draws the cells across at.
func (s Strip) Draw(k Kit, at Box) Box {
	k.put("strip", at)
	rounded(k.gtx, at, k.skin.Radius.Mark, k.skin.Colour.Line.Raised)
	cell := at.W / max(1, len(s.Colours))
	for i, one := range s.Colours {
		fill(k.gtx, Box{X: at.X + i*cell, Y: at.Y + 1, W: cell, H: at.H - 2}, one)
	}
	return at
}

// StripOf is the colours a strip tells a scheme by: its ground, three of its priorities, what is done, and its accent.
func StripOf(c theme.Colours) []color.NRGBA {
	return []color.NRGBA{c.Surface.Back, c.Priority["critical"], c.Priority["high"], c.State.Done, c.Priority["normal"], c.Accent.Base}
}

// SchemeTile is a scheme offered in the chooser: a small board in its own colours, then its strip, its name and its kind; the one in force says so, the one the keyboard is on is ringed, and one that cannot be worn shows why instead of its board.
type SchemeTile struct {
	Colours          theme.Colours
	Name, Kind, Warn string
	Bad              string
	Worn, Keyed      bool
}

const (
	tilePad   = 6
	miniHigh  = 68
	tileHigh  = tilePad + miniHigh + tilePad + lineHigh + tilePad
	miniCard  = 10
	wornWide  = 52
	wornHigh  = 14
	miniInset = 3
)

// Draw draws the tile in at.
func (s SchemeTile) Draw(k Kit, at Box) Box {
	k.put("scheme", at)
	rounded(k.gtx, at, k.skin.Radius.Ground, k.skin.Colour.Surface.Card)
	in := k.inside("scheme")
	mini := Box{X: at.X + tilePad, Y: at.Y + tilePad, W: at.W - 2*tilePad, H: miniHigh}
	if s.Bad != "" {
		Dashed{Radius: k.skin.Radius.Card}.Draw(in, mini)
		sayIn(k.gtx, k.th, k.skin.Colour.State.Wrong, Inside(mini), s.Bad, small, k.skin.Type.Mono, text.Middle)
	} else {
		s.mini(k, mini)
	}
	name := Box{X: mini.X, Y: mini.Y + mini.H + tilePad, W: mini.W, H: lineHigh}
	x := name.X
	if s.Bad == "" {
		Strip{Colours: StripOf(s.Colours)}.Draw(in, Box{X: x, Y: name.Y + (name.H-stripHigh)/2, W: 6 * stripCell, H: stripHigh})
		x += 6*stripCell + between
	}
	kind, ink := s.Kind, k.skin.Colour.Ink.Low
	if s.Warn != "" {
		kind, ink = s.Warn, k.skin.Colour.State.Late
	}
	kindWide := ends(k.measured(), kind, small)
	sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, Box{X: x, Y: name.Y, W: max(0, name.X+name.W-kindWide-between-x), H: name.H}, s.Name, small, font.Medium, 1, text.Start)
	sayIn(k.gtx, k.th, ink, Box{X: name.X + name.W - kindWide, Y: name.Y, W: kindWide, H: name.H}, kind, small, font.Font{}, text.End)
	if s.Worn {
		worn := Box{X: at.X + at.W - tilePad - wornWide - inset, Y: at.Y + tilePad + inset, W: wornWide, H: wornHigh}
		rounded(k.gtx, worn, k.skin.Radius.Full(worn), k.skin.Colour.Accent.Base)
		sayStyled(k.gtx, k.th, k.skin.Colour.Ink.OnAccent, worn, "in force", small-2, font.SemiBold, 1, text.Middle)
	}
	if s.Keyed {
		Ring{Radius: k.skin.Radius.Ground}.Draw(in, at)
	}
	return at
}

func (s SchemeTile) mini(k Kit, at Box) {
	c := s.Colours
	rounded(k.gtx, at, k.skin.Radius.Card, c.Surface.Back)
	cell := (at.W - 4*miniInset) / 3
	priorities := []string{"critical", "normal", "low", "high"}
	for col := range 3 {
		ground := Box{X: at.X + miniInset + col*(cell+miniInset), Y: at.Y + miniInset, W: cell, H: at.H - 2*miniInset}
		rounded(k.gtx, ground, k.skin.Radius.Mark, c.Surface.Ground)
		for row := range 3 - col%2 {
			card := Box{X: ground.X + miniInset, Y: ground.Y + miniInset + row*(miniCard+miniInset), W: ground.W - 2*miniInset, H: miniCard}
			back := c.Surface.Card
			if col == 1 && row == 1 {
				back = c.Surface.Tint
			}
			rounded(k.gtx, card, k.skin.Radius.Stripe, back)
			fill(k.gtx, Box{X: card.X, Y: card.Y, W: 2, H: card.H}, c.Priority[priorities[(col+row)%len(priorities)]])
		}
	}
}
