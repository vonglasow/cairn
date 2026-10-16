package tui

import (
	"fmt"

	"cairn/internal/engine"
)

// Set names one of the two.
type Set string

// The two sets.
const (
	NerdFont Set = "nerdfont"
	Unicode  Set = "unicode"
)

// Sets is every set that can be asked for, for a surface that has to say what it accepts.
var Sets = []Set{NerdFont, Unicode}

// Glyphs is one mark per thing the board has to say without words.
type Glyphs struct {
	Blocked   string
	Watching  string
	Paused    string
	Active    string
	OnHold    string
	Reached   string
	Dropped   string
	Described string
	Steps     string
	Label     string
	InTheWay  string
	Archived  string
}

var nerd = Glyphs{
	Blocked:   "",
	Watching:  "",
	Paused:    "󰏤",
	Active:    "",
	OnHold:    "󰏤",
	Reached:   "",
	Dropped:   "",
	Described: "",
	Steps:     "",
	Label:     "",
	InTheWay:  "",
}

var plain = Glyphs{
	Blocked:   "⊘",
	Watching:  "◉",
	Paused:    "‖",
	Active:    "▶",
	OnHold:    "‖",
	Reached:   "✓",
	Dropped:   "✕",
	Described: "≡",
	Steps:     "▪",
	Label:     "⚑",
	InTheWay:  "↳",
	Archived:  "▭",
}

// ErrNotASet is returned when a set is asked for by a name that is not one of the two.
var ErrNotASet = engine.Refusing("not-a-glyph-set", "glyphs are drawn from one of two sets")

// ParseSet reads which set was asked for, and refuses anything else by name.
func ParseSet(named string) (Set, error) {
	for _, set := range Sets {
		if string(set) == named {
			return set, nil
		}
	}
	return "", fmt.Errorf("%q: %w — nerdfont or unicode", named, ErrNotASet)
}

// GlyphsFor answers the set that was asked for.
func GlyphsFor(s Set) Glyphs {
	if s == NerdFont {
		return nerd
	}
	return plain
}
