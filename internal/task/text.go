package task

import (
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	"cairn/internal/domain"
)

// row is one line of a card: its key, and the value to show for it.
type row struct{ key, value string }

// textMembrane is a task's membrane for the text a person reads: it is told
// every fact and lays each one down as a row, in the order it was told.
// Labels are collected as they arrive and only laid down once LabelsTold says
// there are no more.
type textMembrane struct {
	number   int
	title    string
	rows     []row
	labels   []string
	blockers []string
}

func (c *textMembrane) Numbered(n int)  { c.number = n }
func (c *textMembrane) Titled(s string) { c.title = s }
func (c *textMembrane) InColumn(s domain.State) {
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

// A blocker and the words beside it are one row: what blocks this. A number is
// written as a task is written everywhere else, so a reader can take it
// straight to `task show`.
func (c *textMembrane) BlockedBy(n int) { c.blockers = append(c.blockers, fmt.Sprintf("#%d", n)) }
func (c *textMembrane) NotBlocked()     {}
func (c *textMembrane) BlockersTold()   {}
func (c *textMembrane) Because(s string) {
	c.blockers = append(c.blockers, s)
	c.blocking()
}
func (c *textMembrane) NotBecause() { c.blocking() }

// blocking closes the row, once both kinds of blocker have been told. Written
// on the last of them rather than on BlockersTold, because the words come
// after the numbers and a row closed early would leave them out.
func (c *textMembrane) blocking() {
	c.rows = append(c.rows, row{"blocked by", strings.Join(c.blockers, ", ")})
}

func (c *textMembrane) Due(t time.Time) {
	c.rows = append(c.rows, row{"deadline", domain.FormatDay(t)})
}
func (c *textMembrane) NotDue() { c.rows = append(c.rows, row{"deadline", ""}) }
func (c *textMembrane) Estimated(d time.Duration) {
	c.rows = append(c.rows, row{"estimate", domain.FormatDuration(d)})
}
func (c *textMembrane) NotEstimated() { c.rows = append(c.rows, row{"estimate", ""}) }
func (c *textMembrane) Spent(d time.Duration) {
	c.rows = append(c.rows, row{"time spent", domain.FormatDuration(d)})
}
func (c *textMembrane) NothingSpent() { c.rows = append(c.rows, row{"time spent", ""}) }
func (c *textMembrane) Under(n int)   { c.rows = append(c.rows, row{"parent", fmt.Sprintf("%d", n)}) }
func (c *textMembrane) UnderNothing() { c.rows = append(c.rows, row{"parent", ""}) }
func (c *textMembrane) Serves(n int) {
	c.rows = append(c.rows, row{"objective", fmt.Sprintf("%d", n)})
}
func (c *textMembrane) ServesNothing() { c.rows = append(c.rows, row{"objective", ""}) }

// Text renders a task in full, empty attributes included, the way a person
// reads it. Rendering lives here and only here: the day a second surface
// shows a task, it asks for this rather than inventing a layout of its own.
func Text(t Task, p domain.Progress) string {
	m := &textMembrane{}
	t.Tell(m)

	var out strings.Builder
	_, _ = fmt.Fprintf(&out, "#%d  %s\n\n", m.number, m.title)

	tw := tabwriter.NewWriter(&out, 0, 2, 2, ' ', 0)
	for _, f := range append(m.rows, row{"progress", progressLine(p)}) {
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", f.key, domain.OrNothing(f.value))
	}
	_ = tw.Flush()
	return out.String()
}

// progressLine says how far along a task is, and says nothing at all when
// nothing is under it. Zero of zero is not "none of it done": there is no
// work to be done, and a share of nothing would read as a standstill.
//
// Tasks and sub-tasks are named apart because they are counted apart, while
// the share is over both — three sub-tasks and four more under those is
// never seven, but all seven have to happen.
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
	return fmt.Sprintf("%s are done — the task is %d%% done",
		strings.Join(counts, " and "), p.Percent())
}

// Under renders what hangs beneath a task, and nothing at all when nothing
// does — an empty heading would read as a section somebody forgot to fill.
func Under(tasks []Task) string {
	if len(tasks) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\nsub-tasks\n")
	for _, t := range tasks {
		_, _ = fmt.Fprintf(&out, "  %s\n", Line(t))
	}
	return out.String()
}

// nest is what sets a sub-task apart in a listing. Indentation rather than a
// drawn tree: one level does not need corners, and an indented line still
// reads when someone greps for it.
const nest = "    "

// InFocusMark is what sets apart the task being worked on. Written out rather
// than drawn: a listing is read left to right and grepped afterwards, and a
// glyph is neither searchable nor sure to render.
//
// Exported because a scenario asserts it: a step looking for a string of its
// own would pass the day this one changed.
const InFocusMark = "  (in focus)"

// NothingBlocksMark is what sets apart a blocked task whose every named blocker
// is done — the one thing a listing says about blocking, because it is the one
// thing that needs acting on. What blocks a task is on its card; that nothing
// does any more is what you would otherwise have to go looking for.
//
// Written out and exported for the reasons InFocusMark is.
const NothingBlocksMark = "  (nothing blocks it)"

// NothingInFocus is the number to hand Tree when no task is being worked on.
// Zero is safe to mean none because the store never hands out zero.
const NothingInFocus = 0

// Marks are what a listing says about a task that the task does not carry
// itself: which one is being worked on, and which ones are waiting on nothing
// any more. Both are facts of the store rather than of any task, which is why
// neither is told by a task's own membrane.
//
// One parameter rather than two, and a struct rather than a pair, because a
// listing has gained a mark twice now and will again: a shape that grows a
// field costs nothing, where a signature that grows a parameter costs every
// caller.
type Marks struct {
	// InFocus is the task being worked on, or NothingInFocus.
	InFocus int
	// NothingBlocks are the blocked tasks whose every named blocker is done.
	NothingBlocks []int
}

// Tree renders a listing with each sub-task under the task it hangs from, and
// marks what a listing marks. Everything keeps its own number, so a line found
// here can be acted on directly — the indentation says where it belongs, not
// what it is.
//
// The place a task holds is never shown. The order is what there is to see,
// and a second number beside one a task already carries would only mislead.
func Tree(tasks []Task, marks Marks) string {
	hanging := map[int][]Task{}
	for _, t := range tasks {
		if r := rowOf(t); r.parent != NoParent {
			hanging[r.parent] = append(hanging[r.parent], t)
		}
	}

	var out strings.Builder
	for _, t := range tasks {
		r := rowOf(t)
		if r.parent != NoParent {
			continue
		}
		out.WriteString(marked(t, marks))
		out.WriteString("\n")
		for _, under := range hanging[r.number] {
			out.WriteString(nest)
			out.WriteString(marked(under, marks))
			out.WriteString("\n")
		}
	}
	return out.String()
}

// Lines renders a listing flat, one task per line, marking what a listing
// marks. Where each task sits is not shown, because a listing rendered this way
// was not asked about it.
func Lines(tasks []Task, marks Marks) string {
	var out strings.Builder
	for _, t := range tasks {
		out.WriteString(marked(t, marks))
		out.WriteString("\n")
	}
	return out.String()
}

// marked renders one line of a listing with whatever the listing says about it.
// One task is in focus at a time, so at most one line carries that; any number
// of tasks can be waiting on nothing.
func marked(t Task, marks Marks) string {
	line := Line(t)
	number := rowOf(t).number
	if number == marks.InFocus {
		line += InFocusMark
	}
	if slices.Contains(marks.NothingBlocks, number) {
		line += NothingBlocksMark
	}
	return line
}

// listingRow is a task's membrane for one row of a listing: it is told every
// fact and keeps the three a row needs — its number, its title, and the
// parent it hangs under, which says where it belongs rather than what it is.
type listingRow struct {
	number, parent int
	title          string
}

func (r *listingRow) Numbered(n int)          { r.number = n }
func (r *listingRow) Titled(s string)         { r.title = s }
func (r *listingRow) InColumn(domain.State)   {}
func (r *listingRow) Carries(domain.Priority) {}
func (r *listingRow) Described(string)        {}
func (r *listingRow) NotDescribed()           {}
func (r *listingRow) Labelled(string)         {}
func (r *listingRow) NotLabelled()            {}
func (r *listingRow) LabelsTold()             {}
func (r *listingRow) BlockedBy(int)           {}
func (r *listingRow) NotBlocked()             {}
func (r *listingRow) BlockersTold()           {}
func (r *listingRow) Because(string)          {}
func (r *listingRow) NotBecause()             {}
func (r *listingRow) Due(time.Time)           {}
func (r *listingRow) NotDue()                 {}
func (r *listingRow) Estimated(time.Duration) {}
func (r *listingRow) NotEstimated()           {}
func (r *listingRow) Spent(time.Duration)     {}
func (r *listingRow) NothingSpent()           {}
func (r *listingRow) Under(n int)             { r.parent = n }
func (r *listingRow) UnderNothing()           { r.parent = NoParent }
func (r *listingRow) Serves(int)              {}
func (r *listingRow) ServesNothing()          {}

// rowOf reads back what a listing row needs from a task, without reaching
// into fields that stopped being exported.
func rowOf(t Task) listingRow {
	var r listingRow
	t.Tell(&r)
	return r
}

// Line renders a task as one row of a listing: its number, then its title.
func Line(t Task) string {
	r := rowOf(t)
	return fmt.Sprintf("#%d  %s", r.number, r.title)
}
