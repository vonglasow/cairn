package engine

import (
	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
	"fmt"
	"os"
	"path/filepath"
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

func importInto(t *testing.T, e *Engine, file string) (domain.Held, error) {
	t.Helper()
	from := filepath.Join(t.TempDir(), "held.json")
	if err := os.WriteFile(from, []byte(file), 0o600); err != nil {
		t.Fatal(err)
	}
	read, err := readExport(from)
	if err != nil {
		return domain.Held{}, err
	}
	return e.importing(read)
}

func doneLongAgo(t *testing.T, e *Engine, number int) {
	t.Helper()
	if _, err := e.Ask(TaskEdit, given("number", fmt.Sprint(number), "state", "done")); err != nil {
		t.Fatal(err)
	}
	ago := time.Now().AddDate(0, 0, -10)
	if err := e.history.Record(trace.Made(trace.Carried{Task: number, Did: trace.Changed, Field: "column", Was: "todo", Became: "done", At: ago})); err != nil {
		t.Fatal(err)
	}
}

func objectiveDeleted(e *Engine, given map[string]string) error {
	_, err := e.Ask(ObjectiveDelete, Input{Confirmed: true, Given: given})
	return err
}
