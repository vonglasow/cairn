package task

import (
	"slices"
	"testing"
	"time"

	"cairn/internal/domain"
)

func carrying(t *testing.T, tell func(*AsTold)) Task {
	t.Helper()
	built := spec(t)
	a := From(built)
	tell(&a)
	return a.Task()
}

func TestMatching_findsAWordFromTheTitle(t *testing.T) {
	if !Matching(spec(t), []string{"spec"}) {
		t.Error("Matching = false, want true: the title carries \"spec\"")
	}
}

func TestMatching_ignoresCase(t *testing.T) {
	if !Matching(spec(t), []string{"SPEC"}) {
		t.Error("Matching = false, want true: looking should not care how it was typed")
	}
	got := carrying(t, func(a *AsTold) { a.Described("URGENT REVIEW") })
	if !Matching(got, []string{"urgent"}) {
		t.Error("Matching = false, want true: what is carried is read case-insensitively too")
	}
}

func TestMatching_refusesWhatIsNotThere(t *testing.T) {
	if Matching(spec(t), []string{"nowhere"}) {
		t.Error("Matching = true, want false: nothing carries \"nowhere\"")
	}
}

func TestMatching_reachesTheDescription(t *testing.T) {
	got := carrying(t, func(a *AsTold) { a.Described("the first") })
	if !Matching(got, []string{"first"}) {
		t.Error("Matching = false, want true: the description carries \"first\"")
	}
}

func TestMatching_reachesWhatAListingNarrowsBy(t *testing.T) {
	for _, tc := range []struct {
		what   string
		told   func(*AsTold)
		looked string
	}{
		{"the column", func(a *AsTold) { a.InColumn(domain.Current) }, "current"},
		{"the priority", func(a *AsTold) { a.Carries(domain.High) }, "high"},
		{"a label", func(a *AsTold) {
			a.Labelled("team")
			a.Labelled("urgent")
		}, "urgent"},
	} {
		if !Matching(carrying(t, tc.told), []string{tc.looked}) {
			t.Errorf("looking for %q did not reach %s", tc.looked, tc.what)
		}
	}
}

func TestMatching_reachesTheDeadlineTheWayItIsShown(t *testing.T) {
	deadline, err := domain.ParseDeadline("2026-09-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	got := carrying(t, func(a *AsTold) { a.Due(deadline) })
	if !Matching(got, []string{"2026-09"}) {
		t.Error("Matching = false, want true: the deadline reads 2026-09-01")
	}
}

func TestMatching_reachesADurationTheWayItIsShown(t *testing.T) {
	spent, err := domain.ParseEstimate("90m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	got := carrying(t, func(a *AsTold) { a.Spent(spent) })
	if !Matching(got, []string{"1h30m"}) {
		t.Error("Matching = false, want true: 90 minutes reads 1h30m")
	}
}

func TestMatching_everyNamedThingHasToBeFound(t *testing.T) {
	got := carrying(t, func(a *AsTold) { a.Described("the first") })
	if !Matching(got, []string{"first", "spec"}) {
		t.Error("Matching = false, want true: both \"first\" and \"spec\" are carried")
	}
	if Matching(got, []string{"first", "nowhere"}) {
		t.Error("Matching = true, want false: \"nowhere\" is not carried")
	}
}

func TestMatching_aPhraseWithASpaceIsOneThing(t *testing.T) {
	if !Matching(spec(t), []string{"the spec"}) {
		t.Error("Matching = false, want true: \"the spec\" sits in the title as one phrase")
	}
	got := carrying(t, func(a *AsTold) { a.Described("spec of the thing") })
	if Matching(got, []string{"the thing spec"}) {
		t.Error("Matching = true, want false: that phrase does not sit anywhere as one")
	}
}

func TestMatching_anEmptyNamedThingIsNeverFound(t *testing.T) {
	if Matching(spec(t), []string{""}) {
		t.Error("Matching = true, want false: an empty name is never found")
	}
}

func TestMatching_anAbsenceContributesNothing(t *testing.T) {
	got := carrying(t, func(a *AsTold) {
	})
	if Matching(got, []string{"team"}) {
		t.Error("Matching = true, want false: nothing carries \"team\"")
	}
}

func TestMatching_looksInTheEstimate(t *testing.T) {
	built, err := NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	a := From(built)
	a.Estimated(90 * time.Minute)
	carrying := a.Task()

	if !Matching(carrying, []string{"1h30m"}) {
		t.Error("a task with an estimate of 1h30m is not found by it")
	}
	if Matching(carrying, []string{"5400"}) {
		t.Error("a task is found by the seconds of its estimate, which nobody types")
	}
}

func TestMatching_reachesItsOwnNumberAndNoOtherTasks(t *testing.T) {
	numbered := func(n int) Task {
		return carrying(t, func(a *AsTold) { a.Numbered(n) })
	}
	for _, tc := range []struct {
		what  string
		task  Task
		named string
		want  bool
	}{
		{"its own number", numbered(1), "1", true},
		{"a number it does not carry", numbered(1), "9", false},
		{"its own number inside a longer one", numbered(1228), "228", true},
		{"the number of an objective it serves", carrying(t, func(a *AsTold) {
			a.Numbered(2)
			a.Serves(7)
		}), "7", false},
	} {
		if got := Matching(tc.task, []string{tc.named}); got != tc.want {
			t.Errorf("%s: Matching(%q) = %v, want %v", tc.what, tc.named, got, tc.want)
		}
	}
}

func TestSaid_isWhatASearchLooksIn(t *testing.T) {
	got := carrying(t, func(a *AsTold) {
		a.InColumn(domain.Current)
		a.Carries(domain.High)
		a.Labelled("urgent")
	})
	said := Said(got)
	for _, want := range []string{"current", "high", "urgent"} {
		if !slices.Contains(said, want) {
			t.Errorf("Said = %v, want %q among it", said, want)
		}
	}
}
