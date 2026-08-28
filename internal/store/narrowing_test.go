package store

// What a narrowed read is worth. Asked of the store as a whole — tasks in, a
// fact named, the titles read back — because what a narrowing is worth is what
// comes out of the store, not the SQL it was turned into.

import (
	"strings"
	"testing"

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
func wantKept(t *testing.T, s *Store, narrowing *Narrowing, want ...string) {
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

	narrowing := &Narrowing{}
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

	narrowing := &Narrowing{}
	narrowing.Carrying(string(domain.High))
	wantKept(t, s, narrowing, "Write the spec")
}

// Two labels named keep only what carries both. Naming two things is naming
// two conditions, not a choice between them.
func TestNarrowed_keepsWhatCarriesEveryLabelNamed(t *testing.T) {
	s := openStore(t)
	keepLabelled(t, s, "Write the spec", "team", "urgent")
	keepLabelled(t, s, "Ship it", "team", "later")

	narrowing := &Narrowing{}
	narrowing.Labelled([]string{"team", "urgent"})
	wantKept(t, s, narrowing, "Write the spec")

	one := &Narrowing{}
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

	narrowing := &Narrowing{}
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

	wantKept(t, s, &Narrowing{}, "Ship it", "Write the spec")
}

// A narrowing says whether anything was named at all, because what is done
// with the answer depends on it: a listing nobody narrowed is shown as the
// tree it is, and a narrowed one is shown flat.
func TestNarrowing_narrowsOnlyWhenSomethingWasNamed(t *testing.T) {
	if (&Narrowing{}).Narrows() {
		t.Error("a narrowing nobody named anything on says it narrows")
	}
	named := &Narrowing{}
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
	}))
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	_ = serving
	keep(t, s, "Ship it")

	one := &Narrowing{}
	one.Serving(1)
	wantKept(t, s, one, "Write the spec")

	nobody := &Narrowing{}
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
		named func(*Narrowing)
		want  []string
	}{
		{"a state", func(n *Narrowing) { n.InColumn(string(objective.OnHold)) }, []string{"Ship cairn 1.0"}},
		{"a label", func(n *Narrowing) { n.Labelled([]string{"urgent"}) }, []string{"Ship cairn 1.0"}},
		{"nothing", func(*Narrowing) {}, []string{"Ship cairn 1.0", "Ship the docs"}},
	} {
		narrowing := &Narrowing{}
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
