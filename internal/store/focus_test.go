package store

// What the focus is worth is asked of the store: put a task there, read it
// back, and check that putting another there releases the first.

import (
	"testing"

	"cairn/internal/task"
)

// inFocus reads back which task is being worked on, failing the test if the
// store cannot say.
func inFocus(t *testing.T, s *Store) int {
	t.Helper()
	number, err := s.Focus()
	if err != nil {
		t.Fatalf("Focus: %v", err)
	}
	return number
}

func TestFocus_isNothingToBeginWith(t *testing.T) {
	s := openStore(t)
	keep(t, s, "The only one")
	if got := inFocus(t, s); got != NoFocus {
		t.Errorf("Focus = %d, want %d on a store nobody has worked in", got, NoFocus)
	}
}

func TestSetFocus_makesATaskTheOneBeingWorkedOn(t *testing.T) {
	s := openStore(t)
	only := keep(t, s, "The only one")
	if err := s.SetFocus(factsOf(only).number); err != nil {
		t.Fatalf("SetFocus: %v", err)
	}
	if got, want := inFocus(t, s), factsOf(only).number; got != want {
		t.Errorf("Focus = %d, want %d", got, want)
	}
}

// One row holds the focus, so there is at most one task in it. Asserted rather
// than assumed: this is the whole reason it is kept in a row of its own.
func TestSetFocus_releasesTheOneBefore(t *testing.T) {
	s := openStore(t)
	first := keep(t, s, "First")
	second := keep(t, s, "Second")
	for _, one := range []task.Task{first, second} {
		if err := s.SetFocus(factsOf(one).number); err != nil {
			t.Fatalf("SetFocus: %v", err)
		}
	}
	if got, want := inFocus(t, s), factsOf(second).number; got != want {
		t.Errorf("Focus = %d, want %d — the one before was not released", got, want)
	}
}

func TestClearFocus_leavesNothingInFocus(t *testing.T) {
	s := openStore(t)
	only := keep(t, s, "The only one")
	if err := s.SetFocus(factsOf(only).number); err != nil {
		t.Fatalf("SetFocus: %v", err)
	}
	if err := s.ClearFocus(); err != nil {
		t.Fatalf("ClearFocus: %v", err)
	}
	if got := inFocus(t, s); got != NoFocus {
		t.Errorf("Focus = %d, want %d once let go", got, NoFocus)
	}
}

// Letting go of nothing is not a failure: the state asked for already holds.
func TestClearFocus_onNothingIsNotAFailure(t *testing.T) {
	s := openStore(t)
	if err := s.ClearFocus(); err != nil {
		t.Errorf("ClearFocus with nothing held = %v, want nothing wrong", err)
	}
}

// The task in focus may be deleted, and the focus goes with it. It goes by hand:
// the schema's ON DELETE SET NULL never fires, since foreign keys are off.
func TestDelete_letsGoOfTheTaskItRemoved(t *testing.T) {
	s := openStore(t)
	first := keep(t, s, "First")
	second := keep(t, s, "Second")
	if err := s.SetFocus(factsOf(second).number); err != nil {
		t.Fatalf("SetFocus: %v", err)
	}
	if err := s.Delete(factsOf(second).number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got := inFocus(t, s); got != NoFocus {
		t.Errorf("Focus = %d, want %d once the task was deleted", got, NoFocus)
	}
	// And deleting some other task leaves the focus alone.
	if err := s.SetFocus(factsOf(first).number); err != nil {
		t.Fatalf("SetFocus: %v", err)
	}
	other := keep(t, s, "Another")
	if err := s.Delete(factsOf(other).number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got, want := inFocus(t, s), factsOf(first).number; got != want {
		t.Errorf("Focus = %d, want %d — deleting another task released it", got, want)
	}
}

// A store that has somehow lost its one row still answers, and can still be
// told what is in focus: no row and a row holding nothing say the same thing.
func TestFocus_answersNothingWhenTheRowIsGone(t *testing.T) {
	s := openStore(t)
	only := keep(t, s, "The only one")
	if _, err := s.db.Exec(`DELETE FROM focus`); err != nil {
		t.Fatalf("emptying the focus table: %v", err)
	}
	if got := inFocus(t, s); got != NoFocus {
		t.Errorf("Focus = %d, want %d with no row at all", got, NoFocus)
	}
	if err := s.SetFocus(factsOf(only).number); err != nil {
		t.Fatalf("SetFocus with no row: %v", err)
	}
	if got, want := inFocus(t, s), factsOf(only).number; got != want {
		t.Errorf("Focus = %d, want %d — the row was not made", got, want)
	}
}
