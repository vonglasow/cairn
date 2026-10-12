package gui

// The third page: the time spent, a week or a month at a time. How long the
// work took was a table the command line printed; here a column a day,
// Monday first, each day's total and each task worked on with its time and a
// bar as long as its share of the longest stretch of the week — every task,
// a day scrolling to its last — or a month's six weeks, each day saying what
// it held. A strip along the top walks it, as the arrows and t do, and goes
// between the two. See docs/gui.md.

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
	// arrowWide is how wide an arrow of the strip is, rangeWide a range.
	arrowWide = 32
	rangeWide = 72
)

// timeState is what the time page keeps: the Monday of the week it shows,
// or the first day of the month, the time spent over it, whether a month is
// shown, and how far down each day of a week its tasks are scrolled.
type timeState struct {
	week    time.Time
	sheet   engine.Timesheet
	monthly bool
	dayFrom [7]int
}

// Week is the Monday of the week the time page shows, or the first day of
// the month it shows.
func (w Window) Week() time.Time { return w.week }

// Monthly says the time page shows a month rather than a week.
func (w Window) Monthly() bool { return w.monthly }

// MonthShown is the first day of the month the time page shows.
func (w Window) MonthShown() time.Time { return w.week }

// firstOf is the first day of the month a day is in.
func firstOf(day time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, time.UTC)
}

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
	since, until := w.week, w.week.AddDate(0, 0, 6)
	if w.monthly {
		grid := MonthGrid(w.week)
		since, until = grid[0], grid[len(grid)-1]
	}
	// The engine refuses what has not happened yet: a period in progress is
	// read up to today, and one wholly to come holds nothing to read.
	if today := w.today(); until.After(today) {
		until = today
	}
	if since.After(until) {
		w.sheet = engine.Timesheet{}
		return w
	}
	asked := engine.Input{Given: map[string]string{
		"since": domain.FormatDay(since),
		"until": domain.FormatDay(until),
	}}
	answer, err := w.work.Ask(engine.TimesheetShow, asked)
	if err != nil {
		return w.failing(err)
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
		if w.monthly {
			w.week = monthOn(w.week, -1)
		} else {
			w.week = w.week.AddDate(0, 0, -7)
		}
	case "right":
		if w.monthly {
			w.week = monthOn(w.week, 1)
		} else {
			w.week = w.week.AddDate(0, 0, 7)
		}
	case "t":
		w.week = mondayOf(w.today())
		if w.monthly {
			w.week = firstOf(w.today())
		}
	case "m":
		w = w.turned()
	}
	w.dayFrom = [7]int{}
	return w.readingWeek()
}

// turned goes from a week to its month — the month of today when today is
// in the week, else the month of its Thursday, which holds the most of its
// days — or from a month to the week today is in, or to its first week when
// today is elsewhere.
func (w Window) turned() Window {
	w.monthly = !w.monthly
	if w.monthly {
		today := w.today()
		if today.Before(w.week) || !today.Before(w.week.AddDate(0, 0, 7)) {
			today = w.week.AddDate(0, 0, 3)
		}
		w.week = firstOf(today)
		return w
	}
	if firstOf(w.today()).Equal(w.week) {
		w.week = mondayOf(w.today())
		return w
	}
	w.week = mondayOf(w.week)
	return w
}

// weekKeys is what the time page answers itself; anything else that is not
// the window's own — tab, the backups — reaches nothing on it, since there
// is no card on the page to act on.
func weekKeys(name string) (mine, shared bool) {
	switch name {
	case "left", "right", "t", "m":
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

// TimeStrip is the strip along the top of the time page: the board's whole
// width, where the columns' headers are on the other pages.
func TimeStrip(l Layout) Box {
	left, right := l.Column[0].X, l.Column[Columns-1].X+l.Column[Columns-1].W
	return Box{X: left, Y: l.Head[0].Y, W: right - left, H: barHigh}
}

// TimeChip is where a chip of the strip is: the arrows at its left, the
// ranges at its right, Today before them.
func TimeChip(l Layout, name string) Box {
	strip := TimeStrip(l)
	at := func(x, wide int) Box { return Box{X: x, Y: strip.Y, W: wide, H: strip.H} }
	end := strip.X + strip.W
	switch name {
	case "‹":
		return at(strip.X, arrowWide)
	case "›":
		return at(strip.X+arrowWide+between, arrowWide)
	case "Month":
		return at(end-rangeWide, rangeWide)
	case "Week":
		return at(end-2*rangeWide-between, rangeWide)
	case "Today":
		return at(end-3*rangeWide-2*between, rangeWide)
	}
	return Box{}
}

// TimeChips is every chip of the strip, in the order they are drawn.
var TimeChips = []string{"‹", "›", "Today", "Week", "Month"}

// TimeSaid is where the strip says what is shown, between the arrows and
// the ranges.
func TimeSaid(l Layout) Box {
	after, today := TimeChip(l, "›"), TimeChip(l, "Today")
	x := after.X + arrowWide + padding
	return Box{X: x, Y: after.Y, W: today.X - padding - x, H: after.H}
}

// TimeSays is what is shown, as the strip says it: a week's first and last
// days, or a month.
func TimeSays(w Window) string {
	if w.monthly {
		return w.week.Format("January 2006")
	}
	last := w.week.AddDate(0, 0, 6)
	if w.week.Year() != last.Year() {
		return w.week.Format("2 Jan 2006") + " – " + last.Format("2 Jan 2006")
	}
	if w.week.Month() != last.Month() {
		return w.week.Format("2 Jan") + " – " + last.Format("2 Jan 2006")
	}
	return w.week.Format("2") + " – " + last.Format("2 Jan 2006")
}

// WeekDayAt is where the i-th day of the week is laid out: the board's whole
// width, from under the strip to the foot of the band, cut in seven.
func WeekDayAt(l Layout, i int) Box {
	strip := TimeStrip(l)
	wide, top := (strip.W-6*gap)/7, strip.Y+strip.H+between
	return Box{X: strip.X + i*(wide+gap), Y: top, W: wide, H: l.Band.Y + l.Band.H - top}
}

// DayRoom is how many tasks a day of a week shows at once.
func DayRoom(day Box) int { return max(1, (day.H-2*lineHigh-padding)/weekRow) }

// DayFrom is the first task of the i-th day of the week in sight.
func (w Window) DayFrom(i int) int { return w.dayFrom[i] }

// DayShown is the tasks of the i-th day of the week in sight.
func (w Window) DayShown(i int) []engine.TimesheetTask {
	tasks, from := w.timeOn(i).Tasks, w.dayFrom[i]
	return tasks[min(from, len(tasks)):min(len(tasks), from+DayRoom(WeekDayAt(w.Layout(), i)))]
}

// scrollingDay turns the day of the week under the pointer, a task a notch,
// between its first task and the last screenful.
func (w Window) scrollingDay(p Point, by int) Window {
	l := w.Layout()
	for i := range 7 {
		day := WeekDayAt(l, i)
		if !day.In(p) {
			continue
		}
		// A task's row at a time, what did not come to one kept for the
		// rest of the turn.
		total := w.spareChoosing + by
		rows := total / weekRow
		w.spareChoosing = total % weekRow
		w.dayFrom[i] = max(0, min(w.dayFrom[i]+rows, len(w.timeOn(i).Tasks)-DayRoom(day)))
		return w
	}
	return w
}

// MonthWeekdayAt is where a weekday's name is written over a month.
func MonthWeekdayAt(l Layout, i int) Box {
	cell := MonthCellAt(l, i)
	return Box{X: cell.X, Y: cell.Y - lineHigh, W: cell.W, H: lineHigh}
}

// MonthCellAt is where the i-th day of a month's six weeks is laid out.
func MonthCellAt(l Layout, i int) Box {
	strip := TimeStrip(l)
	top := strip.Y + strip.H + between + lineHigh
	wide, high := (strip.W-6*gap)/7, (l.Band.Y+l.Band.H-top-5*gap)/6
	return Box{X: strip.X + (i%7)*(wide+gap), Y: top + (i/7)*(high+gap), W: wide, H: high}
}

// CellRoom is how many lines a day of a month has under its own.
func CellRoom(cell Box) int { return max(1, (cell.H-lineHigh-padding)/lineHigh) }

// CellLines is what a day of a month says of its tasks in the room it has:
// as many as fit, and how many more there are on the last line when they do
// not all fit.
func CellLines(tasks []engine.TimesheetTask, room int) []string {
	var lines []string
	shown := len(tasks)
	if shown > room {
		shown = room - 1
	}
	for _, one := range tasks[:shown] {
		lines = append(lines, TaskTimeSays(one))
	}
	if shown < len(tasks) {
		lines = append(lines, "+"+strconv.Itoa(len(tasks)-shown)+" more")
	}
	return lines
}

// MonthDay is the sheet's day for the i-th day of the month's six weeks.
func (w Window) MonthDay(i int) engine.TimesheetDay {
	day := domain.FormatDay(MonthGrid(w.week)[i])
	for _, d := range w.sheet.Days {
		if d.Day == day {
			return d
		}
	}
	return engine.TimesheetDay{Day: day}
}

// clickingTime is what a click does on the time page: a chip of the strip
// does what its key does, and a day of a month shows its week.
func (w Window) clickingTime(p Point) Window {
	l := w.Layout()
	keyOf := map[string]string{"‹": "left", "›": "right", "Today": "t"}
	for _, chip := range TimeChips {
		if !TimeChip(l, chip).In(p) {
			continue
		}
		if key, ok := keyOf[chip]; ok {
			return w.onWeek(key)
		}
		if (chip == "Month") != w.monthly {
			return w.onWeek("m")
		}
		return w
	}
	if !w.monthly {
		return w
	}
	for i, day := range MonthGrid(w.week) {
		if MonthCellAt(l, i).In(p) {
			w.monthly, w.week, w.dayFrom = false, mondayOf(day), [7]int{}
			return w.readingWeek()
		}
	}
	return w
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

// DayTrack is the bar beside a day of a week holding more tasks than it
// shows, along the tasks it shows; nothing when they all fit.
func DayTrack(day Box, all int) (Box, bool) {
	room := DayRoom(day)
	if all <= room {
		return Box{}, false
	}
	return Box{X: day.X + day.W - barWide, Y: WeekRowAt(day, 0).Y, W: barWide, H: room * weekRow}, true
}

// CellHead is what a day of a month says at its top: its number, and its
// total once anything was spent on it.
func CellHead(day time.Time, held engine.TimesheetDay) string {
	if held.Minutes <= 0 {
		return strconv.Itoa(day.Day())
	}
	return strconv.Itoa(day.Day()) + "   " + DaySays(held)
}

// CellHeadAt is where a day of a month says its number and its total.
func CellHeadAt(cell Box) Box { return Row(cell, cell.Y+inset, lineHigh) }

// CellLineAt is where the j-th line under a day of a month's head goes.
func CellLineAt(cell Box, j int) Box { return Row(cell, cell.Y+inset+(j+1)*lineHigh, lineHigh) }

// sameMonth says two days are in the same month of the same year.
func sameMonth(a, b time.Time) bool { return a.Year() == b.Year() && a.Month() == b.Month() }
