package trace

// How long the work took, day by day. It is arithmetic over acts, so it lives
// with the acts and not with whoever displays it — and it is given as it was:
// every hour of a day, however many. Sorting the time into the kinds of work a
// timesheet asks for, and bringing a day down to what a sheet accepts, is done
// afterwards by whoever fills the sheet in.

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"time"

	"cairn/internal/domain"
)

// Sheet is how long the work took over some days: every day asked for, in the
// order they ran — one nothing was done on says so, rather than going missing.
type Sheet struct {
	Days []Day
}

// Day is one day of a sheet: the tasks worked on, each rounded, in the order
// their numbers run, and what those come to.
type Day struct {
	Day   time.Time
	Tasks []Worked
	Total time.Duration
}

// Worked is how long one task took on one day, rounded to the quarter hour,
// and whether it was still in focus when the day ended.
type Worked struct {
	Task     int
	Took     time.Duration
	LeftOpen bool
}

// quarter is what a sheet is filled in by. A function and not a constant: a
// constant's arithmetic runs on no line, so nothing would notice it changing.
func quarter() time.Duration { return 15 * time.Minute }

// span is one run of focus on one task, already stopped where a day stops it.
type span struct {
	task     int
	from, to time.Time
	leftOpen bool
}

// Timed works out how long the work took on each day from the first to the
// last, both counted whole, in the location the first is given in.
//
// A run of focus is a task taken up until it is let go. It stops at the first
// midnight after it began, whatever comes after: a focus left open would
// otherwise count the night, and every day after it, as work — and a day it
// crossed says so. One still open today runs to now, since it is work being
// done. Two tasks in focus at once share the minutes they overlap, so a day
// never comes to more than the clock. Time added by hand counts on the day it
// was added for, and less than nothing takes some back, never below nothing.
func Timed(acts []Act, first, last, now time.Time) Sheet {
	facts := make([]fact, 0, len(acts))
	for _, a := range acts {
		var f fact
		a.Tell(&f)
		facts = append(facts, f)
	}
	// In the order things happened, not the order they were written down: a
	// focus put right after the fact is recorded after what followed it.
	sort.SliceStable(facts, func(i, j int) bool { return facts[i].at.Before(facts[j].at) })

	loc := first.Location()
	var spans []span
	open := map[int]time.Time{}
	byHand := map[time.Time]map[int]time.Duration{}
	for _, f := range facts {
		switch f.did {
		case FocusTaken:
			if _, already := open[f.task]; !already {
				open[f.task] = f.at
			}
		case FocusReleased:
			if from, was := open[f.task]; was {
				spans = append(spans, cutAtMidnight(f.task, from, f.at, loc))
				delete(open, f.task)
			}
		case TimeAdded:
			day, err := domain.ParseDeadline(f.was)
			length, bad := time.ParseDuration(f.became)
			if err != nil || bad != nil {
				continue
			}
			on := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
			if byHand[on] == nil {
				byHand[on] = map[int]time.Duration{}
			}
			byHand[on][f.task] += length
		}
	}
	for task, from := range open {
		spans = append(spans, cutAtMidnight(task, from, now, loc))
	}

	var sheet Sheet
	for day := first; !day.After(last); day = day.AddDate(0, 0, 1) {
		took, leftOpen := shared(spans, day, day.AddDate(0, 0, 1))
		for task, length := range byHand[day] {
			took[task] += length
		}
		worked := rounded(took, leftOpen)
		worked.Day = day
		sheet.Days = append(sheet.Days, worked)
	}
	return sheet
}

// cutAtMidnight is a run of focus cut at the first midnight after it began.
func cutAtMidnight(task int, from, to time.Time, loc *time.Location) span {
	start := from.In(loc)
	midnight := time.Date(start.Year(), start.Month(), start.Day()+1, 0, 0, 0, 0, loc)
	if to.After(midnight) {
		return span{task: task, from: from, to: midnight, leftOpen: true}
	}
	return span{task: task, from: from, to: to}
}

// shared is how long each task was in focus between two moments, the minutes
// several were in focus at once shared out between them, and which of them
// were left open on that day.
func shared(spans []span, from, to time.Time) (map[int]time.Duration, map[int]bool) {
	took := map[int]time.Duration{}
	leftOpen := map[int]bool{}
	var edges []time.Time
	for _, s := range spans {
		if s.to.After(from) && s.from.Before(to) {
			edges = append(edges, maxTime(s.from, from), minTime(s.to, to))
			leftOpen[s.task] = leftOpen[s.task] || s.leftOpen
		}
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].Before(edges[j]) })
	for i := 1; i < len(edges); i++ {
		var in []int
		for _, s := range spans {
			if !s.from.After(edges[i-1]) && !s.to.Before(edges[i]) {
				in = append(in, s.task)
			}
		}
		for _, task := range in {
			took[task] += edges[i].Sub(edges[i-1]) / time.Duration(len(in))
		}
	}
	return took, leftOpen
}

// rounded is a day's tasks each brought to the nearest quarter hour, half a
// quarter up and never below nothing, and what those come to. A task that
// rounds to nothing is not written.
func rounded(took map[int]time.Duration, leftOpen map[int]bool) Day {
	var day Day
	// In the order the numbers run, read off the numbers themselves: a
	// comparison written out had a boundary no two tasks of a day can reach.
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

// Hours writes a length of time the way a sheet is filled in: hours and
// minutes, the minutes two digits once there are hours, and nothing of a part
// that is nought — 1h30, 45m, 2h, and 0m for nothing at all.
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

// fact is an act's membrane for what the arithmetic needs: which task, what was
// done, and when.
type fact struct {
	task        int
	did         string
	was, became string
	at          time.Time
}

func (f *fact) On(n int)          { f.task = n }
func (f *fact) Did(what string)   { f.did = what }
func (f *fact) Changing(string)   {}
func (f *fact) NothingNamed()     {}
func (f *fact) Was(v string)      { f.was = v }
func (f *fact) WasNothing()       {}
func (f *fact) Became(v string)   { f.became = v }
func (f *fact) BecameNothing()    {}
func (f *fact) Through(string)    {}
func (f *fact) ByHand()           {}
func (f *fact) At(when time.Time) { f.at = when }
