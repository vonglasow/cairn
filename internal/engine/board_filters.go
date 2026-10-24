package engine

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
)

// Any is the value a filter opens on: it narrows nothing.
const Any = "any"

// Filter is one of the narrowings a board offers on top of its search: what it is called, the argument of task list it stands for, the values it walks through and how each reads, and the one it is on.
type Filter struct {
	Name   string
	Arg    string
	Values []string
	Shown  []string
	At     int
}

// Value is what the filter is on.
func (f Filter) Value() string { return f.Values[f.At] }

// Says is how what the filter is on reads.
func (f Filter) Says() string { return f.Shown[f.At] }

// Filters are a board's filters, in the order it lists them.
type Filters []Filter

var dueIn = []int{0, 7, 30}

// BoardFilters are the filters a board offers over what it read: a column, a priority, a label some card carries, an objective there is, and a day things are due by — today, in a week, in a month — each opening on Any.
func BoardFilters(s Snapshot, today time.Time) Filters {
	var labels []string
	for _, c := range s.Cards {
		for _, l := range c.Labels {
			if !slices.Contains(labels, l) {
				labels = append(labels, l)
			}
		}
	}
	slices.Sort(labels)
	var goals, named []string
	for _, g := range s.Goals {
		goals = append(goals, strconv.Itoa(g.Number))
		named = append(named, strings.TrimSpace("#"+strconv.Itoa(g.Number)+" "+g.Title))
	}
	var days []string
	for _, in := range dueIn {
		days = append(days, domain.FormatDay(today.AddDate(0, 0, in)))
	}
	return Filters{
		offered("column", "state", States(), nil),
		offered("priority", "priority", Priorities(), nil),
		offered("label", "label", labels, nil),
		offered("objective", "objective", goals, named),
		offered("due by", "due-before", days, nil),
	}
}

func offered(name, arg string, values, shown []string) Filter {
	if shown == nil {
		shown = values
	}
	return Filter{Name: name, Arg: arg, Values: append([]string{Any}, values...), Shown: append([]string{Any}, shown...)}
}

// Walked is the filters with one of them moved on, or back, through its values, coming round to Any again past the last.
func (fs Filters) Walked(at, by int) Filters {
	out := slices.Clone(fs)
	n := len(out[at].Values)
	out[at].At = ((out[at].At+by)%n + n) % n
	return out
}

// Cleared is the filters every one back on Any.
func (fs Filters) Cleared() Filters {
	out := slices.Clone(fs)
	for i := range out {
		out[i].At = 0
	}
	return out
}

// Keeping is these filters on the values the ones held were on, wherever such a value is still offered: a label nobody carries any more is let go of rather than narrowing to nothing.
func (fs Filters) Keeping(held Filters) Filters {
	out := slices.Clone(fs)
	for i := range out {
		for _, h := range held {
			if h.Arg != out[i].Arg {
				continue
			}
			if at := slices.Index(out[i].Values, h.Value()); at >= 0 {
				out[i].At = at
			}
		}
	}
	return out
}

// Narrows says whether any filter is on something other than Any.
func (fs Filters) Narrows() bool {
	for _, f := range fs {
		if f.Value() != Any {
			return true
		}
	}
	return false
}

// Said is what the filters narrow by, the way a board writes it beside its search: each one on something, named and what it is on.
func (fs Filters) Said() string {
	var said []string
	for _, f := range fs {
		if f.Value() != Any {
			said = append(said, f.Name+" "+f.Says())
		}
	}
	return strings.Join(said, " · ")
}

// Kept is the cards the filters keep, narrowed by the same rule task list narrows by.
func (fs Filters) Kept(cards []Card) ([]Card, error) {
	in := Input{Given: map[string]string{}, Listed: map[string][]string{}}
	for _, f := range fs {
		if f.Value() == Any {
			continue
		}
		if f.Arg == "label" || f.Arg == "objective" {
			in.Listed[f.Arg] = []string{f.Value()}
			continue
		}
		in.Given[f.Arg] = f.Value()
	}
	kept, err := keeping(in, TaskList.Takes)
	if err != nil {
		return nil, err
	}
	var out []Card
	for _, c := range cards {
		judged := &cardJudged{card: c, kept: true}
		kept.Tell(judged)
		if judged.kept {
			out = append(out, c)
		}
	}
	return out, nil
}

type cardJudged struct {
	card Card
	kept bool
}

func (j *cardJudged) InColumn(state string)    { j.kept = j.kept && j.card.State == state }
func (j *cardJudged) Carrying(priority string) { j.kept = j.kept && j.card.Priority == priority }
func (j *cardJudged) Serving(objective int) {
	j.kept = j.kept && slices.Contains(j.card.Objectives, objective)
}
func (j *cardJudged) Archived()     {}
func (j *cardJudged) WithArchived() {}

func (j *cardJudged) DueBefore(day string) {
	by, err := domain.ParseDeadline(day)
	j.kept = j.kept && err == nil && !j.card.Deadline.IsZero() && !j.card.Deadline.After(by)
}

func (j *cardJudged) Labelled(labels []string) {
	for _, l := range labels {
		j.kept = j.kept && slices.Contains(j.card.Labels, l)
	}
}
