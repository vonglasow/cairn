package engine

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/exchange"
	"cairn/internal/engine/item"
	"cairn/internal/engine/label"
	"cairn/internal/engine/series"
	"fmt"

	"cairn/internal/engine/task"
	"cairn/internal/store/sqlite"
)

func (e *Engine) seriesWritten() ([]exchange.SeriesFiled, error) {
	every, err := e.store.EverySeries()
	if err != nil {
		return nil, err
	}
	out := []exchange.SeriesFiled{}
	for _, kept := range every {
		s, err := e.aSeries(kept)
		if err != nil {
			return nil, err
		}
		out = append(out, exchange.SeriesFiled{Number: kept.Number, Rule: kept.Rule, From: kept.Start, Stopped: kept.Stopped,
			Template: series.Template{Title: kept.Title, Priority: kept.Priority, Labels: listed(kept.Labels), Steps: listed(kept.Steps),
				Estimate: calendar.FormatDuration(kept.Estimate), Objectives: listed(kept.Objectives), Arrives: kept.Arrives},
			Acts: s.trace()})
	}
	return out, nil
}

func checkedSeries(read []exchange.SeriesFiled, objectives map[int]bool) error {
	seen := map[int]bool{}
	for _, s := range read {
		if seen[s.Number] {
			return fmt.Errorf("%w: series %d is in the file twice", exchange.ErrNotAnExport, s.Number)
		}
		seen[s.Number] = true
		for _, o := range s.Template.Objectives {
			if !objectives[o] {
				return fmt.Errorf("%w: series %d names objective %d — objective %d is not in the file", exchange.ErrExportPointsOutside, s.Number, o, o)
			}
		}
		if _, err := templateFiled(s); err != nil {
			return fmt.Errorf("series %d: %w", s.Number, err)
		}
		for _, a := range s.Acts {
			if _, err := exchange.ActFiled(a); err != nil {
				return fmt.Errorf("series %d: %w", s.Number, err)
			}
		}
	}
	return nil
}

func templateFiled(s exchange.SeriesFiled) (sqlite.Series, error) {
	rule, err := series.ParseRule(s.Rule)
	if err != nil {
		return sqlite.Series{}, err
	}
	if s.From == "" {
		return sqlite.Series{}, fmt.Errorf("%w: the day it counts from is missing", calendar.ErrNotADay)
	}
	for _, day := range []string{s.From, s.Stopped} {
		if _, err := calendar.ParseDeadline(day); err != nil {
			return sqlite.Series{}, err
		}
	}
	if _, err := task.NewTask(s.Template.Title); err != nil {
		return sqlite.Series{}, err
	}
	priority, err := item.ParsePriority(s.Template.Priority)
	if err != nil {
		return sqlite.Series{}, err
	}
	estimate, err := calendar.ParseEstimate(s.Template.Estimate)
	if err != nil {
		return sqlite.Series{}, err
	}
	labels, err := label.ParseLabels(s.Template.Labels)
	if err != nil {
		return sqlite.Series{}, err
	}
	return sqlite.Series{Rule: rule.Words(), Start: s.From, Stopped: s.Stopped, Title: s.Template.Title, Priority: string(priority),
		Estimate: estimate, Steps: listed(s.Template.Steps), Labels: labels, Arrives: series.ArrivalOf(item.State(s.Template.Arrives))}, nil
}

func (e *Engine) seriesFiled(read []exchange.SeriesFiled, tasks, objectives map[int]int) error {
	for _, s := range read {
		kept, err := templateFiled(s)
		if err != nil {
			return err
		}
		for _, o := range s.Template.Objectives {
			kept.Objectives = append(kept.Objectives, objectives[o])
		}
		made, err := e.store.KeepSeries(kept)
		if err != nil {
			return err
		}
		for _, name := range made.Labels {
			if err := e.store.KeepLabel(name); err != nil {
				return err
			}
		}
		for _, a := range s.Acts {
			act, _ := exchange.ActFiled(a)
			act.Series, act.Task = made.Number, tasks[a.Task]
			if err := e.store.RecordSeries(act); err != nil {
				return err
			}
		}
	}
	return nil
}
