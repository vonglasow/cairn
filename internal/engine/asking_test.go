package engine

import (
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"cairn/internal/store"
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

	if _, err := e.Ask(BackupNew, Input{}); err != nil {
		t.Fatalf("Ask(BackupNew): %v", err)
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

// Every field of the engine that reaches the store is handed the one it is
// kept by, in a single place: an act runs as one transaction only when every
// door it can reach the store through is that transaction, and a field added
// tomorrow and forgotten would write beside it without a word.
func TestKeptBy_handsEveryDoorToTheStoreTheOneGiven(t *testing.T) {
	kept, err := store.Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = kept.Close() })
	e := (&Engine{}).keptBy(kept)
	value := reflect.ValueOf(*e)
	for i := range value.NumField() {
		field := value.Type().Field(i)
		if field.Type.Kind() != reflect.Interface && field.Type != reflect.TypeOf(kept) {
			continue
		}
		if value.Field(i).IsNil() {
			t.Errorf("Engine.%s is left out of keptBy: an act would reach the store beside its transaction", field.Name)
		}
	}
}

// An engine held open follows a store put back under it: what it writes next
// reaches the store everyone else reads, and its version moves.
func TestAsk_followsAStorePutBackUnderIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	held, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = held.Close() })
	if _, err := held.Ask(TaskNew, given("title", "Write the spec")); err != nil {
		t.Fatal(err)
	}
	made, err := BackupsAt(path).Make()
	if err != nil {
		t.Fatal(err)
	}
	before, err := held.Version()
	if err != nil {
		t.Fatal(err)
	}
	if err := BackupsAt(path).Restore(made.Number); err != nil {
		t.Fatal(err)
	}
	// Higher, not only different: a version only ever grows, so a board that
	// keeps the highest it has seen is never told a store moved back.
	if after, err := held.Version(); err != nil || after <= before {
		t.Errorf("Version = %d, %v after a backup was put back, want it above %d", after, err, before)
	}
	if _, err := held.Ask(TaskNew, given("title", "Added on the board")); err != nil {
		t.Fatal(err)
	}
	listed, err := Do(TaskList, path, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tasks) != 2 {
		t.Errorf("%d tasks in the store, want the one put back and the one added after", len(listed.Tasks))
	}
}
