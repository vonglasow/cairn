package engine

import (
	"errors"
	"testing"
	"time"

	"cairn/internal/domain"
)

func dayNamed(day string) time.Time {
	parsed, err := domain.ParseDeadline(day)
	if err != nil {
		panic(err)
	}
	return parsed
}

// A week starts on the Monday a day is in: a Sunday belongs to the week
// before it, a Monday to its own — whatever the hour.
func TestAWeekStartsOnItsMonday(t *testing.T) {
	for day, want := range map[string]string{"2026-10-14": "2026-10-12", "2026-10-18": "2026-10-12", "2026-10-12": "2026-10-12", "2026-10-19": "2026-10-19"} {
		if got := domain.FormatDay(Monday(dayNamed(day).Add(15 * time.Hour))); got != want {
			t.Errorf("the week of %s starts on %s, want %s", day, got, want)
		}
	}
}

// A month walked from the 31st lands on the last day of a shorter month.
func TestMonthOn_keepsTheDayItCan(t *testing.T) {
	if got := domain.FormatDay(MonthOn(dayNamed("2026-01-31"), 1)); got != "2026-02-28" {
		t.Errorf("a month after the 31st of January is %s", got)
	}
	if got := domain.FormatDay(FirstOfMonth(dayNamed("2026-01-31"))); got != "2026-01-01" {
		t.Errorf("the first of January's month is %s", got)
	}
}

// A period is read up to today: the engine refuses what has not happened, so
// a week in progress stops at today, and one wholly to come holds nothing and
// asks for nothing.
func TestReadPeriod_readsUpToToday(t *testing.T) {
	f := newFake()
	today := dayNamed("2026-10-14")
	if _, err := ReadPeriod(f, dayNamed("2026-10-12"), dayNamed("2026-10-18"), today); err != nil {
		t.Fatal(err)
	}
	if len(f.asked) != 1 || f.asked[0].in.Given["until"] != "2026-10-14" || f.asked[0].in.Given["since"] != "2026-10-12" {
		t.Fatalf("asked %+v, want the week up to today", f.asked)
	}
	if _, err := ReadPeriod(f, dayNamed("2026-10-19"), dayNamed("2026-10-25"), today); err != nil || len(f.asked) != 1 {
		t.Errorf("a week to come asked %d times, err %v", len(f.asked)-1, err)
	}
	f.fail = errNope
	if _, err := ReadPeriod(f, dayNamed("2026-10-12"), dayNamed("2026-10-18"), today); !errors.Is(err, errNope) {
		t.Errorf("a store that does not answer gave %v", err)
	}
}

// A sheet says what one day holds, and a day nothing was done on is still a
// day, holding nothing.
func TestATimesheetSaysWhatADayHolds(t *testing.T) {
	sheet := Timesheet{Days: []TimesheetDay{{Day: "2026-10-13", Minutes: 135, Tasks: []TimesheetTask{
		{Number: 4, Title: "Ship it", Minutes: 120}, {Number: 5, Title: "Spec", Minutes: 15}}}}}
	day := sheet.On(dayNamed("2026-10-13"))
	if day.Says() != "2h15m" || day.Tasks[0].Says() != "2h  #4 Ship it" {
		t.Errorf("the day says %q and its first task %q", day.Says(), day.Tasks[0].Says())
	}
	if empty := sheet.On(dayNamed("2026-10-12")); empty.Day != "2026-10-12" || len(empty.Tasks) != 0 {
		t.Errorf("a day nothing was done on reads %+v", empty)
	}
	if sheet.Longest() != 120 {
		t.Errorf("the longest stretch is %d", sheet.Longest())
	}
}

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
