package gui

// The calendar a deadline is chosen on. A deadline was typed as a date, which
// meant knowing what date next Friday falls on; enter on the field now opens
// a month at a time, and the field still takes a date typed. Both forms have
// a deadline, and both open the same calendar. See docs/gui.md.

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

const (
	// dayHigh is how high a day is drawn, and dayInset what is left between
	// two of them. The days share the panel's width, which its heading and
	// its offers set.
	dayHigh  = 32
	dayInset = 4
	// weeks is how many rows a month is shown on: six hold every month.
	weeks = 6
	// monthWide is how wide each of the arrows walking the months is.
	monthWide = 28
)

// Weekdays is the head of the calendar's columns, Monday first.
var Weekdays = [7]string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}

// calendarState is what the calendar keeps: the day its keyboard is on.
type calendarState struct {
	day time.Time
}

// At is the window told what moment it is: the day a calendar opens on when
// the field holds none, and the day it marks as today.
func (w Window) At(now time.Time) Window {
	w.now = now
	return w
}

// today is the day the window was last told it is.
func (w Window) today() time.Time { return engine.DayOf(w.now) }

// CalendarDay is the day the keyboard is on.
func (w Window) CalendarDay() time.Time { return w.day }

// CalendarFor is the panel the calendar is drawn in at a corner, as wide as
// given: a heading, the weekdays, six weeks and the offers.
func CalendarFor(x, y, wide int) Box {
	high := FirstRow(Box{Y: y}) - y + lineHigh + between + weeks*dayHigh + padding + barHigh + padding
	return Box{X: x, Y: y, W: wide, H: high}
}

// CalendarBox is the calendar's panel in this window: hung under the field it
// answers, the way a list is, rather than in the middle of the window, far
// from the field it fills in.
func (w Window) CalendarBox() Box {
	wide := w.calendarWide()
	at := w.hung(wide, CalendarFor(0, 0, wide).H)
	return CalendarFor(at.X, at.Y, wide)
}

// calendarWide is how wide the calendar is: its offers, or its heading with
// the arrows beside it, either far wider than a week of legible days —
// measured on the longest month, so
// the panel and its arrows do not move from one month to the next.
func (w Window) calendarWide() int {
	longest := w
	longest.day = time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	return max(longest.headWide()+2*monthWide, w.offersWide())
}

// hung is where a panel of a size opened from the field the keyboard is on
// goes; across the middle of the window when no form is open to hang it from.
func (w Window) hung(wide, high int) Point {
	if len(w.form.Fields()) == 0 {
		return Point{X: (w.width - wide) / 2, Y: (w.height - high) / 2}
	}
	return Hung(fieldFrame(w), Over(w.width, w.height), wide, high)
}

// fieldFrame is the frame of the field the keyboard is on.
func fieldFrame(w Window) Box {
	_, value, rows := FormPlace(w.formBox(), w.form.Fields(), w.form.On())
	return Framed(value, rows)
}

// dayWide is the room one day takes across a calendar of a given width.
func dayWide(box Box) int { return (box.W - padding - padding) / 7 }

// WeekdayAt is where a weekday's name is written, above its column.
func WeekdayAt(box Box, i int) Box {
	return Box{X: box.X + padding + i*dayWide(box), Y: FirstRow(box), W: dayWide(box) - dayInset, H: lineHigh}
}

// DayAt is where the i-th day of the grid is drawn.
func DayAt(box Box, i int) Box {
	top := FirstRow(box) + lineHigh + between
	return Box{
		X: box.X + padding + (i%7)*dayWide(box),
		Y: top + (i/7)*dayHigh,
		W: dayWide(box) - dayInset,
		H: dayHigh - dayInset,
	}
}

// MonthAfter is the arrow at the end of the heading that walks to the next
// month; MonthBefore the one before it, walking back. A hand on the mouse
// had no way out of the month the calendar opened on.
func MonthAfter(box Box) Box {
	head := Heading(box)
	return Box{X: box.X + box.W - padding - monthWide, Y: head.Y, W: monthWide, H: titleHi}
}

// MonthBefore is the arrow walking back a month, left of the one walking on.
func MonthBefore(box Box) Box {
	after := MonthAfter(box)
	after.X -= monthWide
	return after
}

// monthClicked walks a month either way when an arrow of the heading is
// under the pointer.
func (w Window) monthClicked(p Point) (Window, bool) {
	box := w.CalendarBox()
	switch {
	case MonthBefore(box).In(p):
		w.day = engine.MonthOn(w.day, -1)
	case MonthAfter(box).In(p):
		w.day = engine.MonthOn(w.day, 1)
	default:
		return w, false
	}
	return w, true
}

// scrollingMonths walks the months under the wheel, a notch a month, keeping
// what did not come to a notch for the next turn.
func (w Window) scrollingMonths(p Point, by int) Window {
	if !w.CalendarBox().In(p) {
		return w
	}
	months, left := turning(w.spareChoosing, by)
	w.day, w.spareChoosing = engine.MonthOn(w.day, months), left
	return w
}

// Faded says a day of the grid lies in the month before or after the one the
// keyboard is in, and is drawn faint: there to finish a week, not to be read.
func (w Window) Faded(day time.Time) bool { return day.Month() != w.day.Month() }

// openingCalendar opens the calendar on the day the field holds, or on today
// when it holds none or something that is not a day.
func (w Window) openingCalendar() Window {
	w.day = w.today()
	if held, err := domain.ParseDeadline(w.form.Fields()[w.form.On()].Value); err == nil && !held.IsZero() {
		w.day = engine.DayOf(held)
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
		w.day = engine.MonthOn(w.day, -1)
	case "shift+right":
		w.day = engine.MonthOn(w.day, 1)
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
	for i, day := range engine.MonthGrid(w.day) {
		if DayAt(box, i).In(p) {
			return w.dated(day), true
		}
	}
	return w, false
}
