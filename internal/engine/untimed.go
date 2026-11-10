package engine

import (
	"cairn/internal/config"
	"cairn/internal/objective"
	"cairn/internal/task"
)

func (e *Engine) timeShown() bool {
	if e.home.Dir == "" {
		return true
	}
	read, err := e.configured()
	return err != nil || read.TimeMode() == config.TimeOn
}

func (a Answer) untimed() Answer {
	a.Untimed = true
	if a.Task != nil {
		one := withoutTime(*a.Task)
		a.Task = &one
	}
	for _, all := range []*[]task.Task{&a.Tasks, &a.Matched, &a.Under, &a.Held} {
		*all = everyWithoutTime(*all)
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
		v.Objectives = append([]ViewObjective(nil), v.Objectives...)
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
		kept := SeriesListed{Series: make([]Series, 0, len(a.SeriesKept.Series))}
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
	s.Occurrences = append([]Occurrence(nil), s.Occurrences...)
	for i := range s.Occurrences {
		s.Occurrences[i].TimeSpent = ""
	}
	return s
}

func everyWithoutTime(all []task.Task) []task.Task {
	if all == nil {
		return nil
	}
	out := make([]task.Task, 0, len(all))
	for _, t := range all {
		out = append(out, withoutTime(t))
	}
	return out
}

func withoutTime(t task.Task) task.Task {
	c := t.Carried()
	c.Spent = 0
	return task.Made(c)
}
