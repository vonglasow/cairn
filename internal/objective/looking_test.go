package objective

import (
	"testing"

	"cairn/internal/domain"
)

// carrying builds an objective from plan(t), told the extra facts given, the
// same way TestText_showsWhatTheObjectiveCarries builds one — through the
// membrane, never through a literal struct.
func carrying(t *testing.T, tell func(*AsTold)) Objective {
	t.Helper()
	built := plan(t)
	var a AsTold
	built.Tell(&a)
	tell(&a)
	return a.Objective()
}

func TestMatching_findsAWordFromTheTitle(t *testing.T) {
	if !Matching(plan(t), []string{"release"}) {
		t.Error("Matching = false, want true: the title carries \"release\"")
	}
}

func TestMatching_ignoresCase(t *testing.T) {
	if !Matching(plan(t), []string{"RELEASE"}) {
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
	if Matching(plan(t), []string{"nowhere"}) {
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
	got := carrying(t, func(a *AsTold) { a.Standing(OnHold) })
	if !Matching(got, []string{"on-hold"}) {
		t.Error("Matching = false, want true: the state carries \"on-hold\"")
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
	estimate, err := domain.ParseEstimate("90m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	got := carrying(t, func(a *AsTold) { a.Estimated(estimate) })
	if !Matching(got, []string{"1h30m"}) {
		t.Error("Matching = false, want true: 90 minutes reads 1h30m")
	}
}

func TestMatching_everyNamedThingHasToBeFound(t *testing.T) {
	got := carrying(t, func(a *AsTold) { a.Carries(domain.High) })
	if !Matching(got, []string{"high", "release"}) {
		t.Error("Matching = false, want true: both \"high\" and \"release\" are carried")
	}
	if Matching(got, []string{"high", "nowhere"}) {
		t.Error("Matching = true, want false: \"nowhere\" is not carried")
	}
}

// TestMatching_aPhraseWithASpaceIsOneThing pins the frozen clause: a thing
// with a space in it is one thing, so it has to sit together rather than
// its words being spread across two.
func TestMatching_aPhraseWithASpaceIsOneThing(t *testing.T) {
	if !Matching(plan(t), []string{"the release"}) {
		t.Error("Matching = false, want true: \"the release\" sits in the title as one phrase")
	}
	got := carrying(t, func(a *AsTold) { a.Described("release of the thing") })
	if Matching(got, []string{"the thing release"}) {
		t.Error("Matching = true, want false: that phrase does not sit anywhere as one")
	}
}

// TestMatching_aNumberIsNeverReached pins the frozen reason: looking for "1"
// must not answer an objective numbered 1, since asking by number is what
// `objective show` is for.
func TestMatching_aNumberIsNeverReached(t *testing.T) {
	if Matching(plan(t), []string{"1"}) {
		t.Error("Matching = true, want false: a number is never reached by looking")
	}
}

// TestMatching_anEmptyNamedThingIsNeverFound guards the other side of the
// same rule: an absent fact contributes nothing, so nothing named as nothing
// may be satisfied by it — an objective with plenty to carry must still not
// be found by looking for "".
func TestMatching_anEmptyNamedThingIsNeverFound(t *testing.T) {
	if Matching(plan(t), []string{""}) {
		t.Error("Matching = true, want false: an empty name is never found")
	}
}

// TestMatching_anAbsenceContributesNothing pins the frozen reason directly:
// what an objective does not carry cannot be what it is found by.
func TestMatching_anAbsenceContributesNothing(t *testing.T) {
	got := carrying(t, func(a *AsTold) {
		a.NotDescribed()
		a.NotDue()
		a.NotEstimated()
		a.NotLabelled()
		a.LabelsTold()
	})
	if Matching(got, []string{"team"}) {
		t.Error("Matching = true, want false: nothing carries \"team\"")
	}
}
