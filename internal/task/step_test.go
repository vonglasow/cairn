package task

import (
	"errors"
	"strings"
	"testing"

	"cairn/internal/domain"
)

func withSteps(state domain.State, steps ...Step) Task {
	var a AsTold
	a.Numbered(1)
	a.Titled("Ship it")
	a.InColumn(state)
	for _, s := range steps {
		a.Step(s.Title, s.Done)
	}
	return a.Task()
}

func TestSteps_areCarriedInTheOrderTheyWereGiven(t *testing.T) {
	got := withSteps(domain.Todo, Step{Title: "Write the spec"}, Step{Title: "Plan it", Done: true}).Carried().Steps
	want := []Step{{Title: "Write the spec"}, {Title: "Plan it", Done: true}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("steps = %v, want %v", got, want)
	}
}

func TestMayFinishAs_refusesAStepStillOpenAndNamesItsPlace(t *testing.T) {
	was := withSteps(domain.Todo, Step{Title: "Plan it", Done: true}, Step{Title: "Write the spec"})
	now := withSteps(domain.Done, Step{Title: "Plan it", Done: true}, Step{Title: "Write the spec"})
	err := was.MayFinishAs(now)
	if !errors.Is(err, ErrStepStillOpen) {
		t.Fatalf("finishing with a step open = %v, want ErrStepStillOpen", err)
	}
	if !strings.Contains(err.Error(), "step 2") {
		t.Errorf("the refusal %q does not name step 2", err)
	}
}

func TestMayFinishAs_letsATaskGoOnceEveryStepIsDone(t *testing.T) {
	was := withSteps(domain.Todo, Step{Title: "Plan it", Done: true})
	if err := was.MayFinishAs(withSteps(domain.Done, Step{Title: "Plan it", Done: true})); err != nil {
		t.Fatalf("finishing with every step done = %v", err)
	}
	if err := withSteps(domain.Todo).MayFinishAs(withSteps(domain.Done)); err != nil {
		t.Fatalf("finishing with no step = %v", err)
	}
}

func TestMayFinishAs_asksNothingOfATaskNotBeingFinished(t *testing.T) {
	open := Step{Title: "Write the spec"}
	if err := withSteps(domain.Todo, open).MayFinishAs(withSteps(domain.Blocked, open)); err != nil {
		t.Fatalf("stopping with a step open = %v", err)
	}
	if err := withSteps(domain.Done, open).MayFinishAs(withSteps(domain.Done, open)); err != nil {
		t.Fatalf("a task already done = %v", err)
	}
}

func TestTicked_isTheTaskWithEveryStepDone(t *testing.T) {
	got := withSteps(domain.Todo, Step{Title: "a"}, Step{Title: "b", Done: true}).Ticked().Carried().Steps
	for _, s := range got {
		if !s.Done {
			t.Fatalf("step %q is not done after ticking: %v", s.Title, got)
		}
	}
}

func TestSaid_reachesWhatAStepIsCalled(t *testing.T) {
	if !Matching(withSteps(domain.Todo, Step{Title: "Read it first"}), []string{"first"}) {
		t.Fatal("a word written only in a step does not reach its task")
	}
}

func TestStepsText_namesEachStepAndWhetherItIsDone(t *testing.T) {
	got := StepsText([]Step{{Title: "Write the spec"}, {Title: "Plan it", Done: true}})
	for _, want := range []string{"steps", "1  Write the spec", DoneMark + "  2  Plan it"} {
		if !strings.Contains(got, want) {
			t.Errorf("%q does not say %q", got, want)
		}
	}
	if StepsText(nil) != "" {
		t.Error("a task with no steps writes a heading for none")
	}
}

func TestNewStep_refusesAStepWithNoTitle(t *testing.T) {
	if _, err := NewStep("   ", false); err == nil {
		t.Fatal("a step with no title was taken")
	}
	got, err := NewStep("  Write it  ", true)
	if err != nil || got != (Step{Title: "Write it", Done: true}) {
		t.Fatalf("NewStep = %v, %v", got, err)
	}
}
