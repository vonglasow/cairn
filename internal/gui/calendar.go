package gui

// The calendar a deadline is chosen on. A deadline was typed as a date, which
// meant knowing what date next Friday falls on; enter on the field now opens
// a month at a time, and the field still takes a date typed. Both forms have
// a deadline, and both open the same calendar. See docs/gui.md.

import (
	"time"

	"cairn/internal/domain"
)

const (
	// dayWide and dayHigh are the room one day takes on the calendar, and
	// dayInset what is left between two of them.
	dayWide  = 60
	dayHigh  = 48
	dayInset = 4
	// weeks is how many rows a month is shown on: six hold every month.
	weeks = 6
)

// Weekdays is the head of the calendar's columns, Monday first.
var Weekdays = [7]string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}

// At is the window told what moment it is: the day a calendar opens on when
// the field holds none, and the day it marks as today.
func (w Window) At(now time.Time) Window {
	w.now = now
	return w
}

// today is the day the window was last told it is.
func (w Window) today() time.Time { return dayOf(w.now) }

// dayOf is a moment's day, at midnight in the zone days are kept in.
func dayOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// CalendarDay is the day the keyboard is on.
func (w Window) CalendarDay() time.Time { return w.day }

// MonthGrid is the six weeks a month is shown as, from the Monday on or
// before its first day: every month fits, and no row moves from one month to
// the next.
func MonthGrid(day time.Time) []time.Time {
	first := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, time.UTC)
	back := (int(first.Weekday()) + 6) % 7
	grid := make([]time.Time, weeks*7)
	for i := range grid {
		grid[i] = first.AddDate(0, 0, i-back)
	}
	return grid
}

// CalendarOver is the panel the calendar is drawn in, in the middle of the
// window: seven days wide, a heading, the weekdays, six weeks and the offers.
func CalendarOver(width, height int) Box {
	w := padding + 7*dayWide + padding
	h := 50 + lineHigh + between + weeks*dayHigh + padding + barHigh + padding
	return Box{X: (width - w) / 2, Y: (height - h) / 2, W: w, H: h}
}

// CalendarBox is the calendar's panel in this window.
func (w Window) CalendarBox() Box { return CalendarOver(w.width, w.height) }

// WeekdayAt is where a weekday's name is written, above its column.
func WeekdayAt(box Box, i int) Box {
	return Box{X: box.X + padding + i*dayWide, Y: FirstRow(box), W: dayWide - dayInset, H: lineHigh}
}

// DayAt is where the i-th day of the grid is drawn.
func DayAt(box Box, i int) Box {
	top := FirstRow(box) + lineHigh + between
	return Box{
		X: box.X + padding + (i%7)*dayWide,
		Y: top + (i/7)*dayHigh,
		W: dayWide - dayInset,
		H: dayHigh - dayInset,
	}
}

// Faded says a day of the grid lies in the month before or after the one the
// keyboard is in, and is drawn faint: there to finish a week, not to be read.
func (w Window) Faded(day time.Time) bool { return day.Month() != w.day.Month() }

// openingCalendar opens the calendar on the day the field holds, or on today
// when it holds none or something that is not a day.
func (w Window) openingCalendar() Window {
	w.day = w.today()
	if held, err := domain.ParseDeadline(w.form.Fields()[w.form.On()].Value); err == nil && !held.IsZero() {
		w.day = dayOf(held)
	}
	w.screen = OnCalendar
	return w
}

// onCalendar is what a key does on the calendar: the arrows walk the days and
// the weeks, shift with them the months, and t goes back to today.
func (w Window) onCalendar(name string) Window {
	switch name {
	case "left":
		w.day = w.day.AddDate(0, 0, -1)
	case "right":
		w.day = w.day.AddDate(0, 0, 1)
	case "up":
		w.day = w.day.AddDate(0, 0, -7)
	case "down":
		w.day = w.day.AddDate(0, 0, 7)
	case "shift+left":
		w.day = monthOn(w.day, -1)
	case "shift+right":
		w.day = monthOn(w.day, 1)
	case "t":
		w.day = w.today()
	}
	return w
}

// dated closes the calendar and writes a day into the field, or none.
func (w Window) dated(day time.Time) Window {
	w.form = w.form.Writes(Ending(domain.FormatDay(day)))
	w.screen = OnForm
	return w
}

// dayClicked takes the day under the pointer, when there is one.
func (w Window) dayClicked(p Point) (Window, bool) {
	box := w.CalendarBox()
	for i, day := range MonthGrid(w.day) {
		if DayAt(box, i).In(p) {
			return w.dated(day), true
		}
	}
	return w, false
}

// monthOn is the same day some months on, or the last day of that month when
// it is shorter: a month on from the thirty-first of January is the end of
// February, not the third of March.
func monthOn(day time.Time, by int) time.Time {
	first := time.Date(day.Year(), day.Month()+time.Month(by), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(day.Day(), last)-1)
}
