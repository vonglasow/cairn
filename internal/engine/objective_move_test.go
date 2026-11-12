package engine

import (
	"errors"
	"slices"
	"testing"

	"cairn/internal/objective"
)

func objectiveOrder(t *testing.T, e *Engine) []int {
	t.Helper()
	all, err := e.listObjectives()
	if err != nil {
		t.Fatalf("listObjectives: %v", err)
	}
	out := []int{}
	for _, o := range all {
		out = append(out, o.Carried().Number)
	}
	return out
}

func moving(t *testing.T, e *Engine, given map[string]string) error {
	t.Helper()
	_, err := e.Ask(ObjectiveMove, Input{Given: given})
	return err
}

func TestObjectiveMove_putsOneFirst(t *testing.T) {
	e, _ := engineOn(t)
	addObjective(t, e, "Home")
	addObjective(t, e, "Work")
	if err := moving(t, e, map[string]string{"number": "2", "first": "true"}); err != nil {
		t.Fatalf("move: %v", err)
	}
	if got := objectiveOrder(t, e); !slices.Equal(got, []int{2, 1}) {
		t.Errorf("listed %v, want 2 then 1", got)
	}
}

func TestObjectiveMove_carriesAParentWithItsSubObjectives(t *testing.T) {
	e, _ := engineOn(t)
	addObjective(t, e, "Home")
	addObjective(t, e, "Work")
	under := "1"
	if _, err := e.addObjective("Garden", ObjectiveChanges{Parent: &under}); err != nil {
		t.Fatalf("add Garden: %v", err)
	}
	if err := moving(t, e, map[string]string{"number": "1", "after": "2"}); err != nil {
		t.Fatalf("move: %v", err)
	}
	if got := objectiveOrder(t, e); !slices.Equal(got, []int{2, 1, 3}) {
		t.Errorf("listed %v, want 2, then 1 with its sub-objective 3", got)
	}
}

func TestObjectiveMove_putsASubObjectiveUnderAnotherParent(t *testing.T) {
	e, _ := engineOn(t)
	addObjective(t, e, "Home")
	addObjective(t, e, "Work")
	under := "1"
	if _, err := e.addObjective("Garden", ObjectiveChanges{Parent: &under}); err != nil {
		t.Fatalf("add Garden: %v", err)
	}
	if err := moving(t, e, map[string]string{"number": "3", "parent": "2"}); err != nil {
		t.Fatalf("move: %v", err)
	}
	moved, err := e.showObjective(3)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if moved.Carried().Parent != 2 {
		t.Errorf("parent = %d, want 2", moved.Carried().Parent)
	}
}

func TestObjectiveMove_refusesAParentUnderAnother(t *testing.T) {
	e, _ := engineOn(t)
	addObjective(t, e, "Home")
	addObjective(t, e, "Work")
	under := "1"
	if _, err := e.addObjective("Garden", ObjectiveChanges{Parent: &under}); err != nil {
		t.Fatalf("add Garden: %v", err)
	}
	err := moving(t, e, map[string]string{"number": "1", "parent": "2"})
	if !errors.Is(err, objective.ErrTooDeep) {
		t.Errorf("err = %v, want ErrTooDeep", err)
	}
}

func TestEditObjective_aNewParentPutsItLastThere(t *testing.T) {
	e, _ := engineOn(t)
	addObjective(t, e, "Home")
	under := "1"
	if _, err := e.addObjective("Garden", ObjectiveChanges{Parent: &under}); err != nil {
		t.Fatalf("add Garden: %v", err)
	}
	addObjective(t, e, "Errand")
	if _, err := e.addObjective("Seed", ObjectiveChanges{Parent: &under}); err != nil {
		t.Fatalf("add Seed: %v", err)
	}
	if _, err := e.editObjective(3, ObjectiveChanges{Parent: &under}); err != nil {
		t.Fatalf("edit: %v", err)
	}
	if got := objectiveOrder(t, e); !slices.Equal(got, []int{1, 2, 4, 3}) {
		t.Errorf("listed %v, want 3 last under 1", got)
	}
}

func TestObjectiveMove_behindAnotherTakesItsParent(t *testing.T) {
	e, _ := engineOn(t)
	addObjective(t, e, "Home")
	under := "1"
	if _, err := e.addObjective("Garden", ObjectiveChanges{Parent: &under}); err != nil {
		t.Fatalf("add Garden: %v", err)
	}
	addObjective(t, e, "Errand")
	if err := moving(t, e, map[string]string{"number": "3", "after": "2"}); err != nil {
		t.Fatalf("move: %v", err)
	}
	moved, _ := e.showObjective(3)
	if moved.Carried().Parent != 1 {
		t.Errorf("parent = %d, want 1, the parent of the objective it went behind", moved.Carried().Parent)
	}
}

func TestObjectiveMove_behindASiblingUnderTheParentNamed(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	for _, title := range []string{"a", "b"} {
		if _, err := e.Ask(ObjectiveNew, given("title", title, "parent", "1")); err != nil {
			t.Fatal(err)
		}
	}
	if err := moving(t, e, map[string]string{"number": "2", "parent": "1", "after": "3"}); err != nil {
		t.Errorf("2 behind its sibling 3, under their parent: %v", err)
	}
}
