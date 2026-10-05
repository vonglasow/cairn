package engine

import (
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// engineOn gives an open engine and the path it was opened on, because what is
// under test is that the second survives inside the first.
func engineOn(t *testing.T) (*Engine, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = e.Close() })
	return e, path
}

func TestAskRunsAnOperationOnAnEngineThatStaysOpen(t *testing.T) {
	e, _ := engineOn(t)
	add(t, e, "Write the spec")

	answer, err := e.Ask(TaskList, Input{})
	if err != nil {
		t.Fatalf("Ask(TaskList): %v", err)
	}
	if len(answer.Tasks) != 1 {
		t.Fatalf("Tasks = %d, want 1", len(answer.Tasks))
	}
}

// The engine has to remember where it was opened, or an operation reaching for
// what sits beside the store is handed nowhere to look.
func TestAskGivesAnOperationThePathItWasOpenedOn(t *testing.T) {
	e, _ := engineOn(t)
	add(t, e, "Write the spec")

	if _, err := e.Ask(BackupAdd, Input{}); err != nil {
		t.Fatalf("Ask(BackupAdd): %v", err)
	}
	answer, err := e.Ask(BackupList, Input{})
	if err != nil {
		t.Fatalf("Ask(BackupList): %v", err)
	}
	if len(answer.Backups) != 1 {
		t.Fatalf("Backups = %d, want 1", len(answer.Backups))
	}
}

func TestAskRefusesAnArgumentTheOperationDoesNotTake(t *testing.T) {
	e, _ := engineOn(t)

	_, err := e.Ask(TaskList, Input{Given: map[string]string{"nonesuch": "1"}})
	if err == nil {
		t.Fatal("Ask: no error, want a refusal")
	}
	if !strings.Contains(err.Error(), "nonesuch") {
		t.Errorf("error = %q, want it to name the argument", err)
	}
}

// An engine held open is asked through an operation and nothing else: Ask
// checks what it is handed and keeps what it writes whole. It used to export
// thirty-odd methods beside it — Add, Edit, Archive — each a second door that
// checked nothing, and the boards' own tests went through them.
func TestEngine_isAskedThroughOperationsAlone(t *testing.T) {
	door := reflect.TypeFor[*Engine]()
	var exported []string
	for i := range door.NumMethod() {
		exported = append(exported, door.Method(i).Name)
	}
	if want := []string{"Ask", "Close", "Version"}; !slices.Equal(exported, want) {
		t.Errorf("an engine exports %v, want only %v", exported, want)
	}
}
