package gui

import (
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/reading"
)

// A month is shown as six weeks from the Monday on or before its first day:
// every month fits, and the rows do not move from one month to the next.
func TestMonthGrid_startsOnTheMondayBeforeTheFirst(t *testing.T) {
	grid := MonthGrid(time.Date(2026, 10, 12, 15, 0, 0, 0, time.UTC))
	if got := domain.FormatDay(grid[0]); got != "2026-09-28" {
		t.Errorf("October 2026 starts its grid on %s, want the Monday 2026-09-28", got)
	}
	if got := domain.FormatDay(grid[len(grid)-1]); got != "2026-11-08" {
		t.Errorf("it ends on %s, want 2026-11-08", got)
	}
	if got := domain.FormatDay(MonthGrid(time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC))[0]); got != "2026-06-01" {
		t.Errorf("a month starting on a Monday starts its grid on %s", got)
	}
	if got := domain.FormatDay(MonthGrid(time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC))[0]); got != "2026-10-26" {
		t.Errorf("a month starting on a Sunday starts its grid on %s, want the Monday before", got)
	}
}

// The calendar's panel and its days, to the pixel: a sum gone wrong by a
// padding puts a click on the day beside the one drawn.
func TestTheCalendarIsLaidOutToThePixel(t *testing.T) {
	box := CalendarOver(1200, 800)
	if want := (Box{X: 378, Y: 191, W: 444, H: 418}); box != want {
		t.Errorf("the calendar is %+v, want %+v", box, want)
	}
	if got, want := DayAt(box, 0), (Box{X: 390, Y: 265, W: 56, H: 44}); got != want {
		t.Errorf("the first day is at %+v, want %+v", got, want)
	}
	if got, want := DayAt(box, 9), (Box{X: 510, Y: 313, W: 56, H: 44}); got != want {
		t.Errorf("the tenth day is at %+v, want %+v", got, want)
	}
	if got, want := WeekdayAt(box, 2), (Box{X: 510, Y: 241, W: 56, H: 18}); got != want {
		t.Errorf("Wednesday is written at %+v, want %+v", got, want)
	}
	if got := len(MonthGrid(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))); got != 42 {
		t.Errorf("a month is shown as %d days, want six weeks", got)
	}
}

// The arrows walk back as they walk on: a day, and a week.
func TestTheCalendarWalksBack(t *testing.T) {
	w := Window{screen: OnCalendar, day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}
	if got := domain.FormatDay(w.onCalendar("left").day); got != "2026-10-11" {
		t.Errorf("a day back is %s", got)
	}
	if got := domain.FormatDay(w.onCalendar("up").day); got != "2026-10-05" {
		t.Errorf("a week back is %s", got)
	}
	if !w.Faded(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)) || w.Faded(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("only the days of the months either side are faint")
	}
}

// A month on from the last day of a longer month is the last day of the
// shorter one, never a day into the month after it.
func TestTheCalendarWalksMonthsWithoutSpillingOver(t *testing.T) {
	w := Window{screen: OnCalendar, day: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)}
	if got := domain.FormatDay(w.onCalendar("shift+right").day); got != "2026-02-28" {
		t.Errorf("a month on from 2026-01-31 is %s, want 2026-02-28", got)
	}
	w.day = time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	if got := domain.FormatDay(w.onCalendar("shift+left").day); got != "2026-02-28" {
		t.Errorf("a month back from 2026-03-31 is %s, want 2026-02-28", got)
	}
}

// The calendar is drawn over the form: the day the keyboard is on carries the
// outline the chosen card does.
func TestTheCalendarIsDrawn(t *testing.T) {
	w := Window{screen: OnCalendar, width: 1200, height: 800, day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}
	w.form = editing(reading.Card{Number: 1, Title: "Write the spec"}, nil)
	shot := drawn(t, w)
	for i, day := range MonthGrid(w.day) {
		edge := Edges(DayAt(w.CalendarBox(), i))[0]
		outlined := ink(shot, edge, Dark().Panel) > 0
		if day.Equal(w.day) && !outlined {
			t.Error("the day the keyboard is on is not outlined")
		}
		if !day.Equal(w.day) && outlined {
			t.Errorf("%s is outlined, and the keyboard is not on it", domain.FormatDay(day))
		}
	}
}
