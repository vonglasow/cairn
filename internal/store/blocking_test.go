package store

// What the store makes of what blocks a task: kept, read back, and released
// when a task on either end of the link goes.

import (
	"testing"

	"cairn/internal/domain"
	"cairn/internal/task"
)

// keepBlocked adds a task saying what is in its way.
func keepBlocked(t *testing.T, s *Store, title string, state domain.State, blockers []int, because string) task.Task {
	t.Helper()
	kept, err := s.Keep(changed(mustBuild(t, title), func(a *task.AsTold) {
		a.InColumn(state)
		for _, one := range blockers {
			a.BlockedBy(one)
		}
		a.BlockersTold()
		a.Because(because)
	}))
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	return kept
}

// What a task says blocks it comes back exactly as it was given, and a task
// that said nothing comes back saying nothing.
func TestBlockers_comeBackAsTheyWereGiven(t *testing.T) {
	s := openStore(t)
	keep(t, s, "Write the spec")
	keep(t, s, "Write the tests")
	blocked := keepBlocked(t, s, "Ship it", domain.Blocked, []int{2, 1}, "the client's answer")

	found, err := s.Find(factsOf(blocked).number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	f := factsOf(found)
	if len(f.blockers) != 2 || f.blockers[0] != 1 || f.blockers[1] != 2 {
		t.Errorf("blockers = %v, want 1 and 2 in that order", f.blockers)
	}
	if f.because != "the client's answer" {
		t.Errorf("because = %q, want the words it was given", f.because)
	}

	bare, err := s.Find(1)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if got := factsOf(bare); len(got.blockers) != 0 || got.because != "" {
		t.Errorf("a task nobody blocked carries %v and %q", got.blockers, got.because)
	}
}

// An edit says what blocks a task, so what it leaves out is taken away: a set
// given is a set replaced, the way the labels are.
func TestBlockers_areReplacedByAnUpdate(t *testing.T) {
	s := openStore(t)
	keep(t, s, "Write the spec")
	keep(t, s, "Write the tests")
	blocked := keepBlocked(t, s, "Ship it", domain.Blocked, []int{1, 2}, "the client")

	if err := s.Update(changed(blocked, func(a *task.AsTold) {
		a.NotBlocked()
		a.BlockedBy(2)
		a.BlockersTold()
		a.NotBecause()
	})); err != nil {
		t.Fatalf("Update: %v", err)
	}
	found, err := s.Find(factsOf(blocked).number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if f := factsOf(found); len(f.blockers) != 1 || f.blockers[0] != 2 || f.because != "" {
		t.Errorf("it carries %v and %q, want just task 2 and no words", f.blockers, f.because)
	}
}

// A task that is gone neither blocks anything nor is blocked by anything, so
// both sides of every link go with it. By hand, since this connection leaves
// the foreign keys off.
func TestBlockers_areReleasedWhenEitherEndGoes(t *testing.T) {
	s := openStore(t)
	blocking := keep(t, s, "Write the spec")
	blocked := keepBlocked(t, s, "Ship it", domain.Blocked, []int{factsOf(blocking).number}, "")
	number := factsOf(blocked).number

	if err := s.Delete(factsOf(blocking).number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	found, err := s.Find(number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if got := factsOf(found).blockers; len(got) != 0 {
		t.Errorf("it is still blocked by %v, and that task is gone", got)
	}

	// And the other way round: the blocked task going leaves nothing behind it.
	again := keep(t, s, "Write the tests")
	third := keepBlocked(t, s, "Announce it", domain.Blocked, []int{factsOf(again).number}, "")
	if err := s.Delete(factsOf(third).number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	left, err := s.blockersOn(factsOf(third).number)
	if err != nil {
		t.Fatalf("blockersOn: %v", err)
	}
	if len(left) != 0 {
		t.Errorf("%v is still recorded as blocking a task that is gone", left)
	}
}

// The three conditions that make "nothing blocks it" honest, each on its own.
func TestNothingBlocks_answersOnlyWhenItIsCertain(t *testing.T) {
	s := openStore(t)
	done := keepInColumn(t, s, "Write the spec", domain.Done)
	open := keepInColumn(t, s, "Write the tests", domain.Todo)
	finished, waiting := factsOf(done).number, factsOf(open).number

	free := keepBlocked(t, s, "Every blocker done", domain.Blocked, []int{finished}, "")
	keepBlocked(t, s, "One still open", domain.Blocked, []int{finished, waiting}, "")
	keepBlocked(t, s, "Blocked by words too", domain.Blocked, []int{finished}, "the client")
	keepBlocked(t, s, "Blocked by nothing named", domain.Blocked, nil, "")
	keepBlocked(t, s, "Not in the blocked column", domain.Todo, []int{finished}, "")

	answered, err := s.NothingBlocks()
	if err != nil {
		t.Fatalf("NothingBlocks: %v", err)
	}
	if len(answered) != 1 || answered[0] != factsOf(free).number {
		t.Errorf("it says %v have nothing in their way, want just task %d",
			answered, factsOf(free).number)
	}
}
