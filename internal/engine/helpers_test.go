package engine

import (
	"cairn/internal/objective"
	"cairn/internal/task"
	"fmt"
	"testing"
	"time"
)

func addObjective(t *testing.T, e *Engine, title string) objective.Objective {
	t.Helper()
	o, err := e.addObjective(title, ObjectiveChanges{})
	if err != nil {
		t.Fatalf("AddObjective(%q): %v", title, err)
	}
	return o
}

func serves(numbers ...int) *[]string {
	named := make([]string, 0, len(numbers))
	for _, one := range numbers {
		named = append(named, fmt.Sprint(one))
	}
	return &named
}

func serving(t *testing.T, e *Engine, title string, objectiveNumber int) task.Task {
	t.Helper()
	added, err := e.add(title, Changes{Objective: serves(objectiveNumber)})
	if err != nil {
		t.Fatalf("Add(%q): %v", title, err)
	}
	return added
}

func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
