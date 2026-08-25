package task

import (
	"testing"

	"cairn/internal/domain"
)

// carrying builds a task from spec(t), told the extra facts given, the same
// way TestText_showsWhatTheTaskCarries builds one — through the membrane,
// never through a literal struct.
func carrying(t *testing.T, tell func(*AsTold)) Task {
	t.Helper()
	built := spec(t)
	var a AsTold
	built.Tell(&a)
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
	// The other direction: what is carried is read the way it is shown, not
	// lower-cased on the way in, so a name typed in lower case still has to
	// reach it.
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

func TestMatching_reachesTheState(t *testing.T) {
	got := carrying(t, func(a *AsTold) { a.InColumn(domain.Current) })
	if !Matching(got, []string{"current"}) {
		t.Error("Matching = false, want true: the column carries \"current\"")
	}
}

func TestMatching_reachesThePriority(t *testing.T) {
	got := carrying(t, func(a *AsTold) { a.Carries(domain.High) })
	if !Matching(got, []string{"high"}) {
		t.Error("Matching = false, want true: the priority carries \"high\"")
	}
}

func TestMatching_reachesALabel(t *testing.T) {
	got := carrying(t, func(a *AsTold) {
		a.Labelled("team")
		a.Labelled("urgent")
		a.LabelsTold()
	})
	if !Matching(got, []string{"urgent"}) {
		t.Error("Matching = false, want true: a label carries \"urgent\"")
	}
}

// TestMatching_reachesTheDeadlineTheWayItIsShown pins the frozen example: a
// deadline is looked for as domain.FormatDay shows it, so a month reaches a
// day within it.
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

// TestMatching_reachesADurationTheWayItIsShown pins the frozen example:
// ninety minutes is shown as 1h30m, so that is what it is found by.
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
	got := carrying(t, func(a *AsTold) { a.Carries(domain.High) })
	if !Matching(got, []string{"high", "spec"}) {
		t.Error("Matching = false, want true: both \"high\" and \"spec\" are carried")
	}
	if Matching(got, []string{"high", "nowhere"}) {
		t.Error("Matching = true, want false: \"nowhere\" is not carried")
	}
}

// TestMatching_aPhraseWithASpaceIsOneThing pins the frozen clause: a thing
// with a space in it is one thing, so it has to sit together rather than
// its words being spread across two.
func TestMatching_aPhraseWithASpaceIsOneThing(t *testing.T) {
	if !Matching(spec(t), []string{"the spec"}) {
		t.Error("Matching = false, want true: \"the spec\" sits in the title as one phrase")
	}
	got := carrying(t, func(a *AsTold) { a.Described("spec of the thing") })
	if Matching(got, []string{"the thing spec"}) {
		t.Error("Matching = true, want false: that phrase does not sit anywhere as one")
	}
}

// TestMatching_aNumberIsNeverReached pins the frozen reason: looking for "1"
// must not answer a task numbered 1, since asking by number is what
// `task show` is for.
func TestMatching_aNumberIsNeverReached(t *testing.T) {
	if Matching(spec(t), []string{"1"}) {
		t.Error("Matching = true, want false: a number is never reached by looking")
	}
}

// TestMatching_anEmptyNamedThingIsNeverFound guards the other side of the
// same rule: an absent fact contributes nothing, so nothing named as nothing
// may be satisfied by it — a task with plenty to carry must still not be
// found by looking for "".
func TestMatching_anEmptyNamedThingIsNeverFound(t *testing.T) {
	if Matching(spec(t), []string{""}) {
		t.Error("Matching = true, want false: an empty name is never found")
	}
}

// TestMatching_anAbsenceContributesNothing pins the frozen reason directly:
// what a task does not carry cannot be what it is found by.
func TestMatching_anAbsenceContributesNothing(t *testing.T) {
	got := carrying(t, func(a *AsTold) {
		a.NotDescribed()
		a.NotDue()
		a.NotEstimated()
		a.NothingSpent()
		a.NotLabelled()
		a.LabelsTold()
	})
	if Matching(got, []string{"team"}) {
		t.Error("Matching = true, want false: nothing carries \"team\"")
	}
}
