package engine

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"cmp"
	"slices"
	"time"

	"cairn/internal/engine/act"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

// objectiveReport writes an objective's review over a period.
var objectiveReport = Operation{
	Thing: "objective",
	Verb:  "report",
	Short: "Report on an objective over a period",
	Does:  "Write an objective's report over a period: how far along it is, the time spent on its tasks in the period month by month and in all, the tasks put in done in the period, and what is left.",
	Example: `cairn objective report 6 --since 2026-07-01
cairn objective report 6 --since 2026-07-01 --until 2026-09-30`,
	Level:   Read,
	Needs:   aStore,
	Answers: review{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "an objective", Means: "the number of the objective"},
		{Name: "since", Kind: day, Required: true, Means: "the first day of the period"},
		{Name: "until", Kind: day, Means: "the last day of the period; today when left out"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, _ := in.whole("number")
		now := time.Now()
		first, err := dayOf(in, "since", now)
		if err != nil {
			return Answer{}, err
		}
		last, err := dayOf(in, "until", now)
		if err != nil {
			return Answer{}, err
		}
		read, err := readObjective(on.Engine, n)
		if err != nil {
			return Answer{}, err
		}
		parts, err := on.Engine.partsOf(n, *read.Family)
		if err != nil {
			return Answer{}, err
		}
		period, err := on.Engine.periodOf(read.Under, parts, first, last)
		if err != nil {
			return Answer{}, err
		}
		read.Period = &period
		return read, nil
	},
}

// review is how an objective's review crosses: the objective as show gives it, the period, the time spent on its tasks in the period and month by month, and the tasks done in it.
type review struct {
	Objective objectiveRow           `json:"objective"`
	Since     string                 `json:"since"`
	Until     string                 `json:"until"`
	TimeSpent string                 `json:"time_spent" jsonschema:"the time spent on its tasks in the period, as 1h30m; empty when none was"`
	Months    []objective.MonthShape `json:"months"`
	Done      []objective.DoneShape  `json:"done"`
}

// ofReview turns an objective's review into what crosses.
func ofReview(o objective.Objective, under []task.Task, p item.Progress, in objective.Period) review {
	r := review{
		Objective: ofObjective(o, under, p),
		Since:     calendar.FormatDay(in.Since),
		Until:     calendar.FormatDay(in.Until),
		TimeSpent: calendar.FormatDuration(in.Spent()),
		Months:    []objective.MonthShape{},
		Done:      []objective.DoneShape{},
	}
	for _, m := range in.Months {
		month := objective.MonthShape{Month: m.Month.Format("2006-01"), TimeSpent: calendar.FormatDuration(m.Spent)}
		for _, part := range m.Split {
			if part.Number == 0 {
				month.Itself = calendar.FormatDuration(part.Spent)
				continue
			}
			month.SubObjectives = append(month.SubObjectives, objective.PartShape{Number: part.Number, TimeSpent: calendar.FormatDuration(part.Spent)})
		}
		r.Months = append(r.Months, month)
	}
	for _, f := range in.Done {
		r.Done = append(r.Done, objective.DoneShape{Number: f.Number, Title: f.Title, On: calendar.FormatDay(f.On)})
	}
	return r
}

type part struct {
	number int
	tasks  []task.Task
}

func (e *Engine) partsOf(n int, f objective.Family) ([]part, error) {
	if len(f.Children) == 0 {
		return nil, nil
	}
	var parts []part
	for _, c := range append(slices.Clone(f.Children), objective.Child{}) {
		number := c.Objective.Carried().Number
		served := number
		if number == 0 {
			served = n
		}
		tasks, err := e.store.TasksServing(served)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part{number: number, tasks: tasks})
	}
	return parts, nil
}

func (e *Engine) periodOf(under []task.Task, parts []part, first, last time.Time) (objective.Period, error) {
	sheet, _, _, err := e.timed(first, last)
	if err != nil {
		return objective.Period{}, err
	}
	acts, err := e.since(first)
	if err != nil {
		return objective.Period{}, err
	}
	months := objective.MonthsOf(sheet, servingOf(under))
	for i, m := range months {
		for _, p := range parts {
			var spent time.Duration
			for _, pm := range objective.MonthsOf(sheet, servingOf(p.tasks)) {
				if pm.Month.Equal(m.Month) {
					spent = pm.Spent
				}
			}
			months[i].Split = append(months[i].Split, objective.Part{Number: p.number, Spent: spent})
		}
	}
	return objective.Period{Since: first, Until: last, Months: months, Done: finishedIn(acts, under, first, last)}, nil
}

func servingOf(tasks []task.Task) map[int]bool {
	serving := map[int]bool{}
	for _, t := range tasks {
		serving[t.Carried().Number] = true
	}
	return serving
}

func finishedIn(acts []act.Act, under []task.Task, first, last time.Time) []objective.Finished {
	end := last.AddDate(0, 0, 1)
	done := map[int]time.Time{}
	for _, act := range acts {
		c := act.Carried()
		if c.Field == "column" && c.Became == string(item.Done) && !c.At.Before(first) && c.At.Before(end) {
			done[c.Task] = c.At
		}
	}
	var out []objective.Finished
	for _, t := range under {
		c := t.Carried()
		if on, found := done[c.Number]; found && c.State == item.Done {
			out = append(out, objective.Finished{Number: c.Number, Title: c.Title, On: on})
		}
	}
	slices.SortFunc(out, func(a, b objective.Finished) int {
		return cmp.Or(a.On.Compare(b.On), cmp.Compare(a.Number, b.Number))
	})
	return out
}
