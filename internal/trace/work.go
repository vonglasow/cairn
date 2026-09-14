package trace

// What a period of acts adds up to. It is arithmetic over acts, so it lives
// with the acts and not with whoever displays it — and it is given raw: how
// much work, on how many tasks, how often it switched. What counts as a bad day
// is a judgement, and cairn does not make it.

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// A stretch is one run of work on one task: a focus taken, closed by that task
// being let go. There is no other definition available, and every duration
// rests on this one.
//
// Stretches overlap. Several tasks are worked on at once — one per agent
// session — so taking up a second does not close the first, and a day's
// stretches can add up to more than the day.
type stretch struct {
	task  int
	from  time.Time
	until time.Time
}

// Work is what a period came to.
//
// The total of the stretches is not the length of the day, and nothing here
// claims it is: nothing is in focus during a meeting, so a day has gaps, and
// cairn knows neither when it began nor when it ended.
type Work struct {
	stretches []stretch
	switches  int
}

// AddedUp works out what a period of acts comes to. A stretch still open at the
// end runs to now: it is work being done, not work that never happened.
func AddedUp(acts []Act, now time.Time) Work {
	var w Work
	// Which stretch is open on each task. Several are open at once, so it is a
	// stretch per task and not one at a time — kept as an index into the
	// stretches so the order they were taken up in survives.
	open := map[int]int{}
	for _, a := range acts {
		var f fact
		a.Tell(&f)
		switch f.did {
		case FocusTaken:
			if _, already := open[f.task]; already {
				// Taking up what is already held is not a change, and the
				// engine does not write it down. Ignored here as well, so a
				// history written before that rule still adds up.
				continue
			}
			// Work taken up while other work is already under way is attention
			// moving. Taken up with nothing else open it is a start.
			if len(open) > 0 {
				w.switches++
			}
			w.stretches = append(w.stretches, stretch{task: f.task, from: f.at})
			open[f.task] = len(w.stretches) - 1
		case FocusReleased:
			if at, was := open[f.task]; was {
				w.stretches[at].until = f.at
				delete(open, f.task)
			}
		}
	}
	// Whatever is still held runs to now: it is work being done, not work that
	// never happened.
	for _, at := range open {
		w.stretches[at].until = now
	}
	return w
}

// Summary renders what a period came to, or nothing at all when no work was
// done in it: a period with acts but no stretches has nothing to add up.
func Summary(w Work) string {
	if len(w.stretches) == 0 {
		return ""
	}
	var out strings.Builder
	_, _ = fmt.Fprintf(&out, "\n%d stretches of work on %d task(s), switched %d time(s)\n",
		len(w.stretches), len(w.onTasks()), w.switches)
	for _, task := range w.onTasks() {
		_, _ = fmt.Fprintf(&out, "  #%d  %s", task, spent(w.on(task)))
	}
	out.WriteString("\n")
	return out.String()
}

// onTasks is every task worked on in the period, in the order their numbers
// run: a summary read twice says the same thing in the same order.
func (w Work) onTasks() []int {
	seen := map[int]bool{}
	tasks := []int{}
	for _, s := range w.stretches {
		if !seen[s.task] {
			seen[s.task] = true
			tasks = append(tasks, s.task)
		}
	}
	sort.Ints(tasks)
	return tasks
}

// on is how long was spent on one task over the period, across every stretch of
// it.
func (w Work) on(task int) time.Duration {
	var total time.Duration
	for _, s := range w.stretches {
		if s.task == task {
			total += s.until.Sub(s.from)
		}
	}
	return total
}

// spent writes a length of work the way a person reads one, down to the minute:
// a stretch measured to the second says more than anybody knows.
func spent(d time.Duration) string {
	minutes := int(d.Round(time.Minute).Minutes())
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dh %02dm", minutes/60, minutes%60)
}

// fact is an act's membrane for what the arithmetic needs: which task, what was
// done, and when.
type fact struct {
	task   int
	did    string
	became string
	at     time.Time
}

func (f *fact) On(n int)          { f.task = n }
func (f *fact) Did(what string)   { f.did = what }
func (f *fact) Changing(string)   {}
func (f *fact) NothingNamed()     {}
func (f *fact) Was(string)        {}
func (f *fact) WasNothing()       {}
func (f *fact) Became(v string)   { f.became = v }
func (f *fact) BecameNothing()    {}
func (f *fact) Through(string)    {}
func (f *fact) ByHand()           {}
func (f *fact) At(when time.Time) { f.at = when }
