package gui

import (
	"strconv"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

const (
	dayHigh       = 38
	dayInset      = 6
	weeks         = 6
	monthWide     = 32
	nameWide      = 120
	yearWide      = 64
	calendarHead  = 32
	calendarLeast = 336
)

type calendarView int

const (
	viewDays calendarView = iota
	viewMonths
	viewYears
)

// Weekdays is the head of the calendar's columns, Monday first.
var Weekdays = [7]string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}

type calendarState struct {
	day  time.Time
	view calendarView
}

// At is the window told what moment it is: the day a calendar opens on when the field holds none, and the day it marks as today.
func (w Window) At(now time.Time) Window {
	w.now = now
	return w
}

func (w Window) today() time.Time { return engine.DayOf(w.now) }

// CalendarFor is the panel the calendar is drawn in at a corner, as wide as given: the month and the arrows, the weekdays, six weeks and the offers.
func CalendarFor(x, y, wide int) Box {
	high := calendarTop(Box{Y: y}) - y + lineHigh + between + weeks*dayHigh + padding + barHigh + padding
	return Box{X: x, Y: y, W: wide, H: high}
}

// CalendarHead is the row across the calendar's top, in place of a heading: the month and the year left, the arrows right.
func CalendarHead(box Box) Box {
	return Box{X: box.X + padding, Y: box.Y + padding, W: box.W - padding - padding, H: calendarHead}
}

func calendarTop(box Box) int {
	head := CalendarHead(box)
	return head.Y + head.H + between
}

// CalendarBox is the calendar's panel in this window: hung under the field it answers, the way a list is, rather than in the middle of the window, far from the field it fills in.
func (w Window) CalendarBox() Box {
	wide := w.calendarWide()
	at := w.hung(wide, CalendarFor(0, 0, wide).H)
	return CalendarFor(at.X, at.Y, wide)
}

func (w Window) calendarWide() int {
	return max(calendarLeast, padding+nameWide+inset+yearWide+between+monthWide+inset+monthWide+padding, w.offersWide())
}

func (w Window) hung(wide, high int) Point {
	if len(w.form.Fields()) == 0 {
		return Point{X: (w.width - wide) / 2, Y: (w.height - high) / 2}
	}
	return Hung(fieldFrame(w), Over(w.width, w.height), wide, high)
}

func fieldFrame(w Window) Box {
	_, value, rows := FormPlace(w.formBox(), w.form.Fields(), w.form.On())
	return Framed(value, rows)
}

func dayWide(box Box) int { return (box.W - padding - padding) / 7 }

// WeekdayAt is where a weekday's name is written, above its column.
func WeekdayAt(box Box, i int) Box {
	return Box{X: box.X + padding + i*dayWide(box), Y: calendarTop(box), W: dayWide(box) - dayInset, H: lineHigh}
}

// DayAt is where the i-th day of the grid is drawn.
func DayAt(box Box, i int) Box {
	top := calendarTop(box) + lineHigh + between
	return Box{
		X: box.X + padding + (i%7)*dayWide(box),
		Y: top + (i/7)*dayHigh,
		W: dayWide(box) - dayInset,
		H: dayHigh - dayInset,
	}
}

// MonthAfter is the arrow at the right end of the calendar's head that walks to the next month.
func MonthAfter(box Box) Box {
	head := CalendarHead(box)
	return Box{X: head.X + head.W - monthWide, Y: head.Y, W: monthWide, H: head.H}
}

// MonthBefore is the arrow walking back a month, left of the one walking on.
func MonthBefore(box Box) Box {
	after := MonthAfter(box)
	after.X -= monthWide + inset
	return after
}

func (w Window) monthClicked(p Point) (Window, bool) {
	box := w.CalendarBox()
	switch {
	case MonthBefore(box).In(p):
		w.day = engine.MonthOn(w.day, -w.pageMonths())
	case MonthAfter(box).In(p):
		w.day = engine.MonthOn(w.day, w.pageMonths())
	case CalendarMonth(box).In(p):
		w.view = viewMonths
	case CalendarYear(box).In(p):
		w.view = viewYears
	default:
		return w, false
	}
	return w, true
}

func (w Window) scrollingMonths(p Point, by int) Window {
	if !w.CalendarBox().In(p) {
		return w
	}
	months, left := turning(w.spareChoosing, by)
	w.day, w.spareChoosing = engine.MonthOn(w.day, months), left
	return w
}

// Faded says a day of the grid lies in the month before or after the one the keyboard is in, and is drawn faint: there to finish a week, not to be read.
func (w Window) Faded(day time.Time) bool { return day.Month() != w.day.Month() }

func (w Window) openingCalendar() Window {
	w.day = w.today()
	if held, err := domain.ParseDeadline(w.form.Fields()[w.form.On()].Value); err == nil && !held.IsZero() {
		w.day = engine.DayOf(held)
	}
	w.screen = OnCalendar
	return w
}

var gridSteps = map[string]int{"left": -1, "right": 1, "up": -3, "down": 3}

func (w Window) cellMonths() int {
	if w.view == viewYears {
		return 12
	}
	return 1
}

func (w Window) pageMonths() int {
	switch w.view {
	case viewYears:
		return 144
	case viewMonths:
		return 12
	}
	return 1
}

func (w Window) backFromGrid() Window {
	if w.view == viewYears {
		w.view = viewMonths
		return w
	}
	w.view = viewDays
	return w
}

func (w Window) onCalendar(name string) Window {
	if by, walking := gridSteps[name]; walking && w.view != viewDays {
		w.day = engine.MonthOn(w.day, by*w.cellMonths())
		return w
	}
	switch name {
	case "m":
		w.view = viewMonths
	case "y":
		w.view = viewYears
	case "left":
		w.day = w.day.AddDate(0, 0, -1)
	case "right":
		w.day = w.day.AddDate(0, 0, 1)
	case "up":
		w.day = w.day.AddDate(0, 0, -7)
	case "down":
		w.day = w.day.AddDate(0, 0, 7)
	case "shift+left":
		w.day = engine.MonthOn(w.day, -1)
	case "shift+right":
		w.day = engine.MonthOn(w.day, 1)
	case "t":
		w.day = w.today()
	}
	return w
}

func (w Window) dated(day time.Time) Window {
	w.form = w.form.Writes(Ending(domain.FormatDay(day)))
	w.screen = OnForm
	return w
}

func (w Window) dayClicked(p Point) (Window, bool) {
	if w.view != viewDays {
		return w.cellClicked(p)
	}
	box := w.CalendarBox()
	for i, day := range engine.MonthGrid(w.day) {
		if DayAt(box, i).In(p) {
			return w.dated(day), true
		}
	}
	return w, false
}

// CalendarMonth is the month's name at the left of the calendar's head, which a click opens the twelve months from.
func CalendarMonth(box Box) Box {
	head := CalendarHead(box)
	return Box{X: head.X, Y: head.Y, W: nameWide, H: head.H}
}

// CalendarYear is the year beside the month's name, which a click opens the years from.
func CalendarYear(box Box) Box {
	month := CalendarMonth(box)
	return Box{X: month.X + month.W + inset, Y: month.Y, W: yearWide, H: month.H}
}

// CalendarCell is the i-th of the twelve cells, three across and four down, a month or a year is chosen from; they take the room of the weekdays and the six weeks.
func CalendarCell(box Box, i int) Box {
	wide := (box.W - padding - padding) / 3
	high := (lineHigh + between + weeks*dayHigh) / 4
	return Box{
		X: box.X + padding + (i%3)*wide,
		Y: calendarTop(box) + (i/3)*high,
		W: wide - dayInset,
		H: high - dayInset,
	}
}

// CalendarCells is what the calendar offers to choose from instead of its days: nothing while it shows days.
func (w Window) CalendarCells() []string { return gridCells(w.view, w.day) }

func gridCells(view calendarView, day time.Time) []string {
	if view == viewYears {
		years := make([]string, 12)
		for i := range years {
			years[i] = strconv.Itoa(firstYear(day) + i)
		}
		return years
	}
	if view != viewMonths {
		return nil
	}
	names := make([]string, 12)
	for i := range names {
		names[i] = time.Month(i + 1).String()[:3]
	}
	return names
}

func firstYear(day time.Time) int { return day.Year() - 4 }

// CalendarMarked is the cell of the grid shown that holds the day the calendar is on.
func (w Window) CalendarMarked() string { return gridMarked(w.view, w.day) }

func gridMarked(view calendarView, day time.Time) string {
	if view == viewYears {
		return strconv.Itoa(day.Year())
	}
	return day.Month().String()[:3]
}

func gridPicked(view calendarView, day time.Time, i int) (time.Time, calendarView) {
	if view == viewYears {
		return engine.MonthOn(day, (firstYear(day)+i-day.Year())*12), viewMonths
	}
	return engine.MonthOn(day, i+1-int(day.Month())), viewDays
}

func (w Window) cellClicked(p Point) (Window, bool) {
	for i := range w.CalendarCells() {
		if CalendarCell(w.CalendarBox(), i).In(p) {
			w.day, w.view = gridPicked(w.view, w.day, i)
			return w, true
		}
	}
	return w, false
}
