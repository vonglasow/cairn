package board

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"cairn/internal/domain"
	"cairn/internal/reading"
)

// An objective, drawn. The same six rows as a task, because a person who has
// learnt one card has learnt both — and one thing a task has not got: how far
// along it is, counted from the columns of everything that serves it. That
// count is the headline, since it is the only fact an objective owns.

// StandingGlyph answers the mark an objective's state is drawn with.
func StandingGlyph(state string, g Glyphs) string {
	switch state {
	case "active":
		return g.Active
	case "on-hold":
		return g.OnHold
	case "reached":
		return g.Reached
	case "dropped":
		return g.Dropped
	default:
		return ""
	}
}

// progress draws how far along an objective is, in colour.
func (b Board) progress(done, total, width int) string {
	plain := Bar(done, total, width)
	filled := strings.Count(plain, "█")
	return b.skin.Filled.Render(strings.Repeat("█", filled)) +
		b.skin.Empty.Render(strings.Repeat("░", len([]rune(plain))-filled))
}

// goalFacts is the bar, the count, and whatever else the width allows, dressed.
func (b Board) goalFacts(g reading.Goal, width int) string {
	room := min(barWidth, width)
	told := []string{}
	dressed := []string{}
	if g.Total > 0 {
		count := fmt.Sprintf("%d/%d", g.Done, g.Total)
		told = append(told, count)
		dressed = append(dressed, b.skin.Bright.Render(count))
	}
	if day := Written(g.Deadline); day != "" {
		told = append(told, day)
		dressed = append(dressed, b.skin.Faint.Render(day))
	}
	for _, label := range g.Labels {
		tag := b.glyphs.Label + " " + label
		told = append(told, tag)
		dressed = append(dressed, b.skin.Says.Render(tag))
	}
	rest := dress(Join("  ", width-room-2, told...), told, dressed)
	if rest == "" {
		return b.progress(g.Done, g.Total, room)
	}
	return b.progress(g.Done, g.Total, room) + "  " + rest
}

// Bar draws how far along an objective is. Nothing under it is not nought per
// cent: an objective nothing serves yet has an empty bar and no count, because
// a count of none out of none says a thing that is not true.
func Bar(done, total, width int) string {
	if width < 1 {
		return ""
	}
	if total <= 0 {
		return strings.Repeat("░", width)
	}
	filled := done * width / total
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// barWidth is how much of a card's width the bar takes. The rest of the row
// carries the count and whatever else fits.
const barWidth = 28

// goalRows is the four lines an objective's card carries, dressed.
func (b Board) goalRows(g reading.Goal, width int) []string {
	var rows []string
	mark := goalMarks(g, b.glyphs)
	if glyph := StandingGlyph(g.State, b.glyphs); glyph != "" {
		mark = strings.Replace(mark, glyph,
			lipgloss.NewStyle().Foreground(b.skin.palette.Standing(g.State)).Render(glyph), 1)
	}
	rows = append(rows, Spread(b.skin.Faint.Render(strconv.Itoa(g.Number)), mark, width))
	for _, line := range Wrap(g.Title, width, titleLines) {
		rows = append(rows, b.skin.Bright.Render(line))
	}
	return append(rows, b.goalFacts(g, width))
}

// goalMarks is the right-hand end of an objective's first row.
func goalMarks(g reading.Goal, gl Glyphs) string {
	var found []string
	if g.Described {
		found = append(found, gl.Described)
	}
	if glyph := StandingGlyph(g.State, gl); glyph != "" {
		found = append(found, glyph)
	}
	return strings.Join(found, " ")
}

// factsRow is the bar, the count, and whatever else the width allows.
func factsRow(g reading.Goal, gl Glyphs, width int) string {
	bar := Bar(g.Done, g.Total, min(barWidth, width))
	told := []string{}
	if g.Total > 0 {
		told = append(told, fmt.Sprintf("%d/%d", g.Done, g.Total))
	}
	for _, label := range g.Labels {
		told = append(told, gl.Label+" "+label)
	}
	rest := Join("  ", width-lipgloss.Width(bar)-2, told...)
	if rest == "" {
		return bar
	}
	return bar + "  " + rest
}

// Written is a day as a person reads it, or nothing when there is none.
func Written(day time.Time) string { return domain.FormatDay(day) }
