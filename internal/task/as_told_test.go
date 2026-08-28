package task

import (
	"testing"
	"time"

	"cairn/internal/domain"
)

// TestAsTold_buildsATaskFromEveryFactItIsTold checks that a task told to a
// builder, fact by fact, is handed back whole: this is the same membrane
// Tell uses, the other way round, so what goes in has to come out unchanged.
func TestAsTold_buildsATaskFromEveryFactItIsTold(t *testing.T) {
	deadline, err := domain.ParseDeadline("2026-01-02")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}

	want := Task{
		number: 3, parent: 1, objectives: []int{2}, title: "Write the spec",
		description: "Say what it does", state: domain.Current, priority: domain.High,
		labels: []string{"backend", "urgent"}, deadline: deadline,
		estimate: 90 * time.Minute, timeSpent: 30 * time.Minute,
	}

	a := &AsTold{}
	want.Tell(a)
	got := a.Task()

	if got.number != want.number {
		t.Errorf("Number = %d, want %d", got.number, want.number)
	}
	if got.parent != want.parent {
		t.Errorf("Parent = %d, want %d", got.parent, want.parent)
	}
	if len(got.objectives) != len(want.objectives) || got.objectives[0] != want.objectives[0] {
		t.Errorf("Objectives = %v, want %v", got.objectives, want.objectives)
	}
	if got.title != want.title {
		t.Errorf("Title = %q, want %q", got.title, want.title)
	}
	if got.description != want.description {
		t.Errorf("Description = %q, want %q", got.description, want.description)
	}
	if got.state != want.state {
		t.Errorf("State = %q, want %q", got.state, want.state)
	}
	if got.priority != want.priority {
		t.Errorf("Priority = %q, want %q", got.priority, want.priority)
	}
	if len(got.labels) != len(want.labels) {
		t.Fatalf("Labels = %v, want %v", got.labels, want.labels)
	}
	for i := range want.labels {
		if got.labels[i] != want.labels[i] {
			t.Errorf("Labels[%d] = %q, want %q", i, got.labels[i], want.labels[i])
		}
	}
	if !got.deadline.Equal(want.deadline) {
		t.Errorf("Deadline = %v, want %v", got.deadline, want.deadline)
	}
	if got.estimate != want.estimate {
		t.Errorf("Estimate = %v, want %v", got.estimate, want.estimate)
	}
	if got.timeSpent != want.timeSpent {
		t.Errorf("TimeSpent = %v, want %v", got.timeSpent, want.timeSpent)
	}
}

// The seven tests below each poison the field first, through the fact that
// carries a value, before telling AsTold the fact says there is none. A
// builder that dropped the absence on the floor would still hold the
// poisoned value afterwards.

func TestAsTold_notDescribedMeansNoDescription(t *testing.T) {
	a := &AsTold{}
	a.Described("Say what it does")
	a.NotDescribed()
	if got := a.Task().description; got != "" {
		t.Errorf("Description = %q, want none", got)
	}
}

func TestAsTold_notDueMeansNoDeadline(t *testing.T) {
	a := &AsTold{}
	a.Due(time.Now())
	a.NotDue()
	if got := a.Task().deadline; !got.IsZero() {
		t.Errorf("Deadline = %v, want none", got)
	}
}

func TestAsTold_notEstimatedMeansNoEstimate(t *testing.T) {
	a := &AsTold{}
	a.Estimated(90 * time.Minute)
	a.NotEstimated()
	if got := a.Task().estimate; got != 0 {
		t.Errorf("Estimate = %v, want none", got)
	}
}

func TestAsTold_nothingSpentMeansNoTimeSpent(t *testing.T) {
	a := &AsTold{}
	a.Spent(30 * time.Minute)
	a.NothingSpent()
	if got := a.Task().timeSpent; got != 0 {
		t.Errorf("TimeSpent = %v, want none", got)
	}
}

func TestAsTold_underNothingMeansNoParent(t *testing.T) {
	a := &AsTold{}
	a.Under(1)
	a.UnderNothing()
	if got := a.Task().parent; got != NoParent {
		t.Errorf("Parent = %d, want NoParent", got)
	}
}

// ServesNothing empties the set, the way NotLabelled and NotBlocked do — and
// like them it is heard before the set is closed, so what was gathered before
// it goes with it.
func TestAsTold_servesNothingEmptiesTheSet(t *testing.T) {
	a := &AsTold{}
	a.Serves(2)
	a.ServesNothing()
	a.ObjectivesTold()
	if got := a.Task().objectives; len(got) != 0 {
		t.Errorf("Objectives = %v, want none", got)
	}
}

func TestAsTold_notLabelledMeansNoLabels(t *testing.T) {
	a := &AsTold{}
	a.Labelled("backend")
	a.NotLabelled()
	a.LabelsTold()
	if got := a.Task().labels; len(got) != 0 {
		t.Errorf("Labels = %v, want none", got)
	}
}
