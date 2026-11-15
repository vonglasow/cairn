package board

import (
	"cairn/internal/engine"
	"cairn/internal/engine/calendar"
	"time"
)

const monthWeeks = 6

// DayOf is a moment's day, at midnight in the zone days are kept in.
func DayOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// Monday is the Monday of the week a moment is in, at midnight.
func Monday(t time.Time) time.Time {
	day := DayOf(t)
	return day.AddDate(0, 0, -((int(day.Weekday()) + 6) % 7))
}

// FirstOfMonth is the first day of the month a moment is in.
func FirstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// MonthOn is the same day so many months on, or the last day of that month when it is shorter.
func MonthOn(day time.Time, by int) time.Time {
	first := time.Date(day.Year(), day.Month()+time.Month(by), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(day.Day(), last)-1)
}

// MonthGrid is the six weeks a month is shown as, from the Monday on or before its first day.
func MonthGrid(day time.Time) []time.Time {
	first := FirstOfMonth(day)
	back := (int(first.Weekday()) + 6) % 7
	grid := make([]time.Time, monthWeeks*7)
	for i := range grid {
		grid[i] = first.AddDate(0, 0, i-back)
	}
	return grid
}

// ReadPeriod asks how long the work took from since to until, read up to today: the engine refuses what has not happened yet, so a period in progress stops at today and one wholly to come holds nothing to read.
func ReadPeriod(w Work, since, until, today time.Time) (engine.Timesheet, error) {
	if until.After(today) {
		until = today
	}
	if since.After(until) {
		return engine.Timesheet{}, nil
	}
	answer, err := w.Ask(engine.TimesheetShow, engine.Input{Given: map[string]string{
		"since": calendar.FormatDay(since),
		"until": calendar.FormatDay(until),
	}})
	if err != nil || answer.Sheet == nil {
		return engine.Timesheet{}, err
	}
	sheet, _ := engine.Of(engine.TimesheetShow, answer).(engine.Timesheet)
	return sheet, nil
}
