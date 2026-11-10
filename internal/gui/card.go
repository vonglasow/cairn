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

// TitleLines is a title laid out on the two lines a card gives it, first wide for the first and room for the second as wide measures them: whole when it fits, broken between words only — a word the first line has no room for moves to the second, a word wider than any line is cut with an ellipsis — and cut with an ellipsis at the end of the second line when it is longer still.
func TitleLines(title string, first, room int, wide func(string) int) []string {
	if first <= 0 || wide(title) <= first {
		return []string{title}
	}
	words := strings.Fields(title)
	if len(words) > 0 && wide(words[0]) > max(first, room) {
		return []string{cutToFit(words[0], first, wide)}
	}
	line, rest := firstLine(words, first, wide)
	if wide(rest) <= room {
		return []string{line, rest}
	}
	return []string{line, cutToFit(rest, room, wide)}
}

func firstLine(words []string, room int, wide func(string) int) (string, string) {
	line := ""
	for i, word := range words {
		next := word
		if line != "" {
			next = line + " " + word
		}
		if wide(next) > room {
			return line, strings.Join(words[i:], " ")
		}
		line = next
	}
	return line, ""
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
	NumberAt Box
	Shown    []string
	Pills    []Box
	More     Box
	Count    string
	Facts    string
	FactsAt  Box
	ServesAt Box
	MarkAt   Box
	Late     bool
	Series   int
	SeriesAt Box
	Start    int
}

// CardRowOf lays a card's row out in row, measuring its words with m and its figures in Go Mono.
func CardRowOf(c engine.Card, row Box, m Measure, now time.Time, pages engine.Pages) CardRow {
	return cardRow(c, row, m, factsOf(c, now, pages))
}

// ReferenceRowOf lays out the row of a reference: why it is there in place of its facts, cut to the room its number leaves, and its labels in what is left.
func ReferenceRowOf(c engine.Card, why string, row Box, m Measure, now time.Time, pages engine.Pages) CardRow {
	why = cutTo(m, why, row.X+row.W-cardRow(c, row, m, nil).Start)
	if why == "" {
		return CardRowOf(c, row, m, now, pages)
	}
	return cardRow(c, row, m, []fact{{said: why, warn: true}})
}

func cutTo(m Measure, said string, room int) string {
	runes := []rune(said)
	for cut := said; len(runes) > 0; cut = strings.TrimSpace(string(runes)) + "…" {
		if mixedWide(m, cut, small) <= room {
			return cut
		}
		runes = runes[:len(runes)-1]
	}
	return ""
}

func cardRow(c engine.Card, row Box, m Measure, pieces []fact) CardRow {
	out := CardRow{Number: "#" + Numbered(c)}
	if c.Described() {
		out.Number += " " + Marks().Described
	}
	out.NumberAt = Box{X: row.X, Y: row.Y, W: monoWide(out.Number, small), H: row.H}
	x := row.X + out.NumberAt.W + 2*pillGap
	if c.Series != nil {
		out.Series = c.Series.Series
		out.SeriesAt = Box{X: x, Y: row.Y, W: SeriesNumberWide(out.Series, small), H: row.H}
		x += out.SeriesAt.W + 2*pillGap
	}
	out.Start = x
	end := row.X + row.W
	serves := Serves(c)
	wideOf := func(pieces []fact) int {
		said, _ := joined(pieces)
		if serves != "" && len(pieces) > 0 && pieces[len(pieces)-1].serves {
			return mixedWide(m, said, small) + servesSide + servesGap
		}
		return mixedWide(m, said, small)
	}
	for len(pieces) > 0 && wideOf(pieces) > end-x {
		pieces = pieces[:len(pieces)-1]
	}
	out.Facts, out.Late = joined(pieces)
	if out.Facts != "" {
		out.FactsAt = FactsAt(row, wideOf(pieces))
		if pieces[len(pieces)-1].serves {
			wide := mixedWide(m, serves, small) + servesSide + servesGap
			out.ServesAt = Box{X: out.FactsAt.X + out.FactsAt.W - wide, Y: row.Y, W: wide, H: row.H}
			out.MarkAt = Box{X: out.ServesAt.X, Y: row.Y + (row.H-servesSide)/2, W: servesSide, H: servesSide}
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

// FocusMark is where the glowing dot of a card in focus stands: at the right end of its title's first line.
func FocusMark(c engine.Card, title Box) (Box, bool) {
	if !c.InFocus {
		return Box{}, false
	}
	return Box{X: title.X + title.W - markSide, Y: title.Y + (lineHigh-markSide)/2, W: markSide, H: markSide}, true
}

// Glow is the mark of a card in focus: a dot, centred, in a faint halo of the same ink.
type Glow struct{ X, Y, Dot, Halo float32 }

// GlowOf is the glow drawn in a mark's box: drawn in fourteen, a dot of 3.4 in a halo of 6, scaled to the box.
func GlowOf(mark Box) Glow {
	scale := float32(mark.W) / 14
	return Glow{X: float32(mark.X) + 7*scale, Y: float32(mark.Y) + 7*scale, Dot: 3.4 * scale, Halo: 6 * scale}
}

// RuleOf is what a card that comes back says under its title: its series' rule; nothing for any other.
func RuleOf(c engine.Card) string {
	if c.Series == nil {
		return ""
	}
	return c.Series.Rule
}

// TitleRoom is how wide a card's title may run on its first line and on its second: the first less the glowing dot's room when the card is in focus.
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

const (
	servesSide = 11
	servesGap  = 3
)

// Serves is what a task contributes to, by name: the first objective it serves, and how many more.
func Serves(c engine.Card) string {
	if len(c.Serves) == 0 {
		return ""
	}
	if len(c.Serves) == 1 {
		return c.Serves[0]
	}
	return c.Serves[0] + " +" + strconv.Itoa(len(c.Serves)-1)
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

// CardFacts is the right-hand end of a card's label row: when it is wanted, what it has cost against what it was reckoned to cost while the time is shown, and what it serves.
func CardFacts(c engine.Card, now time.Time, pages engine.Pages) (said string, warn bool) {
	return joined(factsOf(c, now, pages))
}

type fact struct {
	said         string
	warn, serves bool
}

func factsOf(c engine.Card, now time.Time, pages engine.Pages) []fact {
	if c.Freed {
		return []fact{{said: FreedMark, warn: true}}
	}
	var pieces []fact
	if text, show, late, passed := engine.OccurrenceDay(c, now); show {
		pieces = append(pieces, fact{said: text, warn: late || passed})
	} else if due, late := Due(c, now); due != "" && c.Series == nil {
		pieces = append(pieces, fact{said: due, warn: late})
	}
	if spent, over := pages.Cost(c); spent != "" {
		pieces = append(pieces, fact{said: spent, warn: over})
	}
	if serves := Serves(c); serves != "" {
		pieces = append(pieces, fact{said: serves, serves: true})
	}
	return pieces
}

func joined(pieces []fact) (said string, warn bool) {
	var words []string
	for _, piece := range pieces {
		words = append(words, piece.said)
		warn = warn || piece.warn
	}
	return strings.Join(words, "  "), warn
}

// StandingPill is where an objective's card draws its standing: the first pill of its row, where a task's first label stands.
func StandingPill(face engine.Card, box Box, m Measure) Box {
	if pills := CardRowOf(face, PartsOf(box).Number, m, time.Time{}, engine.Pages{}).Pills; len(pills) > 0 {
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
