package board

import (
	"slices"
	"testing"

	"cairn/internal/engine"
)

func TestMenu_isTheCatalogueGroupedByThing(t *testing.T) {
	groups := Menu(engine.Pages{})
	var things []string
	every := 0
	for _, g := range groups {
		things = append(things, g.Thing)
		every += len(g.Ops)
		for _, op := range g.Ops {
			if op.Thing != g.Thing {
				t.Errorf("%s is in the %s group", op.Line(), g.Thing)
			}
		}
	}
	if !slices.Contains(things, "task") || !slices.Contains(things, "objective") || things[0] != engine.All()[0].Thing {
		t.Errorf("the groups are %v, want the catalogue's things in its order", things)
	}
	if want := len(engine.All()) + len(engine.ForAPersonOnly()); every != want {
		t.Errorf("the menu holds %d operations, want the %d of the catalogue", every, want)
	}
	for _, g := range Menu(engine.PagesOff(engine.PageObjectives)) {
		if g.Thing == "objective" {
			t.Error("the objectives are in the menu with their page off")
		}
	}
}

func TestMenuOf_isWhatActsOnOneElementOfAKind(t *testing.T) {
	ops := MenuOf("task", engine.Pages{})
	if !slices.ContainsFunc(ops, func(o engine.Operation) bool { return o.Line() == engine.TaskEdit.Line() }) ||
		slices.ContainsFunc(ops, func(o engine.Operation) bool { return o.Line() == engine.TaskNew.Line() || o.Thing != "task" }) {
		t.Errorf("the menu of a task holds %v", lines(ops))
	}
}

func TestMenuOf_putsWhatLosesWorkLast(t *testing.T) {
	ops := MenuOf("task", engine.Pages{})
	if last := ops[len(ops)-1]; last.Line() != engine.TaskDelete.Line() {
		t.Errorf("the menu of a task ends with %s, want task delete", last.Line())
	}
}

func lines(ops []engine.Operation) []string {
	var out []string
	for _, op := range ops {
		out = append(out, op.Line())
	}
	return out
}
