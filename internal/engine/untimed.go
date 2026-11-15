package engine

import (
	"cairn/internal/engine/objective"
	"cairn/internal/engine/series"
	"cairn/internal/engine/task"
	"cairn/internal/engine/view"
	"cairn/internal/store/files"
)

func (e *Engine) timeShown() bool {
	if e.home.Dir == "" {
		return true
	}
	read, err := e.configured()
	return err != nil || read.TimeMode() == files.TimeOn
}

func (a Answer) untimed() Answer {
	a.Untimed = true
	if a.Task != nil {
		one := task.WithoutTime(*a.Task)
		a.Task = &one
	}
	for _, all := range []*[]task.Task{&a.Tasks, &a.Matched, &a.Under, &a.Held} {
		*all = task.EveryWithoutTime(*all)
	}
	a.Spent = nil
	if a.Family != nil {
		f := *a.Family
		f.Children = append([]objective.Child(nil), f.Children...)
		for i := range f.Children {
			f.Children[i].Spent = 0
		}
		a.Family = &f
	}
	if a.Period != nil {
		p := *a.Period
		p.Months = nil
		a.Period = &p
	}
	if a.Viewed != nil {
		v := *a.Viewed
		v.Objectives = append([]view.Objective(nil), v.Objectives...)
		for i := range v.Objectives {
			v.Objectives[i].TimeSpent = ""
		}
		a.Viewed = &v
	}
	if a.SeriesShown != nil {
		s := a.SeriesShown.untimed()
		a.SeriesShown = &s
	}
	if a.SeriesKept != nil {
		kept := seriesListed{Series: make([]Series, 0, len(a.SeriesKept.Series))}
		for _, s := range a.SeriesKept.Series {
			kept.Series = append(kept.Series, s.untimed())
		}
		a.SeriesKept = &kept
	}
	return a
}

func (a Answer) review() objective.Review {
	r := reviewOf(a.Under)
	r.Untimed = a.Untimed
	return r
}

func (s Series) untimed() Series {
	s.TimeSpent = ""
	s.Occurrences = append([]series.Occurrence(nil), s.Occurrences...)
	for i := range s.Occurrences {
		s.Occurrences[i].TimeSpent = ""
	}
	return s
}
