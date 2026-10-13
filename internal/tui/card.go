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

func marks(c engine.Card, g Glyphs) string {
	var found []string
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

func facts(c engine.Card, g Glyphs, s Skin, now time.Time, width int) string {
	if engine.Stopped(c.State) {
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
	return withMore(told, dressed, s, width)
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

func inTheWay(c engine.Card, g Glyphs) string {
	if way := c.InTheWay(); way != "" {
		return g.InTheWay + " " + way
	}
	return ""
}

// Rows is the four lines a card carries, before anything is drawn around them.
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
