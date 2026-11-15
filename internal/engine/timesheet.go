package engine

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/task"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"time"

	"cairn/internal/engine/act"
)

func dayOf(in Input, name string, now time.Time) (time.Time, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	given := in.Given[name]
	if given == "" {
		return today, nil
	}
	parsed, err := calendar.ParseDeadline(given)
	if err != nil {
		return time.Time{}, err
	}
	day := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, now.Location())
	if day.After(today) {
		return time.Time{}, fmt.Errorf("%q: %w", given, calendar.ErrDayToCome)
	}
	return day, nil
}

func (e *Engine) addTime(number int, day time.Time, length string) error {
	if _, err := e.tasks.Find(number); err != nil {
		return err
	}
	return e.record(number, act.TimeAdded, "", calendar.FormatDay(day), length)
}

func (e *Engine) timed(first, last time.Time) (act.Sheet, map[int]string, map[int]bool, error) {
	acts, err := e.since(first)
	if err != nil {
		return act.Sheet{}, nil, nil, err
	}
	sheet := act.Timed(act.WithoutTimeTaken(acts), first, last, time.Now())
	worked := map[int]bool{}
	for _, day := range sheet.Days {
		for _, w := range day.Tasks {
			worked[w.Task] = true
		}
	}
	titles, gone := map[int]string{}, map[int]bool{}
	for _, number := range slices.Sorted(maps.Keys(worked)) {
		t, err := e.tasks.Find(number)
		if errors.Is(err, task.ErrNoSuchTask) {
			titles[number], gone[number] = act.TitleWent(acts, number), true
			continue
		}
		if err != nil {
			return act.Sheet{}, nil, nil, err
		}
		titles[number] = t.Carried().Title
	}
	return sheet, titles, gone, nil
}

// TimesheetShow says how long the work took, day by day.
var TimesheetShow = Operation{
	Thing: "timesheet",
	Verb:  "show",
	Short: "Say how long the work took, day by day",
	Does:  "Say how long the work took, day by day and task by task: the time each task was in focus, the minutes several shared split between them, and what was added by hand — each to the minute begun, the day the sum of those, with no ceiling.",
	Example: `cairn timesheet show
cairn timesheet show --since 2026-09-01 --until 2026-09-07`,
	Level:   Read,
	Needs:   aStore,
	Answers: Timesheet{},
	Takes: []Argument{
		{Name: "since", Kind: day, Means: "the first day to say; today when left out"},
		{Name: "until", Kind: day, Means: "the last day to say; today when left out"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		now := time.Now()
		first, err := dayOf(in, "since", now)
		if err != nil {
			return Answer{}, err
		}
		last, err := dayOf(in, "until", now)
		if err != nil {
			return Answer{}, err
		}
		sheet, titles, gone, err := on.Engine.timed(first, last)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Sheet: &sheet, Named: titles, Gone: gone}, nil
	},
}

// TimesheetAdd adds time spent on a task without it being in focus.
var TimesheetAdd = Operation{
	Thing: "timesheet",
	Verb:  "add",
	Short: "Add time spent on a task without it being in focus",
	Does:  "Add time spent on a task without it being in focus — a meeting, an afternoon nobody focused — to the day it was spent. Less than nothing takes some back. Answers that day's timesheet.",
	Example: `cairn timesheet add 3 1h30m
cairn timesheet add 3 --on 2026-09-01 -- -30m`,
	Level:   write,
	Needs:   aStore,
	Again:   doesItAgain,
	Answers: Timesheet{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the task the time was spent on"},
		{Name: "time", Kind: Text, Positional: true, Required: true, Means: "how long: 1h, 45m, 1h30m; a length under zero, -30m, takes some back"},
		{Name: "on", Kind: day, Means: "the day it was spent; today when left out"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, _ := in.whole("number")
		length, err := time.ParseDuration(in.Given["time"])
		if err != nil || length == 0 {
			return Answer{}, fmt.Errorf("%q: %w", in.Given["time"], calendar.ErrNotATime)
		}
		day, err := dayOf(in, "on", time.Now())
		if err != nil {
			return Answer{}, err
		}
		if err := on.Engine.addTime(n, day, in.Given["time"]); err != nil {
			return Answer{}, err
		}
		sheet, titles, gone, err := on.Engine.timed(day, day)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Sheet: &sheet, Named: titles, Gone: gone}, nil
	},
}

// On is what the sheet holds of one day; a day nothing was done on is not in it, and reads as that day holding nothing.
func (s Timesheet) On(day time.Time) TimesheetDay {
	named := calendar.FormatDay(day)
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

// Worked is how many tasks the sheet holds time for, each counted once however many days it was worked on.
func (s Timesheet) Worked() int {
	worked := map[int]bool{}
	for _, d := range s.Days {
		for _, one := range d.Tasks {
			worked[one.Number] = true
		}
	}
	return len(worked)
}

// Says is a day's total, as a duration is written.
func (d TimesheetDay) Says() string {
	return calendar.FormatDuration(time.Duration(d.Minutes) * time.Minute)
}

// Spent is how long one task took on a day, as a duration is written.
func (t TimesheetTask) Spent() string {
	return calendar.FormatDuration(time.Duration(t.Minutes) * time.Minute)
}

// Says is one task's line on a day: its time, then which task.
func (t TimesheetTask) Says() string {
	return t.Spent() + "  #" + strconv.Itoa(t.Number) + " " + t.Shown()
}

// Shown is the task's title as the sheet writes it: said to be deleted when it was.
func (t TimesheetTask) Shown() string {
	if t.Deleted {
		return t.Title + deletedMark
	}
	return t.Title
}

const deletedMark = " (deleted)"
