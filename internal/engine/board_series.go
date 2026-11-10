package engine

import (
	"fmt"
	"math"
	"slices"
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

// RuleRead is a rule typed in a field read back: how often, what it counts from and when it ends, then the days it comes on; or why it is no rule.
type RuleRead struct{ Says, Next, Wrong string }

// RuleReadBack reads back a rule typed, counted from today, the way a form shows it under its field: nothing for nothing typed.
func RuleReadBack(typed string, today time.Time) RuleRead {
	if strings.TrimSpace(typed) == "" {
		return RuleRead{}
	}
	rule, err := domain.ParseRule(typed)
	if err != nil {
		wrong, _, _ := strings.Cut(err.Error(), " — ")
		return RuleRead{Wrong: wrong}
	}
	from := domain.DayOf(today)
	first, ok := rule.First(from)
	if !ok {
		return RuleRead{Says: rule.Said()}
	}
	if rule.AfterDone() {
		return RuleRead{Says: rule.Said(), Next: domain.ShortDay(first) + ", then " + rule.Then()}
	}
	days := []string{domain.ShortDay(first)}
	for _, d := range rule.Dates(first, from, 2) {
		days = append(days, domain.ShortDay(d))
	}
	return RuleRead{Says: rule.Said(), Next: strings.Join(days, " · ")}
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

// Recurring is what a board shows of a snapshot: the whole of it while a task may come back, and no series and no occurrence marked as one while recurrence is off.
func (p Pages) Recurring(s Snapshot) Snapshot {
	if p.Recurs() {
		return s
	}
	s.Series = nil
	cards := make([]Card, len(s.Cards))
	for i, c := range s.Cards {
		c.Series = nil
		cards[i] = c
	}
	s.Cards = cards
	return s
}

// ShownSeries is the series a board lists for what is typed: the active ones, then the stopped ones.
func ShownSeries(all []Series, query string) []Series {
	found := FoundSeries(all, query)
	out := make([]Series, 0, len(found))
	for _, state := range []string{"active", "stopped"} {
		for _, s := range found {
			if s.State == state {
				out = append(out, s)
			}
		}
	}
	return out
}

// SeriesFates is what may become of the tasks of a series deleted, in the order a board offers them: kept, the open one deleted, all deleted.
var SeriesFates = []string{"keep-tasks", "delete-open", "delete-tasks"}

// MadeStill is every task a series made that is still there.
func MadeStill(s Series) []int {
	var out []int
	for _, o := range s.Occurrences {
		if o.Task != 0 && o.Is != "deleted" {
			out = append(out, o.Task)
		}
	}
	return out
}

var templateArgs = []string{"title", "priority", "label", "estimate", "steps", "objective"}

// ChangesTheSeries says whether saving a form on a card asks if its series changes too: the card is an occurrence, and what its series makes the next one from was changed.
func ChangesTheSeries(c Card, fields []Field) bool {
	if c.Series == nil {
		return false
	}
	for _, f := range fields {
		if f.Value != f.Was && slices.Contains(templateArgs, f.Arg) {
			return true
		}
	}
	return false
}

// DaySaid is how a board writes a day of a series: as a card says its occurrence day while that is close and not passed, as its date otherwise; nothing for no day.
func DaySaid(day string, now time.Time) string {
	parsed, err := domain.ParseDeadline(day)
	if err != nil || parsed.IsZero() {
		return ""
	}
	if text, show, _, _ := OccurrenceDay(Card{Series: &Occurs{Day: day}}, now); show && !strings.HasPrefix(text, "−") {
		return text
	}
	return parsed.Format("2 Jan")
}
