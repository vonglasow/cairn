package objective

import (
	"fmt"
	"strings"
	"time"

	"cairn/internal/domain"
)

// Review is what an objective's tasks say of it beyond how far along it is: the time spent on them, and what is left of them.
type Review struct {
	Spent time.Duration
	Left  Left
}

// Left is what remains of an objective: how many of its tasks are not done, and the soonest deadline among them with the task it falls on.
type Left struct {
	Tasks int
	Due   time.Time
	On    int
}

// Line says what is left in a few words, the next deadline named when one is set.
func (l Left) Line() string {
	if l.Tasks == 0 {
		return "nothing"
	}
	many := fmt.Sprintf("%d tasks", l.Tasks)
	next := "the next due"
	if l.Tasks == 1 {
		many, next = "1 task", "due"
	}
	if l.Due.IsZero() {
		return many + ", none due"
	}
	return fmt.Sprintf("%s, %s %s (#%d)", many, next, domain.FormatDay(l.Due), l.On)
}

// Text renders an objective in full, empty attributes included, the way a person reads it.
func Text(o Objective, p domain.Progress, r Review) string {
	c := o.Carried()
	return domain.Card(c.Number, c.Title, []domain.Fact{
		{Key: "state", Value: string(c.State)},
		{Key: "priority", Value: string(c.Priority)},
		{Key: "description", Value: c.Description},
		{Key: "labels", Value: strings.Join(c.Labels, ", ")},
		{Key: "deadline", Value: domain.FormatDay(c.Deadline)},
		{Key: "estimate", Value: domain.FormatDuration(c.Estimate)},
		{Key: "progress", Value: p.Line("the objective")},
		{Key: "time spent", Value: domain.FormatDuration(r.Spent)},
		{Key: "left", Value: r.Left.Line()},
	})
}

// Serving renders the tasks put under an objective as the lines already built for each, and says so when there are none — an objective with an empty heading would read as one whose tasks failed to load.
func Serving(lines []string) string {
	if len(lines) == 0 {
		return "\nnothing is under it\n"
	}
	var out strings.Builder
	out.WriteString("\ntasks\n")
	for _, l := range lines {
		fmt.Fprintf(&out, "  %s\n", l)
	}
	return out.String()
}

// List renders every objective as one row each, with how far along it is when it has anything to count.
func List(all []Objective, along map[int]domain.Progress) string {
	var out strings.Builder
	for _, o := range all {
		r := o.Carried()
		line := fmt.Sprintf("#%d  %s", r.Number, r.Title)
		for _, said := range []string{
			string(r.State), string(r.Priority), domain.FormatDay(r.Deadline),
		} {
			if said != "" {
				line += "  · " + said
			}
		}
		if p := along[r.Number]; p.Any() {
			line += fmt.Sprintf("  · %d%% done", p.Percent())
		}
		fmt.Fprintf(&out, "%s\n", line)
	}
	return out.String()
}

// Period is what a review says of a stretch of days: its first and last, the time spent on the objective's tasks month by month, and the tasks done in it.
type Period struct {
	Since, Until time.Time
	Months       []Month
	Done         []Finished
}

// Month is the time spent in one month, named by its first day.
type Month struct {
	Month time.Time
	Spent time.Duration
}

// Finished is a task put in done within a period, and the day it was.
type Finished struct {
	Number int
	Title  string
	On     time.Time
}

// Spent is the time spent in the whole period.
func (p Period) Spent() time.Duration {
	var all time.Duration
	for _, m := range p.Months {
		all += m.Spent
	}
	return all
}

// Report renders an objective's review over a period: how far along it is, the time spent in the period and in all, what is left, the time month by month, and the tasks done in the period.
func Report(o Objective, p domain.Progress, r Review, in Period) string {
	c := o.Carried()
	heading := c.Title + " — from " + domain.FormatDay(in.Since) + " to " + domain.FormatDay(in.Until)
	var out strings.Builder
	out.WriteString(domain.Card(c.Number, heading, []domain.Fact{
		{Key: "progress", Value: p.Line("the objective")},
		{Key: "deadline", Value: domain.FormatDay(c.Deadline)},
		{Key: "time spent", Value: said(in.Spent()) + " in the period, " + said(r.Spent) + " in all"},
		{Key: "left", Value: r.Left.Line()},
	}))
	if len(in.Months) == 0 {
		out.WriteString("\nnothing was spent in the period\n")
	} else {
		out.WriteString("\nby month\n")
		for _, m := range in.Months {
			fmt.Fprintf(&out, "  %s  %s\n", m.Month.Format("2006-01"), domain.FormatDuration(m.Spent))
		}
	}
	if len(in.Done) == 0 {
		out.WriteString("\nnothing was done in the period\n")
		return out.String()
	}
	out.WriteString("\ndone in the period\n")
	for _, f := range in.Done {
		fmt.Fprintf(&out, "  #%d  %s  · %s\n", f.Number, f.Title, domain.FormatDay(f.On))
	}
	return out.String()
}

func said(d time.Duration) string {
	if d == 0 {
		return domain.Absent
	}
	return domain.FormatDuration(d)
}
