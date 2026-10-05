package store

// What a narrowed read is worth. Asked of the store as a whole — tasks in, a
// fact named, the titles read back — because what a narrowing is worth is what
// comes out of the store, not the SQL it was turned into.

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// keepLabelled adds a task carrying labels, so a test can ask for one of them.
func keepLabelled(t *testing.T, s *Store, title string, labels ...string) task.Task {
	t.Helper()
	kept, err := s.Keep(changed(mustBuild(t, title), func(a *task.AsTold) {
		for _, label := range labels {
			a.Labelled(label)
		}
		a.LabelsTold()
	}))
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	return kept
}

// wantKept fails unless a narrowed read hands back exactly these titles.
func wantKept(t *testing.T, s *Store, narrowing *domain.Narrowing, want ...string) {
	t.Helper()
	kept, err := s.Narrowed(narrowing)
	if err != nil {
		t.Fatalf("Narrowed: %v", err)
	}
	titles := make([]string, 0, len(kept))
	for _, one := range kept {
		titles = append(titles, factsOf(one).title)
	}
	if strings.Join(titles, ",") != strings.Join(want, ",") {
		t.Errorf("it kept %v, want %v", titles, want)
	}
}

// A column named keeps the tasks sitting in it and nothing else.
func TestNarrowed_keepsTheColumnNamed(t *testing.T) {
	s := openStore(t)
	keepInColumn(t, s, "Write the spec", domain.Todo)
	keepInColumn(t, s, "Ship it", domain.Backlog)

	narrowing := &domain.Narrowing{}
	narrowing.InColumn(string(domain.Todo))
	wantKept(t, s, narrowing, "Write the spec")
}

// A priority named keeps the tasks carrying it.
func TestNarrowed_keepsThePriorityNamed(t *testing.T) {
	s := openStore(t)
	high, err := s.Keep(changed(mustBuild(t, "Write the spec"), func(a *task.AsTold) {
		a.Carries(domain.High)
	}))
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	_ = high
	keep(t, s, "Ship it")

	narrowing := &domain.Narrowing{}
	narrowing.Carrying(string(domain.High))
	wantKept(t, s, narrowing, "Write the spec")
}

// Two labels named keep only what carries both. Naming two things is naming
// two conditions, not a choice between them.
func TestNarrowed_keepsWhatCarriesEveryLabelNamed(t *testing.T) {
	s := openStore(t)
	keepLabelled(t, s, "Write the spec", "team", "urgent")
	keepLabelled(t, s, "Ship it", "team", "later")

	narrowing := &domain.Narrowing{}
	narrowing.Labelled([]string{"team", "urgent"})
	wantKept(t, s, narrowing, "Write the spec")

	one := &domain.Narrowing{}
	one.Labelled([]string{"team"})
	wantKept(t, s, one, "Write the spec", "Ship it")
}

// Two facts from different fields both have to hold.
func TestNarrowed_keepsWhatEveryFactNamedHoldsFor(t *testing.T) {
	s := openStore(t)
	both, err := s.Keep(changed(mustBuild(t, "Write the spec"), func(a *task.AsTold) {
		a.InColumn(domain.Todo)
		a.Carries(domain.High)
	}))
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	_ = both
	keepInColumn(t, s, "Ship it", domain.Todo)

	narrowing := &domain.Narrowing{}
	narrowing.InColumn(string(domain.Todo))
	narrowing.Carrying(string(domain.High))
	wantKept(t, s, narrowing, "Write the spec")
}

// Naming nothing narrows nothing: a read told to keep everything reads the
// store in the one order there is, exactly as All does.
func TestNarrowed_namingNothingReadsEverything(t *testing.T) {
	s := openStore(t)
	keepInColumn(t, s, "Write the spec", domain.Todo)
	keepInColumn(t, s, "Ship it", domain.Backlog)

	wantKept(t, s, &domain.Narrowing{}, "Ship it", "Write the spec")
}

// A narrowing says whether anything was named at all, because what is done
// with the answer depends on it: a listing nobody narrowed is shown as the
// tree it is, and a narrowed one is shown flat.
func TestNarrowing_narrowsOnlyWhenSomethingWasNamed(t *testing.T) {
	if (&domain.Narrowing{}).Narrows() {
		t.Error("a narrowing nobody named anything on says it narrows")
	}
	named := &domain.Narrowing{}
	named.InColumn(string(domain.Todo))
	if !named.Narrows() {
		t.Error("a narrowing told a column says it narrows nothing")
	}
}

// An objective named keeps the tasks serving it, and one nothing serves keeps
// nothing rather than failing: a filter says what is wanted, not what exists.
func TestNarrowed_keepsWhatServesTheObjectiveNamed(t *testing.T) {
	s := openStore(t)
	serving, err := s.Keep(changed(mustBuild(t, "Write the spec"), func(a *task.AsTold) {
		a.Serves(1)
		a.ObjectivesTold()
	}))
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	_ = serving
	keep(t, s, "Ship it")

	one := &domain.Narrowing{}
	one.Serving(1)
	wantKept(t, s, one, "Write the spec")

	nobody := &domain.Narrowing{}
	nobody.Serving(9)
	wantKept(t, s, nobody)
}

// An objective is narrowed the same way, on the facts it carries. Asserted
// here as well as through the commands: a narrowing that worked for tasks and
// not for objectives would be caught nowhere in this package.
func TestNarrowedObjectives_keepsWhatWasNamed(t *testing.T) {
	s := openStore(t)
	built, err := objective.NewObjective("Ship cairn 1.0")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	if _, err := s.KeepObjective(changedObjective(built, func(a *objective.AsTold) {
		a.Standing(objective.OnHold)
		a.Labelled("urgent")
		a.LabelsTold()
	})); err != nil {
		t.Fatalf("KeepObjective: %v", err)
	}
	keepObjective(t, s, "Ship the docs")

	for _, tc := range []struct {
		what  string
		named func(*domain.Narrowing)
		want  []string
	}{
		{"a state", func(n *domain.Narrowing) { n.InColumn(string(objective.OnHold)) }, []string{"Ship cairn 1.0"}},
		{"a label", func(n *domain.Narrowing) { n.Labelled([]string{"urgent"}) }, []string{"Ship cairn 1.0"}},
		{"nothing", func(*domain.Narrowing) {}, []string{"Ship cairn 1.0", "Ship the docs"}},
	} {
		narrowing := &domain.Narrowing{}
		tc.named(narrowing)
		kept, err := s.NarrowedObjectives(narrowing)
		if err != nil {
			t.Fatalf("%s: NarrowedObjectives: %v", tc.what, err)
		}
		titles := make([]string, 0, len(kept))
		for _, one := range kept {
			titles = append(titles, factsOfObjective(one).title)
		}
		if strings.Join(titles, ",") != strings.Join(tc.want, ",") {
			t.Errorf("%s: it kept %v, want %v", tc.what, titles, tc.want)
		}
	}
}

// A narrowing leaves archived tasks out unless it is asked for them. Archiving
// is for taking work off the board, and a listing that handed it all back
// would have moved the problem rather than answered it.
//
// Even one nobody narrowed: what is off the board is off it whether or not
// anything else was named.
func TestNarrowing_leavesArchivedOutUnlessAskedFor(t *testing.T) {
	var plain narrowing
	where, _ := plain.whereTasks(taskLabels)
	if !strings.Contains(where, "archived") {
		t.Errorf("a narrowing nobody named gives %q, want it to leave archived out", where)
	}

	// The clause is the same either way; what it holds to is not. Comparing
	// the two strings says nothing, which is what this asked at first.
	_, plainArgs := plain.whereTasks(taskLabels)
	var asked narrowing
	asked.Archived()
	_, askedArgs := asked.whereTasks(taskLabels)
	if len(plainArgs) == 0 || len(askedArgs) == 0 {
		t.Fatalf("args = %v and %v, want each to hold to something", plainArgs, askedArgs)
	}
	if plainArgs[0] != false || askedArgs[0] != true {
		t.Errorf("a plain narrowing holds to %v and one asking holds to %v, want false and true",
			plainArgs[0], askedArgs[0])
	}

	// And asking for them says nothing about what else was named.
	var both narrowing
	both.InColumn("done")
	both.Archived()
	said, args := both.whereTasks(taskLabels)
	if !strings.Contains(said, "state = ?") {
		t.Errorf("where = %q, want it to keep what else was named", said)
	}
	if len(args) == 0 {
		t.Errorf("args = %v, want the column it was given", args)
	}
}

// What is due by a day, the day itself included: a deadline is the last day
// something may be done, so a listing of what is due by the first that left
// out the first would answer a different question. What carries no deadline is
// not due by that day — it is not due at all — and is left out.
func TestNarrowed_keepsWhatIsDueByADay(t *testing.T) {
	s := openStore(t)
	due := func(title, day string) {
		t.Helper()
		when, err := time.Parse("2006-01-02", day)
		if err != nil {
			t.Fatalf("reading %q: %v", day, err)
		}
		if _, err := s.Keep(changed(mustBuild(t, title), func(a *task.AsTold) {
			a.Due(when)
		})); err != nil {
			t.Fatalf("Keep %q: %v", title, err)
		}
	}
	due("Write the spec", "2026-09-10")
	due("On the day itself", "2026-10-01")
	due("Ship it", "2026-12-01")
	keep(t, s, "Someday")

	narrowing := &domain.Narrowing{}
	narrowing.DueBefore("2026-10-01")
	wantKept(t, s, narrowing, "Write the spec", "On the day itself")
}

// Asked for both, what is on the board and what is off it are held to
// nothing: the clause that leaves the archived out is not there at all.
func TestNarrowing_withArchivedHoldsToNeither(t *testing.T) {
	var both narrowing
	both.WithArchived()
	where, args := both.whereTasks(taskLabels)
	if strings.Contains(where, "archived") || len(args) != 0 {
		t.Errorf("where = %q, args = %v, want nothing held about being off the board", where, args)
	}
}
