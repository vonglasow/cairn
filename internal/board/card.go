package board

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

// One task, drawn in six rows. Four of them carry something; the two others are
// the frame. What goes on each row was decided once and holds at every width —
// what changes with the width is how much of it fits, never which row it is on.

// soon is how close a deadline has to be before a card mentions it. A day three
// months out is a fact about a task and not information on a board: it will be
// there tomorrow, and it will be there in ninety days, and saying so on every
// card every day buys nothing.
const soon = 7

// titleLines is how many rows of a card the title gets.
const titleLines = 2

// Stopped says whether a state is one of the three that mean "not moving", and
// therefore whether the task belongs in the band rather than in a column.
func Stopped(state string) bool {
	return state == "blocked" || state == "watching" || state == "paused"
}

// StoppedGlyph answers the mark one of the three stopped states is drawn with.
func StoppedGlyph(state string, g Glyphs) string {
	switch state {
	case "blocked":
		return g.Blocked
	case "watching":
		return g.Watching
	case "paused":
		return g.Paused
	default:
		return ""
	}
}

// Deadline says how a day is written on a card, and whether it is written at
// all. Two answers rather than one string: a day that has passed is drawn in a
// different colour, and a colour is not something a string can carry.
func Deadline(due, now time.Time) (text string, show, late bool) {
	if due.IsZero() {
		return "", false, false
	}
	days := int(due.Sub(startOfDay(now)).Hours() / 24)
	if days < 0 {
		return fmt.Sprintf("−%dd", -days), true, true
	}
	if days > soon {
		return "", false, false
	}
	return fmt.Sprintf("%dd", days), true, false
}

// startOfDay throws away the time of day. A deadline is a day, so the answer to
// "how many days" must not depend on what o'clock it is when the board is
// opened.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// marks is the right-hand end of a card's first row: whether there is a
// description, and, in the band, which kind of stopped this is.
//
// The order is fixed and the state is last, so the eye finds it in the same
// place on every card in the band rather than one column further along
// whenever a task happens to carry a description.
func marks(c Card, g Glyphs) string {
	var found []string
	if c.Described {
		found = append(found, g.Described)
	}
	if glyph := StoppedGlyph(c.State, g); glyph != "" {
		found = append(found, glyph)
	}
	return strings.Join(found, " ")
}

// facts is a card's last row. In a column it says when, how far along, and what
// it is tagged with; in the band it says what is in the way instead, which is
// the only question a stalled task raises.
//
// It is styled here rather than after the fact, because what is dropped when
// the row is full has to be decided on what is drawn and not on the escape
// codes around it — so the widths are measured first and the colour put on
// afterwards, one fact at a time.
func facts(c Card, g Glyphs, s Skin, now time.Time, width int) string {
	if Stopped(c.State) {
		return s.Faint.Render(Clip(inTheWay(c, g), width))
	}
	var told, dressed []string
	if text, show, late := Deadline(c.Deadline, now); show {
		told = append(told, text)
		if late {
			dressed = append(dressed, s.Late.Render(text))
		} else {
			dressed = append(dressed, s.Bright.Render(text))
		}
	}
	if c.SubTotal > 0 {
		count := fmt.Sprintf("%s%d/%d", g.SubTasks, c.SubDone, c.SubTotal)
		told = append(told, count)
		dressed = append(dressed, s.Faint.Render(count))
	}
	for _, label := range c.Labels {
		tag := g.Label + " " + label
		told = append(told, tag)
		dressed = append(dressed, s.Says.Render(tag))
	}
	return dress(Join("  ", width, told...), told, dressed)
}

// dress puts the styling back on a row whose width was worked out plain. It
// takes as many dressed facts as the plain row kept, which is what stops a
// colour deciding whether something fits.
func dress(kept string, told, dressed []string) string {
	if kept == "" {
		return ""
	}
	how := 1
	for i := range told {
		if !strings.Contains(kept, told[i]) {
			break
		}
		how = i + 1
	}
	return strings.Join(dressed[:how], "  ")
}

// inTheWay says what is stopping a task: the tasks it waits on, or the words
// somebody wrote when it waits on something cairn does not hold.
func inTheWay(c Card, g Glyphs) string {
	if len(c.Blockers) > 0 {
		numbers := make([]string, 0, len(c.Blockers))
		for _, blocker := range c.Blockers {
			numbers = append(numbers, strconv.Itoa(blocker))
		}
		return g.InTheWay + " " + strings.Join(numbers, " ")
	}
	if c.Because != "" {
		return g.InTheWay + " " + c.Because
	}
	return ""
}

// Rows is the four lines a card carries, before anything is drawn around them.
// Separated from the frame so a test can read what a card says without reading
// what it is boxed in.
func (c Card) Rows(g Glyphs, s Skin, now time.Time, width int) []string {
	var rows []string
	number := s.Faint.Render(strconv.Itoa(c.Number))
	mark := marks(c, g)
	if glyph := StoppedGlyph(c.State, g); glyph != "" {
		mark = strings.Replace(mark, glyph,
			lipgloss.NewStyle().Foreground(s.palette.Stopped(c.State)).Render(glyph), 1)
	}
	rows = append(rows, Spread(number, mark, width))
	for _, line := range Wrap(c.Title, width, titleLines) {
		rows = append(rows, s.Bright.Render(line))
	}
	return append(rows, facts(c, g, s, now, width))
}
