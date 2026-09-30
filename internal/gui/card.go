package gui

// What a card says, worked out before anything is drawn.

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"cairn/internal/reading"
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

// perSmallGlyph is the same for the small face, which is what a card's labels
// and its number are written in.
const perSmallGlyph = 7

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

// Tagged is the labels a card says it carries, cut to what its width holds.
// What is left out is counted rather than dropped in silence: a card saying
// two of five labels while reading like it says all of them is a card that
// misleads about the work.
func Tagged(c reading.Card, width int) string {
	room := width / perSmallGlyph
	shown := 0
	// Grown one at a time, counting the rest as it goes: what is left over is
	// written on the same line, so a label only fits if it fits beside the
	// count of the ones it pushes out.
	for shown < len(c.Labels) && utf8.RuneCountInString(saying(c.Labels, shown+1)) <= room {
		shown++
	}
	return saying(c.Labels, shown)
}

// saying is the line for the first so many labels, with however many are left
// over counted at its end.
func saying(labels []string, shown int) string {
	line := strings.Join(labels[:shown], "  ")
	left := len(labels) - shown
	if left == 0 {
		return line
	}
	said := "+" + strconv.Itoa(left)
	if line == "" {
		return said
	}
	return line + "  " + said
}

// padding is the room left inside a card's edge, so text never touches its
// border.
const padding = 10

// What a card's number line says besides the number.
//
// Written as a mark rather than drawn as a glyph: the eleven glyph marks the
// terminal board draws were left out of this window on purpose, and this one
// earns its place because it cannot be worked out by looking — a card gives no
// sign that somebody, or some agent session, is on it.
//
// There were two. The other said a card hung under another, and it went when
// sub-tasks left the board: nothing draws one there to mark.
const (
	// InFocusMark says somebody, or some agent session, is on this task.
	InFocusMark = " ●"
)

// Numbered is a card's number line: the number, and what the card cannot say
// for itself.
func Numbered(c reading.Card) string {
	line := strconv.Itoa(c.Number)
	if c.InFocus {
		line += InFocusMark
	}
	return line
}

// Carried is what a card says of what it holds but does not show: that there
// is a description to read, and how far the parts under it have got. The two
// facts the terminal's marks carry that a window had no other way of saying —
// the rest of them it says by where a card sits, what colour its edge is, and
// the labels written across it.
func Carried(c reading.Card, g Glyphs) string {
	said := ""
	if c.Described {
		said = g.Described
	}
	if c.SubTotal > 0 {
		said = strings.TrimSpace(said + " " + g.SubTasks + " " +
			strconv.Itoa(c.SubDone) + "/" + strconv.Itoa(c.SubTotal))
	}
	return said
}

// Glyphs are the marks a card writes. Named rather than written in, because
// the terminal has two sets and the day a window wants the other one there is
// somewhere for it to go.
type Glyphs struct {
	Described, SubTasks string
}

// Marks is the one set a window draws. Nine of the ten shapes the terminal
// uses render in the font Gio ships; only the flag does not, and a window
// writes its labels out rather than flagging them.
func Marks() Glyphs { return Glyphs{Described: "≡", SubTasks: "▪"} }

// Counted is what an objective's card says of the work under it: how many
// tasks serve it, how many are done, and what share of it that is. Empty when
// nothing serves it — "0/0" and a share of nothing are two ways of dividing by
// zero, and a card saying 0% of an objective nobody has started reads as a
// failure rather than as an objective waiting for its first task.
//
// The share is rounded down: an objective one task short of finished must not
// read 100%.
func Counted(g reading.Goal) string {
	if g.Total == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d · %d%%", g.Done, g.Total, g.Done*100/g.Total)
}

// SubLine is what one part of the work reads on its tile: its number, then its
// title, cut to the room there is.
//
// Whether it is done is not written here — the strip down the left says it, and
// a word as well would spend on repetition the room the title needs.
//
// Cutting it to the tile is not done here. The font is proportional, so what
// fits is measured where there is something to measure with, and a count of
// runes would only be close.
func SubLine(c reading.Card) string {
	return "#" + strconv.Itoa(c.Number) + "  " + c.Title
}
