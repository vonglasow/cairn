package objective

import (
	"fmt"
	"strings"
	"time"

	"cairn/internal/domain"
)

// textMembrane is an objective's membrane for the text a person reads: it is
// told every fact and lays each one down as a row, in the order it was told.
// Labels are collected as they arrive and only laid down once LabelsTold
// says there are no more.
type textMembrane struct {
	number int
	title  string
	rows   []domain.Fact
	labels []string
}

func (c *textMembrane) Numbered(n int)  { c.number = n }
func (c *textMembrane) Titled(s string) { c.title = s }
func (c *textMembrane) Standing(s State) {
	c.rows = append(c.rows, domain.Fact{Key: "state", Value: string(s)})
}
func (c *textMembrane) Carries(p domain.Priority) {
	c.rows = append(c.rows, domain.Fact{Key: "priority", Value: string(p)})
}
func (c *textMembrane) Described(s string) {
	c.rows = append(c.rows, domain.Fact{Key: "description", Value: s})
}
func (c *textMembrane) NotDescribed() {
	c.rows = append(c.rows, domain.Fact{Key: "description", Value: ""})
}
func (c *textMembrane) Labelled(s string) { c.labels = append(c.labels, s) }
func (c *textMembrane) NotLabelled()      {}
func (c *textMembrane) LabelsTold() {
	c.rows = append(c.rows, domain.Fact{Key: "labels", Value: strings.Join(c.labels, ", ")})
}
func (c *textMembrane) Due(t time.Time) {
	c.rows = append(c.rows, domain.Fact{Key: "deadline", Value: domain.FormatDay(t)})
}
func (c *textMembrane) NotDue() { c.rows = append(c.rows, domain.Fact{Key: "deadline", Value: ""}) }
func (c *textMembrane) Estimated(d time.Duration) {
	c.rows = append(c.rows, domain.Fact{Key: "estimate", Value: domain.FormatDuration(d)})
}
func (c *textMembrane) NotEstimated() {
	c.rows = append(c.rows, domain.Fact{Key: "estimate", Value: ""})
}

// Text renders an objective in full, empty attributes included, the way a
// person reads it. It is not a task's text with fields removed: an objective
// has no column, no parent and no time spent, and a shared renderer would
// have to be told which half to leave out every time.
func Text(o Objective, p domain.Progress) string {
	m := &textMembrane{}
	o.Tell(m)

	return domain.Card(m.number, m.title, append(m.rows, domain.Fact{Key: "progress", Value: p.Line("the objective")}))
}

// Serving renders the tasks put under an objective as the lines already
// built for each, and says so when there are none — an objective with an
// empty heading would read as one whose tasks failed to load.
//
// It takes lines rather than tasks: an objective may not know of a task, so
// whoever calls this across that boundary renders each line first.
func Serving(lines []string) string {
	if len(lines) == 0 {
		return "\nnothing is under it\n"
	}
	var out strings.Builder
	out.WriteString("\ntasks\n")
	for _, l := range lines {
		_, _ = fmt.Fprintf(&out, "  %s\n", l)
	}
	return out.String()
}

// listingRow is an objective's membrane for one row of a listing: it is told
// every fact and keeps the two a row needs — its number and its title.
type listingRow struct {
	number int
	title  string
	// And where it stands, what it carries, and when it is wanted. A listing
	// that said only a number and a title made somebody open every objective
	// to find out which of them was reached and which was abandoned.
	state    State
	priority domain.Priority
	due      time.Time
}

func (r *listingRow) Numbered(n int)            { r.number = n }
func (r *listingRow) Titled(s string)           { r.title = s }
func (r *listingRow) Standing(s State)          { r.state = s }
func (r *listingRow) Carries(p domain.Priority) { r.priority = p }
func (r *listingRow) Described(string)          {}
func (r *listingRow) NotDescribed()             {}
func (r *listingRow) Labelled(string)           {}
func (r *listingRow) NotLabelled()              {}
func (r *listingRow) LabelsTold()               {}
func (r *listingRow) Due(d time.Time)           { r.due = d }
func (r *listingRow) NotDue()                   {}
func (r *listingRow) Estimated(time.Duration)   {}
func (r *listingRow) NotEstimated()             {}

// List renders every objective as one row each. No tree: what is under an
// objective is shown by `objective show`, and a listing that unfolded every
// task would stop being a listing of objectives.
func List(all []Objective) string {
	var out strings.Builder
	for _, o := range all {
		var r listingRow
		o.Tell(&r)
		line := fmt.Sprintf("#%d  %s", r.number, r.title)
		for _, said := range []string{
			string(r.state), string(r.priority), domain.FormatDay(r.due),
		} {
			if said != "" {
				line += "  · " + said
			}
		}
		_, _ = fmt.Fprintf(&out, "%s\n", line)
	}
	return out.String()
}
