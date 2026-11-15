package engine

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/exchange"
	"cairn/internal/engine/label"
	"fmt"
	"maps"
	"slices"
	"time"

	"cairn/internal/engine/act"
)

func countedFromToday(read exchanged, today time.Time) (exchanged, error) {
	for i, o := range read.Objectives {
		day, err := calendar.FromToday(o.Deadline, today)
		if err != nil {
			return exchanged{}, fmt.Errorf("objective %d: %w", o.Number, err)
		}
		read.Objectives[i].Deadline = day
	}
	for i, t := range read.Tasks {
		day, err := calendar.FromToday(t.Deadline, today)
		if err != nil {
			return exchanged{}, fmt.Errorf("task %d: %w", t.Number, err)
		}
		read.Tasks[i].Deadline = day
		if t.Spent == nil {
			continue
		}
		spent := map[string]string{}
		for said, length := range t.Spent {
			day, err := calendar.FromToday(said, today)
			if err != nil {
				return exchanged{}, fmt.Errorf("task %d: %w", t.Number, err)
			}
			spent[day] = length
		}
		read.Tasks[i].Spent = spent
	}
	series, err := exchange.SeriesFromToday(read.Series, today)
	read.Series = series
	return read, err
}

func (e *Engine) spentOnDays(number int, spent map[string]string) error {
	for _, said := range slices.Sorted(maps.Keys(spent)) {
		day, err := calendar.ParseDeadline(said)
		if err != nil {
			return err
		}
		if err := e.addTime(number, day, spent[said]); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) spentByDay() (map[int]map[string]string, error) {
	acts, err := e.since(time.Time{})
	if err != nil {
		return nil, err
	}
	now := time.Now()
	first := startOfDay(now)
	for _, a := range acts {
		c := a.Carried()
		at := c.At.Local()
		if c.Did == act.TimeAdded {
			if day, err := calendar.ParseDeadline(c.Was); err == nil {
				at = day
			}
		}
		if day := startOfDay(at); day.Before(first) {
			first = day
		}
	}
	out := map[int]map[string]string{}
	for _, day := range act.Timed(act.WithoutTimeTaken(acts), first, startOfDay(now), now).Days {
		for _, w := range day.Tasks {
			if out[w.Task] == nil {
				out[w.Task] = map[string]string{}
			}
			out[w.Task][calendar.FormatDay(day.Day)] = calendar.FormatDuration(w.Took)
		}
	}
	return out, nil
}

func (e *Engine) labelsFiled(labels []exchange.LabelFiled) error {
	held, err := e.everyLabel()
	if err != nil {
		return err
	}
	for _, l := range labels {
		name, _ := label.OneLabel(l.Label)
		if _, there := held[name]; there {
			continue
		}
		if err := e.store.KeepLabel(name); err != nil {
			return err
		}
		if l.Scope != "" {
			if err := e.store.ScopeLabel(name, l.Scope); err != nil {
				return err
			}
		}
		if l.Colour == "" {
			continue
		}
		if err := e.store.ColourLabel(name, l.Colour); err != nil {
			return err
		}
	}
	return nil
}
