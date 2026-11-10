package engine

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
)

func lookAtSeries(w Work) ([]Series, error) {
	listed, err := w.Ask(SeriesList, Input{})
	if err != nil && RefusalOf(err).Reason == RefusalOf(ErrRecurrenceOff).Reason {
		return nil, nil
	}
	if err != nil || listed.SeriesKept == nil {
		return nil, err
	}
	out := make([]Series, 0, len(listed.SeriesKept.Series))
	for _, s := range listed.SeriesKept.Series {
		shown, err := w.Ask(SeriesShow, Numbered(s.Number))
		if err != nil {
			return nil, err
		}
		if shown.SeriesShown != nil {
			s = *shown.SeriesShown
		}
		out = append(out, s)
	}
	return out, nil
}

// SeriesOf is the series carrying that number, and whether there is one.
func (s Snapshot) SeriesOf(n int) (Series, bool) {
	for _, one := range s.Series {
		if one.Number == n {
			return one, true
		}
	}
	return Series{}, false
}

// FoundSeries is the series carrying every word typed in their title, their rule or their number.
func FoundSeries(all []Series, query string) []Series {
	if strings.TrimSpace(query) == "" {
		return all
	}
	var out []Series
	for _, s := range all {
		if Reaches([]string{s.Title, s.Rule, s.Says, "↻" + strconv.Itoa(s.Number)}, query) {
			out = append(out, s)
		}
	}
	return out
}

const occurrenceSoon = 7

// OccurrenceDay is what a card says of the day its series made it for: the time today, or today, a passed time said passed; the weekday and its date within a week; a day passed counted down as a deadline is; nothing further off, nor once it is done.
func OccurrenceDay(c Card, now time.Time) (text string, show, late, passed bool) {
	if c.Series == nil || Finished(c) {
		return "", false, false, false
	}
	day, err := domain.ParseDeadline(c.Series.Day)
	if err != nil || day.IsZero() {
		return "", false, false, false
	}
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, now.Location())
	days := int(math.Round(day.Sub(startOfDay(now)).Hours() / 24))
	switch {
	case days < 0:
		return fmt.Sprintf("−%dd", -days), true, true, false
	case days == 0 && c.Series.At != "":
		return c.Series.At, true, false, c.Series.At < now.Format("15:04")
	case days == 0:
		return "today", true, false, false
	case days <= occurrenceSoon:
		return day.Format("Mon 2"), true, false, false
	}
	return "", false, false, false
}

// RepeatsField is the field a task is given a rule in, in plain words, to become the first occurrence of a new series.
func RepeatsField() Field {
	return Field{Name: "repeats", Arg: "every", Means: meaning(TaskNew, "every"), Example: "every weekday 10:30"}
}

// SeriesFields is the form of a series: what its next occurrence is made from, its rule and the day it counts from.
func SeriesFields(s Series) []Field {
	steps := make([]Part, 0, len(s.Template.Steps))
	for _, title := range s.Template.Steps {
		steps = append(steps, Part{Title: title})
	}
	fields := opened(SeriesEdit, []Field{
		{Name: "title", Arg: "title", Value: s.Template.Title},
		{Name: "priority", Arg: "priority", Value: s.Template.Priority, Choices: Priorities()},
		{Name: "labels", Arg: "label", Value: strings.Join(s.Template.Labels, " ")},
		{Name: "estimate", Arg: "estimate", Value: s.Template.Estimate},
		{Name: "objectives", Arg: "objective", Value: Numbers(s.Template.Objectives)},
		{Name: "repeats", Arg: "every", Value: s.Rule},
		{Name: "from", Arg: "from", Value: s.From},
		{Name: "steps", Arg: "steps", Value: PartsWritten(steps), Long: true},
	})
	for i := range fields {
		if fields[i].Example == "" && len(fields[i].Choices) == 0 {
			fields[i].Example = typedLike["task "+fields[i].Arg]
		}
	}
	return fields
}
