package board

import (
	"cmp"
	"slices"

	"cairn/internal/engine"
)

// Group is the operations acting on one thing, in the catalogue's order.
type Group struct {
	Thing string
	Ops   []engine.Operation
}

// Menu is what can be done, grouped by the thing each operation acts on, in the catalogue's order: the tree the command line and the MCP server show; an operation the pages turn off is not in it.
func Menu(pages engine.Pages) []Group {
	var groups []Group
	at := map[string]int{}
	for _, op := range append(engine.All(), engine.ForAPersonOnly()...) {
		if !pages.Offers(op) {
			continue
		}
		i, seen := at[op.Thing]
		if !seen {
			i, at[op.Thing] = len(groups), len(groups)
			groups = append(groups, Group{Thing: op.Thing})
		}
		groups[i].Ops = append(groups[i].Ops, op)
	}
	return groups
}

// MenuOf is what can be done to one element of a kind: the operations on that thing that name one by its number, what loses work last.
func MenuOf(thing string, pages engine.Pages) []engine.Operation {
	var ops []engine.Operation
	for _, g := range Menu(pages) {
		if g.Thing != thing {
			continue
		}
		for _, op := range g.Ops {
			if namesOne(op) {
				ops = append(ops, op)
			}
		}
	}
	slices.SortStableFunc(ops, func(a, b engine.Operation) int { return cmp.Compare(losesWork(a), losesWork(b)) })
	return ops
}

func namesOne(op engine.Operation) bool {
	return slices.ContainsFunc(op.Takes, func(arg engine.Argument) bool { return arg.Name == "number" && arg.Positional })
}

func losesWork(op engine.Operation) int {
	if op.Level == engine.Full {
		return 1
	}
	return 0
}
