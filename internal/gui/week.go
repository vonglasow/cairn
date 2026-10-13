package gui

import (
	"strconv"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

const (
	weekRow   = 40
	barHighOf = 6
	arrowWide = 32
	rangeWide = 72
)

type timeState struct {
	week    time.Time
	sheet   engine.Timesheet
	monthly bool
	dayFrom [7]int
}

// Monthly says the time page shows a month rather than a week.
func (w Window) Monthly() bool { return w.monthly }

func (w Window) readingWeek() Window {
	if w.week.IsZero() {
		w.week = engine.Monday(w.today())
	}
	since, until := w.week, w.week.AddDate(0, 0, 6)
	if w.monthly {
		grid := engine.MonthGrid(w.week)
		since, until = grid[0], grid[len(grid)-1]
	}
	sheet, err := engine.ReadPeriod(w.work, since, until, w.today())
	if err != nil {
		return w.failing(err)
	}
	w.sheet = sheet
	return w
}

func (w Window) onWeek(name string) Window {
	switch name {
	case "left":
		if w.monthly {
			w.week = engine.MonthOn(w.week, -1)
		} else {
			w.week = w.week.AddDate(0, 0, -7)
		}
	case "right":
		if w.monthly {
			w.week = engine.MonthOn(w.week, 1)
		} else {
			w.week = w.week.AddDate(0, 0, 7)
		}
	case "t":
		w.week = engine.Monday(w.today())
		if w.monthly {
			w.week = engine.FirstOfMonth(w.today())
		}
	case "m":
		w = w.turned()
	}
	w.dayFrom = [7]int{}
	return w.readingWeek()
}

func (w Window) turned() Window {
	w.monthly = !w.monthly
	if w.monthly {
		today := w.today()
		if today.Before(w.week) || !today.Before(w.week.AddDate(0, 0, 7)) {
			today = w.week.AddDate(0, 0, 3)
		}
		w.week = engine.FirstOfMonth(today)
		return w
	}
	if engine.FirstOfMonth(w.today()).Equal(w.week) {
		w.week = engine.Monday(w.today())
		return w
	}
	w.week = engine.Monday(w.week)
	return w
}

func weekKeys(name string) (mine, shared bool) {
	switch name {
	case "left", "right", "t", "m":
		return true, false
	case "tab", "shift+tab", "cmd+b", "shift+b":
		return false, true
	}
	return false, false
}

// TimeStrip is the strip along the top of the time page: the board's whole width, where the columns' headers are on the other pages.
func TimeStrip(l Layout) Box {
	left, right := l.Column[0].X, l.Column[engine.Columns-1].X+l.Column[engine.Columns-1].W
	return Box{X: left, Y: l.Head[0].Y, W: right - left, H: barHigh}
}

// TimeChip is where a chip of the strip is: the arrows at its left, the ranges at its right, Today before them.
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

// TimeSaid is where the strip says what is shown, between the arrows and the ranges.
func TimeSaid(l Layout) Box {
	after, today := TimeChip(l, "›"), TimeChip(l, "Today")
	x := after.X + arrowWide + padding
	return Box{X: x, Y: after.Y, W: today.X - padding - x, H: after.H}
}

// TimeSays is what is shown, as the strip says it: a week's first and last days, or a month.
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

// WeekDayAt is where the i-th day of the week is laid out: the board's whole width, from under the strip to the foot of the band, cut in seven.
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
	tasks, from := w.sheet.On(w.week.AddDate(0, 0, i)).Tasks, w.dayFrom[i]
	return tasks[min(from, len(tasks)):min(len(tasks), from+DayRoom(WeekDayAt(w.Layout(), i)))]
}

func (w Window) scrollingDay(p Point, by int) Window {
	l := w.Layout()
	for i := range 7 {
		day := WeekDayAt(l, i)
		if !day.In(p) {
			continue
		}
		total := w.spareChoosing + by
		rows := total / weekRow
		w.spareChoosing = total % weekRow
		w.dayFrom[i] = max(0, min(w.dayFrom[i]+rows, len(w.sheet.On(w.week.AddDate(0, 0, i)).Tasks)-DayRoom(day)))
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

// CellLines is what a day of a month says of its tasks in the room it has: as many as fit, and how many more there are on the last line when they do not all fit.
func CellLines(tasks []engine.TimesheetTask, room int) []string {
	var lines []string
	shown := len(tasks)
	if shown > room {
		shown = room - 1
	}
	for _, one := range tasks[:shown] {
		lines = append(lines, one.Says())
	}
	if shown < len(tasks) {
		lines = append(lines, "+"+strconv.Itoa(len(tasks)-shown)+" more")
	}
	return lines
}

// MonthDay is the sheet's day for the i-th day of the month's six weeks.
func (w Window) MonthDay(i int) engine.TimesheetDay {
	day := domain.FormatDay(engine.MonthGrid(w.week)[i])
	for _, d := range w.sheet.Days {
		if d.Day == day {
			return d
		}
	}
	return engine.TimesheetDay{Day: day}
}

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
	for i, day := range engine.MonthGrid(w.week) {
		if MonthCellAt(l, i).In(p) {
			w.monthly, w.week, w.dayFrom = false, engine.Monday(day), [7]int{}
			return w.readingWeek()
		}
	}
	return w
}

// WeekHeadAt is where a day says which it is and its total, at its top.
func WeekHeadAt(day Box) Box { return Row(day, day.Y+padding, lineHigh) }

// WeekRowAt is where the j-th task of a day is written, under its heading and its total.
func WeekRowAt(day Box, j int) Box {
	return Box{X: day.X + padding, Y: day.Y + 2*lineHigh + padding + j*weekRow, W: day.W - 2*padding, H: weekRow}
}

// WeekBar is a task's bar: under its line, as long as its share of the longest stretch of the week.
func WeekBar(row Box, minutes, longest int) Box {
	return Box{X: row.X, Y: row.Y + lineHigh + between, W: row.W * minutes / max(1, longest), H: barHighOf}
}

// DayTrack is the bar beside a day of a week holding more tasks than it shows, along the tasks it shows; nothing when they all fit.
func DayTrack(day Box, all int) (Box, bool) {
	room := DayRoom(day)
	if all <= room {
		return Box{}, false
	}
	return Box{X: day.X + day.W - barWide, Y: WeekRowAt(day, 0).Y, W: barWide, H: room * weekRow}, true
}

// CellHead is what a day of a month says at its top: its number, and its total once anything was spent on it.
func CellHead(day time.Time, held engine.TimesheetDay) string {
	if held.Minutes <= 0 {
		return strconv.Itoa(day.Day())
	}
	return strconv.Itoa(day.Day()) + "   " + held.Says()
}

// CellHeadAt is where a day of a month says its number and its total.
func CellHeadAt(cell Box) Box { return Row(cell, cell.Y+inset, lineHigh) }

// CellLineAt is where the j-th line under a day of a month's head goes.
func CellLineAt(cell Box, j int) Box { return Row(cell, cell.Y+inset+(j+1)*lineHigh, lineHigh) }

func sameMonth(a, b time.Time) bool { return a.Year() == b.Year() && a.Month() == b.Month() }
