package tui

import (
	"image/color"

	"cairn/internal/domain"
	"cairn/internal/objective"

	"charm.land/lipgloss/v2"
)

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
		BorderForeground(l.palette.Line.Border()).
		BorderLeftForeground(l.palette.Priority(priority).Border()).
		Width(width).
		Padding(0, 1)
	if chosen {
		return style.BorderForeground(l.palette.Accent.Border())
	}
	return style
}

// Reference is the frame of a card kept in sight only as what is in the way: dashed, and never the one chosen.
func (l Skin) Reference(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.Border{Top: "┄", Bottom: "┄", Left: "┆", Right: "┆",
			TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘"}).
		BorderForeground(l.palette.Line.Border()).
		Width(width).
		Padding(0, 1)
}

// Ink is how one thing is drawn: in one of the terminal's own accents, or in its own foreground, faint or not.
type Ink struct {
	Colour color.Color
	Faint  bool
}

// Style answers a style that draws in the ink.
func (i Ink) Style() lipgloss.Style { return i.On(lipgloss.NewStyle()) }

// On answers the style given, drawing in the ink.
func (i Ink) On(style lipgloss.Style) lipgloss.Style {
	return style.Foreground(i.Border()).Faint(i.Faint)
}

// Border answers the colour a border is drawn in with the ink, the terminal's own foreground when it has none: a border carries no faint.
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
}

var quiet = Ink{Faint: true}

// Default is the palette the board draws with: the terminal's own accents and its own foreground, so that the board wears whatever theme the terminal has.
var Default = Palette{
	Low:      Ink{Colour: lipgloss.Cyan},
	Normal:   quiet,
	High:     Ink{Colour: lipgloss.Yellow},
	Critical: Ink{Colour: lipgloss.Red},
	Blocked:  Ink{Colour: lipgloss.Red},
	Watching: Ink{Colour: lipgloss.Cyan},
	Paused:   quiet,
	Reached:  Ink{Colour: lipgloss.Green},
	Late:     Ink{Colour: lipgloss.Red},
	Faint:    quiet,
	Bright:   Ink{},
	Line:     quiet,
	Accent:   Ink{Colour: lipgloss.Blue},
	Stalled:  Ink{Colour: lipgloss.Yellow},
}

// Faded is the palette a reference is drawn with: every ink quiet, since it stands for a task kept in sight, not one of those counted.
var Faded = Palette{
	Low: quiet, Normal: quiet, High: quiet, Critical: quiet, Blocked: quiet, Watching: quiet, Paused: quiet,
	Reached: quiet, Late: quiet, Faint: quiet, Bright: quiet, Line: quiet, Accent: quiet, Stalled: quiet,
}

// LabelsInTheTerminal are the fourteen label colours, each as the nearest of the terminal's own accents.
var LabelsInTheTerminal = map[string]color.Color{
	"rosewater": lipgloss.BrightMagenta, "flamingo": lipgloss.BrightRed,
	"pink": lipgloss.Magenta, "mauve": lipgloss.BrightMagenta,
	"red": lipgloss.Red, "maroon": lipgloss.BrightRed,
	"peach": lipgloss.BrightYellow, "yellow": lipgloss.Yellow,
	"green": lipgloss.Green, "teal": lipgloss.Cyan,
	"sky": lipgloss.BrightCyan, "sapphire": lipgloss.BrightBlue,
	"blue": lipgloss.Blue, "lavender": lipgloss.BrightBlue,
}

// Priority answers the ink a priority is drawn in.
func (p Palette) Priority(named string) Ink {
	switch named {
	case string(domain.Low):
		return p.Low
	case string(domain.High):
		return p.High
	case string(domain.Critical):
		return p.Critical
	default:
		return p.Normal
	}
}

// Stopped answers the ink one of the three stopped states is drawn in.
func (p Palette) Stopped(named string) Ink {
	switch named {
	case string(domain.Blocked):
		return p.Blocked
	case string(domain.Watching):
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
