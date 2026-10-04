package engine

// How long the work took, day by day — read from the focus taken and let go,
// and from time added by hand. See bdd/timesheet.feature and docs/trace.md.

import (
	"errors"
	"fmt"
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// ErrDayToCome is returned when a day has not happened yet: work cannot be
// said to have been done tomorrow.
var ErrDayToCome = errors.New("that has not happened yet")

// ErrNotATime is returned when time added by hand is not a length of time, or
// is none at all.
var ErrNotATime = errors.New("time added is an amount and a unit, as in 1h or 45m — less than nothing, as in -30m, takes some back")

// dayOf reads a day given, or today, as the start of that day where cairn runs.
func dayOf(in Input, name string, now time.Time) (time.Time, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	given := in.Given[name]
	if given == "" {
		return today, nil
	}
	parsed, err := domain.ParseDeadline(given)
	if err != nil {
		return time.Time{}, err
	}
	day := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, now.Location())
	if day.After(today) {
		return time.Time{}, fmt.Errorf("%q: %w", given, ErrDayToCome)
	}
	return day, nil
}

// AddTime writes down time spent on a task without it being in focus, for the
// day it was spent. Nothing is changed but the history: the time is read back
// from there, like the rest.
//
// The length is kept as it was given, once read: written back from a duration
// it needed a branch for its sign, and a branch nobody can reach the far side
// of, since no time at all is refused before this.
func (e *Engine) AddTime(number int, day time.Time, length string) error {
	if _, err := e.tasks.Find(number); err != nil {
		return err
	}
	return e.record(number, trace.TimeAdded, "", domain.FormatDay(day), length)
}

// Timed is how long the work took from one day to another, both counted whole,
// with what each task worked on is called.
func (e *Engine) Timed(first, last time.Time) (trace.Sheet, map[int]string, error) {
	acts, err := e.Since(first)
	if err != nil {
		return trace.Sheet{}, nil, err
	}
	sheet := trace.Timed(acts, first, last, time.Now())
	titles := map[int]string{}
	for _, day := range sheet.Days {
		for _, w := range day.Tasks {
			// A task deleted since keeps its time and loses its name: the
			// number still says which one it was.
			if t, err := e.tasks.Find(w.Task); err == nil {
				titles[w.Task] = task.OfTask(t).Title
			}
		}
	}
	return sheet, titles, nil
}

// TimesheetShow says how long the work took, day by day.
var TimesheetShow = Operation{
	Thing:   "timesheet",
	Verb:    "show",
	Short:   "Say how long the work took, day by day",
	Does:    "Say how long the work took, day by day and task by task: the time each task was in focus, the minutes several shared split between them, and what was added by hand — each rounded to the quarter hour, the day the sum of those, with no ceiling.",
	Level:   Read,
	Needs:   AStore,
	Answers: Timesheet{},
	Takes: []Argument{
		{Name: "since", Kind: Day, Means: "the first day to say; today when left out"},
		{Name: "until", Kind: Day, Means: "the last day to say; today when left out"},
	},
	Run: func(on On, in Input) (Answer, error) {
		now := time.Now()
		first, err := dayOf(in, "since", now)
		if err != nil {
			return Answer{}, err
		}
		last, err := dayOf(in, "until", now)
		if err != nil {
			return Answer{}, err
		}
		sheet, titles, err := on.Engine.Timed(first, last)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Sheet: &sheet, Named: titles}, nil
	},
}

// TimesheetAdd adds time spent on a task without it being in focus.
var TimesheetAdd = Operation{
	Thing:   "timesheet",
	Verb:    "add",
	Short:   "Add time spent on a task without it being in focus",
	Does:    "Add time spent on a task without it being in focus — a meeting, an afternoon nobody focused — to the day it was spent. Less than nothing takes some back. Answers that day's timesheet.",
	Level:   Write,
	Needs:   AStore,
	Answers: Timesheet{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the task the time was spent on"},
		{Name: "time", Kind: Text, Positional: true, Required: true, Means: "how long: 1h, 45m, 1h30m; -30m takes some back, written after -- on the command line"},
		{Name: "on", Kind: Day, Means: "the day it was spent; today when left out"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		length, err := time.ParseDuration(in.Given["time"])
		if err != nil || length == 0 {
			return Answer{}, fmt.Errorf("%q: %w", in.Given["time"], ErrNotATime)
		}
		day, err := dayOf(in, "on", time.Now())
		if err != nil {
			return Answer{}, err
		}
		if err := on.Engine.AddTime(n, day, in.Given["time"]); err != nil {
			return Answer{}, err
		}
		sheet, titles, err := on.Engine.Timed(day, day)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Sheet: &sheet, Named: titles}, nil
	},
}
