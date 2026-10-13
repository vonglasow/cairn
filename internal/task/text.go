package task

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"cairn/internal/domain"
)

type textMembrane struct {
	number     int
	title      string
	rows       []domain.Fact
	labels     []string
	blockers   []string
	objectives []string
}

// Archived says so on a row of its own.
func (m *textMembrane) Archived() {
	m.rows = append(m.rows, domain.Fact{Key: "archived", Value: "yes"})
}
func (m *textMembrane) NotArchived() {}

func (c *textMembrane) Numbered(n int)  { c.number = n }
func (c *textMembrane) Titled(s string) { c.title = s }
func (c *textMembrane) InColumn(s domain.State) {
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

// A blocker and the words beside it are one row: what blocks this.
func (c *textMembrane) BlockedBy(n int) { c.blockers = append(c.blockers, domain.Ref(n)) }
func (c *textMembrane) NotBlocked()     {}
func (c *textMembrane) BlockersTold()   {}
func (c *textMembrane) Because(s string) {
	c.blockers = append(c.blockers, s)
	c.blocking()
}
func (c *textMembrane) NotBecause() { c.blocking() }

func (c *textMembrane) blocking() {
	c.rows = append(c.rows, domain.Fact{Key: "blocked by", Value: strings.Join(c.blockers, ", ")})
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
func (c *textMembrane) Spent(d time.Duration) {
	c.rows = append(c.rows, domain.Fact{Key: "time spent", Value: domain.FormatDuration(d)})
}
func (c *textMembrane) NothingSpent() {
	c.rows = append(c.rows, domain.Fact{Key: "time spent", Value: ""})
}
func (c *textMembrane) Under(n int) {
	c.rows = append(c.rows, domain.Fact{Key: "parent", Value: fmt.Sprintf("%d", n)})
}
func (c *textMembrane) UnderNothing() { c.rows = append(c.rows, domain.Fact{Key: "parent", Value: ""}) }

// A task serves a set, so the card lays out one row naming all of them, each written the way a task is written everywhere else.
func (c *textMembrane) Serves(n int) {
	c.objectives = append(c.objectives, domain.Ref(n))
}
func (c *textMembrane) ServesNothing() {}
func (c *textMembrane) ObjectivesTold() {
	c.rows = append(c.rows, domain.Fact{Key: "objectives", Value: strings.Join(c.objectives, ", ")})
}

// Text renders a task in full, empty attributes included, the way a person reads it.
func Text(t Task, p domain.Progress) string {
	m := &textMembrane{}
	t.Tell(m)

	return domain.Card(m.number, m.title, append(m.rows, domain.Fact{Key: "progress", Value: p.Line("the task")}))
}

// DoneMark is what sets apart a sub-task that is finished.
const DoneMark = "(done)"

// Under renders what hangs beneath a task, and nothing at all when nothing does — an empty heading would read as a section somebody forgot to fill.
func Under(tasks []Task) string {
	if len(tasks) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\nsub-tasks\n")
	for _, t := range tasks {
		mark := strings.Repeat(" ", len(DoneMark))
		if rowOf(t).state == domain.Done {
			mark = DoneMark
		}
		fmt.Fprintf(&out, "  %s  %s\n", mark, Line(t))
	}
	return out.String()
}

// InFocusMark is what sets apart the task being worked on.
const InFocusMark = "  (in focus)"

// NothingBlocksMark is what sets apart a blocked task whose every named blocker is done — the one thing a listing says about blocking, because it is the one thing that needs acting on.
const NothingBlocksMark = "  (nothing blocks it)"

// Marks are what a listing says about a task that the task does not carry itself: which one is being worked on, and which ones are waiting on nothing any more.
type Marks struct {
	InFocus       []int
	NothingBlocks []int
}

// Tree renders a listing of the work, and marks what a listing marks.
func Tree(tasks []Task, marks Marks) string {
	var out strings.Builder
	for _, t := range tasks {
		if rowOf(t).parent != NoParent {
			continue
		}
		out.WriteString(marked(t, marks))
		out.WriteString("\n")
	}
	return out.String()
}

// Lines renders a listing flat, one task per line, marking what a listing marks.
func Lines(tasks []Task, marks Marks) string {
	var out strings.Builder
	for _, t := range tasks {
		out.WriteString(marked(t, marks))
		out.WriteString("\n")
	}
	return out.String()
}

func marked(t Task, marks Marks) string {
	line := Line(t)
	number := rowOf(t).number
	if slices.Contains(marks.InFocus, number) {
		line += InFocusMark
	}
	if slices.Contains(marks.NothingBlocks, number) {
		line += NothingBlocksMark
	}
	return line
}

type listingRow struct {
	number, parent int
	title          string
	state          domain.State
	archived       bool
	priority       domain.Priority
	due            time.Time
	labels         []string
}

func (r *listingRow) Numbered(n int)            { r.number = n }
func (r *listingRow) Titled(s string)           { r.title = s }
func (r *listingRow) Archived()                 { r.archived = true }
func (r *listingRow) NotArchived()              { r.archived = false }
func (r *listingRow) InColumn(s domain.State)   { r.state = s }
func (r *listingRow) Carries(p domain.Priority) { r.priority = p }
func (r *listingRow) Described(string)          {}
func (r *listingRow) NotDescribed()             {}
func (r *listingRow) Labelled(l string)         { r.labels = append(r.labels, l) }
func (r *listingRow) NotLabelled()              {}
func (r *listingRow) LabelsTold()               {}
func (r *listingRow) BlockedBy(int)             {}
func (r *listingRow) NotBlocked()               {}
func (r *listingRow) BlockersTold()             {}
func (r *listingRow) Because(string)            {}
func (r *listingRow) NotBecause()               {}
func (r *listingRow) Due(d time.Time)           { r.due = d }
func (r *listingRow) NotDue()                   {}
func (r *listingRow) Estimated(time.Duration)   {}
func (r *listingRow) NotEstimated()             {}
func (r *listingRow) Spent(time.Duration)       {}
func (r *listingRow) NothingSpent()             {}
func (r *listingRow) Under(n int)               { r.parent = n }
func (r *listingRow) UnderNothing()             { r.parent = NoParent }
func (r *listingRow) Serves(int)                {}
func (r *listingRow) ServesNothing()            {}
func (r *listingRow) ObjectivesTold()           {}

func rowOf(t Task) listingRow {
	var r listingRow
	t.Tell(&r)
	return r
}

// Line renders a task as one row of a listing: its number, its title, and what a decision is made on — where it stands, what it carries, when it is wanted, and what it is tagged with.
func Line(t Task) string {
	r := rowOf(t)
	line := fmt.Sprintf("#%d  %s", r.number, r.title)
	for _, said := range []string{
		string(r.state), string(r.priority), domain.FormatDay(r.due),
		strings.Join(r.labels, " "),
	} {
		if said != "" {
			line += "  · " + said
		}
	}
	return line
}
