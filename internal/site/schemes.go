package site

import (
	"cairn/internal/tui/design"
	"fmt"
	"image/color"
	"slices"
	"strings"

	"cairn/internal/theme"

	"maps"
)

const (
	darkFirst  = theme.DefaultSlug
	lightFirst = "catppuccin-latte"
)

func hex(c color.NRGBA) string { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }

func schemeRules(s theme.Scheme) string {
	c := s.Colours()
	var rules strings.Builder
	fmt.Fprintf(&rules, `[data-scheme=%q]{`, s.Slug)
	for _, token := range []struct {
		name string
		is   color.NRGBA
	}{
		{"bg", c.Surface.Back}, {"surface", c.Surface.Ground}, {"card", c.Surface.Card}, {"tint", c.Surface.Tint},
		{"rule", c.Line.Rule}, {"ink", c.Ink.Text}, {"faint", c.Ink.Faint}, {"on-accent", c.Ink.OnAccent},
		{"accent", c.Accent.Base}, {"ok", c.State.Done}, {"late", c.State.Late}, {"crit", c.Priority["critical"]},
		{"tfg", s.Slots()[5]}, {"tbg", s.Slots()[0]},
	} {
		fmt.Fprintf(&rules, "--%s:%s;", token.name, hex(token.is))
	}
	for i, slot := range terminal16(s) {
		fmt.Fprintf(&rules, "--t%d:%s;", i, hex(slot))
	}
	inks := design.PaletteOf(c).Inks()
	for _, name := range slices.Sorted(maps.Keys(inks)) {
		fmt.Fprintf(&rules, "--p-%s:%s;", name, hex(color.NRGBAModel.Convert(inks[name]).(color.NRGBA)))
	}
	variant := "dark"
	if s.Light() {
		variant = "light"
	}
	fmt.Fprintf(&rules, "color-scheme:%s}\n", variant)
	return rules.String()
}
