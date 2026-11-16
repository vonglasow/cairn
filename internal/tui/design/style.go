// Package design is how the terminal draws: the palette a colour scheme gives, and every style made from it.
package design

import (
	"cairn/internal/board"
	"cairn/internal/engine/item"
	"cairn/internal/theme"
	"image/color"

	"cairn/internal/engine/objective"

	"charm.land/lipgloss/v2"
)

// Skin carries every style the board draws with.
type Skin struct {
	Palette Palette

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
	Cap    lipgloss.Style
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
		Palette: p,
		TabOn:   pill.Foreground(p.Accent.Border()).Reverse(true).Bold(true),
		TabOff:  p.Faint.On(pill),
		Mode:    pill.Reverse(true).Bold(true),
		Crumb:   p.Accent.Style(),
		Head:    p.Bright.Style().Bold(true),
		HeadOn:  p.Accent.Style().Bold(true),
		Count:   p.Faint.Style(),
		Rule:    p.Line.Style(),
		RuleOn:  p.Accent.Style(),
		Band:    p.Stalled.Style().Bold(true),
		Key:     p.Accent.Style().Bold(true),
		Cap:     lipgloss.NewStyle().Background(p.Line.Border()).Foreground(p.Bright.Border()).Bold(true),
		Says:    p.Faint.Style(),
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Line.Border()).
			Padding(0, 1),
		Title:  p.Accent.Style().Bold(true),
		Faint:  p.Faint.Style(),
		Bright: p.Bright.Style(),
		Late:   p.Late.Style().Bold(true),
		Prompt: p.Accent.Style(),
		Wrong:  p.Late.Style().Bold(true),
		Popup: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Accent.Border()).
			Padding(0, 2),
		Picked: lipgloss.NewStyle().Foreground(p.Accent.Border()).Reverse(true).Bold(true),
		Filled: p.Accent.Style(),
		Empty:  p.Line.Style(),
	}
}

// Card is the frame one card is drawn in: rounded, faint, and carrying its priority on the left edge alone.
func (l Skin) Card(width int, priority string, focused, chosen bool) lipgloss.Style {
	border := lipgloss.ThickBorder()
	if focused {
		border = lipgloss.DoubleBorder()
	}
	style := lipgloss.NewStyle().
		Border(border).
		BorderForeground(l.Palette.Line.Border()).
		BorderLeftForeground(l.Palette.Priority(priority).Border()).
		Width(width).
		Padding(0, 1)
	if chosen {
		return style.BorderForeground(l.Palette.Accent.Border())
	}
	return style
}

// Reference is the frame of a card kept in sight only as what is in the way: dashed, and never the one chosen.
func (l Skin) Reference(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.Border{Top: "┄", Bottom: "┄", Left: "┆", Right: "┆",
			TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘"}).
		BorderForeground(l.Palette.Line.Border()).
		Width(width).
		Padding(0, 1)
}

// Ink is how one thing is drawn: in a colour of the scheme, or in the terminal's own foreground.
type Ink struct {
	Colour color.Color
}

// Style answers a style that draws in the ink.
func (i Ink) Style() lipgloss.Style { return i.On(lipgloss.NewStyle()) }

// On answers the style given, drawing in the ink.
func (i Ink) On(style lipgloss.Style) lipgloss.Style {
	return style.Foreground(i.Border())
}

// Border answers the colour a border is drawn in with the ink, the terminal's own foreground when it has none.
func (i Ink) Border() color.Color {
	if i.Colour == nil {
		return lipgloss.NoColor{}
	}
	return i.Colour
}

// Palette is every ink the board uses.
type Palette struct {
	Low      Ink
	Normal   Ink
	High     Ink
	Critical Ink
	Blocked  Ink
	Watching Ink
	Paused   Ink
	Reached  Ink
	Late     Ink
	Faint    Ink
	Bright   Ink
	Line     Ink
	Accent   Ink
	Stalled  Ink
	Labels   map[string]color.Color
}

// PaletteOf is the palette the board draws with in a colour scheme: the colours the window wears, so that both boards wear the same.
func PaletteOf(c theme.Colours) Palette {
	of := func(c color.NRGBA) Ink { return Ink{Colour: c} }
	labels := map[string]color.Color{}
	for name, said := range c.Labels {
		labels[name] = said
	}
	return Palette{
		Low: of(c.Priority[string(item.Low)]), Normal: of(c.Priority[string(item.Normal)]),
		High: of(c.Priority[string(item.High)]), Critical: of(c.Priority[string(item.Critical)]),
		Blocked: of(c.Priority[string(item.Critical)]), Watching: of(c.Priority[string(item.Normal)]),
		Paused: of(c.Ink.Faint), Reached: of(c.State.Done), Late: of(c.State.Late), Faint: of(c.Ink.Faint),
		Bright: of(c.Ink.Text), Line: of(c.Line.Edge), Accent: of(c.Accent.Base), Stalled: of(c.Priority[string(item.High)]),
		Labels: labels,
	}
}

// Inks are the colours of the palette, each by the name of what it draws, the labels' as label-<colour>.
func (p Palette) Inks() map[string]color.Color {
	named := map[string]color.Color{}
	for name, ink := range map[string]Ink{
		"low": p.Low, "normal": p.Normal, "high": p.High, "critical": p.Critical, "blocked": p.Blocked,
		"watching": p.Watching, "paused": p.Paused, "reached": p.Reached, "late": p.Late, "faint": p.Faint,
		"bright": p.Bright, "line": p.Line, "accent": p.Accent, "stalled": p.Stalled,
	} {
		named[name] = ink.Border()
	}
	for name, c := range p.Labels {
		named["label-"+name] = c
	}
	return named
}

// FadedOf is the palette a reference is drawn with: every ink the scheme's faint, since it stands for a task kept in sight, not one of those counted.
func FadedOf(c theme.Colours) Palette {
	q := Ink{Colour: c.Ink.Faint}
	return Palette{
		Low: q, Normal: q, High: q, Critical: q, Blocked: q, Watching: q, Paused: q,
		Reached: q, Late: q, Faint: q, Bright: q, Line: q, Accent: q, Stalled: q,
	}
}

// Priority answers the ink a priority is drawn in.
func (p Palette) Priority(named string) Ink {
	switch named {
	case string(item.Low):
		return p.Low
	case string(item.High):
		return p.High
	case string(item.Critical):
		return p.Critical
	default:
		return p.Normal
	}
}

// Stopped answers the ink one of the three stopped states is drawn in.
func (p Palette) Stopped(named string) Ink {
	switch named {
	case string(item.Blocked):
		return p.Blocked
	case string(item.Watching):
		return p.Watching
	default:
		return p.Paused
	}
}

// Standing answers the ink one of an objective's four states is drawn in.
func (p Palette) Standing(named string) Ink {
	switch named {
	case string(objective.Active):
		return p.Accent
	case string(objective.Reached):
		return p.Reached
	case string(objective.Dropped):
		return p.Faint
	default:
		return p.Paused
	}
}

// Of is the style a tone of the board is written in.
func (s Skin) Of(t board.Tone) lipgloss.Style {
	return map[board.Tone]lipgloss.Style{
		board.Normal: s.Bright, board.Faint: s.Faint, board.Late: s.Late, board.Passed: s.Band,
		board.Says: s.Says, board.Accent: s.Bright,
	}[t]
}
