package reading

import (
	"testing"

	"cairn/internal/objective"
)

// An absence is a fact that gets told, and this is the one that a reading can
// only learn that way: whether anything is written about an objective at all.
func TestAGoalIsToldItHasADescription(t *testing.T) {
	f := &fake{}
	built, err := objective.NewObjective("Ship it")
	if err != nil {
		t.Fatalf("building an objective: %v", err)
	}
	plainGoal, err := goalOf(f, built)
	if err != nil {
		t.Fatalf("reading it: %v", err)
	}
	if plainGoal.Described {
		t.Fatal("an objective with nothing written about it was said to have a description")
	}

	var told objective.AsTold
	built.Tell(&told)
	told.Described("why it matters")
	described, err := goalOf(f, told.Objective())
	if err != nil {
		t.Fatalf("reading it: %v", err)
	}
	if !described.Described {
		t.Fatal("an objective with something written about it was said to have none")
	}
}
