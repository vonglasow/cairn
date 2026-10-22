package engine

import (
	"fmt"
	"strconv"
	"strings"

	"cairn/internal/domain"
)

// Field is one thing a form asks about.
type Field struct {
	Name    string
	Arg     string
	Value   string
	Was     string
	Means   string
	Choices []string
	Long    bool
	Example string
}

// TaskFields is every field of a task, in the order a form lays them out: what it is called first, and the two that are read far more often than they are written last.
func TaskFields(c Card) []Field {
	return opened(TaskEdit, []Field{
		{Name: "title", Arg: "title", Value: c.Title},
		{Name: "description", Arg: "desc", Value: c.Description, Long: true},
		{Name: "column", Arg: "state", Value: c.State, Choices: States()},
		{Name: "priority", Arg: "priority", Value: c.Priority, Choices: Priorities()},
		{Name: "labels", Arg: "label", Value: strings.Join(c.Labels, " ")},
		{Name: "blockers", Arg: "blocker", Value: Numbers(c.Blockers)},
		{Name: "links", Arg: "link", Value: Numbers(c.Linked)},
		{Name: "because", Arg: "because", Value: c.Because},
		{Name: "deadline", Arg: "deadline", Value: domain.FormatDay(c.Deadline)},
		{Name: "estimate", Arg: "estimate", Value: c.Estimate},
		{Name: "time spent", Arg: "spent", Value: c.Spent},
		{Name: "objectives", Arg: "objective", Value: Numbers(c.Objectives)},
	})
}

func opened(op Operation, fields []Field) []Field {
	for i := range fields {
		fields[i].Was = fields[i].Value
		fields[i].Means = meaning(op, fields[i].Arg)
		if len(fields[i].Choices) == 0 {
			fields[i].Example = typedLike[op.Thing+" "+fields[i].Arg]
		}
	}
	return fields
}

var typedLike = map[string]string{
	"task title":         "Write the spec",
	"task label":         "spec team",
	"task blocker":       "3 5",
	"task link":          "3 5",
	"task because":       "the client's answer",
	"task deadline":      "2026-12-25",
	"task estimate":      "1h30m",
	"task spent":         "+30m",
	"objective title":    "Ship cairn 1.0",
	"objective label":    "product",
	"objective deadline": "2026-12-25",
	"objective estimate": "40h",
}

func meaning(op Operation, arg string) string {
	for _, a := range op.Takes {
		if a.Name == arg {
			return a.Means
		}
	}
	return ""
}

// GoalFields is every field of an objective.
func GoalFields(g Goal) []Field {
	return opened(ObjectiveEdit, []Field{
		{Name: "title", Arg: "title", Value: g.Title},
		{Name: "description", Arg: "desc", Value: g.Description, Long: true},
		{Name: "standing", Arg: "state", Value: g.State, Choices: Standings},
		{Name: "priority", Arg: "priority", Value: g.Priority, Choices: Priorities()},
		{Name: "labels", Arg: "label", Value: strings.Join(g.Labels, " ")},
		{Name: "deadline", Arg: "deadline", Value: domain.FormatDay(g.Deadline)},
		{Name: "estimate", Arg: "estimate", Value: g.Estimate},
	})
}

// States is the closed set a column is picked from, read off the domain rather than listed again here.
func States() []string {
	out := make([]string, 0, len(domain.States))
	for _, s := range domain.States {
		out = append(out, string(s))
	}
	return out
}

// Priorities is the same, for what a task carries.
func Priorities() []string {
	out := make([]string, 0, len(domain.Priorities))
	for _, p := range domain.Priorities {
		out = append(out, string(p))
	}
	return out
}

// NumbersIn reads back what Numbers wrote.
func NumbersIn(said string) []int {
	var all []int
	for _, word := range strings.Fields(said) {
		if n, err := strconv.Atoi(word); err == nil {
			all = append(all, n)
		}
	}
	return all
}

// Numbers writes a list of task or objective numbers the way a form takes them back.
func Numbers(all []int) string {
	out := make([]string, 0, len(all))
	for _, n := range all {
		out = append(out, strconv.Itoa(n))
	}
	return strings.Join(out, " ")
}

// Words splits what was typed into a field the way a shell would have split it: on spaces, and nothing cleverer.
func Words(s string) []string {
	found := strings.Fields(s)
	if found == nil {
		return []string{}
	}
	return found
}

// Part is one step of a task, as a board and a form hold it: its place in the task when it already exists, what it is called, and whether it is done.
type Part struct {
	Number int
	Title  string
	Done   bool
}

// PartsWritten writes steps the way PartsIn reads them: a place when the step has one, its box, its title — one line each, which is how a form reads, ticks, renames, adds to and takes away the steps of its task.
func PartsWritten(parts []Part) string {
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		mark := " "
		if part.Done {
			mark = "x"
		}
		line := fmt.Sprintf("[%s] %s", mark, part.Title)
		if part.Number != 0 {
			line = fmt.Sprintf("#%d %s", part.Number, line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// PartsIn reads those lines back.
func PartsIn(said string) []Part {
	var parts []Part
	for _, line := range strings.Split(said, "\n") {
		rest := strings.TrimSpace(line)
		if rest == "" {
			continue
		}
		part := Part{}
		if number, after, found := strings.Cut(rest, " "); found && strings.HasPrefix(number, "#") {
			if n, err := strconv.Atoi(strings.TrimPrefix(number, "#")); err == nil {
				part.Number, rest = n, strings.TrimSpace(after)
			}
		}
		if mark, after, found := strings.Cut(rest, "]"); found && strings.HasPrefix(mark, "[") {
			part.Done = strings.EqualFold(strings.TrimPrefix(mark, "["), "x")
			rest = strings.TrimSpace(after)
		}
		if rest == "" {
			continue
		}
		part.Title = rest
		parts = append(parts, part)
	}
	return parts
}
