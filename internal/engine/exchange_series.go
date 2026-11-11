package engine

import (
	"fmt"
	"slices"
	"time"

	"cairn/internal/domain"
	"cairn/internal/store"
	"cairn/internal/task"
)

type seriesFiled struct {
	Number   int         `json:"number"`
	Rule     string      `json:"rule"`
	From     string      `json:"from"`
	Stopped  string      `json:"stopped,omitempty"`
	Template Template    `json:"template"`
	Acts     []SeriesAct `json:"acts,omitempty"`
}

var seriesDid = []string{"made", "missed", "skipped", "changed", "stopped", "resumed", "deleted"}

func (e *Engine) seriesWritten() ([]seriesFiled, error) {
	every, err := e.store.EverySeries()
	if err != nil {
		return nil, err
	}
	out := []seriesFiled{}
	for _, kept := range every {
		s, err := e.aSeries(kept)
		if err != nil {
			return nil, err
		}
		out = append(out, seriesFiled{Number: kept.Number, Rule: kept.Rule, From: kept.Start, Stopped: kept.Stopped,
			Template: Template{Title: kept.Title, Priority: kept.Priority, Labels: listed(kept.Labels), Steps: listed(kept.Steps),
				Estimate: domain.FormatDuration(kept.Estimate), Objectives: listed(kept.Objectives), Arrives: kept.Arrives},
			Acts: s.trace()})
	}
	return out, nil
}

func seriesFromToday(read []seriesFiled, today time.Time) ([]seriesFiled, error) {
	for i, s := range read {
		for _, day := range []*string{&read[i].From, &read[i].Stopped} {
			counted, err := domain.FromToday(*day, today)
			if err != nil {
				return nil, fmt.Errorf("series %d: %w", s.Number, err)
			}
			*day = counted
		}
		for j, a := range s.Acts {
			counted, err := domain.FromToday(a.Day, today)
			if err != nil {
				return nil, fmt.Errorf("series %d: %w", s.Number, err)
			}
			read[i].Acts[j].Day = counted
		}
	}
	return read, nil
}

func checkedSeries(read []seriesFiled, objectives map[int]bool) error {
	seen := map[int]bool{}
	for _, s := range read {
		if seen[s.Number] {
			return fmt.Errorf("%w: series %d is in the file twice", ErrNotAnExport, s.Number)
		}
		seen[s.Number] = true
		for _, o := range s.Template.Objectives {
			if !objectives[o] {
				return fmt.Errorf("%w: series %d names objective %d — objective %d is not in the file", ErrExportPointsOutside, s.Number, o, o)
			}
		}
		if _, err := templateFiled(s); err != nil {
			return fmt.Errorf("series %d: %w", s.Number, err)
		}
		for _, a := range s.Acts {
			if _, err := actFiled(a); err != nil {
				return fmt.Errorf("series %d: %w", s.Number, err)
			}
		}
	}
	return nil
}

func templateFiled(s seriesFiled) (store.Series, error) {
	rule, err := domain.ParseRule(s.Rule)
	if err != nil {
		return store.Series{}, err
	}
	if s.From == "" {
		return store.Series{}, fmt.Errorf("%w: the day it counts from is missing", domain.ErrNotADay)
	}
	for _, day := range []string{s.From, s.Stopped} {
		if _, err := domain.ParseDeadline(day); err != nil {
			return store.Series{}, err
		}
	}
	if _, err := task.NewTask(s.Template.Title); err != nil {
		return store.Series{}, err
	}
	priority, err := domain.ParsePriority(s.Template.Priority)
	if err != nil {
		return store.Series{}, err
	}
	estimate, err := domain.ParseEstimate(s.Template.Estimate)
	if err != nil {
		return store.Series{}, err
	}
	labels, err := domain.ParseLabels(s.Template.Labels)
	if err != nil {
		return store.Series{}, err
	}
	return store.Series{Rule: rule.Words(), Start: s.From, Stopped: s.Stopped, Title: s.Template.Title, Priority: string(priority),
		Estimate: estimate, Steps: listed(s.Template.Steps), Labels: labels, Arrives: arrivalOf(domain.State(s.Template.Arrives))}, nil
}

func actFiled(a SeriesAct) (store.SeriesAct, error) {
	if !slices.Contains(seriesDid, a.Did) {
		return store.SeriesAct{}, fmt.Errorf("%w: %q is nothing a series does — %v", ErrNotAnExport, a.Did, seriesDid)
	}
	if _, err := domain.ParseDeadline(a.Day); err != nil {
		return store.SeriesAct{}, err
	}
	at := time.Now()
	if a.At != "" {
		read, err := time.Parse(time.RFC3339, a.At)
		if err != nil {
			return store.SeriesAct{}, fmt.Errorf("%w: the moment %q is written %s", ErrNotAnExport, a.At, time.RFC3339)
		}
		at = read
	}
	return store.SeriesAct{Did: a.Did, Task: a.Task, Day: a.Day, Field: a.Field, Was: a.Was, Became: a.Became,
		Through: a.Through, By: a.By, At: at}, nil
}

func (e *Engine) seriesFiled(read []seriesFiled, tasks, objectives map[int]int) error {
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
		for _, label := range made.Labels {
			if err := e.store.KeepLabel(label); err != nil {
				return err
			}
		}
		for _, a := range s.Acts {
			act, _ := actFiled(a)
			act.Series, act.Task = made.Number, tasks[a.Task]
			if err := e.store.RecordSeries(act); err != nil {
				return err
			}
		}
	}
	return nil
}
