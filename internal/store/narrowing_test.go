package store

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

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

func TestNarrowed_keepsTheColumnNamed(t *testing.T) {
	s := openStore(t)
	keepInColumn(t, s, "Write the spec", domain.Todo)
	keepInColumn(t, s, "Ship it", domain.Backlog)

	narrowing := &domain.Narrowing{}
	narrowing.InColumn(string(domain.Todo))
	wantKept(t, s, narrowing, "Write the spec")
}

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

func TestNarrowed_namingNothingReadsEverything(t *testing.T) {
	s := openStore(t)
	keepInColumn(t, s, "Write the spec", domain.Todo)
	keepInColumn(t, s, "Ship it", domain.Backlog)

	wantKept(t, s, &domain.Narrowing{}, "Ship it", "Write the spec")
}

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

func TestNarrowing_leavesArchivedOutUnlessAskedFor(t *testing.T) {
	var plain narrowing
	where, _ := plain.whereTasks(taskLabels)
	if !strings.Contains(where, "archived") {
		t.Errorf("a narrowing nobody named gives %q, want it to leave archived out", where)
	}

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

func TestNarrowing_withArchivedHoldsToNeither(t *testing.T) {
	var both narrowing
	both.WithArchived()
	where, args := both.whereTasks(taskLabels)
	if strings.Contains(where, "archived") || len(args) != 0 {
		t.Errorf("where = %q, args = %v, want nothing held about being off the board", where, args)
	}
}

func TestNarrowed_takesAPartOffTheBoardWithItsWork(t *testing.T) {
	s := openStore(t)
	work := keep(t, s, "Dépôt APP")
	built, err := task.NewTask("Préparation")
	if err != nil {
		t.Fatal(err)
	}
	var p task.AsTold
	built.Tell(&p)
	p.Under(1)
	if _, err := s.Keep(p.Task()); err != nil {
		t.Fatal(err)
	}
	var a task.AsTold
	work.Tell(&a)
	a.Archived()
	if err := s.Update(a.Task()); err != nil {
		t.Fatal(err)
	}

	numbers := func(kept *domain.Narrowing) []int {
		t.Helper()
		found, err := s.Narrowed(kept)
		if err != nil {
			t.Fatal(err)
		}
		var out []int
		for _, one := range found {
			out = append(out, factsOf(one).number)
		}
		return out
	}
	if got := numbers(&domain.Narrowing{}); len(got) != 0 {
		t.Errorf("on the board: %v, want nothing — the part went with its work", got)
	}
	off := domain.Narrowing{}
	off.Archived()
	if got := numbers(&off); len(got) != 2 {
		t.Errorf("off the board: %v, want the work and its part", got)
	}
	both := domain.Narrowing{}
	both.WithArchived()
	if got := numbers(&both); len(got) != 2 {
		t.Errorf("both: %v, want the work and its part", got)
	}
}
