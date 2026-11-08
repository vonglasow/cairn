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

// Filter is one of the narrowings a board offers on top of its search: what it is called, the argument of task list it stands for, the values it walks through and how each reads, the one it is on, and for a filter that takes several, the values ticked together.
type Filter struct {
	Name   string
	Arg    string
	Values []string
	Shown  []string
	At     int
	Many   bool
	Ticked []string
}

// Taken is every value the filter is on: none on Any, the one it is on, or every value ticked on a filter that takes several.
func (f Filter) Taken() []string {
	if len(f.Ticked) > 0 {
		return f.Ticked
	}
	if f.At == 0 {
		return nil
	}
	return []string{f.Values[f.At]}
}

// ReachesGoals says whether the filter means anything to an objective: what it carries — its labels, its priority — or which one it is.
func (f Filter) ReachesGoals() bool {
	return f.Arg == "label" || f.Arg == "priority" || f.Arg == "objective"
}

func (f Filter) shownOf(value string) string {
	if at := slices.Index(f.Values, value); at >= 0 {
		return f.Shown[at]
	}
	return value
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
		several(offered("label", "label", labels, nil)),
		offered("objective", "objective", goals, named),
		offered("due by", "due-before", days, nil),
	}
}

func several(f Filter) Filter {
	f.Many = true
	return f
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
	out[at].Ticked = nil
	return out
}

// Toggled is the filters with one value of a filter that takes several ticked, or let go of when it was ticked already.
func (fs Filters) Toggled(at int, value string) Filters {
	out := slices.Clone(fs)
	taken := slices.Clone(out[at].Taken())
	if i := slices.Index(taken, value); i >= 0 {
		taken = slices.Delete(taken, i, i+1)
	} else {
		taken = append(taken, value)
	}
	out[at].At, out[at].Ticked = 0, taken
	if len(taken) == 1 {
		out[at].At, out[at].Ticked = max(0, slices.Index(out[at].Values, taken[0])), nil
	}
	return out
}

// Cleared is the filters every one back on Any.
func (fs Filters) Cleared() Filters {
	out := slices.Clone(fs)
	for i := range out {
		out[i].At, out[i].Ticked = 0, nil
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
			var still []string
			for _, value := range h.Taken() {
				if slices.Contains(out[i].Values, value) {
					still = append(still, value)
				}
			}
			if len(still) > 1 {
				out[i].Ticked = still
			} else if len(still) == 1 {
				out[i].At = slices.Index(out[i].Values, still[0])
			}
		}
	}
	return out
}

// Narrows says whether any filter is on something other than Any.
func (fs Filters) Narrows() bool {
	for _, f := range fs {
		if len(f.Taken()) > 0 {
			return true
		}
	}
	return false
}

// Said is what the filters narrow by, the way a board writes it beside its search: each one on something, named and what it is on.
func (fs Filters) Said() string {
	var said []string
	for _, f := range fs {
		if taken := f.Taken(); len(taken) > 0 {
			said = append(said, f.Name+" "+f.TakenSays())
		}
	}
	return strings.Join(said, " · ")
}

// TakenSays is how what the filter is on reads: the values taken, joined by a plus.
func (f Filter) TakenSays() string {
	var shown []string
	for _, value := range f.Taken() {
		shown = append(shown, f.shownOf(value))
	}
	return strings.Join(shown, " + ")
}

// Counts is how many of the cards each value of a filter keeps, Any first, under every other filter as it stands.
func (fs Filters) Counts(cards []Card, at int) ([]int, error) {
	counts := make([]int, 0, len(fs[at].Values))
	for i := range fs[at].Values {
		one := slices.Clone(fs)
		one[at].At, one[at].Ticked = i, nil
		kept, err := one.Kept(cards)
		if err != nil {
			return nil, err
		}
		counts = append(counts, len(kept))
	}
	return counts, nil
}

// KeptGoals is the objectives the filters keep: by what an objective carries — its labels, its priority — or by which one it is; a filter that means nothing to an objective keeps every one.
func (fs Filters) KeptGoals(goals []Goal) []Goal {
	var out []Goal
	for _, g := range goals {
		if fs.keepsGoal(g) {
			out = append(out, g)
		}
	}
	return out
}

func (fs Filters) keepsGoal(g Goal) bool {
	for _, f := range fs {
		for _, value := range f.Taken() {
			switch f.Arg {
			case "label":
				if !slices.Contains(g.Labels, value) {
					return false
				}
			case "priority":
				if g.Priority != value {
					return false
				}
			case "objective":
				if strconv.Itoa(g.Number) != value {
					return false
				}
			}
		}
	}
	return true
}

// Kept is the cards the filters keep, narrowed by the same rule task list narrows by.
func (fs Filters) Kept(cards []Card) ([]Card, error) {
	in := Input{Given: map[string]string{}, Listed: map[string][]string{}}
	for _, f := range fs {
		taken := f.Taken()
		if len(taken) == 0 {
			continue
		}
		if f.Arg == "label" || f.Arg == "objective" {
			in.Listed[f.Arg] = taken
			continue
		}
		in.Given[f.Arg] = taken[0]
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
