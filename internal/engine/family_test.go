package engine

import (
	"errors"
	"slices"
	"testing"

	"cairn/internal/objective"
)

func TestDeleteObjectives_passesOverASubObjectiveDeletedWithItsParent(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	for _, title := range []string{"Write", "Test"} {
		if _, err := e.Ask(ObjectiveNew, given("title", title, "parent", "1")); err != nil {
			t.Fatal(err)
		}
	}
	went, err := e.deleteObjectives([]int{1, 2}, deletion{subs: "keep"})
	if err != nil || !slices.Equal(went.Kept, []int{3}) {
		t.Errorf("deleting 1 and its sub-objective 2 kept %v, %v; want 3", went.Kept, err)
	}
}

func TestProposedBy_aTopObjectiveReachedProposesNothing(t *testing.T) {
	e := openEngine(t)
	c := builtObjective(t, 1, "Ship").Carried()
	c.State = objective.Reached
	if proposed, err := e.proposedBy(nil, objective.Made(c)); proposed != nil || err != nil {
		t.Errorf("proposed %v, %v", proposed, err)
	}
}

func TestNamesOnly_refusesATopObjectiveAsASubObjective(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	addObjective(t, e, "Other")
	tr, err := e.tree()
	if err != nil {
		t.Fatal(err)
	}
	if err := (deletion{keep: []int{2}}).namesOnly([]int{1}, tr); !errors.Is(err, ErrNotASubObjective) {
		t.Errorf("a top objective kept as a sub-objective of 1: %v", err)
	}
}

func TestNamesOnly_refusesASubObjectiveOfAnotherObjective(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	addObjective(t, e, "Other")
	if _, err := e.Ask(ObjectiveNew, given("title", "Elsewhere", "parent", "2")); err != nil {
		t.Fatal(err)
	}
	tr, err := e.tree()
	if err != nil {
		t.Fatal(err)
	}
	if err := (deletion{keep: []int{3}}).namesOnly([]int{1}, tr); !errors.Is(err, ErrNotASubObjective) {
		t.Errorf("a sub-objective of 2 kept as one of 1: %v", err)
	}
}
