package gui

import (
	"cairn/internal/engine"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const perGlyph = 8

const perSmallGlyph = 7

// TitleOf is a title cut to what a card of this width can show, ending in an ellipsis when it had to be cut.
func TitleOf(title string, width int) string {
	room := (width - padding - padding) / perGlyph
	if room < 2 || utf8.RuneCountInString(title) <= room {
		return title
	}
	kept := []rune(title)[:room-1]
	return strings.TrimRight(string(kept), " ") + "…"
}

// Tagged is the labels a card says it carries, cut to what its width holds.
func Tagged(c engine.Card, width int) string {
	room := width / perSmallGlyph
	shown := 0
	for shown < len(c.Labels) && utf8.RuneCountInString(saying(c.Labels, shown+1)) <= room {
		shown++
	}
	said := saying(c.Labels, shown)
	if utf8.RuneCountInString(said) > room {
		return ""
	}
	return said
}

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

const padding = 12

const (
	eyeWide  = 16
	eyeHigh  = 10
	eyeSpace = 5
)

// FocusEye is where the eye of a card in focus is drawn: after its number, on the number line.
func FocusEye(c engine.Card, number Box, numberWide int) (Box, bool) {
	if !c.InFocus {
		return Box{}, false
	}
	return Box{X: number.X + numberWide + eyeSpace, Y: number.Y + (number.H-eyeHigh)/2, W: eyeWide, H: eyeHigh}, true
}

// Numbered is a card's number line: its number.
func Numbered(c engine.Card) string { return strconv.Itoa(c.Number) }

// Carried is what a card says of what it holds but does not show: that there is a description to read, and how far its steps have got.
func Carried(c engine.Card, g Glyphs) string {
	said := ""
	if c.Described() {
		said = g.Described
	}
	if done := c.PartsDone(); done != "" {
		said = strings.TrimSpace(said + " " + g.Steps + " " + done)
	}
	return said
}

// Glyphs are the marks a card writes.
type Glyphs struct {
	Described, Steps string
}

// Marks is the one set a window draws.
func Marks() Glyphs { return Glyphs{Described: "≡", Steps: "▪"} }

// Counted is what an objective's card says of the work under it: how many tasks serve it, how many are done, and what share of it that is.
func Counted(g engine.Goal) string {
	if g.Total == 0 {
		return ""
	}
	return g.Counted() + " · " + strconv.Itoa(g.Percent()) + "%"
}

// SubLine is what a task serving an objective reads on its tile: its number, then its title, cut to the room there is.
func SubLine(c engine.Card) string {
	if c.Number == 0 {
		return c.Title
	}
	return "#" + strconv.Itoa(c.Number) + "  " + c.Title
}

// StepLine is what the tile of a step says: its title alone, since the place it is named by is no task's number.
func StepLine(c engine.Card) string { return c.Title }

// Due is what a card says about when it is wanted, and whether that day has passed.
func Due(c engine.Card, now time.Time) (said string, late bool) {
	text, show, over := engine.CardDeadline(c, now)
	if !show {
		return "", false
	}
	return text, over
}

// FreedMark is what a blocked task wears once every task it waits on is done.
const FreedMark = "✓ nothing blocks it"

// ServesMark is what stands in front of the objectives a task serves.
const ServesMark = "◇"

// Serves is what a task contributes to, which the detail carried and no overview ever summarised.
func Serves(c engine.Card) string {
	if len(c.Objectives) == 0 {
		return ""
	}
	return ServesMark + engine.Numbers(c.Objectives)
}

// Standing is the mark an objective's card wears for where it stands.
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

// Facts is the right-hand end of a card's label row: when it is wanted, what it has cost against what it was reckoned to cost, and what it serves.
func Facts(c engine.Card, now time.Time) (said string, warn bool) {
	if c.Freed {
		return FreedMark, true
	}
	var pieces []string
	due, late := Due(c, now)
	if due != "" {
		pieces = append(pieces, due)
	}
	spent, over := c.Cost()
	if spent != "" {
		pieces = append(pieces, spent)
	}
	if serves := Serves(c); serves != "" {
		pieces = append(pieces, serves)
	}
	return strings.Join(pieces, "  "), late || over
}

// Standings is where an objective stands, put where a card writes its labels.
func Standings(g engine.Goal) []string {
	if mark := Standing(g); mark != "" {
		return []string{mark + " " + g.State}
	}
	return nil
}
