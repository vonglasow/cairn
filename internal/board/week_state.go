package board

import (
	"cairn/internal/engine"
	"cairn/internal/engine/calendar"
	"strconv"
	"time"
)

// Holds is how much a page holds, which is what its tab says beside its name.
func (b Board) Holds(tab Tab) int {
	if tab == Objectives {
		return len(b.Showing().Goals)
	}
	if tab == Series {
		return len(b.SeriesListed())
	}
	if tab == Time {
		return b.Sheet.Worked()
	}
	return len(Laid(b.Showing().Cards, b.Archived))
}

// Faded says a day of the grid lies in the month before or after the one the keyboard is in, and is drawn faint: there to finish a week, not to be read.
func (b Board) Faded(day time.Time) bool { return day.Month() != b.Day.Month() }

// CellMonths is how many months one cell of the calendar holds.
func (b Board) CellMonths() int {
	if b.CalendarView == ViewYears {
		return 12
	}
	return 1
}

// PageMonths is how many months one page of the calendar moves by.
func (b Board) PageMonths() int {
	switch b.CalendarView {
	case ViewYears:
		return 144
	case ViewMonths:
		return 12
	}
	return 1
}

// TabCount is what a page's tab counts: what it shows, and of how many when the search or the filters narrow it.
func (b Board) TabCount(page Tab) string {
	shown := strconv.Itoa(b.Holds(page))
	if page == Series && (b.SeriesIn != 0 || !b.Looking.Empty()) {
		return shown + "/" + strconv.Itoa(len(b.Snap.Series))
	}
	if page == Series ||
		page == Time || (b.Looking.Empty() && !b.Filters.Narrows() && b.Viewing == "") {
		return shown
	}
	if page == Objectives {
		if all := len(b.Reading().Goals); b.Holds(page) != all {
			return shown + "/" + strconv.Itoa(all)
		}
		return shown
	}
	return shown + "/" + strconv.Itoa(len(Laid(b.Reading().Cards, b.Archived)))
}

// TimeShown is the time spent as the page shows it: every task of the period, or only those the search reaches, each day's total counting only those.
func (b Board) TimeShown() engine.Timesheet {
	if b.Looking.Empty() && !b.Filters.Narrows() {
		return b.Sheet
	}
	reached := map[int]bool{}
	for _, c := range b.Showing().Cards {
		reached[c.Number] = true
	}
	shown := engine.Timesheet{}
	for _, d := range b.Sheet.Days {
		kept := engine.TimesheetDay{Day: d.Day}
		for _, one := range d.Tasks {
			if reached[one.Number] {
				kept.Tasks = append(kept.Tasks, one)
				kept.Minutes += one.Minutes
			}
		}
		shown.Days = append(shown.Days, kept)
	}
	return shown
}

// WeekDays is the days of the week in sight that the page lays out.
func (b Board) WeekDays() []time.Time {
	days := make([]time.Time, b.Pages.WeekDays())
	for i := range days {
		days[i] = b.Week.AddDate(0, 0, i)
	}
	return days
}

// DayHeld is the i-th day of the week in sight, as the page shows it.
func (b Board) DayHeld(i int) engine.TimesheetDay {
	return b.TimeShown().On(b.Week.AddDate(0, 0, i))
}

// Stepped is the period the time tab moves to, some pages on or back.
func (b Board) Stepped(by int) time.Time {
	switch {
	case b.Monthly && b.Grid == ViewYears:
		return MonthOn(b.Week, 144*by)
	case b.Monthly && b.Grid == ViewMonths:
		return MonthOn(b.Week, 12*by)
	case b.Monthly:
		return MonthOn(b.Week, by)
	case b.Daily:
		return b.Week.AddDate(0, 0, by)
	}
	return b.Week.AddDate(0, 0, 7*by)
}

// Ranging is what the time tab shows a page of: a day, a week or a month.
func (b Board) Ranging() string {
	switch {
	case b.Monthly:
		return "Month"
	case b.Daily:
		return "Day"
	}
	return "Week"
}

// DayFrom is the first task of the i-th day of the week in sight.
func (b Board) DayFrom(i int) int { return b.DaysFrom[i] }

// MonthDay is the sheet's day for the i-th day of the month's six weeks.
func (b Board) MonthDay(i int) engine.TimesheetDay {
	day := calendar.FormatDay(MonthGrid(b.Week)[i])
	for _, d := range b.TimeShown().Days {
		if d.Day == day {
			return d
		}
	}
	return engine.TimesheetDay{Day: day}
}

// GoneOnTime says whether a task the timesheet names was deleted since.
func (b Board) GoneOnTime(number int) bool {
	for _, day := range b.Sheet.Days {
		for _, one := range day.Tasks {
			if one.Number == number {
				return one.Deleted
			}
		}
	}
	return false
}

// PeriodSays is the time spent over the whole period shown, and nothing when none was.
func (b Board) PeriodSays() string {
	minutes := 0
	for _, day := range b.TimeShown().Days {
		minutes += day.Minutes
	}
	if minutes == 0 {
		return ""
	}
	return calendar.FormatDuration(time.Duration(minutes) * time.Minute)
}
