package board

import (
	"testing"

	"cairn/internal/engine"
)

// A field says which engine argument it fills, so nothing has to translate one
// vocabulary into the other on the way to the store. These two tests are what
// stops the two drifting: a field renamed on screen still names its argument,
// and an argument renamed in the engine breaks here rather than at runtime.

func TestEveryTaskFieldNamesAnArgumentTaskEditTakes(t *testing.T) {
	takes := map[string]bool{}
	for _, a := range engine.TaskEdit.Takes {
		takes[a.Name] = true
	}
	fields := taskFields(Card{})
	if len(fields) == 0 {
		t.Fatal("taskFields is empty")
	}
	for _, f := range fields {
		if f.Arg == "" {
			t.Errorf("field %q names no engine argument", f.Name)
			continue
		}
		if !takes[f.Arg] {
			t.Errorf("field %q names %q, which `task edit` does not take", f.Name, f.Arg)
		}
	}
}

func TestEveryGoalFieldNamesAnArgumentObjectiveEditTakes(t *testing.T) {
	takes := map[string]bool{}
	for _, a := range engine.ObjectiveEdit.Takes {
		takes[a.Name] = true
	}
	fields := goalFields(Goal{})
	if len(fields) == 0 {
		t.Fatal("goalFields is empty")
	}
	for _, f := range fields {
		if f.Arg == "" {
			t.Errorf("field %q names no engine argument", f.Name)
			continue
		}
		if !takes[f.Arg] {
			t.Errorf("field %q names %q, which `objective edit` does not take", f.Name, f.Arg)
		}
	}
}
