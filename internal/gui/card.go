package gui

import (
	"unicode"
	"unicode/utf8"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"strconv"
	"strings"
	"time"
)

const perGlyph = 8

const perSmallGlyph = 7

// TitleLines is a title laid out on the two lines a card gives it, first wide for the first and room for the second as wide measures them: whole when it fits, broken between words, and cut with an ellipsis at the end of the second line when it is longer still.
func TitleLines(title string, first, room int, wide func(string) int) []string {
	if first <= 0 || wide(title) <= first {
		return []string{title}
	}
	line, rest := firstLine(title, first, wide)
	if wide(rest) <= room {
		return []string{line, rest}
	}
	return []string{line, cutToFit(rest, room, wide)}
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

// CardRow is how a card's row is laid out: the number it starts with, each label's pill, the pill counting the labels left out, and its facts at the right end.
type CardRow struct {
	Number   string
	Shown    []string
	Pills    []Box
	More     Box
	Count    string
	Facts    string
	FactsAt  Box
	ServesAt Box
	Late     bool
}

// CardRowOf lays a card's row out in row, measuring its words with m and its figures in Go Mono.
func CardRowOf(c engine.Card, row Box, m Measure, now time.Time) CardRow {
	out := CardRow{Number: "#" + Numbered(c)}
	if c.Described() {
		out.Number += " " + Marks().Described
	}
	x := row.X + monoWide(out.Number, small) + 2*pillGap
	end := row.X + row.W
	out.Facts, out.Late = Facts(c, now)
	if out.Facts != "" {
		out.FactsAt = FactsAt(row, mixedWide(m, out.Facts, small))
		if serves := Serves(c); serves != "" {
			wide := mixedWide(m, serves, small)
			out.ServesAt = Box{X: out.FactsAt.X + out.FactsAt.W - wide, Y: row.Y, W: wide, H: row.H}
		}
		end = out.FactsAt.X - 2*pillGap
	}
	wide := func(said string) int { return ends(m, said, small) }
	shown, more := Pills(c.Labels, end-x, wide)
	top := row.Y + (row.H-pillHigh)/2
	for _, label := range shown {
		pill := Box{X: x, Y: top, W: wide(label) + 2*pillPad, H: pillHigh}
		out.Shown, out.Pills = append(out.Shown, label), append(out.Pills, pill)
		x += pill.W + pillGap
	}
	if more > 0 {
		out.Count = "+" + strconv.Itoa(more)
		out.More = Box{X: x, Y: top, W: wide(out.Count) + 2*pillPad, H: pillHigh}
	}
	return out
}

// mixedWide is how wide said is written in Go Mono at size, a mark Go Mono has no glyph for measured in the face that stands in for it.
func mixedWide(m Measure, said string, size int) int {
	wide := 0
	for _, r := range said {
		if r < unicode.MaxLatin1 || r == '−' {
			wide += monoWide(string(r), size)
			continue
		}
		wide += ends(m, string(r), size)
	}
	return wide
}

// monoWide is how wide Go Mono writes said at size: every glyph of it advances three fifths of an em, rounded up, so the last glyph is never pushed out of its box.
func monoWide(said string, size int) int { return (utf8.RuneCountInString(said)*size*3 + 4) / 5 }

const padding = 12

const (
	markSide  = 13
	markSpace = 5
)

// FocusMark is where the cairn of a card in focus stands: at the right end of its title's first line.
func FocusMark(c engine.Card, title Box) (Box, bool) {
	if !c.InFocus {
		return Box{}, false
	}
	return Box{X: title.X + title.W - markSide, Y: title.Y + (lineHigh-markSide)/2, W: markSide, H: markSide}, true
}

// TitleRoom is how wide a card's title may run on its first line and on its second: the first less the cairn's room when the card is in focus.
func TitleRoom(c engine.Card, title Box) (int, int) {
	if c.InFocus {
		return title.W - markSide - markSpace, title.W
	}
	return title.W, title.W
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

// StandingPill is where an objective's card draws its standing: the first pill of its row, where a task's first label stands.
func StandingPill(face engine.Card, box Box, m Measure) Box {
	if pills := CardRowOf(face, PartsOf(box).Number, m, time.Time{}).Pills; len(pills) > 0 {
		return pills[0]
	}
	return Box{}
}

// Standings is where an objective stands, put where a card writes its labels.
func Standings(g engine.Goal) []string {
	if mark := Standing(g); mark != "" {
		return []string{mark + " " + g.State}
	}
	return nil
}
