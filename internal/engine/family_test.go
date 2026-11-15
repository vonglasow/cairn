package engine

import (
	"slices"
	"testing"

	"cairn/internal/engine/objective"
)

func TestDeleteObjectives_passesOverASubObjectiveDeletedWithItsParent(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	for _, title := range []string{"Write", "Test"} {
		if _, err := e.Ask(ObjectiveNew, given("title", title, "parent", "1")); err != nil {
			t.Fatal(err)
		}
	}
	went, err := e.deleteObjectives([]int{1, 2}, objective.Deletion{Subs: "keep"})
	if err != nil || !slices.Equal(went.Kept, []int{3}) {
		t.Errorf("deleting 1 and its sub-objective 2 kept %v, %v; want 3", went.Kept, err)
	}
}
