package tui

import (
	"cairn/internal/board"
	"cairn/internal/engine"
	"cairn/internal/engine/item"
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

const titleLines = 2

// StoppedGlyph answers the mark one of the three stopped states is drawn with.
func StoppedGlyph(state string, g Glyphs) string {
	switch state {
	case string(item.Blocked):
		return g.Blocked
	case string(item.Watching):
		return g.Watching
	case string(item.Paused):
		return g.Paused
	default:
		return ""
	}
}

func marks(c board.Card, g Glyphs) string {
	var found []string
	for _, m := range board.MarksOf(c) {
		found = append(found, g.Of(m))
	}
	return strings.Join(found, " ")
}

func facts(c board.Card, g Glyphs, s Skin, now time.Time, pages engine.Pages, width int) string {
	var told, dressed []string
	said := board.FactsOf(c, now, pages)
	for _, f := range said {
		told = append(told, g.said(f))
		dressed = append(dressed, s.Of(f.Tone).Render(g.said(f)))
	}
	if row := withMore(told, dressed, s, width); row != "" || len(told) == 0 || (said[0].Mark != board.InTheWay && said[0].Mark != board.Freed) {
		return row
	}
	return s.Of(said[0].Tone).Render(Clip(told[0], width))
}

func dress(kept string, told, dressed []string) string {
	if kept == "" {
		return ""
	}
	return strings.Join(dressed[:max(fit(told, lipgloss.Width(kept)), 1)], "  ")
}

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

func fit(told []string, width int) int {
	kept := Join("  ", width, told...)
	for i := range told {
		if !strings.Contains(kept, told[i]) {
			return i
		}
	}
	return len(told)
}

// Rows is the four lines a card carries, before anything is drawn around them.
func Rows(c board.Card, g Glyphs, s Skin, now time.Time, pages engine.Pages, width int) []string {
	var rows []string
	number := s.Faint.Render(strconv.Itoa(c.Number))
	mark := marks(c, g)
	if glyph := StoppedGlyph(c.State, g); glyph != "" {
		mark = strings.Replace(mark, glyph,
			s.palette.Stopped(c.State).Style().Render(glyph), 1)
	}
	rows = append(rows, Spread(number, mark, width))
	for i, line := range Wrap(c.Title, width, titleLines) {
		if i == titleLines-1 && line == "" && c.Series != nil {
			rows = append(rows, s.Faint.Render(Clip(c.Series.Rule, width)))
			continue
		}
		rows = append(rows, s.Bright.Render(line))
	}
	return append(rows, facts(c, g, s, now, pages, width))
}
