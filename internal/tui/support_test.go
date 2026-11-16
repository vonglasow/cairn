package tui

import (
	"cairn/internal/theme"
	"cairn/internal/tui/design"
	"strings"
)

func (f Form) typedInto(name, value string) Form {
	if held, there := f.held[name]; there {
		*held = value
	}
	return f
}

// Showing is how many whole cards a run of rows has room for.

// PlainSkin is a skin with no colour in it.
func PlainSkin() design.Skin { return design.Skin{Palette: design.PaletteOf(theme.Default())} }

func flat(s string) string { return strings.Join(strings.Fields(s), " ") }
