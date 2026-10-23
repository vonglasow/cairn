package gui

import (
	"strconv"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

const (
	weekRow   = CardHeight + lineHigh
	barHighOf = 6
	spentWide = 48
	arrowWide = 32
	rangeWide = 72
)

type timeState struct {
	week    time.Time
	sheet   engine.Timesheet
	monthly bool
	daily   bool
	dayFrom [7]int
}

// Monthly says the time page shows a month rather than a week.
func (w Window) Monthly() bool { return w.monthly }

// Daily says the time page shows one day rather than a week.
func (w Window) Daily() bool { return w.daily }

// TimeShown is the time spent as the page shows it: every task of the period, or only those the search reaches, each day's total counting only those.
func (w Window) TimeShown() engine.Timesheet {
	if w.looking.Empty() {
		return w.sheet
	}
	reached := map[int]bool{}
	for _, c := range engine.Found(w.snap.Cards, w.looking.Reads(), w.archived) {
		reached[c.Number] = true
	}
	shown := engine.Timesheet{}
	for _, d := range w.sheet.Days {
		kept := engine.TimesheetDay{Day: d.Day}
		for _, one := range d.Tasks {
			if reached[one.Number] {
				kept.Tasks = append(kept.Tasks, one)
				kept.Minutes += one.Minutes
			}
		}
		shown.Days = append(shown.Days, kept)
	}
	return shown
}

// DayHeld is the i-th day of the week in sight, as the page shows it.
func (w Window) DayHeld(i int) engine.TimesheetDay {
	return w.TimeShown().On(w.week.AddDate(0, 0, i))
}

func (w Window) readingWeek() Window {
	if w.week.IsZero() {
		w.week = engine.Monday(w.today())
	}
	since, until := w.week, w.week.AddDate(0, 0, 6)
	if w.daily {
		until = w.week
	}
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
		w.week = w.stepped(-1)
	case "right":
		w.week = w.stepped(1)
	case "t":
		w.week = engine.Monday(w.today())
		if w.monthly {
			w.week = engine.FirstOfMonth(w.today())
		}
		if w.daily {
			w.week = w.today()
		}
	case "m":
		if w.monthly {
			w = w.ranged("Week")
		} else {
			w = w.ranged("Month")
		}
	case "d":
		w = w.ranged("Day")
	}
	w.dayFrom = [7]int{}
	return w.readingWeek()
}

func (w Window) stepped(by int) time.Time {
	switch {
	case w.monthly:
		return engine.MonthOn(w.week, by)
	case w.daily:
		return w.week.AddDate(0, 0, by)
	}
	return w.week.AddDate(0, 0, 7*by)
}

func (w Window) ranged(to string) Window {
	today, on := w.today(), w.week
	switch {
	case w.monthly:
		if sameMonth(today, w.week) {
			on = today
		}
	case !w.daily:
		if !today.Before(w.week) && today.Before(w.week.AddDate(0, 0, 7)) {
			on = today
		} else if to == "Month" {
			on = w.week.AddDate(0, 0, 3)
		}
	}
	w.daily, w.monthly = to == "Day", to == "Month"
	switch to {
	case "Day":
		w.week = on
	case "Month":
		w.week = engine.FirstOfMonth(on)
	default:
		w.week = engine.Monday(on)
	}
	return w
}

func (w Window) ranging() string {
	switch {
	case w.monthly:
		return "Month"
	case w.daily:
		return "Day"
	}
	return "Week"
}

func weekKeys(name string) (mine, shared bool) {
	switch name {
	case "left", "right", "t", "m", "d":
		return true, false
	case "tab", "shift+tab", "cmd+b", "shift+b", "shift+l":
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
	case "Day":
		return at(end-3*rangeWide-2*between, rangeWide)
	case "Today":
		return at(end-4*rangeWide-4*between, rangeWide)
	}
	return Box{}
}

// TimeChips is every chip of the strip, in the order they are drawn.
var TimeChips = []string{"‹", "›", "Today", "Day", "Week", "Month"}

// ChipSaid is where a chip writes its name: in its middle, a line high and as wide as the name.
func ChipSaid(chip Box, wide int) Box {
	line := Box{X: chip.X, Y: chip.Y + (chip.H-lineHigh)/2, W: chip.W, H: lineHigh}
	if wide >= chip.W {
		return line
	}
	return Box{X: chip.X + (chip.W-wide)/2, Y: line.Y, W: wide, H: line.H}
}

// TimeSaid is where the strip says what is shown, between the arrows and the ranges.
func TimeSaid(l Layout) Box {
	after, today := TimeChip(l, "›"), TimeChip(l, "Today")
	x := after.X + arrowWide + padding
	return Box{X: x, Y: after.Y, W: today.X - padding - x, H: after.H}
}

// TimeSays is what is shown, as the strip says it: a week's first and last days, or a month.
func TimeSays(w Window) string {
	if w.daily {
		return w.week.Format("Monday 2 January 2006")
	}
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
	tasks, from := w.DayHeld(i).Tasks, w.dayFrom[i]
	return tasks[min(from, len(tasks)):min(len(tasks), from+DayRoom(WeekDayAt(w.Layout(), i)))]
}

// AloneShown is the tasks of the day shown alone that are in sight.
func (w Window) AloneShown() []engine.TimesheetTask {
	tasks, from := w.DayHeld(0).Tasks, w.dayFrom[0]
	return tasks[min(from, len(tasks)):min(len(tasks), from+DayRoom(DayAloneAt(w.Layout())))]
}

// DayAloneAt is where one day is laid out when it is shown alone: the whole of what a week takes.
func DayAloneAt(l Layout) Box {
	strip, first := TimeStrip(l), WeekDayAt(l, 0)
	return Box{X: strip.X, Y: first.Y, W: strip.W, H: first.H}
}

// AloneCardAt is where the j-th task of a day shown alone is drawn: the card a week draws, down the day's left and a third of it wide, so its title reads.
func AloneCardAt(alone Box, j int) Box {
	row := WeekRowAt(alone, j)
	return Box{X: row.X, Y: row.Y, W: alone.W/3 - padding, H: CardHeight}
}

// AloneSpentAt is where a task of a day shown alone says how long it took: beside its card.
func AloneSpentAt(alone Box, j int) Box {
	card := AloneCardAt(alone, j)
	return Box{X: card.X + card.W + padding, Y: card.Y + (CardHeight-gap-lineHigh)/2, W: spentWide, H: lineHigh}
}

// AloneBar is a task's bar on a day shown alone: after its time, across what is left of the day, as long as its share of the day's longest.
func AloneBar(alone Box, j, minutes, longest int) Box {
	spent := AloneSpentAt(alone, j)
	x := spent.X + spent.W
	room := alone.X + alone.W - padding - barWide - x
	return Box{X: x, Y: spent.Y + (lineHigh-barHighOf)/2, W: room * minutes / max(1, longest), H: barHighOf}
}

func (w Window) scrollingDay(p Point, by int) Window {
	l := w.Layout()
	if w.daily {
		alone := DayAloneAt(l)
		if !alone.In(p) {
			return w
		}
		total := w.spareChoosing + by
		w.spareChoosing = total % weekRow
		w.dayFrom[0] = max(0, min(w.dayFrom[0]+total/weekRow, len(w.DayHeld(0).Tasks)-DayRoom(alone)))
		return w
	}
	for i := range 7 {
		day := WeekDayAt(l, i)
		if !day.In(p) {
			continue
		}
		total := w.spareChoosing + by
		rows := total / weekRow
		w.spareChoosing = total % weekRow
		w.dayFrom[i] = max(0, min(w.dayFrom[i]+rows, len(w.DayHeld(i).Tasks)-DayRoom(day)))
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
	shown := cellShown(tasks, room)
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
	for _, d := range w.TimeShown().Days {
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
		if chip != w.ranging() {
			w = w.ranged(chip)
			w.dayFrom = [7]int{}
			return w.readingWeek()
		}
		return w
	}
	if number := w.timeCardAt(p); number != 0 {
		if w.goneOnTime(number) {
			return w.telling("task " + strconv.Itoa(number) + " was deleted")
		}
		return w.choosingAt(number)
	}
	if !w.monthly && !w.daily {
		for i := range 7 {
			if WeekHeadAt(WeekDayAt(l, i)).In(p) {
				w.daily, w.week, w.dayFrom = true, w.week.AddDate(0, 0, i), [7]int{}
				return w.readingWeek()
			}
		}
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

// WeekRowAt is where the j-th task of a day is laid out, under its heading and its total: its card, then the line saying how long it took that day.
func WeekRowAt(day Box, j int) Box {
	return Box{X: day.X + padding, Y: day.Y + 2*lineHigh + padding + j*weekRow, W: day.W - 2*padding, H: weekRow}
}

// DayCardAt is where the j-th task of a day is drawn as the card the board draws it as.
func DayCardAt(day Box, j int) Box {
	row := WeekRowAt(day, j)
	return Box{X: row.X, Y: row.Y, W: row.W, H: CardHeight}
}

// DaySpentAt is where a task of a day says how long it took that day: under its card, before its bar.
func DaySpentAt(row Box) Box {
	return Box{X: row.X, Y: row.Y + CardHeight - gap, W: spentWide, H: lineHigh}
}

// WeekBar is a task's bar: under its card, after its time, as long as its share of the longest stretch of the week.
func WeekBar(row Box, minutes, longest int) Box {
	spent := DaySpentAt(row)
	room := row.W - spent.W
	return Box{X: spent.X + spent.W, Y: spent.Y + (lineHigh-barHighOf)/2, W: room * minutes / max(1, longest), H: barHighOf}
}

// DayCard is a task of the time page as the board holds it, so it is drawn the way the board draws it; a task the board no longer holds is drawn from what the sheet says of it.
func (w Window) DayCard(one engine.TimesheetTask) engine.Card {
	if c, found := w.snap.Card(one.Number); found {
		return c
	}
	return engine.Card{Number: one.Number, Title: one.Shown()}
}

func (w Window) goneOnTime(number int) bool {
	for _, day := range w.sheet.Days {
		for _, one := range day.Tasks {
			if one.Number == number {
				return one.Deleted
			}
		}
	}
	return false
}

func cellShown(tasks []engine.TimesheetTask, room int) int {
	if len(tasks) > room {
		return room - 1
	}
	return len(tasks)
}

func (w Window) timeCardAt(p Point) int {
	l := w.Layout()
	if w.daily {
		for j, one := range w.AloneShown() {
			if AloneCardAt(DayAloneAt(l), j).In(p) {
				return one.Number
			}
		}
		return 0
	}
	if w.monthly {
		for i := range engine.MonthGrid(w.week) {
			cell := MonthCellAt(l, i)
			if !cell.In(p) {
				continue
			}
			tasks := w.MonthDay(i).Tasks
			for j := range cellShown(tasks, CellRoom(cell)) {
				if CellLineAt(cell, j).In(p) {
					return tasks[j].Number
				}
			}
			return 0
		}
		return 0
	}
	for i := range 7 {
		for j, one := range w.DayShown(i) {
			if DayCardAt(WeekDayAt(l, i), j).In(p) {
				return one.Number
			}
		}
	}
	return 0
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
