package store

import (
	"errors"
	"slices"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/task"
)

func titlesOf(t *testing.T, s *Store) []string {
	t.Helper()
	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	titles := []string{}
	for _, each := range all {
		titles = append(titles, factsOf(each).Title)
	}
	return titles
}

func threeInTheBacklog(t *testing.T) *Store {
	t.Helper()
	s := openStore(t)
	for _, title := range []string{"one", "two", "three"} {
		keep(t, s, title)
	}
	return s
}

func moved(t *testing.T, s *Store, went bool, err error, want bool, order ...string) {
	t.Helper()
	if err != nil {
		t.Fatalf("moving: %v", err)
	}
	if went != want {
		t.Errorf("the move said it went somewhere: %v, want %v", went, want)
	}
	if got := titlesOf(t, s); !slices.Equal(got, order) {
		t.Errorf("the order is %v, want %v", got, order)
	}
}

func TestMoveFirst_putsATaskAtTheHead(t *testing.T) {
	s := threeInTheBacklog(t)
	went, err := s.MoveFirst(3)
	moved(t, s, went, err, true, "three", "one", "two")
}

func TestMoveFirst_saysATaskAlreadyFirstWentNowhere(t *testing.T) {
	s := threeInTheBacklog(t)
	went, err := s.MoveFirst(1)
	moved(t, s, went, err, false, "one", "two", "three")
}

func TestMoveLast_putsATaskBehindTheOthers(t *testing.T) {
	s := threeInTheBacklog(t)
	went, err := s.MoveLast(1)
	moved(t, s, went, err, true, "two", "three", "one")
}

func TestMoveAfter_putsATaskFromAboveDirectlyBehindAnother(t *testing.T) {
	s := threeInTheBacklog(t)
	went, err := s.MoveAfter(1, 2)
	moved(t, s, went, err, true, "two", "one", "three")
}

func TestMoveAfter_putsATaskFromBelowDirectlyBehindAnother(t *testing.T) {
	s := threeInTheBacklog(t)
	went, err := s.MoveAfter(3, 1)
	moved(t, s, went, err, true, "one", "three", "two")
}

func TestMoveAfter_aTaskBehindItselfGoesNowhere(t *testing.T) {
	s := threeInTheBacklog(t)
	went, err := s.MoveAfter(2, 2)
	moved(t, s, went, err, false, "one", "two", "three")
}

func TestMove_saysThereIsNoSuchTask(t *testing.T) {
	s := threeInTheBacklog(t)
	if _, err := s.MoveAfter(1, 99); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("moving behind task 99 answered %v", err)
	}
	if _, err := s.MoveFirst(99); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("moving task 99 answered %v", err)
	}
}

func TestUpdate_keepsATaskWhereItSitsWhenItStaysInItsColumn(t *testing.T) {
	s := threeInTheBacklog(t)
	first, err := s.Find(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Update(changed(first, func(a *task.AsTold) { a.Titled("still one") })); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got := titlesOf(t, s); !slices.Equal(got, []string{"still one", "two", "three"}) {
		t.Errorf("the order is %v, want the task where it sat", got)
	}
}

func TestUpdate_putsATaskLastInTheColumnItEntersAndClosesTheGap(t *testing.T) {
	s := threeInTheBacklog(t)
	for _, title := range []string{"four", "five"} {
		if _, err := s.Keep(changed(mustTask(t, title), func(a *task.AsTold) { a.InColumn(domain.Todo) })); err != nil {
			t.Fatalf("Keep: %v", err)
		}
	}
	first, err := s.Find(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Update(changed(first, func(a *task.AsTold) { a.InColumn(domain.Todo) })); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got := titlesOf(t, s); !slices.Equal(got, []string{"two", "three", "four", "five", "one"}) {
		t.Errorf("the order is %v, want the task last in the column it entered", got)
	}
	if went, err := s.MoveFirst(2); err != nil || went {
		t.Errorf("the first left in the backlog moved: %v, %v — the gap was not closed", went, err)
	}
}

func TestUpdate_keepsWhatATaskCarriesWhenItChangesColumn(t *testing.T) {
	s := openStore(t)
	built := keep(t, s, "Write the spec")
	goal := keepObjective(t, s, "Ship it")
	built = changed(built, func(a *task.AsTold) {
		a.InColumn(domain.Todo)
		a.Labelled("spec")
		a.Serves(goal.Carried().Number)
	})
	if err := s.Update(built); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err := s.Find(1)
	if err != nil {
		t.Fatal(err)
	}
	facts := factsOf(got)
	if !slices.Equal(facts.Labels, []string{"spec"}) || !slices.Equal(facts.Objectives, []int{1}) {
		t.Errorf("read back %v serving %v, want [spec] serving [1]", facts.Labels, facts.Objectives)
	}
}
