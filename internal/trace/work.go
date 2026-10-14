package trace

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"time"

	"cairn/internal/domain"
)

// Sheet is how long the work took over some days: every day asked for, in the order they ran — one nothing was done on says so, rather than going missing.
type Sheet struct {
	Days []Day
}

// Day is one day of a sheet: the tasks worked on, each rounded, in the order their numbers run, and what those come to.
type Day struct {
	Day   time.Time
	Tasks []Worked
	Total time.Duration
}

// Worked is how long one task took on one day, rounded to the quarter hour, and whether it was still in focus when the day ended.
type Worked struct {
	Task     int
	Took     time.Duration
	LeftOpen bool
}

// LastNamed is the moment the last of these acts naming a field was done, the creation included, or nothing when none named it.
func LastNamed(acts []Act, field string) time.Time {
	var last time.Time
	for _, a := range acts {
		if f := a.Carried(); f.Field == field {
			last = f.At
		}
	}
	return last
}

func quarter() time.Duration { return 15 * time.Minute }

type span struct {
	task     int
	from, to time.Time
	leftOpen bool
}

type work struct {
	spans   []span
	byHand  map[time.Time]map[int]time.Duration
	undated map[int]time.Duration
}

func workedFrom(acts []Act, loc *time.Location, now time.Time) work {
	facts := make([]Carried, 0, len(acts))
	for _, a := range acts {
		facts = append(facts, a.Carried())
	}
	sort.SliceStable(facts, func(i, j int) bool { return facts[i].At.Before(facts[j].At) })

	w := work{byHand: map[time.Time]map[int]time.Duration{}, undated: map[int]time.Duration{}}
	open := map[int]time.Time{}
	for _, f := range facts {
		switch f.Did {
		case FocusTaken:
			if _, already := open[f.Task]; !already {
				open[f.Task] = f.At
			}
		case FocusReleased:
			if from, was := open[f.Task]; was {
				w.spans = append(w.spans, cutAtMidnight(f.Task, from, f.At, loc))
				delete(open, f.Task)
			}
		case TimeAdded:
			w.addedByHand(f, loc)
		}
	}
	for task, from := range open {
		w.spans = append(w.spans, cutAtMidnight(task, from, now, loc))
	}
	return w
}

func (w work) addedByHand(f Carried, loc *time.Location) {
	if f.Was == "" {
		if length, err := time.ParseDuration(f.Became); err == nil {
			w.undated[f.Task] += length
		}
		return
	}
	if on, length, read := addedOn(f, loc); read {
		if w.byHand[on] == nil {
			w.byHand[on] = map[int]time.Duration{}
		}
		w.byHand[on][f.Task] += length
	}
}

// Timed works out how long the work took on each day from the first to the last, both counted whole, in the location the first is given in.
func Timed(acts []Act, first, last, now time.Time) Sheet {
	w := workedFrom(acts, first.Location(), now)
	var sheet Sheet
	for day := first; !day.After(last); day = day.AddDate(0, 0, 1) {
		took, leftOpen := shared(w.spans, day, day.AddDate(0, 0, 1))
		for task, length := range w.byHand[day] {
			took[task] += length
		}
		worked := rounded(took, leftOpen)
		worked.Day = day
		sheet.Days = append(sheet.Days, worked)
	}
	return sheet
}

// Spent is the time spent on each task, all told: the time it was in focus, shared with any other in focus at once, and the time added by hand with a day or without one — a minute begun counted whole, and never less than nothing.
func Spent(acts []Act, now time.Time) map[int]time.Duration {
	w := workedFrom(acts, now.Location(), now)
	var first time.Time
	for _, s := range w.spans {
		if first.IsZero() || s.from.Before(first) {
			first = s.from
		}
	}
	took, _ := shared(w.spans, first, now)
	for _, day := range w.byHand {
		for task, length := range day {
			took[task] += length
		}
	}
	for task, length := range w.undated {
		took[task] += length
	}
	spent := map[int]time.Duration{}
	for task, length := range took {
		if length > 0 {
			spent[task] = minuteBegun(length)
		}
	}
	return spent
}

func minuteBegun(d time.Duration) time.Duration {
	whole := d.Truncate(time.Minute)
	if whole < d {
		whole += time.Minute
	}
	return whole
}

func addedOn(f Carried, loc *time.Location) (time.Time, time.Duration, bool) {
	day, err := domain.ParseDeadline(f.Was)
	if err != nil {
		return time.Time{}, 0, false
	}
	length, err := time.ParseDuration(f.Became)
	if err != nil {
		return time.Time{}, 0, false
	}
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc), length, true
}

func cutAtMidnight(task int, from, to time.Time, loc *time.Location) span {
	start := from.In(loc)
	midnight := time.Date(start.Year(), start.Month(), start.Day()+1, 0, 0, 0, 0, loc)
	if to.After(midnight) {
		return span{task: task, from: from, to: midnight, leftOpen: true}
	}
	return span{task: task, from: from, to: to}
}

type edge struct {
	at    time.Time
	task  int
	opens bool
}

func shared(spans []span, from, to time.Time) (map[int]time.Duration, map[int]bool) {
	took := map[int]time.Duration{}
	leftOpen := map[int]bool{}
	var edges []edge
	for _, s := range spans {
		if s.to.After(from) && s.from.Before(to) {
			edges = append(edges, edge{at: maxTime(s.from, from), task: s.task, opens: true}, edge{at: minTime(s.to, to), task: s.task})
			leftOpen[s.task] = leftOpen[s.task] || s.leftOpen
		}
	}
	sort.SliceStable(edges, func(i, j int) bool { return edges[i].at.Before(edges[j].at) })
	in := map[int]int{}
	for i, e := range edges {
		if i > 0 && len(in) > 0 {
			length := e.at.Sub(edges[i-1].at) / time.Duration(len(in))
			for task := range in {
				took[task] += length
			}
		}
		if e.opens {
			in[e.task]++
			continue
		}
		if in[e.task]--; in[e.task] == 0 {
			delete(in, e.task)
		}
	}
	return took, leftOpen
}

func rounded(took map[int]time.Duration, leftOpen map[int]bool) Day {
	var day Day
	for _, task := range slices.Sorted(maps.Keys(took)) {
		length := took[task]
		quarters := (max(0, length) + quarter()/2) / quarter()
		if quarters == 0 {
			continue
		}
		day.Tasks = append(day.Tasks, Worked{Task: task, Took: quarters * quarter(), LeftOpen: leftOpen[task]})
		day.Total += quarters * quarter()
	}
	return day
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// Hours writes a length of time the way a sheet is filled in: hours and minutes, the minutes two digits once there are hours, and nothing of a part that is nought — 1h30, 45m, 2h, and 0m for nothing at all.
func Hours(d time.Duration) string {
	minutes := int(d / time.Minute)
	h, m := minutes/60, minutes%60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%02d", h, m)
}
