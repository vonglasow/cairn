package sqlite

import (
	"cairn/internal/engine/backup"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func held(t *testing.T, path string, titles ...string) {
	t.Helper()
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for _, title := range titles {
		keep(t, s, title)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func vaultOn(t *testing.T, titles ...string) (*Vault, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	held(t, path, titles...)
	return NewVault(path), path
}

func TestVault_reportsAPlaceItCannotUse(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "nothing", "here", "cairn.db"))
	if _, err := v.Make(); err == nil {
		t.Error("Make should report a store it cannot backup")
	}
}

func TestHeldOn_readsTheOlderShapeThatHasNoObjectives(t *testing.T) {
	older := filepath.Join(t.TempDir(), "older.db")
	db, err := sql.Open("sqlite", older)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(
		`CREATE TABLE tasks (number INTEGER PRIMARY KEY); INSERT INTO tasks VALUES (1), (2)`); err != nil {
		t.Fatalf("laying down the older shape: %v", err)
	}

	held, err := heldOn(db)
	if err != nil {
		t.Fatalf("heldOn: %v", err)
	}
	if held.Tasks != 2 || held.Objectives != 0 {
		t.Errorf("it holds %+v, want 2 tasks and no objectives", held)
	}
}

func TestNumberOf_takesNoNameButOurs(t *testing.T) {
	for _, tc := range []struct {
		name   string
		number int
		ours   bool
	}{
		{"cairn-0004.db", 4, true},
		{"notes.txt", 0, false},
		{"cairn-0004.sqlite", 0, false},
		{"cairn-four.db", 0, false},
		{"cairn-.db", 0, false},
	} {
		number, ours := numberOf(tc.name)
		if ours != tc.ours || number != tc.number {
			t.Errorf("numberOf(%q) = %d, %v — want %d, %v",
				tc.name, number, ours, tc.number, tc.ours)
		}
	}
}

func TestVault_backsUpADamagedStoreAsItLies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, []byte("not a database at all"), 0o600); err != nil {
		t.Fatal(err)
	}
	made, err := NewVault(path).Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	kept, err := os.ReadFile(made.Path)
	if err != nil || string(kept) != "not a database at all" {
		t.Errorf("the backup holds %q, %v, want the file as it lay", kept, err)
	}
}

func TestVault_skipsANumberAnotherCairnTookMeanwhile(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	if _, err := v.Make(); err != nil {
		t.Fatalf("Make: %v", err)
	}
	theirs := v.pathOf(2)
	if err := os.WriteFile(theirs, []byte("theirs"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	made, err := v.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if made.Number != 3 {
		t.Errorf("Number = %d, want 3: a number another cairn took is not taken again", made.Number)
	}
	if got, _ := os.ReadFile(theirs); string(got) != "theirs" {
		t.Errorf("the other cairn's backup was overwritten")
	}
}

func TestVault_refusesToPutBackABackupThatHoldsNoTasks(t *testing.T) {
	v, path := vaultOn(t, "Write the spec")
	made, err := v.Make()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(made.Path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := v.Restore(made.Number); !errors.Is(err, backup.ErrDamagedBackup) {
		t.Fatalf("Restore of an empty file = %v, want ErrDamagedBackup", err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if held, _ := s.Held(); held.Tasks != 1 {
		t.Errorf("%d tasks, want the store left as it was", held.Tasks)
	}
}

func TestSnapshot_saysWhyAReadableStoreWasNotBackedUp(t *testing.T) {
	_, path := vaultOn(t, "Write the spec")
	taken := filepath.Join(t.TempDir(), "taken.db")
	if err := os.WriteFile(taken, []byte("somebody else's"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := snapshot(path, taken); err == nil {
		t.Error("snapshot copied a readable store by hand rather than saying why it could not read it")
	}
	if got, _ := os.ReadFile(taken); string(got) != "somebody else's" {
		t.Error("what was in the way was overwritten")
	}
}
