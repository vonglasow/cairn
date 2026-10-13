package task

import (
	"fmt"
	"slices"
	"strings"

	"cairn/internal/domain"
)

// Text renders a task in full, empty attributes included, the way a person reads it.
func Text(t Task, p domain.Progress) string {
	c := t.Carried()
	rows := []domain.Fact{{Key: "state", Value: string(c.State)}}
	if c.Archived {
		rows = append(rows, domain.Fact{Key: "archived", Value: "yes"})
	}
	var blocked, serves []string
	for _, n := range c.Blockers {
		blocked = append(blocked, domain.Ref(n))
	}
	if c.Because != "" {
		blocked = append(blocked, c.Because)
	}
	for _, n := range c.Objectives {
		serves = append(serves, domain.Ref(n))
	}
	parent := ""
	if c.Parent != NoParent {
		parent = fmt.Sprintf("%d", c.Parent)
	}
	rows = append(rows,
		domain.Fact{Key: "priority", Value: string(c.Priority)},
		domain.Fact{Key: "description", Value: c.Description},
		domain.Fact{Key: "labels", Value: strings.Join(c.Labels, ", ")},
		domain.Fact{Key: "blocked by", Value: strings.Join(blocked, ", ")},
		domain.Fact{Key: "deadline", Value: domain.FormatDay(c.Deadline)},
		domain.Fact{Key: "estimate", Value: domain.FormatDuration(c.Estimate)},
		domain.Fact{Key: "time spent", Value: domain.FormatDuration(c.Spent)},
		domain.Fact{Key: "parent", Value: parent},
		domain.Fact{Key: "objectives", Value: strings.Join(serves, ", ")},
		domain.Fact{Key: "progress", Value: p.Line("the task")},
	)
	return domain.Card(c.Number, c.Title, rows)
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
		if t.state == domain.Done {
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
		if t.parent != NoParent {
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
	number := t.number
	if slices.Contains(marks.InFocus, number) {
		line += InFocusMark
	}
	if slices.Contains(marks.NothingBlocks, number) {
		line += NothingBlocksMark
	}
	return line
}

// Line renders a task as one row of a listing: its number, its title, and what a decision is made on — where it stands, what it carries, when it is wanted, and what it is tagged with.
func Line(t Task) string {
	r := t.Carried()
	line := fmt.Sprintf("#%d  %s", r.Number, r.Title)
	for _, said := range []string{
		string(r.State), string(r.Priority), domain.FormatDay(r.Deadline),
		strings.Join(r.Labels, " "),
	} {
		if said != "" {
			line += "  · " + said
		}
	}
	return line
}
