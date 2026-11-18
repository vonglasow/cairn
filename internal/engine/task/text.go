package task

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"fmt"
	"slices"
	"strings"
)

// BlocksMark is what stands before how many open tasks a task is in the way of.
const BlocksMark = "⊳"

// ServesMark is what stands before the objectives a task serves, in a listing.
const ServesMark = "◇"

// BlocksSaid is how many open tasks a task is in the way of, as every surface writes it.
func BlocksSaid(n int) string { return fmt.Sprintf("%s%d", BlocksMark, n) }

// Text renders a task in full, empty attributes included, the way a person reads it; the time spent only while the time is shown, and the open tasks it is in the way of only when there are some, and each objective by its parent when chosen names it.
func Text(t Task, p item.Progress, timed bool, chosen Choosing) string {
	blocks := chosen.Blocks
	c := t.Carried()
	rows := []item.Fact{{Key: "state", Value: string(c.State)}}
	if c.Archived {
		rows = append(rows, item.Fact{Key: "archived", Value: "yes"})
	}
	var blocked, serves, linked []string
	for _, n := range c.Blockers {
		blocked = append(blocked, item.Ref(n))
	}
	if c.Because != "" {
		blocked = append(blocked, c.Because)
	}
	for _, n := range c.Linked {
		linked = append(linked, item.Ref(n))
	}
	for i, n := range c.Objectives {
		if i < len(chosen.Serves) {
			serves = append(serves, item.Ref(n)+" "+chosen.Serves[i])
			continue
		}
		serves = append(serves, item.Ref(n))
	}
	rows = append(rows,
		item.Fact{Key: "priority", Value: string(c.Priority)},
		item.Fact{Key: "description", Value: c.Description},
		item.Fact{Key: "labels", Value: strings.Join(c.Labels, ", ")},
		item.Fact{Key: "blocked by", Value: strings.Join(blocked, ", ")},
	)
	if len(blocks) > 0 {
		var refs []string
		for _, n := range blocks {
			refs = append(refs, item.Ref(n))
		}
		rows = append(rows, item.Fact{Key: "blocks", Value: BlocksSaid(len(blocks)) + "  " + strings.Join(refs, ", ")})
	}
	rows = append(rows,
		item.Fact{Key: "linked", Value: strings.Join(linked, ", ")},
		item.Fact{Key: "deadline", Value: calendar.FormatDay(c.Deadline)},
		item.Fact{Key: "estimate", Value: calendar.FormatDuration(c.Estimate)},
	)
	if timed {
		rows = append(rows, item.Fact{Key: "time spent", Value: calendar.FormatDuration(c.Spent)})
	}
	rows = append(rows,
		item.Fact{Key: "objectives", Value: strings.Join(serves, ", ")},
		item.Fact{Key: "progress", Value: p.Line("the task")},
	)
	return item.Card(c.Number, c.Title, rows)
}

// DoneMark is what sets apart a step that is done.
const DoneMark = "(done)"

// StepsText renders a task's steps, each with its place, and nothing at all when it has none — an empty heading would read as a section somebody forgot to fill.
func StepsText(steps []Step) string {
	if len(steps) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\nsteps\n")
	for i, s := range steps {
		mark := strings.Repeat(" ", len(DoneMark))
		if s.Done {
			mark = DoneMark
		}
		fmt.Fprintf(&out, "  %s  %d  %s\n", mark, i+1, s.Title)
	}
	return out.String()
}

// InFocusMark is what sets apart the task being worked on.
const InFocusMark = "  (in focus)"

// NothingBlocksMark is what sets apart a blocked task whose every named blocker is done — the one thing a listing says about blocking, because it is the one thing that needs acting on.
const NothingBlocksMark = "  (nothing blocks it)"

// Marks are what a listing says about a task that the task does not carry itself: how many open tasks each is in the way of, what each serves, which one is being worked on, which ones are waiting on nothing any more, and how often an occurrence of a series comes back.
type Marks struct {
	Blocks        map[int]int
	InFocus       []int
	NothingBlocks []int
	Repeats       map[int]string
	Serves        map[int][]string
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
	if n := marks.Blocks[number]; n > 0 {
		line += "  · " + BlocksSaid(n)
	}
	if paths := marks.Serves[number]; len(paths) > 0 {
		line += "  · " + ServesMark + " " + strings.Join(paths, ", ")
	}
	if repeats, ok := marks.Repeats[number]; ok {
		line += "  · " + repeats
	}
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
		string(r.State), string(r.Priority), calendar.FormatDay(r.Deadline),
		strings.Join(r.Labels, " "),
	} {
		if said != "" {
			line += "  · " + said
		}
	}
	return line
}
