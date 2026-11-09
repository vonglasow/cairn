package gui

import (
	"image/color"

	"gioui.org/font"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/theme"
)

// Skin is every token the window draws with, grouped the way a design system groups them, so a second theme is a second Skin and not a search through the drawing.
type Skin struct {
	Colour theme.Colours
	Space  Spaces
	Radius Radii
	Size   Extents
	Type   Types
	Shadow []Shadow
}

// Shadow is one layer of what something raised casts: how far down, how far out, and in what.
type Shadow struct {
	Down, Spread int
	Ink          color.NRGBA
}

// Spaces is every gap the drawing reads that no layout lays out.
type Spaces struct{ Hair, Halo, DashGap int }

// Radii is every corner the window rounds.
type Radii struct{ Mark, Check, Stripe, Chip, Today, Row, Control, Card, Ground, Panel int }

// Full is the radius that rounds a box's ends into half circles.
func (Radii) Full(b Box) int { return b.H / 2 }

// Extents is every size the drawing reads that no layout lays out.
type Extents struct {
	Rule, Track, Dash int
	Stroke            float32
}

// Types is every face and size a text is drawn in that the layout does not measure with.
type Types struct {
	Mono, MonoBold font.Font
	Figure         int
}

// Skinned is the skin wearing the colours given, every other token as Dark has it.
func Skinned(colours theme.Colours) Skin {
	s := Dark()
	s.Colour = colours
	return s
}

// Dark is the window as cairn has always drawn it, the colours of the default scheme written out token by token.
func Dark() Skin {
	return Skin{
		Colour: theme.Colours{
			Surface: theme.Surfaces{
				Back:             color.NRGBA{R: 12, G: 13, B: 15, A: 255},
				Ground:           color.NRGBA{R: 23, G: 24, B: 27, A: 255},
				Card:             color.NRGBA{R: 31, G: 32, B: 36, A: 255},
				Hover:            color.NRGBA{R: 37, G: 38, B: 43, A: 255},
				Tint:             color.NRGBA{R: 30, G: 36, B: 60, A: 255},
				Raised:           color.NRGBA{R: 38, G: 40, B: 45, A: 255},
				Field:            color.NRGBA{R: 19, G: 20, B: 23, A: 255},
				Control:          color.NRGBA{R: 48, G: 50, B: 56, A: 255},
				ControlLit:       color.NRGBA{R: 66, G: 68, B: 74, A: 255},
				ChipOnControl:    color.NRGBA{R: 26, G: 28, B: 34, A: 255},
				ChipOnControlLit: color.NRGBA{R: 44, G: 46, B: 52, A: 255},
				Ghost:            color.NRGBA{R: 42, G: 51, B: 80, A: 255},
				Veil:             color.NRGBA{R: 6, G: 7, B: 9, A: 158},
				VeilLight:        color.NRGBA{R: 6, G: 7, B: 9, A: 79},
			},
			Line: theme.Strokes{
				Rule:     color.NRGBA{R: 40, G: 41, B: 46, A: 255},
				Raised:   color.NRGBA{R: 52, G: 54, B: 60, A: 255},
				Edge:     color.NRGBA{R: 96, G: 98, B: 106, A: 255},
				Selected: color.NRGBA{R: 141, G: 143, B: 156, A: 255},
			},
			Ink: theme.Inks{
				Text:     color.NRGBA{R: 231, G: 231, B: 234, A: 255},
				Faint:    color.NRGBA{R: 154, G: 155, B: 162, A: 255},
				Low:      color.NRGBA{R: 105, G: 106, B: 113, A: 255},
				OnTint:   color.NRGBA{R: 195, G: 206, B: 255, A: 255},
				OnAccent: color.NRGBA{R: 12, G: 13, B: 15, A: 255},
			},
			Accent: theme.Accents{
				Base:    color.NRGBA{R: 123, G: 147, B: 255, A: 255},
				Lit:     color.NRGBA{R: 141, G: 165, B: 255, A: 255},
				Halo:    color.NRGBA{R: 123, G: 147, B: 255, A: 38},
				Glow:    color.NRGBA{R: 123, G: 147, B: 255, A: 76},
				Chip:    color.NRGBA{R: 101, G: 125, B: 233, A: 255},
				LitChip: color.NRGBA{R: 119, G: 143, B: 233, A: 255},
			},
			State: theme.States{
				Late:          color.NRGBA{R: 240, G: 186, B: 96, A: 255},
				Done:          color.NRGBA{R: 127, G: 182, B: 159, A: 255},
				Wrong:         color.NRGBA{R: 255, G: 155, B: 152, A: 255},
				Danger:        color.NRGBA{R: 58, G: 31, B: 34, A: 255},
				DangerLit:     color.NRGBA{R: 76, G: 49, B: 52, A: 255},
				DangerChip:    color.NRGBA{R: 36, G: 9, B: 12, A: 255},
				DangerLitChip: color.NRGBA{R: 54, G: 27, B: 30, A: 255},
			},
			Priority: map[string]color.NRGBA{
				string(domain.Critical): {R: 229, G: 96, B: 94, A: 255},
				string(domain.High):     {R: 229, G: 162, B: 74, A: 255},
				string(domain.Normal):   {R: 79, G: 174, B: 140, A: 255},
				string(domain.Low):      {R: 125, G: 130, B: 150, A: 255},
			},
			Standing: map[string]theme.Inked{
				"active":  {Back: color.NRGBA{R: 29, G: 51, B: 40, A: 255}, Ink: color.NRGBA{R: 143, G: 211, B: 180, A: 255}},
				"on-hold": {Back: color.NRGBA{R: 51, G: 42, B: 26, A: 255}, Ink: color.NRGBA{R: 232, G: 191, B: 122, A: 255}},
				"reached": {Back: color.NRGBA{R: 30, G: 40, B: 62, A: 255}, Ink: color.NRGBA{R: 160, G: 180, B: 255, A: 255}},
				"dropped": {Back: color.NRGBA{R: 48, G: 50, B: 56, A: 255}, Ink: color.NRGBA{R: 154, G: 155, B: 162, A: 255}},
			},
			Labels: labelColours,
			Stones: []theme.Shade{
				{Light: color.NRGBA{R: 104, G: 102, B: 110, A: 255}, Dark: color.NRGBA{R: 64, G: 62, B: 70, A: 255}},
				{Light: color.NRGBA{R: 122, G: 119, B: 126, A: 255}, Dark: color.NRGBA{R: 82, G: 79, B: 86, A: 255}},
				{Light: color.NRGBA{R: 140, G: 137, B: 142, A: 255}, Dark: color.NRGBA{R: 100, G: 97, B: 102, A: 255}},
			},
		},
		Space:  Spaces{Hair: 2, Halo: 3, DashGap: 3},
		Radius: Radii{Mark: 2, Check: 3, Stripe: 3, Chip: 4, Today: 5, Row: 6, Control: 7, Card: 8, Ground: 10, Panel: 12},
		Size:   Extents{Rule: 1, Track: 3, Dash: 4, Stroke: 1.5},
		Type: Types{
			Mono:     font.Font{Typeface: MonoFace},
			MonoBold: font.Font{Typeface: MonoFace, Weight: font.SemiBold},
			Figure:   HeadingSize + 4,
		},
		Shadow: []Shadow{
			{Down: 18, Spread: 10, Ink: color.NRGBA{A: 18}},
			{Down: 10, Spread: 4, Ink: color.NRGBA{A: 35}},
			{Down: 4, Spread: 0, Ink: color.NRGBA{A: 60}},
		},
	}
}

var labelColours = func() map[string]color.NRGBA {
	inked := map[string]color.NRGBA{}
	for name, ink := range domain.LabelInks {
		inked[name] = color.NRGBA{R: ink[0], G: ink[1], B: ink[2], A: 255}
	}
	return inked
}()

// LabelPill is the ground a label of that colour is drawn on and the ink its name is written in, as the skin's colours say.
func (s Skin) LabelPill(colour string) (ground, ink color.NRGBA) { return s.Colour.LabelPill(colour) }

// Writing is the ink a card's title is written in.
func (s Skin) Writing(c engine.Card) color.NRGBA {
	if c.Archived {
		return s.Colour.Ink.Faint
	}
	return s.Colour.Ink.Text
}

func (s Skin) stoneOf(i int) theme.Shade { return s.Colour.Stones[i%len(s.Colour.Stones)] }
