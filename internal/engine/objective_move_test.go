package engine

import (
	"errors"
	"slices"
	"testing"

	"cairn/internal/engine/objective"
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
