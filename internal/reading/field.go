package reading

// What a form asks about, without knowing what a form looks like. Which fields
// a task has, in what order, which of them hold a closed set of values, and
// which engine argument each one fills — a terminal and a window differ in how
// they draw a box, not in what goes in it.

import (
	"strconv"
	"strings"

	"cairn/internal/domain"
)

// Field is one thing a form asks about.
type Field struct {
	// Name is what a person reads beside the box: "labels", "time spent".
	Name string
	// Arg is what the engine calls the same thing: "label", "spent". Both are
	// kept because neither is wrong — a repeatable flag reads singular on a
	// command line, and a box holding all of them reads plural on a screen.
	// Saying it here is what stops a translation table existing somewhere else.
	Arg string
	// Value is what the box starts with.
	Value string
	// Choices is the closed set the value is picked from, empty when there is
	// none.
	Choices []string
	// Long says the value wants more than one line.
	Long bool
}

// TaskFields is every field of a task, in the order a form lays them out: what
// it is called first, and the two that are read far more often than they are
// written last.
func TaskFields(c Card) []Field {
	return []Field{
		{Name: "title", Arg: "title", Value: c.Title},
		{Name: "description", Arg: "desc", Value: c.Description, Long: true},
		{Name: "column", Arg: "state", Value: c.State, Choices: States()},
		{Name: "priority", Arg: "priority", Value: c.Priority, Choices: Priorities()},
		{Name: "labels", Arg: "label", Value: strings.Join(c.Labels, " ")},
		{Name: "blockers", Arg: "blocker", Value: Numbers(c.Blockers)},
		{Name: "because", Arg: "because", Value: c.Because},
		{Name: "deadline", Arg: "deadline", Value: domain.FormatDay(c.Deadline)},
		{Name: "estimate", Arg: "estimate", Value: c.Estimate},
		{Name: "time spent", Arg: "spent", Value: c.Spent},
		{Name: "parent", Arg: "parent", Value: Maybe(c.Parent)},
		{Name: "objectives", Arg: "objective", Value: Numbers(c.Objectives)},
	}
}

// GoalFields is every field of an objective. Fewer, and none of them about
// what is in the way: an objective is not blocked, the work under it is.
func GoalFields(g Goal) []Field {
	return []Field{
		{Name: "title", Arg: "title", Value: g.Title},
		{Name: "description", Arg: "desc", Value: g.Description, Long: true},
		{Name: "standing", Arg: "state", Value: g.State, Choices: Standings},
		{Name: "priority", Arg: "priority", Value: g.Priority, Choices: Priorities()},
		{Name: "labels", Arg: "label", Value: strings.Join(g.Labels, " ")},
		{Name: "deadline", Arg: "deadline", Value: domain.FormatDay(g.Deadline)},
		{Name: "estimate", Arg: "estimate", Value: g.Estimate},
	}
}

// States is the closed set a column is picked from, read off the domain rather
// than listed again here.
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

// NumbersIn reads back what Numbers wrote. What is not a number is passed
// over rather than refused: this reads a field somebody may have typed into,
// and a form that threw away everything because one word was wrong would lose
// the choices that were right.
func NumbersIn(said string) []int {
	var all []int
	for _, word := range strings.Fields(said) {
		if n, err := strconv.Atoi(word); err == nil {
			all = append(all, n)
		}
	}
	return all
}

// Numbers writes a list of task or objective numbers the way a form takes them
// back.
func Numbers(all []int) string {
	out := make([]string, 0, len(all))
	for _, n := range all {
		out = append(out, strconv.Itoa(n))
	}
	return strings.Join(out, " ")
}

// Maybe writes a number, or nothing when there is none.
func Maybe(n int) string {
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}

// Words splits what was typed into a field the way a shell would have split
// it: on spaces, and nothing cleverer. A form is not a place to invent a
// syntax.
func Words(s string) []string {
	found := strings.Fields(s)
	if found == nil {
		return []string{}
	}
	return found
}
