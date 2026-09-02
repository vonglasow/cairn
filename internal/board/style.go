package board

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// How the board looks. Every style is built once, here, so a colour is chosen
// in one place and a change is made in one place — and so that nothing at a
// call site reaches for a colour of its own.

// Skin carries every style the board draws with.
type Skin struct {
	palette Palette

	TabOn  lipgloss.Style
	TabOff lipgloss.Style
	Mode   lipgloss.Style
	Crumb  lipgloss.Style
	Head   lipgloss.Style
	HeadOn lipgloss.Style
	Count  lipgloss.Style
	Rule   lipgloss.Style
	RuleOn lipgloss.Style
	Band   lipgloss.Style
	Key    lipgloss.Style
	Says   lipgloss.Style
	Panel  lipgloss.Style
	Title  lipgloss.Style
	Faint  lipgloss.Style
	Bright lipgloss.Style
	Late   lipgloss.Style
	Prompt lipgloss.Style
	Wrong  lipgloss.Style
	Popup  lipgloss.Style
	Picked lipgloss.Style
	Filled lipgloss.Style
	Empty  lipgloss.Style
}

// SkinOf builds the styles from a palette.
func SkinOf(p Palette) Skin {
	pill := lipgloss.NewStyle().Padding(0, 1)
	return Skin{
		palette: p,
		TabOn:   pill.Foreground(p.Ink).Background(p.Accent).Bold(true),
		TabOff:  pill.Foreground(p.Faint),
		Mode:    pill.Foreground(p.Ink).Background(p.Faint).Bold(true),
		Crumb:   lipgloss.NewStyle().Foreground(p.Accent),
		Head:    lipgloss.NewStyle().Foreground(p.Bright).Bold(true),
		HeadOn:  lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		Count:   lipgloss.NewStyle().Foreground(p.Faint),
		Rule:    lipgloss.NewStyle().Foreground(p.Line),
		RuleOn:  lipgloss.NewStyle().Foreground(p.Accent),
		Band:    lipgloss.NewStyle().Foreground(p.Stalled).Bold(true),
		Key:     lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		Says:    lipgloss.NewStyle().Foreground(p.Faint),
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Line).
			Padding(0, 1),
		Title:  lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		Faint:  lipgloss.NewStyle().Foreground(p.Faint),
		Bright: lipgloss.NewStyle().Foreground(p.Bright),
		Late:   lipgloss.NewStyle().Foreground(p.Late).Bold(true),
		Prompt: lipgloss.NewStyle().Foreground(p.Accent),
		Wrong:  lipgloss.NewStyle().Foreground(p.Late).Bold(true),
		Popup: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Accent).
			Padding(0, 2),
		Picked: lipgloss.NewStyle().Foreground(p.Ink).Background(p.Accent).Bold(true),
		Filled: lipgloss.NewStyle().Foreground(p.Accent),
		Empty:  lipgloss.NewStyle().Foreground(p.Line),
	}
}

// Card is the frame one card is drawn in: rounded, faint, and carrying its
// priority on the left edge alone.
func (l Skin) Card(width int, priority string, focused, chosen bool) lipgloss.Style {
	// Thick rather than thin: a card is the thing the eye lands on, and a
	// hairline round four sides of it disappears against everything else on
	// screen. The one in focus is double, which reads as heavier still without
	// being the same weight twice.
	border := lipgloss.ThickBorder()
	if focused {
		border = lipgloss.DoubleBorder()
	}
	style := lipgloss.NewStyle().
		Border(border).
		BorderForeground(l.palette.Line).
		BorderLeftForeground(l.palette.Priority(priority)).
		// Width counts the border and the padding, measured rather than
		// assumed: given the interior alone, a card's first row wrapped onto
		// its second and nothing lined up.
		Width(width).
		Padding(0, 1)
	if chosen {
		return style.BorderForeground(l.palette.Accent)
	}
	return style
}

// Palette is every colour the board uses.
type Palette struct {
	Low      color.Color
	Normal   color.Color
	High     color.Color
	Critical color.Color
	Blocked  color.Color
	Watching color.Color
	Paused   color.Color
	Reached  color.Color
	Late     color.Color
	Faint    color.Color
	Bright   color.Color
	Line     color.Color
	Accent   color.Color
	Stalled  color.Color
	Ink      color.Color
}

// Default is the palette the board draws with.
var Default = Palette{
	Low:      lipgloss.Color("39"),  // a cool blue
	Normal:   lipgloss.Color("245"), // grey
	High:     lipgloss.Color("214"), // amber
	Critical: lipgloss.Color("203"), // red
	Blocked:  lipgloss.Color("203"),
	Watching: lipgloss.Color("39"),
	Paused:   lipgloss.Color("245"),
	Reached:  lipgloss.Color("42"), // green
	Late:     lipgloss.Color("203"),
	Faint:    lipgloss.Color("245"),
	Bright:   lipgloss.Color("252"),
	Line:     lipgloss.Color("238"),
	Accent:   lipgloss.Color("111"), // periwinkle
	Stalled:  lipgloss.Color("214"),
	Ink:      lipgloss.Color("235"),
}

// Priority answers the colour a priority is drawn in. An unknown one is drawn
// like a normal one: a board that refused to draw a card because it did not
// recognise a word would hide the card and the word at once.
func (p Palette) Priority(named string) color.Color {
	switch named {
	case "low":
		return p.Low
	case "high":
		return p.High
	case "critical":
		return p.Critical
	default:
		return p.Normal
	}
}

// Stopped answers the colour one of the three stopped states is drawn in.
func (p Palette) Stopped(named string) color.Color {
	switch named {
	case "blocked":
		return p.Blocked
	case "watching":
		return p.Watching
	default:
		return p.Paused
	}
}

// Standing answers the colour one of an objective's four states is drawn in.
func (p Palette) Standing(named string) color.Color {
	switch named {
	case "active":
		return p.Accent
	case "reached":
		return p.Reached
	case "dropped":
		return p.Faint
	default:
		return p.Paused
	}
}

// PlainSkin is a skin with no colour in it. It exists so that a scenario can
// read what a card says rather than what it looks like: how the board is
// dressed is checked where it is decided, against a pinned drawing.
func PlainSkin() Skin { return Skin{palette: Default} }
