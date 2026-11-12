package store

import (
	"slices"
	"testing"

	"cairn/internal/objective"
)

func keepUnder(t *testing.T, s *Store, title string, parent int) objective.Objective {
	t.Helper()
	built, err := objective.NewObjective(title)
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	told := objective.From(built)
	told.Under(parent)
	kept, err := s.KeepObjective(told.Objective())
	if err != nil {
		t.Fatalf("KeepObjective: %v", err)
	}
	return kept
}

func listedTitles(t *testing.T, s *Store) []string {
	t.Helper()
	all, err := s.AllObjectives()
	if err != nil {
		t.Fatalf("AllObjectives: %v", err)
	}
	return objectiveTitles(all)
}

func TestAllObjectives_listsThemInTheOrderTheyWereKept(t *testing.T) {
	s := openStore(t)
	keepObjective(t, s, "First")
	keepObjective(t, s, "Second")
	if got := listedTitles(t, s); !slices.Equal(got, []string{"First", "Second"}) {
		t.Errorf("listed %q, want First then Second", got)
	}
}

func TestPlaceObjective_putsOneFirstAmongItsSiblings(t *testing.T) {
	s := openStore(t)
	keepObjective(t, s, "First")
	keepObjective(t, s, "Second")
	keepObjective(t, s, "Third")
	if err := s.PlaceObjective(3, 0, 0); err != nil {
		t.Fatalf("PlaceObjective: %v", err)
	}
	if got := listedTitles(t, s); !slices.Equal(got, []string{"Third", "First", "Second"}) {
		t.Errorf("listed %q, want Third, First, Second", got)
	}
}

func TestPlaceObjective_putsOneBehindAnother(t *testing.T) {
	s := openStore(t)
	keepObjective(t, s, "First")
	keepObjective(t, s, "Second")
	keepObjective(t, s, "Third")
	if err := s.PlaceObjective(1, 0, 3); err != nil {
		t.Fatalf("PlaceObjective: %v", err)
	}
	if got := listedTitles(t, s); !slices.Equal(got, []string{"Second", "Third", "First"}) {
		t.Errorf("listed %q, want Second, Third, First", got)
	}
}

func TestPlaceObjective_movesOneUnderAnotherParent(t *testing.T) {
	s := openStore(t)
	keepObjective(t, s, "Home")
	keepObjective(t, s, "Work")
	keepUnder(t, s, "Garden", 1)
	keepUnder(t, s, "Spec", 2)
	if err := s.PlaceObjective(3, 2, 4); err != nil {
		t.Fatalf("PlaceObjective: %v", err)
	}
	moved, err := s.FindObjective(3)
	if err != nil {
		t.Fatalf("FindObjective: %v", err)
	}
	if moved.Carried().Parent != 2 {
		t.Errorf("parent = %d, want 2", moved.Carried().Parent)
	}
	if got := listedTitles(t, s); !slices.Equal(got, []string{"Home", "Work", "Spec", "Garden"}) {
		t.Errorf("listed %q, want Home, Work, Spec, Garden", got)
	}
}

func positionsUnder(t *testing.T, s *Store, parent int) []int {
	t.Helper()
	got, err := rowsOf(s, one[int], `SELECT position FROM objectives WHERE COALESCE(parent, 0) = ? ORDER BY position`, parent)
	if err != nil {
		t.Fatalf("reading the positions under %d: %v", parent, err)
	}
	return got
}

func TestPlaceObjective_numbersTheSiblingsItLeavesAndJoinsFromOne(t *testing.T) {
	s := openStore(t)
	keepObjective(t, s, "Home")
	keepObjective(t, s, "Work")
	keepUnder(t, s, "Garden", 1)
	keepUnder(t, s, "Roof", 1)
	keepUnder(t, s, "Cellar", 1)
	keepUnder(t, s, "Spec", 2)
	if err := s.PlaceObjective(3, 2, -1); err != nil {
		t.Fatalf("PlaceObjective: %v", err)
	}
	if got := positionsUnder(t, s, 1); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("the objectives left under Home sit at %v, want 1 and 2", got)
	}
	if got := positionsUnder(t, s, 2); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("the objectives under Work sit at %v, want 1 and 2", got)
	}
}
