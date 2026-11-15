package board

import (
	"cairn/internal/engine"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"fmt"
	"strconv"
	"strings"
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

// taskFields is every field of a task, in the order a form lays them out: what it is called first, and the two that are read far more often than they are written last.
func taskFields(c Card) []Field {
	return opened(engine.TaskEdit, []Field{
		{Name: "title", Arg: "title", Value: c.Title},
		{Name: "description", Arg: "desc", Value: c.Description, Long: true},
		{Name: "column", Arg: "state", Value: c.State, Choices: engine.States()},
		{Name: "priority", Arg: "priority", Value: c.Priority, Choices: Priorities()},
		{Name: "labels", Arg: "label", Value: strings.Join(c.Labels, " ")},
		{Name: "blockers", Arg: "blocker", Value: Numbers(c.Blockers)},
		{Name: "links", Arg: "link", Value: Numbers(c.Linked)},
		{Name: "because", Arg: "because", Value: c.Because},
		{Name: "deadline", Arg: "deadline", Value: calendar.FormatDay(c.Deadline)},
		{Name: "estimate", Arg: "estimate", Value: c.Estimate},
		{Name: "time spent", Arg: "spent", Value: c.Spent},
		{Name: "objectives", Arg: "objective", Value: Numbers(c.Objectives)},
	})
}

func opened(op engine.Operation, fields []Field) []Field {
	return openedAs(op, op.Thing, fields)
}

func openedAs(op engine.Operation, like string, fields []Field) []Field {
	for i := range fields {
		fields[i].Was = fields[i].Value
		fields[i].Means = meaning(op, fields[i].Arg)
		if len(fields[i].Choices) == 0 {
			fields[i].Example = typedLike[like+" "+fields[i].Arg]
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
	"objective parent":   "4",
	"objective waits-on": "3 5",
	"objective estimate": "40h",
}

func meaning(op engine.Operation, arg string) string {
	for _, a := range op.Takes {
		if a.Name == arg {
			return a.Means
		}
	}
	return ""
}

// GoalFields is every field of an objective.
func GoalFields(g Goal) []Field {
	return opened(engine.ObjectiveEdit, []Field{
		{Name: "title", Arg: "title", Value: g.Title},
		{Name: "description", Arg: "desc", Value: g.Description, Long: true},
		{Name: "state", Arg: "state", Value: g.State, Choices: standings},
		{Name: "priority", Arg: "priority", Value: g.Priority, Choices: Priorities()},
		{Name: "labels", Arg: "label", Value: strings.Join(g.Labels, " ")},
		{Name: "parent", Arg: "parent", Value: parentWritten(g.Parent)},
		{Name: "waits on", Arg: "waits-on", Value: Numbers(g.WaitsOn)},
		{Name: "deadline", Arg: "deadline", Value: calendar.FormatDay(g.Deadline)},
		{Name: "estimate", Arg: "estimate", Value: g.Estimate},
	})
}

func parentWritten(n int) string {
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}

// Priorities is the same, for what a task carries.
func Priorities() []string {
	out := make([]string, 0, len(item.Priorities))
	for _, p := range item.Priorities {
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

// PartsWritten writes steps the way PartsIn reads them: a place when the step has one, its box, its title — one line each, which is how a form reads, ticks, renames, adds to and takes away the steps of its task.
func PartsWritten(parts []engine.Part) string {
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
