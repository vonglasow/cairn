package engine

import (
	"testing"

	"cairn/internal/task"
)

func TestUntimed_leavesNoListingWhereThereWasNoneAndKeepsEveryTask(t *testing.T) {
	if got := (Answer{}).untimed().Tasks; got != nil {
		t.Errorf("no listing became %v", got)
	}
	if got := (Answer{Tasks: []task.Task{builtTask(t, 1, "a")}}).untimed().Tasks; len(got) != 1 {
		t.Errorf("one task became %d", len(got))
	}
}
