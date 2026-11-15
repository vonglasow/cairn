package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
	"cairn/internal/gui/geom"
	"strconv"
	"time"
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

// Weekdays is the head of the calendar's columns, Monday first.
var Weekdays = [7]string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}

// At is the window told what moment it is: the day a calendar opens on when the field holds none, and the day it marks as today.
func (w Window) At(now time.Time) Window {
	w.Now = now
	if w.GoalDragging() {
		return w.hoveringGoal(w.held.At)
	}
	return w
}

// CalendarFor is the panel the calendar is drawn in at a corner, as wide as given: the month and the arrows, the weekdays, six weeks and the offers.
func CalendarFor(x, y, wide int) geom.Box {
	high := calendarTop(geom.Box{Y: y}) - y + lineHigh + between + weeks*dayHigh + padding + geom.BarHigh + padding
	return geom.Box{X: x, Y: y, W: wide, H: high}
}

// CalendarHead is the row across the calendar's top, in place of a heading: the month and the year left, the arrows right.
func CalendarHead(box geom.Box) geom.Box {
	return geom.Box{X: box.X + padding, Y: box.Y + padding, W: box.W - padding - padding, H: calendarHead}
}

func calendarTop(box geom.Box) int {
	head := CalendarHead(box)
	return head.Y + head.H + between
}

// CalendarBox is the calendar's panel in this window: hung under the field it answers, the way a list is, rather than in the middle of the window, far from the field it fills in.
func (w Window) CalendarBox() geom.Box {
	wide := w.calendarWide()
	at := w.hung(wide, CalendarFor(0, 0, wide).H)
	return CalendarFor(at.X, at.Y, wide)
}

func (w Window) calendarWide() int {
	return max(calendarLeast, padding+nameWide+inset+yearWide+between+monthWide+inset+monthWide+padding, w.offersWide())
}

func (w Window) hung(wide, high int) geom.Point {
	if len(w.Form.Fields) == 0 {
		return geom.Point{X: (w.width - wide) / 2, Y: (w.height - high) / 2}
	}
	return Hung(fieldFrame(w), Over(w.width, w.height), wide, high)
}

func fieldFrame(w Window) geom.Box {
	_, value, rows := FormPlace(w.formBox(), w.Form.Fields, w.Form.On())
	return Framed(value, rows)
}

func dayWide(box geom.Box) int { return (box.W - padding - padding) / 7 }

// WeekdayAt is where a weekday's name is written, above its column.
func WeekdayAt(box geom.Box, i int) geom.Box {
	return geom.Box{X: box.X + padding + i*dayWide(box), Y: calendarTop(box), W: dayWide(box) - dayInset, H: lineHigh}
}

// DayAt is where the i-th day of the grid is drawn.
func DayAt(box geom.Box, i int) geom.Box {
	top := calendarTop(box) + lineHigh + between
	return geom.Box{
		X: box.X + padding + (i%7)*dayWide(box),
		Y: top + (i/7)*dayHigh,
		W: dayWide(box) - dayInset,
		H: dayHigh - dayInset,
	}
}

// MonthAfter is the arrow at the right end of the calendar's head that walks to the next month.
func MonthAfter(box geom.Box) geom.Box {
	head := CalendarHead(box)
	return geom.Box{X: head.X + head.W - monthWide, Y: head.Y, W: monthWide, H: head.H}
}

// MonthBefore is the arrow walking back a month, left of the one walking on.
func MonthBefore(box geom.Box) geom.Box {
	after := MonthAfter(box)
	after.X -= monthWide + inset
	return after
}

func (w Window) monthClicked(p geom.Point) (Window, bool) {
	box := w.CalendarBox()
	switch {
	case MonthBefore(box).In(p):
		w.Day = board.MonthOn(w.Day, -w.PageMonths())
	case MonthAfter(box).In(p):
		w.Day = board.MonthOn(w.Day, w.PageMonths())
	case CalendarMonth(box).In(p):
		w.CalendarView = board.ViewMonths
	case CalendarYear(box).In(p):
		w.CalendarView = board.ViewYears
	default:
		return w, false
	}
	return w, true
}

func (w Window) scrollingMonths(p geom.Point, by int) Window {
	if !w.CalendarBox().In(p) {
		return w
	}
	months, left := turning(w.spareChoosing, by)
	w.Day, w.spareChoosing = board.MonthOn(w.Day, months), left
	return w
}

func (w Window) openingCalendar() Window {
	w.Day = w.Today()
	if held, err := calendar.ParseDeadline(w.Form.Fields[w.Form.On()].Value); err == nil && !held.IsZero() {
		w.Day = board.DayOf(held)
	}
	w.Screen = board.OnCalendar
	return w
}

var gridSteps = map[string]int{"left": -1, "right": 1, "up": -3, "down": 3}

func (w Window) backFromGrid() Window {
	if w.CalendarView == board.ViewYears {
		w.CalendarView = board.ViewMonths
		return w
	}
	w.CalendarView = board.ViewDays
	return w
}

func (w Window) onCalendar(name string) Window {
	if by, walking := gridSteps[name]; walking && w.CalendarView != board.ViewDays {
		w.Day = board.MonthOn(w.Day, by*w.CellMonths())
		return w
	}
	switch name {
	case "m":
		w.CalendarView = board.ViewMonths
	case "y":
		w.CalendarView = board.ViewYears
	case "left":
		w.Day = w.Day.AddDate(0, 0, -1)
	case "right":
		w.Day = w.Day.AddDate(0, 0, 1)
	case "up":
		w.Day = w.Day.AddDate(0, 0, -7)
	case "down":
		w.Day = w.Day.AddDate(0, 0, 7)
	case "shift+left":
		w.Day = board.MonthOn(w.Day, -1)
	case "shift+right":
		w.Day = board.MonthOn(w.Day, 1)
	case "t":
		w.Day = w.Today()
	}
	return w
}

func (w Window) dated(day time.Time) Window {
	if key := w.calendarKey; key != "" {
		w.calendarKey, w.Screen = "", board.OnConfiguration
		if day.IsZero() {
			return w.unsetting(key)
		}
		return w.setting(key, calendar.FormatDay(day))
	}
	w.Form = w.Form.Writes(board.Ending(calendar.FormatDay(day)))
	w.Screen = board.OnForm
	return w
}

func (w Window) openingCalendarFor(key, value string) Window {
	w.Day = w.Today()
	if held, err := calendar.ParseDeadline(value); err == nil && !held.IsZero() {
		w.Day = board.DayOf(held)
	}
	w.calendarKey, w.Screen = key, board.OnCalendar
	return w
}

func (w Window) leavingCalendar() Window {
	if w.calendarKey != "" {
		w.calendarKey, w.Screen = "", board.OnConfiguration
		return w
	}
	w.Screen = board.OnForm
	return w
}

func (w Window) dayClicked(p geom.Point) (Window, bool) {
	if w.CalendarView != board.ViewDays {
		return w.cellClicked(p)
	}
	box := w.CalendarBox()
	for i, day := range board.MonthGrid(w.Day) {
		if DayAt(box, i).In(p) {
			return w.dated(day), true
		}
	}
	return w, false
}

// CalendarMonth is the month's name at the left of the calendar's head, which a click opens the twelve months from.
func CalendarMonth(box geom.Box) geom.Box {
	head := CalendarHead(box)
	return geom.Box{X: head.X, Y: head.Y, W: nameWide, H: head.H}
}

// CalendarYear is the year beside the month's name, which a click opens the years from.
func CalendarYear(box geom.Box) geom.Box {
	month := CalendarMonth(box)
	return geom.Box{X: month.X + month.W + inset, Y: month.Y, W: yearWide, H: month.H}
}

// CalendarCell is the i-th of the twelve cells, three across and four down, a month or a year is chosen from; they take the room of the weekdays and the six weeks.
func CalendarCell(box geom.Box, i int) geom.Box {
	wide := (box.W - padding - padding) / 3
	high := (lineHigh + between + weeks*dayHigh) / 4
	return geom.Box{
		X: box.X + padding + (i%3)*wide,
		Y: calendarTop(box) + (i/3)*high,
		W: wide - dayInset,
		H: high - dayInset,
	}
}

// CalendarCells is what the calendar offers to choose from instead of its days: nothing while it shows days.
func (w Window) CalendarCells() []string { return gridCells(w.CalendarView, w.Day) }

func gridCells(view board.CalendarView, day time.Time) []string {
	if view == board.ViewYears {
		years := make([]string, 12)
		for i := range years {
			years[i] = strconv.Itoa(firstYear(day) + i)
		}
		return years
	}
	if view != board.ViewMonths {
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
func (w Window) CalendarMarked() string { return gridMarked(w.CalendarView, w.Day) }

func gridMarked(view board.CalendarView, day time.Time) string {
	if view == board.ViewYears {
		return strconv.Itoa(day.Year())
	}
	return day.Month().String()[:3]
}

func gridPicked(view board.CalendarView, day time.Time, i int) (time.Time, board.CalendarView) {
	if view == board.ViewYears {
		return board.MonthOn(day, (firstYear(day)+i-day.Year())*12), board.ViewMonths
	}
	return board.MonthOn(day, i+1-int(day.Month())), board.ViewDays
}

func (w Window) cellClicked(p geom.Point) (Window, bool) {
	for i := range w.CalendarCells() {
		if CalendarCell(w.CalendarBox(), i).In(p) {
			w.Day, w.CalendarView = gridPicked(w.CalendarView, w.Day, i)
			return w, true
		}
	}
	return w, false
}
