package engine

import (
	"time"

	"cairn/internal/domain"
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
	Does:  "List what was done from a day onward, today when none is given, across every task.",
	Example: `cairn trace list
cairn trace list --since 2026-09-01`,
	Level:   Read,
	Needs:   AStore,
	Answers: Acts{},
	Takes: []Argument{
		{Name: "since", Kind: Day, Means: "the first day to look at; today when left out"},
	},
	Run: func(on On, in Input) (Answer, error) {
		from, err := sinceWhen(in)
		if err != nil {
			return Answer{}, err
		}
		acts, err := on.Engine.since(from)
		if err != nil {
			return Answer{}, err
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
	return domain.ParseDeadline(given)
}
