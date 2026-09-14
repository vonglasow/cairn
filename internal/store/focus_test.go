package store

// What the focus is worth is asked of the store: take up tasks, read them
// back, let go of one and of all of them.

import (
	"slices"
	"testing"
)

// inFocus reads back which tasks are being worked on, failing the test if the
// store cannot say.
func inFocus(t *testing.T, s *Store) []int {
	t.Helper()
	held, err := s.Focus()
	if err != nil {
		t.Fatalf("Focus: %v", err)
	}
	return held
}

func TestFocus_isNothingToBeginWith(t *testing.T) {
	s := openStore(t)
	keep(t, s, "The only one")
	if got := inFocus(t, s); len(got) != 0 {
		t.Errorf("Focus = %v, want nothing on a store nobody has worked in", got)
	}
}

func TestAddFocus_takesUpATask(t *testing.T) {
	s := openStore(t)
	only := factsOf(keep(t, s, "The only one")).number
	if err := s.AddFocus(only); err != nil {
		t.Fatalf("AddFocus: %v", err)
	}
	if got := inFocus(t, s); !slices.Equal(got, []int{only}) {
		t.Errorf("Focus = %v, want %v", got, []int{only})
	}
}

// Several at once, in the order they were taken up. The order matters: it is
// what a listing reads back, and the first thing taken up is the first thing
// somebody started on.
func TestAddFocus_holdsSeveralInTheOrderTheyWereTakenUp(t *testing.T) {
	s := openStore(t)
	first := factsOf(keep(t, s, "First")).number
	second := factsOf(keep(t, s, "Second")).number
	third := factsOf(keep(t, s, "Third")).number

	for _, n := range []int{third, first, second} {
		if err := s.AddFocus(n); err != nil {
			t.Fatalf("AddFocus(%d): %v", n, err)
		}
	}
	if got, want := inFocus(t, s), []int{third, first, second}; !slices.Equal(got, want) {
		t.Errorf("Focus = %v, want %v", got, want)
	}
}

// Taking up what is already held is not a change: it is held once, and it
// keeps the place in the order it already had.
func TestAddFocus_twiceHoldsATaskOnce(t *testing.T) {
	s := openStore(t)
	first := factsOf(keep(t, s, "First")).number
	second := factsOf(keep(t, s, "Second")).number
	for _, n := range []int{first, second, first} {
		if err := s.AddFocus(n); err != nil {
			t.Fatalf("AddFocus(%d): %v", n, err)
		}
	}
	if got, want := inFocus(t, s), []int{first, second}; !slices.Equal(got, want) {
		t.Errorf("Focus = %v, want %v", got, want)
	}
}

func TestDropFocus_letsGoOfOneAndLeavesTheRest(t *testing.T) {
	s := openStore(t)
	first := factsOf(keep(t, s, "First")).number
	second := factsOf(keep(t, s, "Second")).number
	for _, n := range []int{first, second} {
		if err := s.AddFocus(n); err != nil {
			t.Fatalf("AddFocus(%d): %v", n, err)
		}
	}
	if err := s.DropFocus(first); err != nil {
		t.Fatalf("DropFocus: %v", err)
	}
	if got, want := inFocus(t, s), []int{second}; !slices.Equal(got, want) {
		t.Errorf("Focus = %v, want %v", got, want)
	}
}

// Letting go of a task that was never taken up is not a failure: the state
// asked for already holds.
func TestDropFocus_onATaskThatWasNotHeldIsNotAFailure(t *testing.T) {
	s := openStore(t)
	only := factsOf(keep(t, s, "The only one")).number
	if err := s.AddFocus(only); err != nil {
		t.Fatalf("AddFocus: %v", err)
	}
	if err := s.DropFocus(99); err != nil {
		t.Errorf("DropFocus on a task that was not held = %v, want nothing wrong", err)
	}
	if got, want := inFocus(t, s), []int{only}; !slices.Equal(got, want) {
		t.Errorf("Focus = %v, want %v", got, want)
	}
}

func TestClearFocus_leavesNothingInFocus(t *testing.T) {
	s := openStore(t)
	for _, title := range []string{"First", "Second"} {
		if err := s.AddFocus(factsOf(keep(t, s, title)).number); err != nil {
			t.Fatalf("AddFocus: %v", err)
		}
	}
	if err := s.ClearFocus(); err != nil {
		t.Fatalf("ClearFocus: %v", err)
	}
	if got := inFocus(t, s); len(got) != 0 {
		t.Errorf("Focus = %v, want nothing once let go", got)
	}
}

// Letting go of nothing is not a failure: the state asked for already holds.
func TestClearFocus_onNothingIsNotAFailure(t *testing.T) {
	s := openStore(t)
	if err := s.ClearFocus(); err != nil {
		t.Errorf("ClearFocus with nothing held = %v, want nothing wrong", err)
	}
}

// A task in focus may be deleted, and it leaves the focus with it while the
// others stay. It goes by hand: the schema's reference never fires, since
// foreign keys are off.
func TestDelete_letsGoOfTheTaskItRemoved(t *testing.T) {
	s := openStore(t)
	first := factsOf(keep(t, s, "First")).number
	second := factsOf(keep(t, s, "Second")).number
	for _, n := range []int{first, second} {
		if err := s.AddFocus(n); err != nil {
			t.Fatalf("AddFocus(%d): %v", n, err)
		}
	}
	if err := s.Delete(second); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got, want := inFocus(t, s), []int{first}; !slices.Equal(got, want) {
		t.Errorf("Focus = %v, want %v once the task was deleted", got, want)
	}
	// And deleting some other task leaves the focus alone.
	other := factsOf(keep(t, s, "Another")).number
	if err := s.Delete(other); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got, want := inFocus(t, s), []int{first}; !slices.Equal(got, want) {
		t.Errorf("Focus = %v, want %v — deleting another task released it", got, want)
	}
}
