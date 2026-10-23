package engine

import (
	"cmp"
	"slices"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// ObjectiveReport writes an objective's review over a period.
var ObjectiveReport = Operation{
	Thing: "objective",
	Verb:  "report",
	Short: "Review an objective over a period",
	Does:  "Write an objective's review over a period: how far along it is, the time spent on its tasks in the period month by month and in all, the tasks put in done in the period, and what is left.",
	Example: `cairn objective report 6 --since 2026-07-01
cairn objective report 6 --since 2026-07-01 --until 2026-09-30`,
	Level:   Read,
	Needs:   AStore,
	Answers: Review{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "an objective", Means: "the number of the objective"},
		{Name: "since", Kind: Day, Required: true, Means: "the first day of the period"},
		{Name: "until", Kind: Day, Means: "the last day of the period; today when left out"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		now := time.Now()
		first, err := dayOf(in, "since", now)
		if err != nil {
			return Answer{}, err
		}
		last, err := dayOf(in, "until", now)
		if err != nil {
			return Answer{}, err
		}
		o, under, p, err := readObjective(on.Engine, n)
		if err != nil {
			return Answer{}, err
		}
		period, err := on.Engine.periodOf(under, first, last)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Objective: &o, Under: under, Progress: p, Period: &period}, nil
	},
}

// Review is how an objective's review crosses: the objective as show gives it, the period, the time spent on its tasks in the period and month by month, and the tasks done in it.
type Review struct {
	Objective Objective    `json:"objective"`
	Since     string       `json:"since"`
	Until     string       `json:"until"`
	TimeSpent string       `json:"time_spent" jsonschema:"the time spent on its tasks in the period, as 1h30m; empty when none was"`
	Months    []MonthShape `json:"months"`
	Done      []DoneShape  `json:"done"`
}

// MonthShape is the time spent in one month of the period.
type MonthShape struct {
	Month     string `json:"month" jsonschema:"the month, as 2026-08"`
	TimeSpent string `json:"time_spent"`
}

// DoneShape is a task put in done within the period, and the day it was.
type DoneShape struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	On     string `json:"on"`
}

// OfReview turns an objective's review into what crosses.
func OfReview(o objective.Objective, under []task.Task, p domain.Progress, in objective.Period) Review {
	r := Review{
		Objective: OfObjective(o, under, p),
		Since:     domain.FormatDay(in.Since),
		Until:     domain.FormatDay(in.Until),
		TimeSpent: domain.FormatDuration(in.Spent()),
		Months:    []MonthShape{},
		Done:      []DoneShape{},
	}
	for _, m := range in.Months {
		r.Months = append(r.Months, MonthShape{Month: m.Month.Format("2006-01"), TimeSpent: domain.FormatDuration(m.Spent)})
	}
	for _, f := range in.Done {
		r.Done = append(r.Done, DoneShape{Number: f.Number, Title: f.Title, On: domain.FormatDay(f.On)})
	}
	return r
}

func (e *Engine) periodOf(under []task.Task, first, last time.Time) (objective.Period, error) {
	sheet, _, _, err := e.timed(first, last)
	if err != nil {
		return objective.Period{}, err
	}
	acts, err := e.since(first)
	if err != nil {
		return objective.Period{}, err
	}
	serving := map[int]bool{}
	for _, t := range under {
		serving[t.Carried().Number] = true
	}
	return objective.Period{Since: first, Until: last, Months: monthsOf(sheet, serving), Done: finishedIn(acts, under, first, last)}, nil
}

func monthsOf(sheet trace.Sheet, serving map[int]bool) []objective.Month {
	var months []objective.Month
	for _, day := range sheet.Days {
		var spent time.Duration
		for _, w := range day.Tasks {
			if serving[w.Task] {
				spent += w.Took
			}
		}
		if spent == 0 {
			continue
		}
		month := time.Date(day.Day.Year(), day.Day.Month(), 1, 0, 0, 0, 0, day.Day.Location())
		if n := len(months); n > 0 && months[n-1].Month.Equal(month) {
			months[n-1].Spent += spent
			continue
		}
		months = append(months, objective.Month{Month: month, Spent: spent})
	}
	return months
}

func finishedIn(acts []trace.Act, under []task.Task, first, last time.Time) []objective.Finished {
	end := last.AddDate(0, 0, 1)
	done := map[int]time.Time{}
	for _, act := range acts {
		c := act.Carried()
		if c.Field == "column" && c.Became == string(domain.Done) && !c.At.Before(first) && c.At.Before(end) {
			done[c.Task] = c.At
		}
	}
	var out []objective.Finished
	for _, t := range under {
		c := t.Carried()
		if on, found := done[c.Number]; found && c.State == domain.Done {
			out = append(out, objective.Finished{Number: c.Number, Title: c.Title, On: on})
		}
	}
	slices.SortFunc(out, func(a, b objective.Finished) int {
		return cmp.Or(a.On.Compare(b.On), cmp.Compare(a.Number, b.Number))
	})
	return out
}
