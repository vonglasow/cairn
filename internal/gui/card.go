package gui

// What a card says, worked out before anything is drawn.

import (
	"cairn/internal/engine"
	"strconv"
	"strings"
	"time"
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
func Tagged(c engine.Card, width int) string {
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
// border. On the scale everything else is spaced on, which the window did not
// used to be: 10 here where a margin and a gap were both 12.
const padding = 12

// What a card's number line says besides the number.
//
// The one mark a card carries before its number: an eye, drawn rather than
// written, saying somebody, or some agent session, is on this task — it
// cannot be worked out by looking. It was a dot, which read as nothing in
// particular. The eleven glyph marks the terminal board draws were left out of
// this window on purpose; there was another, saying a card hung under
// another, and it went when sub-tasks left the board.
const (
	eyeWide  = 16
	eyeHigh  = 10
	eyeSpace = 5
)

// FocusEye is where a card's eye is drawn in its number line, and whether it
// has one: only a card in focus does.
func FocusEye(c engine.Card, number Box) (Box, bool) {
	if !c.InFocus {
		return Box{}, false
	}
	return Box{X: number.X, Y: number.Y + (number.H-eyeHigh)/2, W: eyeWide, H: eyeHigh}, true
}

// NumberText is where a card's number is written: along the number line, or
// after the eye when there is one.
func NumberText(c engine.Card, number Box) Box {
	if _, ok := FocusEye(c, number); !ok {
		return number
	}
	moved := eyeWide + eyeSpace
	return Box{X: number.X + moved, Y: number.Y, W: number.W - moved, H: number.H}
}

// Numbered is a card's number line: its number. What a card cannot say for
// itself is drawn beside it — see FocusEye.
func Numbered(c engine.Card) string { return strconv.Itoa(c.Number) }

// Carried is what a card says of what it holds but does not show: that there
// is a description to read, and how far the parts under it have got. The two
// facts the terminal's marks carry that a window had no other way of saying —
// the rest of them it says by where a card sits, what colour its edge is, and
// the labels written across it.
func Carried(c engine.Card, g Glyphs) string {
	said := ""
	if c.Described() {
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
func Counted(g engine.Goal) string {
	if g.Total == 0 {
		return ""
	}
	return g.Counted() + " · " + strconv.Itoa(g.Percent()) + "%"
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
//
// A part typed into a form and not yet saved has no number, and reads as its
// title alone: a "#0" would be a number nobody gave it.
func SubLine(c engine.Card) string {
	if c.Number == 0 {
		return c.Title
	}
	return "#" + strconv.Itoa(c.Number) + "  " + c.Title
}

// Due is what a card says about when it is wanted, and whether that day has
// passed. The window knew nothing about time: no deadline, no notion of late,
// where the terminal has written it on the card all along. Deciding what to do
// without knowing what is due is not deciding.
//
// The answer is engine.Deadline's and not a second reckoning of its own: two
// boards saying different things about the same day would be worse than one
// board saying nothing.
func Due(c engine.Card, now time.Time) (said string, late bool) {
	text, show, over := engine.CardDeadline(c, now)
	if !show {
		return "", false
	}
	return text, over
}

// Spent is how long a task was reckoned to take against how long it has, and
// whether it has gone past. Empty when neither was ever written: a card saying
// "0h/0h" says nothing and takes the room of something that would.
func Spent(c engine.Card) (said string, over bool) {
	if c.Spent == "" && c.Estimate == "" {
		return "", false
	}
	if c.Estimate == "" {
		return c.Spent, false
	}
	if c.Spent == "" {
		return c.Estimate, false
	}
	return c.Spent + "/" + c.Estimate, c.Overrun
}

// FreedMark is what a blocked task wears once every task it waits on is done.
const FreedMark = "✓ nothing blocks it"

// ServesMark is what stands in front of the objectives a task serves.
const ServesMark = "◇"

// Serves is what a task contributes to, which the detail carried and no
// overview ever summarised.
func Serves(c engine.Card) string {
	if len(c.Objectives) == 0 {
		return ""
	}
	return ServesMark + engine.Numbers(c.Objectives)
}

// Standing is the mark an objective's card wears for where it stands. An
// objective was drawn exactly like a task, so a list mixed what is running with
// what was abandoned and nothing on either said which was which.
func Standing(g engine.Goal) string {
	switch g.State {
	case "active":
		return "▶"
	case "on-hold":
		return "‖"
	case "reached":
		return "✓"
	case "dropped":
		return "✕"
	}
	return ""
}

// Facts is the right-hand end of a card's label row: when it is wanted, what it
// has cost against what it was reckoned to cost, and what it serves. Three
// facts the detail carried and no overview ever summarised.
//
// One string and one warning rather than a piece per ink: a card has one row
// left and the eye reads it as one run. The warning is late or over, which are
// the two states that ask to be acted on.
func Facts(c engine.Card, now time.Time) (said string, warn bool) {
	// Waiting on nothing any more is the one thing about a blocked task that
	// has to be acted on, and it crowds out the rest: a task that can start is
	// not asking how long it was reckoned to take.
	if c.Freed {
		return FreedMark, true
	}
	var pieces []string
	due, late := Due(c, now)
	if due != "" {
		pieces = append(pieces, due)
	}
	spent, over := Spent(c)
	if spent != "" {
		pieces = append(pieces, spent)
	}
	if serves := Serves(c); serves != "" {
		pieces = append(pieces, serves)
	}
	return strings.Join(pieces, "  "), late || over
}

// Standings is where an objective stands, put where a card writes its labels.
// An objective was drawn exactly like a task and said nothing about whether it
// was running, on hold, reached or abandoned — so a list mixed all four.
func Standings(g engine.Goal) []string {
	if mark := Standing(g); mark != "" {
		return []string{mark + " " + g.State}
	}
	return nil
}
