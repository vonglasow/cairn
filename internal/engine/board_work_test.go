package engine

import (
	"testing"

	"cairn/internal/objective"
)

func TestAGoalIsToldItHasADescription(t *testing.T) {
	built, err := objective.NewObjective("Ship it")
	if err != nil {
		t.Fatalf("building an objective: %v", err)
	}
	plainGoal, err := goalOf(built, nil)
	if err != nil {
		t.Fatalf("reading it: %v", err)
	}
	if plainGoal.Described() {
		t.Fatal("an objective with nothing written about it was said to have a description")
	}

	var told objective.AsTold
	built.Tell(&told)
	told.Described("why it matters")
	described, err := goalOf(told.Objective(), nil)
	if err != nil {
		t.Fatalf("reading it: %v", err)
	}
	if !described.Described() {
		t.Fatal("an objective with something written about it was said to have none")
	}
}
