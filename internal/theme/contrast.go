package theme

import (
	"fmt"
	"image/color"
	"math"

	"cairn/internal/domain"
)

// Contrast is how far two colours stand apart, by the measure a person reading a screen is held to: 21 at most, 1 when they are the same.
func Contrast(a, b color.NRGBA) float64 {
	one, two := lit(a), lit(b)
	return (max(one, two) + 0.05) / (min(one, two) + 0.05)
}

func lit(c color.NRGBA) float64 {
	return 0.2126*straight(c.R) + 0.7152*straight(c.G) + 0.0722*straight(c.B)
}

func straight(v uint8) float64 {
	f := float64(v) / 255
	if v <= 10 {
		return f / 12.92
	}
	return math.Pow((f+0.055)/1.055, 2.4)
}

// Shortfall is one colour that does not stand far enough from what it is drawn on: both named, how far it stands, and how far it has to.
type Shortfall struct {
	Said, On     string
	Ratio, Needs float64
}

func (s Shortfall) String() string {
	return fmt.Sprintf("%s on %s reads at %.2f, and needs %g", s.Said, s.On, s.Ratio, s.Needs)
}

var held = []struct {
	said, on string
	needs    float64
}{
	{"ink.text", "surface.back", 4.5},
	{"ink.text", "surface.raised", 4.5},
	{"ink.text", "surface.ground", 4.5},
	{"ink.faint", "surface.card", 4.5},
	{"ink.faint", "surface.raised", 4.5},
	{"ink.faint", "surface.back", 4.5},
	{"ink.faint", "surface.ground", 4.5},
	{"ink.low", "surface.card", 3},
	{"ink.on_tint", "surface.tint", 4.5},
	{"accent.base", "surface.back", 4.5},
	{"accent.base", "surface.card", 3},
	{"state.wrong", "surface.back", 4.5},
	{"state.late", "surface.back", 4.5},
	{"state.late", "surface.raised", 4.5},
	{"state.late", "surface.card", 3},
	{"line.edge", "surface.back", 3},
	{"line.edge", "surface.raised", 2},
	{"surface.raised", "surface.back", 1.3},
	{"surface.ground", "surface.back", 1.08},
	{"surface.raised", "surface.ground", 1.15},
	{"line.rule", "surface.ground", 1.2},
	{"line.raised", "surface.raised", 1.2},
}

const pillNeeds = 3

// Shortfalls is every pair of colours that stands closer than it is held to: a text, an edge or a surface against what it is drawn on, and a label's name on its pill.
func Shortfalls(c Colours) []Shortfall {
	var short []Shortfall
	for _, h := range held {
		if ratio := Contrast(c.token(h.said), c.token(h.on)); ratio < h.needs {
			short = append(short, Shortfall{Said: h.said, On: h.on, Ratio: ratio, Needs: h.needs})
		}
	}
	for _, name := range domain.LabelColours {
		ground, ink := c.LabelPill(name)
		if ratio := Contrast(ink, ground); ratio < pillNeeds {
			short = append(short, Shortfall{Said: labelToken + name, On: "its pill", Ratio: ratio, Needs: pillNeeds})
		}
	}
	return short
}
