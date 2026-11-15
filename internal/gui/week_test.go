package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
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
