package gui

import (
	"errors"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

func dayNamed(day string) time.Time {
	parsed, err := domain.ParseDeadline(day)
	if err != nil {
		panic(err)
	}
	return parsed
}

func TestTheWeekIsLaidOutToThePixel(t *testing.T) {
	l := LayoutFor(1200, 800)
	first, last := WeekDayAt(l, 0), WeekDayAt(l, 6)
	if strip := TimeStrip(l); first.X != l.Column[0].X || first.Y != strip.Y+strip.H+between || first.Y+first.H != l.Band.Y+l.Band.H {
		t.Errorf("the first day is %+v, not from under the strip to the foot of the band", first)
	}
	right := l.Column[engine.Columns-1].X + l.Column[engine.Columns-1].W
	if gapLeft := right - (last.X + last.W); gapLeft < 0 || gapLeft >= 7 {
		t.Errorf("the last day ends %d short of the board's right edge", gapLeft)
	}
	if WeekDayAt(l, 1).X != first.X+first.W+gap {
		t.Errorf("the second day does not start a gap after the first")
	}
	if got, want := WeekHeadAt(Box{X: 100, Y: 50, W: 200, H: 400}), (Box{X: 112, Y: 62, W: 176, H: lineHigh}); got != want {
		t.Errorf("a day's heading is at %+v, want %+v", got, want)
	}
	row := WeekRowAt(Box{X: 100, Y: 50, W: 200, H: 400}, 1)
	if want := (Box{X: 112, Y: 50 + 2*lineHigh + padding + weekRow, W: 176, H: weekRow}); row != want {
		t.Errorf("the second task is at %+v, want %+v", row, want)
	}
	spent := Box{X: 112, Y: row.Y + CardHeight - gap, W: spentWide, H: lineHigh}
	if got := DaySpentAt(row); got != spent {
		t.Errorf("the time of the second task is said at %+v, want %+v", got, spent)
	}
	if got, want := WeekBar(row, 30, 120), (Box{X: 112 + spentWide, Y: spent.Y + (lineHigh-barHighOf)/2, W: (176 - spentWide) / 4, H: barHighOf}); got != want {
		t.Errorf("a quarter of the longest is drawn %+v, want %+v", got, want)
	}
}

func TestTheWeekCountsWhatWasWorkedOn(t *testing.T) {
	w := Window{timeState: timeState{week: dayNamed("2026-09-28"), sheet: engine.Timesheet{Days: []engine.TimesheetDay{
		{Day: "2026-09-29", Minutes: 135, Tasks: []engine.TimesheetTask{{Number: 1, Minutes: 90}, {Number: 2, Minutes: 45}}},
		{Day: "2026-10-02", Minutes: 120, Tasks: []engine.TimesheetTask{{Number: 2, Minutes: 120}}},
	}}}}
	if got := w.sheet.Longest(); got != 120 {
		t.Errorf("the longest stretch is %d, want 120", got)
	}
	if got := w.Holds(Time); got != 2 {
		t.Errorf("the page says %d tasks, want 2", got)
	}
	if got := w.sheet.On(w.week.AddDate(0, 0, 1)); got.Minutes != 135 {
		t.Errorf("Tuesday holds %d minutes", got.Minutes)
	}
	if got := w.sheet.On(w.week.AddDate(0, 0, 0)); got.Day != "2026-09-28" || len(got.Tasks) != 0 {
		t.Errorf("a day nothing was done on reads %+v", got)
	}
	shot := drawn(t, Window{tab: Time, width: 1200, height: 800, timeState: timeState{week: w.week, sheet: w.sheet}})
	l := LayoutFor(1200, 800)
	if got := ink(shot, WeekBar(WeekRowAt(WeekDayAt(l, 4), 0), 120, 120), Dark().Panel); got == 0 {
		t.Error("Friday's bar is not drawn")
	}
}

func TestTheWeekAnswersItsOwnKeys(t *testing.T) {
	for name, want := range map[string][2]bool{"left": {true, false}, "t": {true, false}, "tab": {false, true}, "shift+b": {false, true}, "e": {false, false}} {
		if mine, shared := weekKeys(name); mine != want[0] || shared != want[1] {
			t.Errorf("%s is the page's own %v, shared %v", name, mine, shared)
		}
	}
	f := newFake()
	f.fail = errors.New("the timesheet cannot be read")
	if got := (Window{work: f, tab: Time}).readingWeek(); got.Wrong() == nil {
		t.Error("a timesheet that cannot be read says nothing")
	}
	if got := (Window{work: newFake(), tab: Time}).Key("e"); got.Screen() != OnTheBoard {
		t.Errorf("a key with no card to act on left the window on %v", got.Screen())
	}
}
