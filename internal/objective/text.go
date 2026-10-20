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
