package engine

import (
	"slices"
	"time"

	"cairn/internal/domain"
	"cairn/internal/trace"
)

// TraceShow says what has been done to a task.
var TraceShow = Operation{
	Thing:   "trace",
	Verb:    "show",
	Short:   "Show what has been done to a task",
	Does:    "Show everything done to a task, oldest first, with what each change went from and to.",
	Example: "cairn trace show 1",
	Level:   Read,
	Needs:   AStore,
	Answers: Acts{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		acts, err := on.Engine.historyOf(n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Acts: acts}, nil
	},
}

// TraceList says what was done over a period, across every task.
var TraceList = Operation{
	Thing: "trace",
	Verb:  "list",
	Short: "List what was done over a period",
	Does:  "List what was done from a day onward, today when none is given, up to a last day when one is, across every task, each act saying who did it and through what — or only what one author did.",
	Example: `cairn trace list
cairn trace list --since 2026-09-01
cairn trace list --since 2026-09-29 --until 2026-09-29
cairn trace list --by sam`,
	Level:   Read,
	Needs:   AStore,
	Answers: Acts{},
	Takes: []Argument{
		{Name: "since", Kind: Day, Means: "the first day to look at; today when left out"},
		{Name: "until", Kind: Day, Means: "the last day to look at, the whole of it; no end when left out"},
		{Name: "by", Kind: Text, Means: "keep only what this author did: a name, a login, an agent, a Claude session"},
	},
	Run: func(on On, in Input) (Answer, error) {
		from, err := sinceWhen(in)
		if err != nil {
			return Answer{}, err
		}
		end, err := endOf(in)
		if err != nil {
			return Answer{}, err
		}
		acts, err := on.Engine.since(from)
		if err != nil {
			return Answer{}, err
		}
		acts = slices.DeleteFunc(acts, func(a trace.Act) bool { return !a.Carried().At.Before(end) })
		if by := in.Given["by"]; by != "" {
			acts = slices.DeleteFunc(acts, func(a trace.Act) bool { return a.Carried().By != by })
		}
		return Answer{Since: acts}, nil
	},
}

func sinceWhen(in Input) (time.Time, error) {
	given := in.Given["since"]
	if given == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	}
	return dayStarting(given)
}

func endOf(in Input) (time.Time, error) {
	given := in.Given["until"]
	if given == "" {
		return time.Date(9999, 12, 31, 0, 0, 0, 0, time.Local), nil
	}
	last, err := dayStarting(given)
	return last.AddDate(0, 0, 1), err
}

func dayStarting(given string) (time.Time, error) {
	parsed, err := domain.ParseDeadline(given)
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.Local), err
}
