package theme

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

// ErrNotAScheme is what a file is said to be when it cannot be read as YAML holding a base16 scheme.
var ErrNotAScheme = errors.New("is not a base16 scheme cairn can make sense of")

// ErrNotAColour is what a value is said to be when it is not a colour written in hexadecimal.
var ErrNotAColour = errors.New("is not a colour, written #rrggbb")

// ErrSlotMissing is what one of the sixteen slots of a palette is said to be when the file leaves it out.
var ErrSlotMissing = errors.New("is not said, and a base16 palette says all sixteen")

// ErrNoSuchToken is what a name under cairn.tokens or cairn.labels is said to be when it names nothing the window draws with.
var ErrNoSuchToken = errors.New("is not a token cairn knows")

// Scheme is a colour scheme as its file says it: its name, who made it, whether it is light, its sixteen slots and what it says of cairn's tokens and labels.
type Scheme struct {
	Slug, Name, Author, From string
	variant                  string
	palette                  [16]color.NRGBA
	tokens, labels           map[string]color.NRGBA
}

type file struct {
	Name    string            `yaml:"name"`
	Author  string            `yaml:"author"`
	Variant string            `yaml:"variant"`
	Palette map[string]string `yaml:"palette"`
	Cairn   struct {
		Tokens map[string]string `yaml:"tokens"`
		Labels map[string]string `yaml:"labels"`
	} `yaml:"cairn"`
}

// Parse reads a scheme written as tinted-theming writes base16 and base24, with cairn's own block of tokens and labels; slug and from name it in what it says when it cannot.
func Parse(slug, from string, data []byte) (Scheme, error) {
	wrong := func(err error) error { return fmt.Errorf("%s (%s): %w", slug, from, err) }
	var f file
	if err := yaml.Unmarshal(data, &f); err != nil {
		return Scheme{}, wrong(fmt.Errorf("%w: %w", ErrNotAScheme, err))
	}
	s := Scheme{Slug: slug, Name: f.Name, Author: f.Author, From: from, variant: f.Variant}
	for i := range s.palette {
		slot := fmt.Sprintf("base%02X", i)
		said, ok := f.Palette[slot]
		if !ok {
			return Scheme{}, wrong(fmt.Errorf("palette.%s %w", slot, ErrSlotMissing))
		}
		c, err := colourOf(said)
		if err != nil {
			return Scheme{}, wrong(fmt.Errorf("palette.%s %w", slot, err))
		}
		s.palette[i] = c
	}
	var err error
	if s.tokens, err = coloursOf("cairn.tokens.", f.Cairn.Tokens); err != nil {
		return Scheme{}, wrong(err)
	}
	if s.labels, err = coloursOf("cairn.labels.", f.Cairn.Labels); err != nil {
		return Scheme{}, wrong(err)
	}
	if err := s.known(); err != nil {
		return Scheme{}, wrong(err)
	}
	return s, nil
}

func coloursOf(under string, said map[string]string) (map[string]color.NRGBA, error) {
	read := map[string]color.NRGBA{}
	for name, value := range said {
		c, err := colourOf(value)
		if err != nil {
			return nil, fmt.Errorf("%s%s %w", under, name, err)
		}
		read[name] = c
	}
	return read, nil
}

func (s Scheme) known() error {
	names := map[string]bool{}
	c := s.derived()
	c.each(func(name string, _ *color.NRGBA) { names[name] = true })
	for name := range s.tokens {
		if !names[name] || strings.HasPrefix(name, labelToken) {
			return fmt.Errorf("cairn.tokens.%s %w", name, ErrNoSuchToken)
		}
	}
	for name := range s.labels {
		if !names[labelToken+name] {
			return fmt.Errorf("cairn.labels.%s %w", name, ErrNoSuchToken)
		}
	}
	return nil
}

func colourOf(said string) (color.NRGBA, error) {
	hex := strings.TrimPrefix(said, "#")
	if len(hex) != 6 && len(hex) != 8 {
		return color.NRGBA{}, fmt.Errorf("%q %w", said, ErrNotAColour)
	}
	value, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("%q %w", said, ErrNotAColour)
	}
	if len(hex) == 6 {
		value = value<<8 | 0xff
	}
	return color.NRGBA{R: uint8(value >> 24), G: uint8(value >> 16), B: uint8(value >> 8), A: uint8(value)}, nil
}

// Light says whether the scheme is a light one: as its variant says, or, when it says neither, whether its ground is lighter than its ink.
func (s Scheme) Light() bool {
	switch s.variant {
	case "light":
		return true
	case "dark":
		return false
	}
	return lit(s.palette[0]) > lit(s.palette[5])
}

func mix(a, b color.NRGBA, t float64) color.NRGBA {
	toward := func(x, y uint8) uint8 { return uint8(math.Floor(float64(x) + (float64(y)-float64(x))*t + 0.5)) }
	return color.NRGBA{R: toward(a.R, b.R), G: toward(a.G, b.G), B: toward(a.B, b.B), A: 255}
}

func alpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}

// Colours is every colour the window draws with, worked out from the scheme's sixteen slots, its labels and its tokens over them.
func (s Scheme) Colours() Colours {
	c := s.derived()
	c.each(func(name string, at *color.NRGBA) {
		if said, ok := s.tokens[name]; ok {
			*at = said
		}
	})
	return c
}

func (s Scheme) derived() Colours {
	p := s.palette
	bg, fg, accent := p[0x0], p[0x5], p[0xD]
	ramp := func(t float64) color.NRGBA { return mix(bg, fg, t) }
	faint, low := 0.78, 0.58
	if s.Light() {
		faint, low = 0.86, 0.70
	}
	control, controlLit, danger := ramp(0.19), ramp(0.27), mix(bg, p[0x8], 0.18)
	dangerLit, accentLit, veil := mix(danger, fg, 0.08), mix(accent, fg, 0.15), mix(bg, color.NRGBA{A: 255}, 0.5)
	c := Colours{
		Surface: Surfaces{
			Back: bg, Ground: ramp(0.04), Card: ramp(0.08), Hover: ramp(0.11), Tint: mix(bg, accent, 0.15),
			Raised: ramp(0.16), Field: ramp(0.02), Control: control, ControlLit: controlLit,
			ChipOnControl: ramp(0.06), ChipOnControlLit: ramp(0.14), Ghost: mix(bg, accent, 0.28),
			Veil: alpha(veil, 158), VeilLight: alpha(veil, 79),
		},
		Line: Strokes{Rule: ramp(0.09), Raised: ramp(0.25), Edge: ramp(0.45), Selected: ramp(0.55)},
		Ink:  Inks{Text: fg, Faint: ramp(faint), Low: ramp(low), OnTint: mix(fg, accent, 0.40), OnAccent: bg},
		Accent: Accents{
			Base: accent, Lit: accentLit, Halo: alpha(accent, 38), Glow: alpha(accent, 76),
			Chip: mix(accent, bg, 0.15), LitChip: mix(accentLit, bg, 0.15),
		},
		State: States{
			Late: p[0x9], Done: p[0xB], Wrong: mix(p[0x8], fg, 0.5), Danger: danger, DangerLit: dangerLit,
			DangerChip: mix(danger, bg, 0.5), DangerLitChip: mix(dangerLit, bg, 0.5),
		},
		Priority: map[string]color.NRGBA{
			string(domain.Critical): p[0x8], string(domain.High): p[0xA], string(domain.Normal): p[0xC], string(domain.Low): ramp(0.5),
		},
		Standing: map[string]Inked{},
		Labels:   map[string]color.NRGBA{},
	}
	for state, slot := range map[objective.State]color.NRGBA{objective.Active: p[0xB], objective.OnHold: p[0xA], objective.Reached: p[0xD]} {
		c.Standing[string(state)] = Inked{Back: mix(bg, slot, 0.18), Ink: slot}
	}
	c.Standing[string(objective.Dropped)] = Inked{Back: control, Ink: c.Ink.Faint}
	for _, t := range []float64{0.30, 0.38, 0.46} {
		c.Stones = append(c.Stones, Shade{Light: ramp(t), Dark: mix(ramp(t), bg, 0.4)})
	}
	for name, said := range s.labelled() {
		if s.Light() {
			said = mix(said, fg, 0.45)
		}
		c.Labels[name] = said
	}
	return c
}

func (s Scheme) labelled() map[string]color.NRGBA {
	p, fg := s.palette, s.palette[0x5]
	labels := map[string]color.NRGBA{
		"red": p[0x8], "peach": p[0x9], "yellow": p[0xA], "green": p[0xB], "teal": p[0xC], "blue": p[0xD], "mauve": p[0xE], "flamingo": p[0xF],
		"maroon": mix(p[0x8], p[0x9], 0.4), "pink": mix(p[0x8], p[0xE], 0.5), "rosewater": mix(p[0x8], fg, 0.55),
		"sky": mix(p[0xC], p[0xD], 0.5), "sapphire": mix(p[0xC], p[0xD], 0.75), "lavender": mix(mix(p[0xD], p[0xE], 0.4), fg, 0.3),
	}
	for name, said := range s.labels {
		labels[name] = said
	}
	return labels
}
