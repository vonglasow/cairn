package objective

import (
	"slices"
	"testing"

	"cairn/internal/domain"
)

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

func TestMatching_reachesWhatAListingNarrowsBy(t *testing.T) {
	for _, tc := range []struct {
		what   string
		told   func(*AsTold)
		looked string
	}{
		{"where it stands", func(a *AsTold) { a.Standing(OnHold) }, "on-hold"},
		{"the priority", func(a *AsTold) { a.Carries(domain.High) }, "high"},
		{"a label", func(a *AsTold) {
			a.Labelled("team")
			a.Labelled("urgent")
			a.LabelsTold()
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
	got := carrying(t, func(a *AsTold) { a.Described("the first") })
	if !Matching(got, []string{"first", "release"}) {
		t.Error("Matching = false, want true: both \"first\" and \"release\" are carried")
	}
	if Matching(got, []string{"first", "nowhere"}) {
		t.Error("Matching = true, want false: \"nowhere\" is not carried")
	}
}

func TestMatching_aPhraseWithASpaceIsOneThing(t *testing.T) {
	if !Matching(plan(t), []string{"the release"}) {
		t.Error("Matching = false, want true: \"the release\" sits in the title as one phrase")
	}
	got := carrying(t, func(a *AsTold) { a.Described("release of the thing") })
	if Matching(got, []string{"the thing release"}) {
		t.Error("Matching = true, want false: that phrase does not sit anywhere as one")
	}
}

func TestMatching_reachesItsNumber(t *testing.T) {
	if !Matching(plan(t), []string{"1"}) {
		t.Error("Matching = false, want true: the objective is numbered 1")
	}
}

func TestMatching_anEmptyNamedThingIsNeverFound(t *testing.T) {
	if Matching(plan(t), []string{""}) {
		t.Error("Matching = true, want false: an empty name is never found")
	}
}

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

func TestSaid_isWhatASearchLooksIn(t *testing.T) {
	got := carrying(t, func(a *AsTold) {
		a.Standing(OnHold)
		a.Carries(domain.High)
		a.Labelled("urgent")
		a.LabelsTold()
	})
	said := Said(got)
	for _, want := range []string{"on-hold", "high", "urgent"} {
		if !slices.Contains(said, want) {
			t.Errorf("Said = %v, want %q among it", said, want)
		}
	}
}
