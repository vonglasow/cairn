package engine

import (
	"cairn/internal/engine/backup"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
	"fmt"
	"os"
	"path/filepath"
	"testing"
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

func importInto(t *testing.T, e *Engine, file string) (backup.Held, error) {
	t.Helper()
	from := filepath.Join(t.TempDir(), "held.json")
	if err := os.WriteFile(from, []byte(file), 0o600); err != nil {
		t.Fatal(err)
	}
	read, err := readExport(from)
	if err != nil {
		return backup.Held{}, err
	}
	return e.importing(read)
}

func objectiveDeleted(e *Engine, given map[string]string) error {
	_, err := e.Ask(ObjectiveDelete, Input{Confirmed: true, Given: given})
	return err
}
