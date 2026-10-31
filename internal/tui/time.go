package tui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
)

// Sheeted carries the time spent over the week the tab shows.
type Sheeted struct {
	Sheet engine.Timesheet
	Err   error
}

// Answered carries what a command said, and the board read back after it.
type Answered struct {
	Text string
	Looked
}

func (b Board) walkingWeeks(key string) (Board, bool) {
	switch key {
	case "left", "h":
		b.week = b.week.AddDate(0, 0, -7)
	case "right", "l":
		b.week = b.week.AddDate(0, 0, 7)
	case "t":
		b.week = engine.Monday(b.clock())
	case "j", "down":
		b.timeAt++
		return b, true
	case "k", "up":
		b.timeAt = max(0, b.timeAt-1)
		return b, true
	default:
		return b, false
	}
	b.timeAt = 0
	return b, true
}

func (b Board) readingWeek() tea.Cmd {
	since, until, today := b.week, b.week.AddDate(0, 0, 6), engine.DayOf(b.clock())
	return func() tea.Msg {
		sheet, err := engine.ReadPeriod(b.work, since, until, today)
		return Sheeted{Sheet: sheet, Err: err}
	}
}

func (b Board) timeWeek(l Layout) []string {
	var lines []string
	for i := range b.pages.WeekDays() {
		day := b.week.AddDate(0, 0, i)
		on := timeOn(b.sheet, day)
		lines = append(lines, " "+b.skin.Title.Render(day.Format("Mon 02/01"))+"  "+b.skin.Faint.Render(on.Total))
		for _, one := range on.Tasks {
			lines = append(lines, "    "+Clip(one, b.width-5))
		}
	}
	room := max(1, l.Top+l.Rows+2)
	from := min(b.timeAt, max(0, len(lines)-room))
	return lines[from:min(len(lines), from+room)]
}

type spent struct {
	Total string
	Tasks []string
}

func timeOn(sheet engine.Timesheet, day time.Time) spent {
	on := sheet.On(day)
	out := spent{Total: on.Says()}
	if out.Total == "" {
		out.Total = "0m"
	}
	for _, one := range on.Tasks {
		out.Tasks = append(out.Tasks, one.Says())
	}
	return out
}

func (b Board) answering(text string) Board {
	b.said, b.saidAt = text, 0
	if text != "" {
		b.wasScreen, b.screen = OnTheBoard, OnSaid
	}
	return b
}

func (b Board) drawSaid() string {
	lines := strings.Split(strings.TrimRight(b.said, "\n"), "\n")
	room := max(1, b.height-3)
	from := min(b.saidAt, max(0, len(lines)-room))
	shown := lines[from:min(len(lines), from+room)]
	out := append([]string{b.skin.Title.Render(" what it answered"), ""}, shown...)
	return strings.Join(append(out, b.skin.Faint.Render(" esc  back")), "\n")
}
