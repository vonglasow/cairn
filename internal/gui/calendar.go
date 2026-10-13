package gui

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

const (
	dayHigh   = 32
	dayInset  = 4
	weeks     = 6
	monthWide = 28
)

// Weekdays is the head of the calendar's columns, Monday first.
var Weekdays = [7]string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}

type calendarState struct {
	day time.Time
}

// At is the window told what moment it is: the day a calendar opens on when the field holds none, and the day it marks as today.
func (w Window) At(now time.Time) Window {
	w.now = now
	return w
}

func (w Window) today() time.Time { return engine.DayOf(w.now) }

// CalendarDay is the day the keyboard is on.
func (w Window) CalendarDay() time.Time { return w.day }

// CalendarFor is the panel the calendar is drawn in at a corner, as wide as given: a heading, the weekdays, six weeks and the offers.
func CalendarFor(x, y, wide int) Box {
	high := FirstRow(Box{Y: y}) - y + lineHigh + between + weeks*dayHigh + padding + barHigh + padding
	return Box{X: x, Y: y, W: wide, H: high}
}

// CalendarBox is the calendar's panel in this window: hung under the field it answers, the way a list is, rather than in the middle of the window, far from the field it fills in.
func (w Window) CalendarBox() Box {
	wide := w.calendarWide()
	at := w.hung(wide, CalendarFor(0, 0, wide).H)
	return CalendarFor(at.X, at.Y, wide)
}

func (w Window) calendarWide() int {
	longest := w
	longest.day = time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	return max(longest.headWide()+2*monthWide, w.offersWide())
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

// MonthAfter is the arrow at the end of the heading that walks to the next month; MonthBefore the one before it, walking back.
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

func (w Window) dated(day time.Time) Window {
	w.form = w.form.Writes(Ending(domain.FormatDay(day)))
	w.screen = OnForm
	return w
}

func (w Window) dayClicked(p Point) (Window, bool) {
	box := w.CalendarBox()
	for i, day := range engine.MonthGrid(w.day) {
		if DayAt(box, i).In(p) {
			return w.dated(day), true
		}
	}
	return w, false
}
