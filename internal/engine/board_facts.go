package engine

// What a card and an objective say about themselves, worked out here once so
// that the two boards only choose which of it to show and how to draw it.

import (
	"strconv"
	"strings"
)

// InTheWay is what stops a task: the tasks it waits on, or the words somebody
// wrote when it waits on something cairn does not hold. The tasks come first —
// they are what can be acted on.
func (c Card) InTheWay() string {
	if len(c.Blockers) > 0 {
		return Numbers(c.Blockers)
	}
	return c.Because
}

// Cost is how long a task has taken against how long it was reckoned to take,
// and whether it has gone past. Nothing when neither was ever written: "0h/0h"
// says nothing and takes the room of something that would.
func (c Card) Cost() (said string, over bool) {
	switch {
	case c.Spent == "" && c.Estimate == "":
		return "", false
	case c.Estimate == "":
		return c.Spent, false
	case c.Spent == "":
		return c.Estimate, false
	}
	return c.Spent + "/" + c.Estimate, c.Overrun
}

// PartsDone is how many of the parts under a task are done out of how many
// there are, and nothing when there are none.
func (c Card) PartsDone() string {
	if c.SubTotal == 0 {
		return ""
	}
	return strconv.Itoa(c.SubDone) + "/" + strconv.Itoa(c.SubTotal)
}

// Served is the objectives a task counts towards, as this reading holds them,
// in the order the task names them. One the reading never heard of is left out.
func (s Snapshot) Served(c Card) []Goal {
	var out []Goal
	for _, n := range c.Objectives {
		if g, found := s.Goal(n); found {
			out = append(out, g)
		}
	}
	return out
}

// Named is what a field reads as, which is not always what it carries. A field
// that carries numbers — the objectives a task serves, the task it hangs under
// — reads as what each one is called: nobody knows an objective by its number.
// A task this reading never heard of reads as its number again; an objective,
// as nothing after its number.
func (s Snapshot) Named(f Field) string {
	var said []string
	switch f.Arg {
	case "objective":
		for _, n := range NumbersIn(f.Value) {
			g, _ := s.Goal(n)
			said = append(said, "#"+strconv.Itoa(n)+" "+g.Title)
		}
	case "parent":
		for _, n := range NumbersIn(f.Value) {
			title := strconv.Itoa(n)
			if c, found := s.Card(n); found {
				title = c.Title
			}
			said = append(said, "#"+strconv.Itoa(n)+" "+title)
		}
	default:
		return f.Value
	}
	return strings.Join(said, ", ")
}
