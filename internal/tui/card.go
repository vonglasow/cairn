package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"

	"charm.land/lipgloss/v2"
)

// One task, drawn in six rows. Four of them carry something; the two others are
// the frame. What goes on each row was decided once and holds at every width —
// what changes with the width is how much of it fits, never which row it is on.

// titleLines is how many rows of a card the title gets.
const titleLines = 2

// StoppedGlyph answers the mark one of the three stopped states is drawn with.
func StoppedGlyph(state string, g Glyphs) string {
	switch state {
	case string(domain.Blocked):
		return g.Blocked
	case string(domain.Watching):
		return g.Watching
	case string(domain.Paused):
		return g.Paused
	default:
		return ""
	}
}

// marks is the right-hand end of a card's first row: whether there is a
// description, and, in the band, which kind of stopped this is.
//
// The order is fixed and the state is last, so the eye finds it in the same
// place on every card in the band rather than one column further along
// whenever a task happens to carry a description.
func marks(c engine.Card, g Glyphs) string {
	var found []string
	// First, and before anything else a card carries: an archived task is on a
	// board that behaves differently from the one it looks like it is on, and
	// nothing else on the row says so.
	if c.Archived {
		found = append(found, g.Archived)
	}
	if c.Described() {
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
func facts(c engine.Card, g Glyphs, s Skin, now time.Time, width int) string {
	if engine.Stopped(c.State) {
		// A task waiting on nothing any more is the one thing about blocking
		// that has to be acted on, so it is said in the bright ink and in place
		// of the list of what used to be in the way.
		if c.Freed {
			return s.Bright.Render(Clip(g.Reached+" nothing blocks it", width))
		}
		return s.Faint.Render(Clip(inTheWay(c, g), width))
	}
	var told, dressed []string
	if text, show, late := engine.CardDeadline(c, now); show {
		told = append(told, text)
		if late {
			dressed = append(dressed, s.Late.Render(text))
		} else {
			dressed = append(dressed, s.Bright.Render(text))
		}
	}
	if done := c.PartsDone(); done != "" {
		count := g.SubTasks + done
		told = append(told, count)
		dressed = append(dressed, s.Faint.Render(count))
	}
	for _, label := range c.Labels {
		tag := g.Label + " " + label
		told = append(told, tag)
		dressed = append(dressed, s.Says.Render(tag))
	}
	// A row that ran out of width dropped whatever did not fit without a word,
	// so a task carrying four labels read as a task carrying one.
	return withMore(told, dressed, s, width)
}

// dress puts the styling back on a row whose width was worked out plain. It
// takes as many dressed facts as the plain row kept, which is what stops a
// colour deciding whether something fits.
func dress(kept string, told, dressed []string) string {
	if kept == "" {
		return ""
	}
	return strings.Join(dressed[:max(fit(told, lipgloss.Width(kept)), 1)], "  ")
}

// withMore says how many facts a row had no room for, and drops one more fact
// to make room for saying it. Nothing is added when everything fitted: a row
// that said "+0" would be a row saying something about nothing.
func withMore(told, dressed []string, s Skin, width int) string {
	shown := fit(told, width)
	if shown == len(told) {
		return dress(Join("  ", width, told...), told, dressed)
	}
	for shown > 0 {
		more := fmt.Sprintf("+%d", len(told)-shown)
		row := append(append([]string{}, told[:shown]...), more)
		if lipgloss.Width(strings.Join(row, "  ")) <= width {
			kept := append(append([]string{}, dressed[:shown]...), s.Faint.Render(more))
			return strings.Join(kept, "  ")
		}
		shown--
	}
	return ""
}

// fit says how many of the facts a row of this width holds.
func fit(told []string, width int) int {
	kept := Join("  ", width, told...)
	for i := range told {
		if !strings.Contains(kept, told[i]) {
			return i
		}
	}
	return len(told)
}

// inTheWay is what the engine says is stopping a task, behind its mark.
func inTheWay(c engine.Card, g Glyphs) string {
	if way := c.InTheWay(); way != "" {
		return g.InTheWay + " " + way
	}
	return ""
}

// Rows is the four lines a card carries, before anything is drawn around them.
// Separated from the frame so a test can read what a card says without reading
// what it is boxed in.
func Rows(c engine.Card, g Glyphs, s Skin, now time.Time, width int) []string {
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
