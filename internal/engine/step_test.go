package engine

import (
	"errors"
	"slices"
	"testing"

	"cairn/internal/engine/task"
)

func stepsOf(t *testing.T, path string, number string) []task.Step {
	t.Helper()
	answer, err := Do(stepList, path, given("number", number))
	if err != nil {
		t.Fatalf("step list: %v", err)
	}
	return answer.Task.Carried().Steps
}

func TestStepEdit_ticksRenamesAndUnticksOneStep(t *testing.T) {
	path := storeAt(t, "Ship it")
	if _, err := Do(StepNew, path, given("number", "1", "title", "Write the spec")); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(StepEdit, path, given("number", "1", "step", "1", "done", "true", "title", "Write it")); err != nil {
		t.Fatalf("step edit: %v", err)
	}
	if got := stepsOf(t, path, "1"); !slices.Equal(got, []task.Step{{Title: "Write it", Done: true}}) {
		t.Errorf("steps = %v", got)
	}
	if _, err := Do(StepEdit, path, given("number", "1", "step", "1", "not-done", "true")); err != nil {
		t.Fatalf("step edit: %v", err)
	}
	if got := stepsOf(t, path, "1"); got[0].Done {
		t.Errorf("the step is still done: %v", got)
	}
}

func TestStepEdit_refusesAPlaceTheTaskHasNoStepAt(t *testing.T) {
	path := storeAt(t, "Ship it")
	_, err := Do(StepEdit, path, given("number", "1", "step", "1", "done", "true"))
	if !errors.Is(err, task.ErrNoSuchStep) {
		t.Fatalf("step edit of nothing = %v, want ErrNoSuchStep", err)
	}
}

func TestStepEdit_refusesDoneAndNotDoneAtOnce(t *testing.T) {
	path := storeAt(t, "Ship it")
	if _, err := Do(StepNew, path, given("number", "1", "title", "Write the spec")); err != nil {
		t.Fatal(err)
	}
	_, err := Do(StepEdit, path, given("number", "1", "step", "1", "done", "true", "not-done", "true"))
	if !errors.Is(err, task.ErrDoneAndNotDone) {
		t.Fatalf("both at once = %v, want ErrDoneAndNotDone", err)
	}
}

func TestStepDelete_closesTheGapItLeaves(t *testing.T) {
	path := storeAt(t, "Ship it")
	for _, title := range []string{"Write the spec", "Write the tests"} {
		if _, err := Do(StepNew, path, given("number", "1", "title", title)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Do(StepDelete, path, given("number", "1", "step", "1")); err != nil {
		t.Fatalf("step delete: %v", err)
	}
	if got := stepsOf(t, path, "1"); !slices.Equal(got, []task.Step{{Title: "Write the tests"}}) {
		t.Errorf("steps = %v", got)
	}
}

func TestStepList_answersAMachineTheStepsWithTheirPlaces(t *testing.T) {
	path := storeAt(t, "Ship it")
	if _, err := Do(StepNew, path, given("number", "1", "title", "Write the spec", "done", "true")); err != nil {
		t.Fatal(err)
	}
	answer, err := Do(stepList, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	said, ok := Of(stepList, answer).(task.Steps)
	if !ok || said.Task != 1 || len(said.Steps) != 1 || said.Steps[0] != (task.StepShape{Place: 1, Title: "Write the spec", Done: true}) {
		t.Errorf("a machine reads %+v", Of(stepList, answer))
	}
}

func TestStepList_saysATaskHasNoSteps(t *testing.T) {
	path := storeAt(t, "Ship it")
	answer, err := Do(stepList, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if steps := (*answer.Task).Carried().Steps; len(steps) != 0 {
		t.Errorf("step list answers %v", steps)
	}
}
