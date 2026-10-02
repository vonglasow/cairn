package objective

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"cairn/internal/domain"
)

// row is one line of a card: its key, and the value to show for it.
type row struct{ key, value string }

// textMembrane is an objective's membrane for the text a person reads: it is
// told every fact and lays each one down as a row, in the order it was told.
// Labels are collected as they arrive and only laid down once LabelsTold
// says there are no more.
type textMembrane struct {
	number int
	title  string
	rows   []row
	labels []string
}

func (c *textMembrane) Numbered(n int)  { c.number = n }
func (c *textMembrane) Titled(s string) { c.title = s }
func (c *textMembrane) Standing(s State) {
	c.rows = append(c.rows, row{"state", string(s)})
}
func (c *textMembrane) Carries(p domain.Priority) {
	c.rows = append(c.rows, row{"priority", string(p)})
}
func (c *textMembrane) Described(s string) { c.rows = append(c.rows, row{"description", s}) }
func (c *textMembrane) NotDescribed()      { c.rows = append(c.rows, row{"description", ""}) }
func (c *textMembrane) Labelled(s string)  { c.labels = append(c.labels, s) }
func (c *textMembrane) NotLabelled()       {}
func (c *textMembrane) LabelsTold() {
	c.rows = append(c.rows, row{"labels", strings.Join(c.labels, ", ")})
}
func (c *textMembrane) Due(t time.Time) {
	c.rows = append(c.rows, row{"deadline", domain.FormatDay(t)})
}
func (c *textMembrane) NotDue() { c.rows = append(c.rows, row{"deadline", ""}) }
func (c *textMembrane) Estimated(d time.Duration) {
	c.rows = append(c.rows, row{"estimate", domain.FormatDuration(d)})
}
func (c *textMembrane) NotEstimated() { c.rows = append(c.rows, row{"estimate", ""}) }

// Text renders an objective in full, empty attributes included, the way a
// person reads it. It is not a task's text with fields removed: an objective
// has no column, no parent and no time spent, and a shared renderer would
// have to be told which half to leave out every time.
func Text(o Objective, p domain.Progress) string {
	m := &textMembrane{}
	o.Tell(m)

	var out strings.Builder
	_, _ = fmt.Fprintf(&out, "#%d  %s\n\n", m.number, m.title)

	tw := tabwriter.NewWriter(&out, 0, 2, 2, ' ', 0)
	for _, f := range append(m.rows, row{"progress", progressLine(p)}) {
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", f.key, domain.OrNothing(f.value))
	}
	_ = tw.Flush()
	return out.String()
}

// progressLine says how far along an objective is, and says nothing at all
// when nothing is under it. Zero of zero is not "none of it done": there is
// no work to be done, and a share of nothing would read as a standstill.
//
// Tasks and sub-tasks are named apart because they are counted apart, while
// the share is over both — three tasks and four sub-tasks is never seven
// tasks, but all seven have to happen.
func progressLine(p domain.Progress) string {
	if !p.Any() {
		return ""
	}
	var counts []string
	if p.Tasks.Total > 0 {
		counts = append(counts, fmt.Sprintf("%d of %d tasks", p.Tasks.Done, p.Tasks.Total))
	}
	if p.SubTasks.Total > 0 {
		counts = append(counts, fmt.Sprintf("%d of %d sub-tasks", p.SubTasks.Done, p.SubTasks.Total))
	}
	return fmt.Sprintf("%s are done — the objective is %d%% done",
		strings.Join(counts, " and "), p.Percent())
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
