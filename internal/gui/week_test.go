package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"errors"
	"testing"
	"time"

	"cairn/internal/engine"
)

func dayNamed(day string) time.Time {
	parsed, err := calendar.ParseDeadline(day)
	if err != nil {
		panic(err)
	}
	return parsed
}

func TestTheWeekIsLaidOutToThePixel(t *testing.T) {
	l := geom.LayoutFor(1200, 800)
	first, last := WeekDayAt(l, 0), WeekDayAt(l, 6)
	if strip := TimeStrip(l); first.X != l.Column[0].X || first.Y != strip.Y+strip.H+between || first.Y+first.H != l.Band.Y+l.Band.H {
		t.Errorf("the first day is %+v, not from under the strip to the foot of the band", first)
	}
	right := l.Column[board.Columns-1].X + l.Column[board.Columns-1].W
	if gapLeft := right - (last.X + last.W); gapLeft < 0 || gapLeft >= 7 {
		t.Errorf("the last day ends %d short of the board's right edge", gapLeft)
	}
	if WeekDayAt(l, 1).X != first.X+first.W+geom.Gap {
		t.Errorf("the second day does not start a gap after the first")
	}
	if got, want := WeekHeadAt(geom.Box{X: 100, Y: 50, W: 200, H: 400}), (geom.Box{X: 112, Y: 62, W: 176, H: lineHigh}); got != want {
		t.Errorf("a day's heading is at %+v, want %+v", got, want)
	}
	row := WeekRowAt(geom.Box{X: 100, Y: 50, W: 200, H: 400}, 1)
	if want := (geom.Box{X: 112, Y: 50 + 2*lineHigh + padding + weekRow, W: 176, H: weekRow}); row != want {
		t.Errorf("the second task is at %+v, want %+v", row, want)
	}
	spent := geom.Box{X: 112, Y: row.Y + geom.CardHeight - geom.Gap, W: spentWide, H: lineHigh}
	if got := DaySpentAt(row); got != spent {
		t.Errorf("the time of the second task is said at %+v, want %+v", got, spent)
	}
	if got, want := WeekBar(row, 30, 120), (geom.Box{X: 112 + spentWide, Y: spent.Y + (lineHigh-barHighOf)/2, W: (176 - spentWide) / 4, H: barHighOf}); got != want {
		t.Errorf("a quarter of the longest is drawn %+v, want %+v", got, want)
	}
}

func TestTheWeekCountsWhatWasWorkedOn(t *testing.T) {
	w := Window{Board: board.Board{TimeState: board.TimeState{Week: dayNamed("2026-09-28"), Sheet: engine.Timesheet{Days: []engine.TimesheetDay{
		{Day: "2026-09-29", Minutes: 135, Tasks: []engine.TimesheetTask{{Number: 1, Minutes: 90}, {Number: 2, Minutes: 45}}},
		{Day: "2026-10-02", Minutes: 120, Tasks: []engine.TimesheetTask{{Number: 2, Minutes: 120}}},
	}}}}}
	if got := w.Sheet.Longest(); got != 120 {
		t.Errorf("the longest stretch is %d, want 120", got)
	}
	if got := w.Holds(board.Time); got != 2 {
		t.Errorf("the page says %d tasks, want 2", got)
	}
	if got := w.Sheet.On(w.Week.AddDate(0, 0, 1)); got.Minutes != 135 {
		t.Errorf("Tuesday holds %d minutes", got.Minutes)
	}
	if got := w.Sheet.On(w.Week.AddDate(0, 0, 0)); got.Day != "2026-09-28" || len(got.Tasks) != 0 {
		t.Errorf("a day nothing was done on reads %+v", got)
	}
	shot := drawn(t, Window{Board: board.Board{Tab: board.Time, TimeState: board.TimeState{Week: w.Week, Sheet: w.Sheet}}, width: 1200, height: 800})
	l := geom.LayoutFor(1200, 800)
	if got := ink(shot, WeekBar(WeekRowAt(WeekDayAt(l, 4), 0), 120, 120), design.Dark().Colour.Surface.Raised); got == 0 {
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
	if got := (Window{Board: board.Board{Tab: board.Time, Now: dayNamed("2026-09-30")}, work: f}).readingWeek(); got.Wrong() == nil {
		t.Error("a timesheet that cannot be read says nothing")
	}
	if got := (Window{Board: board.Board{Tab: board.Time}, work: newFake()}).Key("e"); got.Screen != board.OnTheBoard {
		t.Errorf("a key with no card to act on left the window on %v", got.Screen)
	}
}

func TestAWeekWithoutItsWeekendIsCutInFive(t *testing.T) {
	w := Window{Board: board.Board{Tab: board.Time, Pages: engine.Pages{}.WithoutWeekends(), TimeState: board.TimeState{Week: dayNamed("2026-09-28")}}, width: 1200, height: 800}
	days := w.WeekDays()
	if len(days) != 5 || calendar.FormatDay(days[0]) != "2026-09-28" || calendar.FormatDay(days[4]) != "2026-10-02" {
		t.Fatalf("the week shows %v, want Monday to Friday", days)
	}
	l := w.Layout()
	first, last := WeekDayAt(l, 0), WeekDayAt(l, 4)
	right := l.Column[board.Columns-1].X + l.Column[board.Columns-1].W
	if gapLeft := right - (last.X + last.W); gapLeft < 0 || gapLeft >= 5 {
		t.Errorf("the fifth day ends %d short of the board's right edge", gapLeft)
	}
	if first.W <= WeekDayAt(geom.LayoutFor(1200, 800), 0).W {
		t.Errorf("a day is %d wide, no wider than one of seven", first.W)
	}
	if saturday := MonthCellAt(l, 5); saturday != (geom.Box{}) {
		t.Errorf("a Saturday of a month is laid out at %+v", saturday)
	}
	if MonthCellAt(l, 7).X != MonthCellAt(l, 0).X || MonthCellAt(l, 4).X+MonthCellAt(l, 4).W > right {
		t.Errorf("a month's weeks are not laid out in five columns")
	}
	if got := TimeSays(w); got != "28 Sep – 2 Oct 2026" {
		t.Errorf("the strip says %q, want the week to end on its Friday", got)
	}
	drawn(t, w)
	w.Monthly = true
	drawn(t, w)
}

func TestAWindowWithNoClockYetReadsNoWeek(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800).Look()
	if !w.Week.IsZero() {
		t.Errorf("a window that does not know the day settled on the week of %v", w.Week)
	}
	if got := f.what("timesheet", "show"); got != 0 {
		t.Errorf("a window that does not know the day read %d timesheets", got)
	}
}
