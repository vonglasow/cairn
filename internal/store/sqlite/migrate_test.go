package sqlite

import (
	"cairn/internal/engine/backup"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/engine/migration"
)

func olderStore(t *testing.T, title string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	if _, err := db.Exec(factsOfChange(migration.Known[0]).Does); err != nil {
		t.Fatalf("writing the older shape: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO tasks (title, state, priority) VALUES (?, 'backlog', 'normal')`, title); err != nil {
		t.Fatalf("writing a task: %v", err)
	}
	return path
}

func TestMigrations_refusesALedgerThatDoesNotSayWhen(t *testing.T) {
	s := openStore(t)
	if _, err := s.db.Exec(`UPDATE migrations SET applied_at = 'yesterday'`); err != nil {
		t.Fatalf("spoiling the ledger: %v", err)
	}
	if _, err := s.Migrations(); err == nil || !strings.Contains(err.Error(), "yesterday") {
		t.Errorf("Migrations = %v, want the moment it could not read named", err)
	}
}

func TestApply_refusesAChangeAlreadyApplied(t *testing.T) {
	if _, err := openStore(t).Apply(1); !errors.Is(err, migration.ErrAlreadyApplied) {
		t.Errorf("err = %v, want ErrAlreadyApplied", err)
	}
}

func TestRollBack_refusesTheShapeAStoreStartsWith(t *testing.T) {
	s := openStore(t)
	for range migration.Known[1:] {
		if _, err := s.RollBack(); err != nil {
			t.Fatalf("RollBack: %v", err)
		}
	}
	if _, err := s.RollBack(); !errors.Is(err, migration.ErrNothingToRollBack) {
		t.Errorf("err = %v, want ErrNothingToRollBack", err)
	}
}

func TestMigrations_reportsAStoreItCannotRead(t *testing.T) {
	s := closed(t)
	if _, err := s.Migrations(); err == nil {
		t.Error("Migrations on a closed store should fail")
	}
	if _, err := s.Waiting(); err == nil {
		t.Error("Waiting on a closed store should fail")
	}
	if _, err := s.Apply(2); err == nil {
		t.Error("Apply on a closed store should fail")
	}
	if _, err := s.ApplyAll(); err == nil {
		t.Error("ApplyAll on a closed store should fail")
	}
	if _, err := s.RollBack(); err == nil {
		t.Error("RollBack on a closed store should fail")
	}
}

func holdsTable(t *testing.T, s *Store, name string) bool {
	t.Helper()
	var found string
	err := s.db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, name,
	).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}
	if err != nil {
		t.Fatalf("looking for table %q: %v", name, err)
	}
	return true
}

func TestHeld_reportsAStoreItCannotRead(t *testing.T) {
	s := openStore(t)
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := s.Held(); err == nil {
		t.Error("Held on a closed store returned no error")
	}
}

func knowing(does, undoes string) []migration.Migration {
	return []migration.Migration{migration.Known[0], migration.Made(migration.Carried{Name: "a change written to fail half way", Does: does, Undoes: undoes})}
}

func TestApplyAll_leavesNothingOfAChangeThatFailedHalfWay(t *testing.T) {
	known := knowing(`CREATE TABLE half (x); CREATE TABLE half (x);`, `DROP TABLE half`)
	s, err := OpenBehindKnowing(olderStore(t, "Write the spec"), known)
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()

	if _, err := s.ApplyAll(); err == nil {
		t.Fatal("ApplyAll succeeded, want the second statement refused")
	}
	if holdsTable(t, s, "half") {
		t.Error("the first statement was kept: a change is made whole or not at all")
	}
}

func TestRollBack_leavesTheChangeWholeWhenUndoingItFailsHalfWay(t *testing.T) {
	known := knowing(`CREATE TABLE half (x)`, `DROP TABLE half; DROP TABLE nowhere;`)
	s, err := OpenBehindKnowing(olderStore(t, "Write the spec"), known)
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()
	if _, err := s.Apply(2); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if _, err := s.RollBack(); err == nil {
		t.Fatal("RollBack succeeded, want the second statement refused")
	}
	if !holdsTable(t, s, "half") {
		t.Error("the table was dropped though the rollback failed: undone whole or not at all")
	}
	if waiting, _ := s.Waiting(); len(waiting) != 0 {
		t.Errorf("%d changes waiting, want the change still applied", len(waiting))
	}
}

func TestMake_saysAChangeAnotherCairnAppliedFirstIsApplied(t *testing.T) {
	s, err := OpenBehind(olderStore(t, "Write the spec"))
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()
	if _, err := s.Apply(2); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if err := s.make(migration.Known[1]); !errors.Is(err, migration.ErrAlreadyApplied) {
		t.Errorf("make of a change already applied = %v, want ErrAlreadyApplied", err)
	}
}

func TestRollBack_saysAChangeAnotherCairnUndidFirstWasUndone(t *testing.T) {
	s := openStore(t)
	stale, err := s.Migrations()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.RollBack(); err != nil {
		t.Fatalf("RollBack: %v", err)
	}
	if _, err := s.rollBackFrom(stale); !errors.Is(err, migration.ErrUndoneMeanwhile) {
		t.Errorf("a rollback read before another one = %v, want ErrUndoneMeanwhile", err)
	}
}

func TestOpen_refusesAnotherApplicationsDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theirs.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE invoices (id INTEGER PRIMARY KEY, total REAL)`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	if _, err := Open(path); !errors.Is(err, backup.ErrNotAStore) {
		t.Fatalf("Open = %v, want ErrNotAStore", err)
	}
	check, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = check.Close() }()
	var n int
	if err := check.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type = 'table'`).Scan(&n); err != nil || n != 1 {
		t.Errorf("%d tables, %v: want the other application's one table and nothing of cairn's", n, err)
	}
}
