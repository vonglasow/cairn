package tui

import (
	"fmt"

	"cairn/internal/engine"
)

// The two sets of glyphs the board may draw with, and nothing between them.
//
// Nerd Font says more in one cell than a word says in eight, which is what
// makes a card readable at sixty columns. It needs a font installed, and
// detecting one is not reliable enough to guess at — so which set is in use is
// something the person says, and the fallback is drawn by any terminal.
//
// Every Nerd Font code point below was read off the project's own glyph table
// rather than remembered. They all sit in the octicon and material ranges,
// which are one cell wide: a two-cell glyph would slide every card's right edge
// out of line.

// Set names one of the two.
type Set string

// The two sets. A third would be a third table to keep true.
const (
	NerdFont Set = "nerdfont"
	Unicode  Set = "unicode"
)

// Sets is every set that can be asked for, for a surface that has to say what
// it accepts.
var Sets = []Set{NerdFont, Unicode}

// Glyphs is one mark per thing the board has to say without words. It is a
// value rather than a lookup, so a card asks for a field and cannot ask for a
// mark that does not exist.
type Glyphs struct {
	Blocked   string
	Watching  string
	Paused    string
	Active    string
	OnHold    string
	Reached   string
	Dropped   string
	Described string
	SubTasks  string
	Label     string
	InTheWay  string
	Archived  string
}

// nerd is the full set.
var nerd = Glyphs{
	Blocked:   "",
	Watching:  "",
	Paused:    "󰏤",
	Active:    "",
	OnHold:    "󰏤",
	Reached:   "",
	Dropped:   "",
	Described: "",
	SubTasks:  "",
	Label:     "",
	InTheWay:  "",
}

// plain is what any terminal draws. It says the same things, less precisely,
// and it keeps the same metaphor in every case — a pause is two bars in both
// sets, so somebody who switches does not have to relearn one mark.
var plain = Glyphs{
	Blocked:   "⊘",
	Watching:  "◉",
	Paused:    "‖",
	Active:    "▶",
	OnHold:    "‖",
	Reached:   "✓",
	Dropped:   "✕",
	Described: "≡",
	SubTasks:  "▪",
	Label:     "⚑",
	InTheWay:  "↳",
	Archived:  "▭",
}

// ErrNotASet is returned when a set is asked for by a name that is not one of
// the two. Named, so a surface can tell a request it understood and refused
// from one it could not read at all.
var ErrNotASet = engine.Refusing("not-a-glyph-set", "glyphs are drawn from one of two sets")

// ParseSet reads which set was asked for, and refuses anything else by name.
// Written here rather than at the surface: which sets exist is the board's to
// know, and a surface that listed them would be a second list to keep true.
func ParseSet(named string) (Set, error) {
	for _, set := range Sets {
		if string(set) == named {
			return set, nil
		}
	}
	return "", fmt.Errorf("%q: %w — nerdfont or unicode", named, ErrNotASet)
}

// GlyphsFor answers the set that was asked for. Anything that is not the full
// set gets the plain one: a board that refused to draw because it did not
// recognise a word would be a board nobody can open to find out why.
func GlyphsFor(s Set) Glyphs {
	if s == NerdFont {
		return nerd
	}
	return plain
}
