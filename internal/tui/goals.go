package tui

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/objective"

	"charm.land/lipgloss/v2"
)

// StandingGlyph answers the mark an objective's state is drawn with.
func StandingGlyph(state string, g Glyphs) string {
	switch state {
	case string(objective.Active):
		return g.Active
	case string(objective.OnHold):
		return g.OnHold
	case string(objective.Reached):
		return g.Reached
	case string(objective.Dropped):
		return g.Dropped
	default:
		return ""
	}
}

func (b Board) progress(done, total, width int) string {
	plain := Bar(done, total, width)
	filled := strings.Count(plain, "█")
	return b.skin.Filled.Render(strings.Repeat("█", filled)) +
		b.skin.Empty.Render(strings.Repeat("░", len([]rune(plain))-filled))
}

func (b Board) goalFacts(g engine.Goal, width int) string {
	room := min(barWidth, width)
	told := []string{}
	dressed := []string{}
	if g.Total > 0 {
		count := g.Counted()
		told = append(told, count)
		dressed = append(dressed, b.skin.Bright.Render(count))
	}
	if g.Spent > 0 {
		spent := domain.FormatDuration(g.Spent)
		told = append(told, spent)
		dressed = append(dressed, b.skin.Faint.Render(spent))
	}
	if day := Written(g.Deadline); day != "" {
		told = append(told, day)
		dressed = append(dressed, b.skin.Faint.Render(day))
	}
	if waits := b.waitsOn(g, false); waits != "" {
		told = append(told, waits)
		dressed = append(dressed, b.skin.Band.Render(waits))
	}
	rest := dress(Join("  ", width-room-2, told...), told, dressed)
	if rest == "" {
		return b.progress(g.Done, g.Total, room)
	}
	return b.progress(g.Done, g.Total, room) + "  " + rest
}

func (b Board) waitsOn(g engine.Goal, numbered bool) string {
	var names []string
	for _, n := range g.WaitsOn {
		on, _ := b.snap.Goal(n)
		if numbered {
			names = append(names, "#"+strconv.Itoa(n)+" "+on.Title)
			continue
		}
		names = append(names, on.Title)
	}
	if len(names) == 0 {
		return ""
	}
	return "waits on " + strings.Join(names, ", ")
}

func (b Board) typeLabels(g engine.Goal) (string, string) {
	var told, dressed []string
	for _, label := range g.Labels {
		tag := b.glyphs.Label + " " + label
		told = append(told, tag)
		dressed = append(dressed, b.skin.Says.Render(tag))
	}
	return strings.Join(told, "  "), strings.Join(dressed, "  ")
}

// Bar draws how far along an objective is.
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

const barWidth = 28

const rowIndent = 3

func (b Board) goalRows(g engine.Goal, width int) []string {
	mark := b.skin.palette.Standing(g.State).Style().Render(pad(StandingGlyph(g.State, b.glyphs), 1))
	told, dressed := b.typeLabels(g)
	if lipgloss.Width(told) > width/3 {
		told, dressed = "", ""
	}
	hidden := ""
	if slices.Contains(b.folded, g.Number) {
		hidden = "  +" + strconv.Itoa(len(g.Children))
	}
	room := max(1, width-rowIndent-len(strconv.Itoa(g.Number))-2-lipgloss.Width(told)-2-len(hidden))
	title := b.skin.Bright.Render(Clip(g.Title, room))
	if g.Number == b.goal && b.tab == Objectives {
		title = b.skin.Title.Render(Clip(g.Title, room))
	}
	head := mark + "  " + b.skin.Faint.Render(strconv.Itoa(g.Number)) + "  " + title + b.skin.Faint.Render(hidden)
	return []string{
		Spread(head, dressed, width),
		strings.Repeat(" ", rowIndent) + b.goalFacts(g, width-rowIndent),
	}
}

// Written is a day as a person reads it, or nothing when there is none.
func Written(day time.Time) string { return domain.FormatDay(day) }
