package engine

import (
	"time"

	"cairn/internal/config"
	"cairn/internal/domain"
	"cairn/internal/store"
	"cairn/internal/trace"
)

const throughSprint = "sprint"

// TodayVariable is the variable that names the day cairn takes for today when it sweeps the sprints, for a script or a scenario to stand on another day.
const TodayVariable = "CAIRN_TODAY"

func today(h Home) time.Time {
	if named, ok := h.lookup()(TodayVariable); ok {
		if day, err := time.ParseInLocation(time.DateOnly, named, time.Local); err == nil {
			return day
		}
	}
	return time.Now()
}

func (e *Engine) sprintSwept() error {
	if e.home.Dir == "" {
		return nil
	}
	read, err := config.Read(e.home.Dir, e.home.lookup())
	if err != nil {
		return nil
	}
	said, days, ok := read.Sprint()
	if !ok {
		return nil
	}
	start, err := time.ParseInLocation(time.DateOnly, said, time.Local)
	if err != nil {
		return err
	}
	return e.sweptBySprint(start, days, today(e.home))
}

func (e *Engine) sweptBySprint(start time.Time, days int, now time.Time) error {
	begun, ok := domain.SprintUnderWay(start, days, now)
	if !ok {
		return nil
	}
	return e.store.Within(func(kept *store.Store) error {
		in := e.keptBy(kept)
		in.acting.Through = throughSprint
		var done domain.Narrowing
		done.InColumn(string(domain.Done))
		found, err := in.tasks.Narrowed(&done)
		if err != nil {
			return err
		}
		for _, one := range found {
			number := factsOf(one).Number
			acts, err := in.history.ActsOn(number)
			if err != nil {
				return err
			}
			if went := trace.LastNamed(acts, "column"); went.IsZero() || !went.Before(begun) {
				continue
			}
			if err := in.archive(number); err != nil {
				return err
			}
		}
		return nil
	})
}
