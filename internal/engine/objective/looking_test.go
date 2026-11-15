package objective

import (
	"cairn/internal/engine/calendar"
	"testing"
)

func carrying(t *testing.T, tell func(*AsTold)) Objective {
	t.Helper()
	built := plan(t)
	a := From(built)
	tell(&a)
	return a.Objective()
}

func TestMatching_reachesTheDeadlineTheWayItIsShown(t *testing.T) {
	deadline, err := calendar.ParseDeadline("2026-09-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	got := carrying(t, func(a *AsTold) { a.Due(deadline) })
	if !Matching(got, []string{"2026-09"}) {
		t.Error("Matching = false, want true: the deadline reads 2026-09-01")
	}
}
