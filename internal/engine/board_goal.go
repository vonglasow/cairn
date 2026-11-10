package engine

import (
	"slices"
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

// Standings are the four an objective can be in, in the order a list keeps them: what is being worked towards first, what was finished with last.
var Standings = []string{string(objective.Active), string(objective.OnHold), string(objective.Reached), string(objective.Dropped)}

// Stepped is what moves an objective one place up or down among its siblings, a parent with its sub-objectives; false at the first or the last, since crossing to another parent is a deliberate act.
func Stepped(goals []Goal, n int, up bool) (Input, bool) {
	parent := -1
	for _, g := range goals {
		if g.Number == n {
			parent = g.Parent
		}
	}
	var siblings []int
	for _, g := range goals {
		if g.Parent == parent {
			siblings = append(siblings, g.Number)
		}
	}
	at := slices.Index(siblings, n)
	if at < 0 || (up && at == 0) || (!up && at == len(siblings)-1) {
		return Input{}, false
	}
	given := map[string]string{"number": strconv.Itoa(n)}
	switch {
	case up && at == 1:
		given["first"] = "true"
	case up:
		given["after"] = strconv.Itoa(siblings[at-2])
	default:
		given["after"] = strconv.Itoa(siblings[at+1])
	}
	return Input{Given: given}, true
}

// DropZone is where on a row an objective is let go: in its top quarter, its middle half, or its bottom quarter.
type DropZone int

// The three parts of a row an objective can be let go over.
const (
	DropBefore DropZone = iota
	DropInto
	DropAfter
)

// Dropped is what moves an objective let go over another, over in its zone, or below every row when over is 0; false when it would land where it is. A parent goes by the parent of the row it is over, and nothing is put into a sub-objective or into the one dragged.
func Dropped(goals []Goal, dragged, over int, zone DropZone) (Input, bool) {
	given := map[string]string{"number": strconv.Itoa(dragged)}
	byNumber := map[int]Goal{}
	for _, g := range goals {
		byNumber[g.Number] = g
	}
	carried := byNumber[dragged]
	if over == 0 {
		given["parent"], given["last"] = "none", "true"
		return Input{Given: given}, carried.Parent != 0 || !lastOf(goals, 0, dragged)
	}
	target := byNumber[over]
	if len(carried.Children) > 0 && zone == DropInto && target.Parent == 0 && target.Number != dragged {
		return Input{}, false
	}
	if len(carried.Children) > 0 && target.Parent != 0 {
		target, zone = byNumber[target.Parent], DropAfter
	}
	if target.Number == dragged {
		return Input{}, false
	}
	if zone == DropInto {
		if target.Parent == 0 && len(carried.Children) == 0 {
			given["parent"], given["last"] = strconv.Itoa(target.Number), "true"
			return Input{Given: given}, true
		}
		zone = DropAfter
	}
	if zone == DropAfter {
		given["after"] = strconv.Itoa(target.Number)
		return Input{Given: given}, true
	}
	var before int
	for _, g := range goals {
		if g.Parent != target.Parent || g.Number == dragged {
			continue
		}
		if g.Number == target.Number {
			break
		}
		before = g.Number
	}
	if before == 0 {
		parent := "none"
		if target.Parent != 0 {
			parent = strconv.Itoa(target.Parent)
		}
		given["parent"], given["first"] = parent, "true"
		return Input{Given: given}, true
	}
	given["after"] = strconv.Itoa(before)
	return Input{Given: given}, true
}

func lastOf(goals []Goal, parent, n int) bool {
	last := 0
	for _, g := range goals {
		if g.Parent == parent {
			last = g.Number
		}
	}
	return last == n
}

// Sorted puts the objectives in the order a list draws them: the order a person gave them, as the store keeps it.
func Sorted(goals []Goal) []Goal {
	return slices.Clone(goals)
}

// ReachesGoal says whether what was typed reaches an objective, by the same rule a card is reached by.
func ReachesGoal(g Goal, query string) bool {
	return Reaches(g.Said, query)
}

// Counted is how much of an objective is done, the way a board writes it: its tasks, then the steps of its tasks apart, a step being no task.
func (g Goal) Counted() string {
	tasks := strconv.Itoa(g.Done-g.Steps.Done) + "/" + strconv.Itoa(g.Total-g.Steps.Total)
	if g.Steps.Total == 0 {
		return tasks
	}
	return tasks + " · " + strconv.Itoa(g.Steps.Done) + "/" + strconv.Itoa(g.Steps.Total) + " steps"
}

// Percent is the share of it done, by the domain's rule: 0 and 100 are kept for nothing done and everything done.
func (g Goal) Percent() int {
	return domain.Progress{Tasks: domain.Tally{Total: g.Total, Done: g.Done}}.Percent()
}
