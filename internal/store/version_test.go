package store

import (
	"path/filepath"
	"testing"

	"cairn/internal/task"
)

// The claim the board rests on: an integer that moves when somebody else
// writes, and does not move when nobody does. Measured rather than assumed —
// PRAGMA data_version is standard SQLite, and this driver is not the C one.

func TestVersion_DoesNotMoveOnItsOwn(t *testing.T) {
	s := openStore(t)
	first, err := s.Version()
	if err != nil {
		t.Fatalf("reading the version: %v", err)
	}
	again, err := s.Version()
	if err != nil {
		t.Fatalf("reading the version again: %v", err)
	}
	if first != again {
		t.Fatalf("the version moved with nothing happening: %d then %d", first, again)
	}
}

func TestVersion_MovesWhenSomebodyElseWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	mine, err := Open(path)
	if err != nil {
		t.Fatalf("opening the store: %v", err)
	}
	t.Cleanup(func() { _ = mine.Close() })

	before, err := mine.Version()
	if err != nil {
		t.Fatalf("reading the version: %v", err)
	}

	theirs, err := Open(path)
	if err != nil {
		t.Fatalf("opening the store a second time: %v", err)
	}
	kept, err := task.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("making a task: %v", err)
	}
	if _, err := theirs.Keep(kept); err != nil {
		t.Fatalf("keeping a task from the other connection: %v", err)
	}
	if err := theirs.Close(); err != nil {
		t.Fatalf("closing the other connection: %v", err)
	}

	after, err := mine.Version()
	if err != nil {
		t.Fatalf("reading the version again: %v", err)
	}
	if before == after {
		t.Fatalf("the version did not move after another connection wrote: still %d", after)
	}
}

func TestVersion_RefusesOnAClosedStore(t *testing.T) {
	s := openStore(t)
	if err := s.Close(); err != nil {
		t.Fatalf("closing the store: %v", err)
	}
	if _, err := s.Version(); err == nil {
		t.Fatal("a closed store answered a version")
	}
}
