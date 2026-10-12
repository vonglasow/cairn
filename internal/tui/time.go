package tui

// The time spent and what a command answers, on the terminal board. The window
// gained a page for the time spent and a way to the backups; the terminal had
// neither, and a command typed at its prompt said nothing of what it
// answered. A third tab lays the week out, b and B back the store up and list
// the backups, and a command's answer is said until esc puts it away. See
// docs/tui.md.

import (
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/domain"
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

// mondayOf is the Monday of the week a moment is in, at midnight.
func mondayOf(at time.Time) time.Time {
	day := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.UTC)
	return day.AddDate(0, 0, -((int(day.Weekday()) + 6) % 7))
}

// walkingWeeks is what a key does on the time tab: the arrows and h l walk the
// weeks, t comes back to today's. Anything else is not the tab's.
func (b Board) walkingWeeks(key string) (Board, bool) {
	switch key {
	case "left", "h":
		b.week = b.week.AddDate(0, 0, -7)
	case "right", "l":
		b.week = b.week.AddDate(0, 0, 7)
	case "t":
		b.week = mondayOf(b.clock())
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

// readingWeek asks how long the work took over the week shown, up to today:
// the engine refuses what has not happened yet.
func (b Board) readingWeek() tea.Cmd {
	since, until := b.week, b.week.AddDate(0, 0, 6)
	if today := mondayOf(b.clock()).AddDate(0, 0, (int(b.clock().Weekday())+6)%7); until.After(today) {
		until = today
	}
	return func() tea.Msg {
		if since.After(until) {
			return Sheeted{}
		}
		answer, err := b.work.Ask(engine.TimesheetShow, engine.Input{Given: map[string]string{
			"since": domain.FormatDay(since), "until": domain.FormatDay(until)}})
		if err != nil || answer.Sheet == nil {
			return Sheeted{Err: err}
		}
		sheet, _ := engine.Of(engine.TimesheetShow, answer).(engine.Timesheet)
		return Sheeted{Sheet: sheet}
	}
}

// timeWeek is the week laid out down the screen: each day, its total, and a
// line for each task worked on with its time — read whole rather than cut to a
// seventh of a terminal's width — scrolled with j and k.
func (b Board) timeWeek(l Layout) []string {
	var lines []string
	for i := range 7 {
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

// spent is one day of the week as the tab writes it.
type spent struct {
	Total string
	Tasks []string
}

// timeOn is what a day of the sheet says: its total, and a line a task.
func timeOn(sheet engine.Timesheet, day time.Time) spent {
	for _, d := range sheet.Days {
		if d.Day != domain.FormatDay(day) {
			continue
		}
		out := spent{Total: domain.FormatDuration(time.Duration(d.Minutes) * time.Minute)}
		for _, one := range d.Tasks {
			out.Tasks = append(out.Tasks, domain.FormatDuration(time.Duration(one.Minutes)*time.Minute)+"  #"+strconv.Itoa(one.Number)+" "+one.Title)
		}
		return out
	}
	return spent{Total: "0m"}
}

// answering opens what a command answered, over the board.
func (b Board) answering(text string) Board {
	b.said, b.saidAt = text, 0
	if text != "" {
		b.wasScreen, b.screen = OnTheBoard, OnSaid
	}
	return b
}

// drawSaid is what a command answered, a screen of it, scrolled with the
// arrows, until esc.
func (b Board) drawSaid() string {
	lines := strings.Split(strings.TrimRight(b.said, "\n"), "\n")
	room := max(1, b.height-3)
	from := min(b.saidAt, max(0, len(lines)-room))
	shown := lines[from:min(len(lines), from+room)]
	out := append([]string{b.skin.Title.Render(" what it answered"), ""}, shown...)
	return strings.Join(append(out, b.skin.Faint.Render(" esc  back")), "\n")
}
