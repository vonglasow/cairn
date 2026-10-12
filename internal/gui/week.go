package gui

// The third page: the time spent, a week at a time. How long the work took
// was a table the command line printed; here a column a day, Monday first,
// each day's total and each task worked on with its time and a bar as long
// as its share of the longest stretch of the week. The arrows walk the weeks
// and t comes back to this one. See docs/gui.md.

import (
	"strconv"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

const (
	// weekRow is the room one task takes in a day: its line, and its bar.
	weekRow = 40
	// barHighOf is how thick a task's bar is.
	barHighOf = 6
)

// Week is the Monday of the week the time page shows.
func (w Window) Week() time.Time { return w.week }

// Sheet is the time spent over that week, as the engine reckoned it.
func (w Window) Sheet() engine.Timesheet { return w.sheet }

// mondayOf is the Monday of the week a day is in.
func mondayOf(day time.Time) time.Time {
	day = dayOf(day)
	return day.AddDate(0, 0, -((int(day.Weekday()) + 6) % 7))
}

// readingWeek asks how long the work took over the week shown.
func (w Window) readingWeek() Window {
	if w.week.IsZero() {
		w.week = mondayOf(w.today())
	}
	asked := engine.Input{Given: map[string]string{
		"since": domain.FormatDay(w.week),
		"until": domain.FormatDay(w.week.AddDate(0, 0, 6)),
	}}
	answer, err := w.work.Ask(engine.TimesheetShow, asked)
	if err != nil {
		w.err, w.wrongFor = err, 0
		return w
	}
	w.sheet = engine.Timesheet{}
	if answer.Sheet != nil {
		w.sheet, _ = engine.Of(engine.TimesheetShow, answer).(engine.Timesheet)
	}
	return w
}

// onWeek is what a key does on the time page: the arrows walk the weeks, and
// t comes back to the one today is in.
func (w Window) onWeek(name string) Window {
	switch name {
	case "left":
		w.week = w.week.AddDate(0, 0, -7)
	case "right":
		w.week = w.week.AddDate(0, 0, 7)
	case "t":
		w.week = mondayOf(w.today())
	}
	return w.readingWeek()
}

// weekKeys is what the time page answers itself; anything else that is not
// the window's own — tab, the backups — reaches nothing on it, since there
// is no card on the page to act on.
func weekKeys(name string) (mine, shared bool) {
	switch name {
	case "left", "right", "t":
		return true, false
	case "tab", "shift+tab", "cmd+b", "shift+b":
		return false, true
	}
	return false, false
}

// timeOn is the day of the week shown that a day of the sheet is, and the
// sheet's day for each of the seven: a day nothing was done on is not in it.
func (w Window) timeOn(i int) engine.TimesheetDay {
	day := domain.FormatDay(w.week.AddDate(0, 0, i))
	for _, d := range w.sheet.Days {
		if d.Day == day {
			return d
		}
	}
	return engine.TimesheetDay{Day: day}
}

// longest is the longest stretch any task had on any day of the week: what
// every bar is a share of.
func (w Window) longest() int {
	most := 0
	for _, d := range w.sheet.Days {
		for _, one := range d.Tasks {
			most = max(most, one.Minutes)
		}
	}
	return most
}

// DaySays is a day's total, as a duration is written.
func DaySays(d engine.TimesheetDay) string {
	return domain.FormatDuration(time.Duration(d.Minutes) * time.Minute)
}

// TaskTimeSays is one task's line on a day: its time, then which task.
func TaskTimeSays(one engine.TimesheetTask) string {
	return domain.FormatDuration(time.Duration(one.Minutes)*time.Minute) + "  #" + strconv.Itoa(one.Number) + " " + one.Title
}

// WeekDayAt is where the i-th day of the week is laid out: the board's whole
// width, from its headers to the foot of the band, cut in seven.
func WeekDayAt(l Layout, i int) Box {
	left, right := l.Column[0].X, l.Column[Columns-1].X+l.Column[Columns-1].W
	wide := (right - left - 6*gap) / 7
	return Box{X: left + i*(wide+gap), Y: l.Head[0].Y, W: wide, H: l.Band.Y + l.Band.H - l.Head[0].Y}
}

// WeekHeadAt is where a day says which it is and its total, at its top.
func WeekHeadAt(day Box) Box { return Row(day, day.Y+padding, lineHigh) }

// WeekRowAt is where the j-th task of a day is written, under its heading
// and its total.
func WeekRowAt(day Box, j int) Box {
	return Box{X: day.X + padding, Y: day.Y + 2*lineHigh + padding + j*weekRow, W: day.W - 2*padding, H: weekRow}
}

// WeekBar is a task's bar: under its line, as long as its share of the
// longest stretch of the week.
func WeekBar(row Box, minutes, longest int) Box {
	return Box{X: row.X, Y: row.Y + lineHigh + between, W: row.W * minutes / max(1, longest), H: barHighOf}
}
