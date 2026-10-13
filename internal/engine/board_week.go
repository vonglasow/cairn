package engine

import (
	"strconv"
	"time"

	"cairn/internal/domain"
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
func ReadPeriod(w Work, since, until, today time.Time) (Timesheet, error) {
	if until.After(today) {
		until = today
	}
	if since.After(until) {
		return Timesheet{}, nil
	}
	answer, err := w.Ask(TimesheetShow, Input{Given: map[string]string{
		"since": domain.FormatDay(since),
		"until": domain.FormatDay(until),
	}})
	if err != nil || answer.Sheet == nil {
		return Timesheet{}, err
	}
	sheet, _ := Of(TimesheetShow, answer).(Timesheet)
	return sheet, nil
}

// On is what the sheet holds of one day; a day nothing was done on is not in it, and reads as that day holding nothing.
func (s Timesheet) On(day time.Time) TimesheetDay {
	named := domain.FormatDay(day)
	for _, d := range s.Days {
		if d.Day == named {
			return d
		}
	}
	return TimesheetDay{Day: named}
}

// Longest is the longest stretch any task had on any day: what a bar of time is a share of.
func (s Timesheet) Longest() int {
	most := 0
	for _, d := range s.Days {
		for _, one := range d.Tasks {
			most = max(most, one.Minutes)
		}
	}
	return most
}

// Says is a day's total, as a duration is written.
func (d TimesheetDay) Says() string {
	return domain.FormatDuration(time.Duration(d.Minutes) * time.Minute)
}

// Says is one task's line on a day: its time, then which task.
func (t TimesheetTask) Says() string {
	return domain.FormatDuration(time.Duration(t.Minutes)*time.Minute) + "  #" + strconv.Itoa(t.Number) + " " + t.Title
}
