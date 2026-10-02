package objective

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
)

func plan(t *testing.T) Objective {
	t.Helper()
	built, err := NewObjective("Ship the release")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	var a AsTold
	built.Tell(&a)
	a.Numbered(1)
	return a.Objective()
}

func TestText_leadsWithTheNumberAndTheTitle(t *testing.T) {
	got := Text(plan(t), domain.Progress{})
	if first := strings.SplitN(got, "\n", 2)[0]; !strings.Contains(first, "1") || !strings.Contains(first, "Ship the release") {
		t.Errorf("first line = %q, want the number and the title", first)
	}
}

func TestText_namesEveryAttributeEvenTheEmptyOnes(t *testing.T) {
	got := Text(plan(t), domain.Progress{})
	for _, key := range []string{"state", "priority", "description", "labels", "deadline", "estimate"} {
		if !strings.Contains(got, key) {
			t.Errorf("text is missing %q:\n%s", key, got)
		}
	}
}

func TestText_showsTheDefaultsAnObjectiveStartsWith(t *testing.T) {
	got := Text(plan(t), domain.Progress{})
	if !strings.Contains(got, "active") || !strings.Contains(got, "normal") {
		t.Errorf("text should show active and normal:\n%s", got)
	}
}

func TestText_showsWhatTheObjectiveCarries(t *testing.T) {
	built := plan(t)
	deadline, err := domain.ParseDeadline("2026-10-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	estimate, err := domain.ParseEstimate("5h")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	var a AsTold
	built.Tell(&a)
	a.Described("the whole thing")
	a.Labelled("launch")
	a.Labelled("team")
	a.LabelsTold()
	a.Due(deadline)
	a.Estimated(estimate)
	built = a.Objective()

	got := Text(built, domain.Progress{})
	for _, want := range []string{"the whole thing", "launch, team", "2026-10-01", "5h"} {
		if !strings.Contains(got, want) {
			t.Errorf("text is missing %q:\n%s", want, got)
		}
	}
}

func TestText_showsHowFarAlongItIs(t *testing.T) {
	got := Text(plan(t), domain.Progress{Tasks: domain.Tally{Total: 3, Done: 1}})
	if !strings.Contains(got, "1 of 3 tasks are done") {
		t.Errorf("text is missing the progress line:\n%s", got)
	}
}

// TestText_countsSubTasksApartFromTasks pins that the sub-tasks tally is
// only said when there are sub-tasks to say it about — the mirror of what
// TestText_showsHowFarAlongItIs pins for tasks.
func TestText_countsSubTasksApartFromTasks(t *testing.T) {
	got := Text(plan(t), domain.Progress{SubTasks: domain.Tally{Total: 4, Done: 2}})
	if !strings.Contains(got, "2 of 4 sub-tasks are done") {
		t.Errorf("text is missing the progress line:\n%s", got)
	}
	if strings.Contains(got, "0 of 0 tasks") {
		t.Errorf("text should not count tasks it does not have:\n%s", got)
	}
}

func TestList_writesEveryObjectiveAsOneRow(t *testing.T) {
	first, err := NewObjective("First")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	second, err := NewObjective("Second")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	var a, b AsTold
	first.Tell(&a)
	a.Numbered(1)
	second.Tell(&b)
	b.Numbered(2)

	got := List([]Objective{a.Objective(), b.Objective()})
	if !strings.Contains(got, "#1  First") || !strings.Contains(got, "#2  Second") {
		t.Errorf("List is missing a row:\n%s", got)
	}
}

func TestServing_saysNothingIsUnderItForNoLines(t *testing.T) {
	if got := Serving(nil); !strings.Contains(got, "nothing is under it") {
		t.Errorf("Serving(nil) = %q, want it to say nothing is under it", got)
	}
}

func TestServing_listsTheLinesItIsGiven(t *testing.T) {
	got := Serving([]string{"#3  Write the tests"})
	if !strings.Contains(got, "tasks") || !strings.Contains(got, "#3  Write the tests") {
		t.Errorf("Serving is missing the line:\n%s", got)
	}
}

// A listing of objectives said only a number and a title: whether one was
// reached, abandoned or still running was not in it.
func TestAListingSaysWhereAnObjectiveStands(t *testing.T) {
	var a AsTold
	a.Numbered(3)
	a.Titled("Ship cairn")
	a.Standing(Reached)
	a.Carries(domain.High)
	a.Due(time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC))
	listed := List([]Objective{a.Objective()})
	for _, said := range []string{"Ship cairn", "reached", "high", "2026-12-01"} {
		if !strings.Contains(listed, said) {
			t.Errorf("the listing does not say %q: %q", said, listed)
		}
	}
}
