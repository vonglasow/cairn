package gui

import (
	"cairn/internal/domain"
	"cairn/internal/engine"
	"strconv"
	"strings"
	"time"
)

const perGlyph = 8

const perSmallGlyph = 7

// TitleLines is a title laid out on the two lines a card gives it, at room wide each as wide measures it: whole when it fits, broken between words, and cut with an ellipsis at the end of the second line when it is longer still.
func TitleLines(title string, room int, wide func(string) int) []string {
	if room <= 0 || wide(title) <= room {
		return []string{title}
	}
	first, rest := firstLine(title, room, wide)
	if wide(rest) <= room {
		return []string{first, rest}
	}
	return []string{first, cutToFit(rest, room, wide)}
}

func firstLine(title string, room int, wide func(string) int) (string, string) {
	words := strings.Fields(title)
	line := ""
	for i, word := range words {
		next := word
		if line != "" {
			next = line + " " + word
		}
		if wide(next) > room {
			if line == "" {
				kept := runesFitting([]rune(word), room, wide)
				return string(kept), strings.Join(append([]string{string([]rune(word)[len(kept):])}, words[i+1:]...), " ")
			}
			return line, strings.Join(words[i:], " ")
		}
		line = next
	}
	return line, ""
}

func runesFitting(runes []rune, room int, wide func(string) int) []rune {
	n := len(runes)
	for n > 1 && wide(string(runes[:n])) > room {
		n--
	}
	return runes[:n]
}

func cutToFit(said string, room int, wide func(string) int) string {
	runes := []rune(said)
	for n := len(runes); n > 0; n-- {
		cut := strings.TrimRight(string(runes[:n]), " ") + "…"
		if wide(cut) <= room {
			return cut
		}
	}
	return "…"
}

// Pills is how many of a card's labels its row has room for as pills, two at most, and how many it leaves out: the count gets a pill of its own, so a label is only shown when the count still fits beside it.
func Pills(labels []string, room int, wide func(string) int) ([]string, int) {
	pill := func(said string) int { return wide(said) + 2*pillPad }
	count := func(left int) int { return pill("+" + strconv.Itoa(left)) }
	used, shown := 0, 0
	for i := range min(pillsAtMost, len(labels)) {
		need := used + pill(labels[i])
		if i > 0 {
			need += pillGap
		}
		with := need
		if left := len(labels) - i - 1; left > 0 {
			with += pillGap + count(left)
		}
		if with > room {
			break
		}
		used, shown = need, i+1
	}
	more := len(labels) - shown
	if more > 0 && shown == 0 && count(more) > room {
		more = 0
	}
	if shown == 0 {
		return nil, more
	}
	return labels[:shown], more
}

const (
	pillsAtMost = 2
	pillPad     = 7
	pillGap     = 4
	pillHigh    = 16
)

const padding = 12

const (
	markSide  = 13
	markSpace = 5
)

// FocusMark is where the cairn of a card in focus stands: after the last line of its title, as wide as that line is.
func FocusMark(c engine.Card, title Box, lines, lastWide int) (Box, bool) {
	if !c.InFocus {
		return Box{}, false
	}
	return Box{X: title.X + lastWide + markSpace, Y: title.Y + (lines-1)*lineHigh + (lineHigh-markSide)/2, W: markSide, H: markSide}, true
}

// TitleRoom is how wide a card's title may run: the card's room for it, less the cairn's when the card is in focus.
func TitleRoom(c engine.Card, title Box) int {
	if c.InFocus {
		return title.W - markSide - markSpace
	}
	return title.W
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
	said := g.Counted() + " · " + strconv.Itoa(g.Percent()) + "%"
	if g.Spent > 0 {
		said += " · " + domain.FormatDuration(g.Spent)
	}
	return said
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
