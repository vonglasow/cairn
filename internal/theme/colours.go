// Package theme reads a colour scheme, a base16 file, and turns it into every colour the window draws with.
package theme

import (
	"fmt"
	"image/color"
	"reflect"
	"strings"
	"unicode"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

// Colours is every colour of a skin: its surfaces, lines, inks, accent and states, and the maps a name picks its colour from.
type Colours struct {
	Surface  Surfaces
	Line     Strokes
	Ink      Inks
	Accent   Accents
	State    States
	Priority map[string]color.NRGBA
	Standing map[string]Inked
	Labels   map[string]color.NRGBA
	Stones   []Shade
}

// Surfaces is every ground something is drawn on, from the window's back to a control lit under the pointer.
type Surfaces struct {
	Back, Ground, Card, Hover, Tint, Raised, Field color.NRGBA
	Control, ControlLit                            color.NRGBA
	ChipOnControl, ChipOnControlLit                color.NRGBA
	Ghost, Veil, VeilLight                         color.NRGBA
}

// Strokes is every colour a line is drawn in: a rule on the board, a rule on something raised, an edge, and the ring of what is chosen.
type Strokes struct{ Rule, Raised, Edge, Selected color.NRGBA }

// Inks is every colour a text is written in.
type Inks struct{ Text, Faint, Low, OnTint, OnAccent color.NRGBA }

// Accents is the accent and what is made of it: lit, the halo of a field written into, the glow of a card in focus, and the chips drawn on it.
type Accents struct{ Base, Lit, Halo, Glow, Chip, LitChip color.NRGBA }

// States is the colour of what is late, done, wrong or dangerous, and what is made of them.
type States struct{ Late, Done, Wrong, Danger, DangerLit, DangerChip, DangerLitChip color.NRGBA }

// Inked is a ground and the ink written on it.
type Inked struct{ Back, Ink color.NRGBA }

// Shade is the light top and the dark foot of something drawn in a gradient.
type Shade struct{ Light, Dark color.NRGBA }

// LabelPill is the ground a label of that colour is drawn on, a fifth of its colour over a card's ground, and the ink its name is written in, the colour itself.
func (c Colours) LabelPill(colour string) (ground, ink color.NRGBA) {
	ink, ok := c.Labels[colour]
	if !ok {
		return c.Surface.Control, c.Ink.Faint
	}
	card := c.Surface.Card
	fifth := func(a, b uint8) uint8 { return uint8((int(a)*1 + int(b)*4) / 5) }
	return color.NRGBA{R: fifth(ink.R, card.R), G: fifth(ink.G, card.G), B: fifth(ink.B, card.B), A: 255}, ink
}

var standings = []objective.State{objective.Active, objective.OnHold, objective.Reached, objective.Dropped}

func (c *Colours) each(do func(name string, at *color.NRGBA)) {
	groups := reflect.ValueOf(c).Elem()
	for i := range groups.NumField() {
		group := groups.Field(i)
		if group.Kind() != reflect.Struct {
			continue
		}
		for j := range group.NumField() {
			do(snake(groups.Type().Field(i).Name)+"."+snake(group.Type().Field(j).Name), group.Field(j).Addr().Interface().(*color.NRGBA))
		}
	}
	for _, p := range domain.Priorities {
		if said, moved := visited(do, "priority."+string(p), c.Priority[string(p)]); moved {
			c.Priority[string(p)] = said
		}
	}
	for _, s := range standings {
		back, movedBack := visited(do, "standing."+string(s)+".back", c.Standing[string(s)].Back)
		ink, movedInk := visited(do, "standing."+string(s)+".ink", c.Standing[string(s)].Ink)
		if movedBack || movedInk {
			c.Standing[string(s)] = Inked{Back: back, Ink: ink}
		}
	}
	for i, stone := range c.Stones {
		light, movedLight := visited(do, fmt.Sprintf("stones.%d.light", i), stone.Light)
		dark, movedDark := visited(do, fmt.Sprintf("stones.%d.dark", i), stone.Dark)
		if movedLight || movedDark {
			c.Stones[i] = Shade{Light: light, Dark: dark}
		}
	}
	for _, name := range domain.LabelColours {
		if said, moved := visited(do, labelToken+name, c.Labels[name]); moved {
			c.Labels[name] = said
		}
	}
}

func visited(do func(string, *color.NRGBA), name string, was color.NRGBA) (color.NRGBA, bool) {
	said := was
	do(name, &said)
	return said, said != was
}

const labelToken = "label."

func (c Colours) token(name string) color.NRGBA {
	var found color.NRGBA
	c.each(func(each string, at *color.NRGBA) {
		if each == name {
			found = *at
		}
	})
	return found
}

func snake(name string) string {
	var said strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 {
			said.WriteByte('_')
		}
		said.WriteRune(unicode.ToLower(r))
	}
	return said.String()
}
