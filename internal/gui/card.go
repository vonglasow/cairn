package gui

// What a card says, worked out before anything is drawn.

import (
	"strings"
	"unicode/utf8"
)

// perGlyph is how wide one character is reckoned to be, in dp, at the size a
// card's title is drawn.
//
// Reckoned rather than measured, and that is a real approximation: a
// proportional face gives "iii" and "WWW" very different widths, and this gives
// them the same. It is here so that what a card shows can be decided without a
// shaper, which is what keeps the decision testable with no window open — and
// the drawing clips as well, so a title this lets through and the face does not
// is cut by the face rather than spilling.
const perGlyph = 8

// TitleOf is a title cut to what a card of this width can show, ending in an
// ellipsis when it had to be cut. A card is a fixed height and a title is one
// line of it: something has to give, and it is the tail of a long name rather
// than the shape of the board.
func TitleOf(title string, width int) string {
	room := (width - padding - padding) / perGlyph
	if room < 2 || utf8.RuneCountInString(title) <= room {
		return title
	}
	kept := []rune(title)[:room-1]
	return strings.TrimRight(string(kept), " ") + "…"
}

// padding is the room left inside a card's edge, so text never touches its
// border.
const padding = 10
