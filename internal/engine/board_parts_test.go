package engine

import (
	"slices"
	"testing"

	"cairn/internal/task"
)

func TestStepsIn_readsEveryLine(t *testing.T) {
	steps, err := StepsIn("[x] A\n[ ] B")
	if err != nil {
		t.Fatal(err)
	}
	if want := []task.Step{{Title: "A", Done: true}, {Title: "B"}}; !slices.Equal(steps, want) {
		t.Errorf("StepsIn = %v, want %v", steps, want)
	}
}
