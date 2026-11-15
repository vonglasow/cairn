package board

import (
	"cairn/internal/engine"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/series"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

func lookAtSeries(w Work) ([]engine.Series, error) {
	listed, err := w.Ask(engine.SeriesList, engine.Input{})
	if err != nil && engine.RefusalOf(err).Reason == engine.RefusalOf(engine.ErrRecurrenceOff).Reason {
		return nil, nil
	}
	if err != nil || listed.SeriesKept == nil {
		return nil, err
	}
	out := make([]engine.Series, 0, len(listed.SeriesKept.Series))
	for _, s := range listed.SeriesKept.Series {
		shown, err := w.Ask(engine.SeriesShow, Numbered(s.Number))
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
func (s Snapshot) SeriesOf(n int) (engine.Series, bool) {
	for _, one := range s.Series {
		if one.Number == n {
			return one, true
		}
	}
	return engine.Series{}, false
}

// foundSeries is the series carrying every word typed in their title, their rule or their number.
func foundSeries(all []engine.Series, query string) []engine.Series {
	if strings.TrimSpace(query) == "" {
		return all
	}
	var out []engine.Series
	for _, s := range all {
		if engine.Reaches([]string{s.Title, s.Rule, s.Says, "↻" + strconv.Itoa(s.Number)}, query) {
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
	day, err := calendar.ParseDeadline(c.Series.Day)
	if err != nil || day.IsZero() {
		return "", false, false, false
	}
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, now.Location())
	days := int(math.Round(day.Sub(calendar.StartOfDay(now)).Hours() / 24))
	if days < 0 {
		return fmt.Sprintf("−%dd", -days), true, true, false
	}
	if days == 0 && c.Series.At != "" {
		return c.Series.At, true, false, c.Series.At < now.Format("15:04")
	}
	if days == 0 {
		return "today", true, false, false
	}
	if days <= occurrenceSoon {
		return day.Format("Mon 2"), true, false, false
	}
	return "", false, false, false
}

// ruleRead is a rule typed in a field read back: how often, what it counts from and when it ends, then the days it comes on; or why it is no rule.
type ruleRead struct{ Says, Next, Wrong string }

// RuleReadBack reads back a rule typed, counted from today, the way a form shows it under its field: nothing for nothing typed.
func RuleReadBack(typed string, today time.Time) ruleRead {
	if strings.TrimSpace(typed) == "" {
		return ruleRead{}
	}
	rule, err := series.ParseRule(typed)
	if err != nil {
		wrong, _, _ := strings.Cut(err.Error(), " — ")
		return ruleRead{Wrong: wrong}
	}
	from := series.DayOf(today)
	first, ok := rule.First(from)
	if !ok {
		return ruleRead{Says: rule.Said()}
	}
	if rule.AfterDone() {
		return ruleRead{Says: rule.Said(), Next: series.ShortDay(first) + ", then " + rule.Then()}
	}
	days := []string{series.ShortDay(first)}
	for _, d := range rule.Dates(first, from, 2) {
		days = append(days, series.ShortDay(d))
	}
	return ruleRead{Says: rule.Said(), Next: strings.Join(days, " · ")}
}

// SeriesStates are the states the series' tab narrows to, every series first.
var SeriesStates = []string{"all", "active", "stopped"}

// InState is the series in that state; all of them for "all".
func InState(all []engine.Series, state string) []engine.Series {
	if state == SeriesStates[0] {
		return all
	}
	var out []engine.Series
	for _, s := range all {
		if s.State == state {
			out = append(out, s)
		}
	}
	return out
}

// MissedDays is every day a series missed, the latest first, written the short way.
func MissedDays(s engine.Series) []string {
	var out []string
	for _, o := range s.Occurrences {
		if o.Is == "missed" {
			out = append(out, series.ShortDays([]string{o.Day}))
		}
	}
	return out
}

// BeforeIn is what a series did before it made a task, the latest first, each on a line the way its trace writes it.
func BeforeIn(s engine.Series, task int) []string {
	var out []string
	for _, a := range s.Acts {
		if a.Did == "made" && a.Task == task {
			slices.Reverse(out)
			return out
		}
		at, _ := time.Parse(time.RFC3339, a.At)
		out = append(out, calendar.FormatMoment(at)+"  "+series.ActSaid(a))
	}
	return nil
}

// EachSaid is how long one occurrence took, on average over the ones done; nothing before one is.
func EachSaid(s engine.Series) string {
	spent, err := calendar.ParseEstimate(s.TimeSpent)
	if err != nil || s.Done == 0 || spent == 0 {
		return ""
	}
	return calendar.FormatDuration((spent / time.Duration(s.Done)).Round(time.Minute)) + " each"
}

// RuleListed is a rule the way a row of a list writes it; the words as they are when they are no rule.
func RuleListed(words string) string {
	rule, err := series.ParseRule(words)
	if err != nil {
		return words
	}
	return rule.Listed()
}

// NextDaySaid is how a list writes the next day of a series: as a board writes its day, with its time when that day is today; nothing when there is none.
func NextDaySaid(s engine.Series, now time.Time) string {
	if len(s.Next) == 0 {
		return ""
	}
	said := DaySaid(s.Next[0], now)
	if rule, err := series.ParseRule(s.Rule); err == nil && said == "today" && rule.At() != "" {
		said += " " + rule.At()
	}
	return said
}

// FromField is the field the day a new series counts from is given in, beside its rule.
func FromField() Field {
	return Field{Name: "from", Arg: "from", Means: meaning(engine.TaskNew, "from")}
}

// RepeatsField is the field a task is given a rule in, in plain words, to become the first occurrence of a new series.
func RepeatsField() Field {
	return Field{Name: "rule", Arg: "every", Means: meaning(engine.TaskNew, "every"), Example: "every weekday 10:30"}
}

// SeriesFields is the form of a series: what its next occurrence is made from, its rule and the day it counts from.
func SeriesFields(s engine.Series) []Field {
	steps := make([]engine.Part, 0, len(s.Template.Steps))
	for _, title := range s.Template.Steps {
		steps = append(steps, engine.Part{Title: title})
	}
	return openedAs(engine.SeriesEdit, "task", []Field{
		{Name: "title", Arg: "title", Value: s.Template.Title},
		{Name: "priority", Arg: "priority", Value: s.Template.Priority, Choices: Priorities()},
		{Name: "column", Arg: "state", Value: s.Template.Arrives, Choices: series.Arrivals},
		{Name: "labels", Arg: "label", Value: strings.Join(s.Template.Labels, " ")},
		{Name: "estimate", Arg: "estimate", Value: s.Template.Estimate},
		{Name: "objectives", Arg: "objective", Value: Numbers(s.Template.Objectives)},
		{Name: "rule", Arg: "every", Value: s.Rule},
		{Name: "from", Arg: "from", Value: s.From},
		{Name: "steps", Arg: "steps", Value: PartsWritten(steps), Long: true},
	})
}

// Recurring is what a board shows of a snapshot: the whole of it while a task may come back, and no series and no occurrence marked as one while recurrence is off.
func Recurring(p engine.Pages, s Snapshot) Snapshot {
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
func ShownSeries(all []engine.Series, query string) []engine.Series {
	found := foundSeries(all, query)
	out := make([]engine.Series, 0, len(found))
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
func MadeStill(s engine.Series) []int {
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
	parsed, err := calendar.ParseDeadline(day)
	if err != nil || parsed.IsZero() {
		return ""
	}
	if text, show, _, _ := OccurrenceDay(Card{Series: &series.Occurs{Day: day}}, now); show && !strings.HasPrefix(text, "−") {
		return text
	}
	return parsed.Format("2 Jan")
}
