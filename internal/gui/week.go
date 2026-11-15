package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"strconv"
	"time"

	"cairn/internal/engine"
)

const (
	weekRow   = geom.CardHeight + lineHigh
	barHighOf = 6
	spentWide = 48
	arrowWide = 32
	rangeWide = 72
)

func (w Window) readingWeek() Window {
	if w.Now.IsZero() {
		return w
	}
	if w.Week.IsZero() {
		w.Week = board.Monday(w.Today())
	}
	since, until := w.Week, w.Week.AddDate(0, 0, 6)
	if w.Daily {
		until = w.Week
	}
	if w.Monthly {
		grid := board.MonthGrid(w.Week)
		since, until = grid[0], grid[len(grid)-1]
	}
	sheet, err := board.ReadPeriod(w.work, since, until, w.Today())
	if err != nil {
		return w.failing(err)
	}
	w.Sheet = sheet
	return w
}

func (w Window) onWeek(name string) Window {
	switch name {
	case "left":
		w.Week = w.Stepped(-1)
	case "right":
		w.Week = w.Stepped(1)
	case "t":
		w.Grid = board.ViewDays
		w.Week = board.Monday(w.Today())
		if w.Monthly {
			w.Week = board.FirstOfMonth(w.Today())
		}
		if w.Daily {
			w.Week = w.Today()
		}
	case "1":
		w = w.ranged("Day")
	case "2":
		w = w.ranged("Week")
	case "3":
		w = w.ranged("Month")
	}
	w.DaysFrom = [7]int{}
	return w.readingWeek()
}

func (w Window) ranged(to string) Window {
	today, on := w.Today(), w.Week
	switch {
	case w.Monthly:
		if sameMonth(today, w.Week) {
			on = today
		}
	case !w.Daily:
		if !today.Before(w.Week) && today.Before(w.Week.AddDate(0, 0, 7)) {
			on = today
		} else if to == "Month" {
			on = w.Week.AddDate(0, 0, 3)
		}
	}
	w.Daily, w.Monthly, w.Grid = to == "Day", to == "Month", board.ViewDays
	switch to {
	case "Day":
		w.Week = on
	case "Month":
		w.Week = board.FirstOfMonth(on)
	default:
		w.Week = board.Monday(on)
	}
	return w
}

func weekKeys(name string) (mine, shared bool) {
	switch name {
	case "left", "right", "t", "1", "2", "3":
		return true, false
	case "tab", "shift+tab", "cmd+b", "shift+b", "shift+l":
		return false, true
	}
	return false, false
}

// TimeStrip is the strip along the top of the time page: the board's whole width, where the columns' headers are on the other pages.
func TimeStrip(l geom.Layout) geom.Box {
	left, right := l.Column[0].X, l.Column[board.Columns-1].X+l.Column[board.Columns-1].W
	return geom.Box{X: left, Y: l.Head[0].Y, W: right - left, H: geom.BarHigh}
}

// TimeChip is where a chip of the strip is: the arrows at its left, the ranges at its right, Today before them.
func TimeChip(l geom.Layout, name string) geom.Box {
	strip := TimeStrip(l)
	at := func(x, wide int) geom.Box { return geom.Box{X: x, Y: strip.Y, W: wide, H: strip.H} }
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
	case "the month":
		return at(TimeSaid(l).X, nameWide)
	case "the year":
		return at(TimeSaid(l).X+nameWide+inset, yearWide)
	}
	return geom.Box{}
}

// TimeCells is what the month of the time page offers to choose from instead of its days, as the date picker does: nothing while it shows days.
func (w Window) TimeCells() []string {
	if !w.Monthly {
		return nil
	}
	return gridCells(w.Grid, w.Week)
}

// TimeCellAt is the i-th of those twelve cells, three across and four down, in the room the month's days take.
func TimeCellAt(l geom.Layout, i int) geom.Box {
	strip := TimeStrip(l)
	top := strip.Y + strip.H + between
	wide, high := (strip.W-2*geom.Gap)/3, (l.Band.Y+l.Band.H-top-3*geom.Gap)/4
	return geom.Box{X: strip.X + (i%3)*(wide+geom.Gap), Y: top + (i/3)*(high+geom.Gap), W: wide, H: high}
}

// RangesControl is the ground Day, Week and Month stand on together, one control of three segments rather than three buttons.
func RangesControl(l geom.Layout) geom.Box {
	day, month := TimeChip(l, "Day"), TimeChip(l, "Month")
	return geom.Box{X: day.X - segmentInset, Y: day.Y, W: month.X + month.W + segmentInset - day.X + segmentInset, H: day.H}
}

const segmentInset = 2

// TimeChips is every chip of the strip, in the order they are drawn.
var TimeChips = []string{"‹", "›", "Today", "Day", "Week", "Month"}

// ButtonIn is what an offer is drawn as in the room it is given: as tall, and a gap short of the next one.
func ButtonIn(slot geom.Box) geom.Box {
	return geom.Box{X: slot.X, Y: slot.Y, W: slot.W - inset, H: slot.H}
}

// ButtonParts is where a button writes its word and draws its key's chip: the two side by side in its middle, the word first.
func ButtonParts(button geom.Box, m Measure, word, key string) (geom.Box, geom.Box) {
	wide := buttonWordWide(m, word)
	chip := KeyChipWide(key)
	if chip > 0 {
		chip += hintSpace
	}
	x := button.X + max(0, (button.W-wide-chip)/2)
	said := geom.Box{X: x, Y: button.Y + (button.H-lineHigh)/2, W: wide, H: lineHigh}
	if key == "" {
		return said, geom.Box{}
	}
	return said, geom.Box{X: x + wide + hintSpace, Y: button.Y + (button.H-chipHigh)/2, W: KeyChipWide(key), H: chipHigh}
}

// buttonWordWide is how wide a button's word is written: measured in the face's plain weight, and a twentieth more for the medium weight a button writes in.
func buttonWordWide(m Measure, word string) int { return ends(m, word, reads)*21/20 + textSlack }

// ButtonSaid is where a button writes its words: a line high, across it, in its middle.
func ButtonSaid(button geom.Box) geom.Box { return ChipSaid(button, button.W) }

// ChipSaid is where a chip writes its name: in its middle, a line high and as wide as the name.
func ChipSaid(chip geom.Box, wide int) geom.Box {
	line := geom.Box{X: chip.X, Y: chip.Y + (chip.H-lineHigh)/2, W: chip.W, H: lineHigh}
	if wide >= chip.W {
		return line
	}
	return geom.Box{X: chip.X + (chip.W-wide)/2, Y: line.Y, W: wide, H: line.H}
}

// TimeSaid is where the strip says what is shown, between the arrows and the ranges.
func TimeSaid(l geom.Layout) geom.Box {
	after, today := TimeChip(l, "›"), TimeChip(l, "Today")
	x := after.X + arrowWide + padding
	return geom.Box{X: x, Y: after.Y, W: today.X - padding - x, H: after.H}
}

// TimeSays is what is shown, as the strip says it: a week's first and last days, or a month, then the time spent over it when there is any.
func TimeSays(w Window) string {
	if total := w.PeriodSays(); total != "" {
		return periodNamed(w) + " · " + total
	}
	return periodNamed(w)
}

func periodNamed(w Window) string {
	if w.Daily {
		return w.Week.Format("Monday 2 January 2006")
	}
	if w.Monthly {
		return w.Week.Format("January 2006")
	}
	last := w.Week.AddDate(0, 0, w.Pages.WeekDays()-1)
	if w.Week.Year() != last.Year() {
		return w.Week.Format("2 Jan 2006") + " – " + last.Format("2 Jan 2006")
	}
	if w.Week.Month() != last.Month() {
		return w.Week.Format("2 Jan") + " – " + last.Format("2 Jan 2006")
	}
	return w.Week.Format("2") + " – " + last.Format("2 Jan 2006")
}

// WeekDayAt is where the i-th day of the week is laid out: the board's whole width, from under the strip to the foot of the band, cut in as many days as a week shows.
func WeekDayAt(l geom.Layout, i int) geom.Box {
	strip := TimeStrip(l)
	wide, top := (strip.W-(l.Days-1)*geom.Gap)/l.Days, strip.Y+strip.H+between
	return geom.Box{X: strip.X + i*(wide+geom.Gap), Y: top, W: wide, H: l.Band.Y + l.Band.H - top}
}

// DayRoom is how many tasks a day of a week shows at once.
func DayRoom(day geom.Box) int { return max(1, (day.H-2*lineHigh-padding)/weekRow) }

// DayShown is the tasks of the i-th day of the week in sight.
func (w Window) DayShown(i int) []engine.TimesheetTask {
	tasks, from := w.DayHeld(i).Tasks, w.DaysFrom[i]
	return tasks[min(from, len(tasks)):min(len(tasks), from+DayRoom(WeekDayAt(w.Layout(), i)))]
}

// AloneShown is the tasks of the day shown alone that are in sight.
func (w Window) AloneShown() []engine.TimesheetTask {
	tasks, from := w.DayHeld(0).Tasks, w.DaysFrom[0]
	return tasks[min(from, len(tasks)):min(len(tasks), from+DayRoom(DayAloneAt(w.Layout())))]
}

// DayAloneAt is where one day is laid out when it is shown alone: the whole of what a week takes.
func DayAloneAt(l geom.Layout) geom.Box {
	strip, first := TimeStrip(l), WeekDayAt(l, 0)
	return geom.Box{X: strip.X, Y: first.Y, W: strip.W, H: first.H}
}

// AloneCardAt is where the j-th task of a day shown alone is drawn: the card a week draws, down the day's left and a third of it wide, so its title reads.
func AloneCardAt(alone geom.Box, j int) geom.Box {
	row := WeekRowAt(alone, j)
	return geom.Box{X: row.X, Y: row.Y, W: alone.W/3 - padding, H: geom.CardHeight}
}

// AloneSpentAt is where a task of a day shown alone says how long it took: beside its card.
func AloneSpentAt(alone geom.Box, j int) geom.Box {
	card := AloneCardAt(alone, j)
	return geom.Box{X: card.X + card.W + padding, Y: card.Y + (geom.CardHeight-geom.Gap-lineHigh)/2, W: spentWide, H: lineHigh}
}

// AloneBar is a task's bar on a day shown alone: after its time, across what is left of the day, as long as its share of the day's longest.
func AloneBar(alone geom.Box, j, minutes, longest int) geom.Box {
	spent := AloneSpentAt(alone, j)
	x := spent.X + spent.W
	room := alone.X + alone.W - padding - barWide - x
	return geom.Box{X: x, Y: spent.Y + (lineHigh-barHighOf)/2, W: room * minutes / max(1, longest), H: barHighOf}
}

func (w Window) scrollingDay(p geom.Point, by int) Window {
	l := w.Layout()
	if w.Daily {
		alone := DayAloneAt(l)
		if !alone.In(p) {
			return w
		}
		total := w.spareChoosing + by
		w.spareChoosing = total % weekRow
		w.DaysFrom[0] = max(0, min(w.DaysFrom[0]+total/weekRow, len(w.DayHeld(0).Tasks)-DayRoom(alone)))
		return w
	}
	for i := range w.WeekDays() {
		day := WeekDayAt(l, i)
		if !day.In(p) {
			continue
		}
		total := w.spareChoosing + by
		rows := total / weekRow
		w.spareChoosing = total % weekRow
		w.DaysFrom[i] = max(0, min(w.DaysFrom[i]+rows, len(w.DayHeld(i).Tasks)-DayRoom(day)))
		return w
	}
	return w
}

// MonthWeekdayAt is where a weekday's name is written over a month.
func MonthWeekdayAt(l geom.Layout, i int) geom.Box {
	cell := MonthCellAt(l, i)
	return geom.Box{X: cell.X, Y: cell.Y - lineHigh, W: cell.W, H: lineHigh}
}

// MonthCellAt is where the i-th day of a month's six weeks is laid out; a day the week does not show is laid out nowhere.
func MonthCellAt(l geom.Layout, i int) geom.Box {
	if i%7 >= l.Days {
		return geom.Box{}
	}
	strip := TimeStrip(l)
	top := strip.Y + strip.H + between + lineHigh
	wide, high := (strip.W-(l.Days-1)*geom.Gap)/l.Days, (l.Band.Y+l.Band.H-top-5*geom.Gap)/6
	return geom.Box{X: strip.X + (i%7)*(wide+geom.Gap), Y: top + (i/7)*(high+geom.Gap), W: wide, H: high}
}

// CellRoom is how many lines a day of a month has under its own.
func CellRoom(cell geom.Box) int { return max(1, (cell.H-lineHigh-padding)/lineHigh) }

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

func (w Window) clickingTime(p geom.Point) Window {
	l := w.Layout()
	keyOf := map[string]string{"‹": "left", "›": "right", "Today": "t"}
	for _, chip := range TimeChips {
		if !TimeChip(l, chip).In(p) {
			continue
		}
		if key, ok := keyOf[chip]; ok {
			return w.onWeek(key)
		}
		if chip != w.Ranging() {
			w = w.ranged(chip)
			w.DaysFrom = [7]int{}
			return w.readingWeek()
		}
		return w
	}
	if w.Monthly {
		if picked, on := w.timeGridClicked(l, p); on {
			return picked
		}
	}
	if number := w.timeCardAt(p); number != 0 {
		if w.GoneOnTime(number) {
			return w.telling("task " + strconv.Itoa(number) + " was deleted")
		}
		return w.choosingAt(number)
	}
	if !w.Monthly && !w.Daily {
		for i := range w.WeekDays() {
			if WeekHeadAt(WeekDayAt(l, i)).In(p) {
				w.Daily, w.Week, w.DaysFrom = true, w.Week.AddDate(0, 0, i), [7]int{}
				return w.readingWeek()
			}
		}
	}
	if !w.Monthly {
		return w
	}
	for i, day := range board.MonthGrid(w.Week) {
		if MonthCellAt(l, i).In(p) {
			w.Monthly, w.Week, w.DaysFrom = false, board.Monday(day), [7]int{}
			return w.readingWeek()
		}
	}
	return w
}

// WeekHeadAt is where a day says which it is and its total, at its top.
func WeekHeadAt(day geom.Box) geom.Box { return RowIn(day, day.Y+padding, lineHigh) }

// WeekTotalAt is where a day of the week says how long its work took: on its date's line, from past the date to its right end, written against that end — or on the line under the date when the longest total has no room beside it.
func WeekTotalAt(day geom.Box, m Measure) geom.Box {
	head := WeekHeadAt(day)
	pill := TodayPill(head, m)
	from := pill.X + pill.W + pillGap
	if head.X+head.W-from < monoWide(widestTime, small) {
		return geom.Box{X: head.X, Y: head.Y + lineHigh, W: head.W, H: lineHigh}
	}
	return geom.Box{X: from, Y: head.Y, W: head.X + head.W - from, H: head.H}
}

const widestTime = "00h00m"

// WeekRowAt is where the j-th task of a day is laid out, under its heading and its total: its card, then the line saying how long it took that day.
func WeekRowAt(day geom.Box, j int) geom.Box {
	return geom.Box{X: day.X + padding, Y: day.Y + 2*lineHigh + padding + j*weekRow, W: day.W - 2*padding, H: weekRow}
}

// DayCardAt is where the j-th of all the tasks of a day is drawn as the card the board draws it as: across the day as a card is across its column, short of the bar when the day scrolls.
func DayCardAt(day geom.Box, j, all int) geom.Box {
	row := WeekRowAt(day, j)
	wide := day.W
	if track, more := DayTrack(day, all); more {
		wide = track.X - day.X
	}
	return geom.Box{X: day.X, Y: row.Y, W: wide, H: geom.CardHeight}
}

// DaySpentAt is where a task of a day says how long it took that day: under its card, before its bar.
func DaySpentAt(row geom.Box) geom.Box {
	return geom.Box{X: row.X, Y: row.Y + geom.CardHeight - geom.Gap, W: spentWide, H: lineHigh}
}

// WeekBar is a task's bar: under its card, after its time, as long as its share of the longest stretch of the week.
func WeekBar(row geom.Box, minutes, longest int) geom.Box {
	spent := DaySpentAt(row)
	room := row.W - spent.W
	return geom.Box{X: spent.X + spent.W, Y: spent.Y + (lineHigh-barHighOf)/2, W: room * minutes / max(1, longest), H: barHighOf}
}

func cellShown(tasks []engine.TimesheetTask, room int) int {
	if len(tasks) > room {
		return room - 1
	}
	return len(tasks)
}

func (w Window) timeCardAt(p geom.Point) int {
	l := w.Layout()
	if w.Daily {
		for j, one := range w.AloneShown() {
			if AloneCardAt(DayAloneAt(l), j).In(p) {
				return one.Number
			}
		}
		return 0
	}
	if w.Monthly {
		for i := range board.MonthGrid(w.Week) {
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
	for i := range w.WeekDays() {
		for j, one := range w.DayShown(i) {
			if DayCardAt(WeekDayAt(l, i), j, len(w.DayHeld(i).Tasks)).In(p) {
				return one.Number
			}
		}
	}
	return 0
}

// DayTrack is the bar beside a day of a week holding more tasks than it shows, along the tasks it shows; nothing when they all fit.
func DayTrack(day geom.Box, all int) (geom.Box, bool) {
	room := DayRoom(day)
	if all <= room {
		return geom.Box{}, false
	}
	return geom.Box{X: day.X + day.W - barWide, Y: WeekRowAt(day, 0).Y, W: barWide, H: room * weekRow}, true
}

// CellHead is what a day of a month says at its top: its number, and its total once anything was spent on it.
func CellHead(day time.Time, held engine.TimesheetDay) string {
	if held.Minutes <= 0 {
		return strconv.Itoa(day.Day())
	}
	return strconv.Itoa(day.Day()) + "   " + held.Says()
}

// CellHeadAt is where a day of a month says its number and its total.
func CellHeadAt(cell geom.Box) geom.Box { return RowIn(cell, cell.Y+inset, lineHigh) }

// CellLineAt is where the j-th line under a day of a month's head goes.
func CellLineAt(cell geom.Box, j int) geom.Box {
	return RowIn(cell, cell.Y+inset+(j+1)*lineHigh, lineHigh)
}

func sameMonth(a, b time.Time) bool { return a.Year() == b.Year() && a.Month() == b.Month() }

func (w Window) timeGridClicked(l geom.Layout, p geom.Point) (Window, bool) {
	for name, view := range map[string]board.CalendarView{"the month": board.ViewMonths, "the year": board.ViewYears} {
		if !TimeChip(l, name).In(p) {
			continue
		}
		if w.Grid == view {
			view = board.ViewDays
		}
		w.Grid = view
		return w, true
	}
	for i := range w.TimeCells() {
		if TimeCellAt(l, i).In(p) {
			w.Week, w.Grid = gridPicked(w.Grid, w.Week, i)
			w.DaysFrom = [7]int{}
			return w.readingWeek(), true
		}
	}
	return w, w.Grid != board.ViewDays
}

// PeriodAt is where a month shown says its total: after its year, short of Today.
func PeriodAt(w Window, l geom.Layout) geom.Box {
	year, today := TimeChip(l, "the year"), TimeChip(l, "Today")
	x := year.X + year.W + padding
	return geom.Box{X: x, Y: year.Y, W: max(0, today.X-padding-x), H: year.H}
}

// TodayPill is where today's date stands on the accent in its day's head: around the date, as wide as it reads.
func TodayPill(head geom.Box, m Measure) geom.Box {
	return geom.Box{X: head.X - pillPad/2, Y: head.Y, W: ends(m, todayFormat, reads) + pillPad, H: head.H}
}

const todayFormat = "Mon 02/01"
