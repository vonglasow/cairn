package engine

import (
	"errors"
	"strings"
	"testing"
)

func heldIn(t *testing.T, path string) []int {
	t.Helper()
	e, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.Close() }()
	held, err := e.focus()
	if err != nil {
		t.Fatal(err)
	}
	return held
}

func TestFocus_isNeverHeldByATaskInDone(t *testing.T) {
	path := storeAt(t, "one", "two")
	for _, n := range []string{"1", "2"} {
		if _, err := Do(TaskFocus, path, given("number", n)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Do(TaskEdit, path, given("number", "1", "state", "done")); err != nil {
		t.Fatal(err)
	}
	if held := heldIn(t, path); len(held) != 1 || held[0] != 2 {
		t.Errorf("in focus after one was done: %v, want [2]", held)
	}
	if _, err := Do(TaskMove, path, given("number", "2", "state", "done", "first", "true")); err != nil {
		t.Fatal(err)
	}
	if held := heldIn(t, path); len(held) != 0 {
		t.Errorf("in focus after both were done: %v, want nothing", held)
	}
	if _, err := Do(TaskFocus, path, given("number", "1")); !errors.Is(err, ErrDoneHoldsNoFocus) {
		t.Errorf("focusing a task in done: %v, want %v", err, ErrDoneHoldsNoFocus)
	}
	if held := heldIn(t, path); len(held) != 0 {
		t.Errorf("in focus after the refusal: %v, want nothing", held)
	}
}

func TestDone_confirmedTicksEveryStep(t *testing.T) {
	path := storeAt(t, "ship it")
	for _, title := range []string{"write the spec", "write the docs"} {
		if _, err := Do(StepNew, path, given("number", "1", "title", title)); err != nil {
			t.Fatal(err)
		}
	}
	_, err := Do(TaskMove, path, given("number", "1", "state", "done", "first", "true"))
	if !errors.Is(err, ErrStepStillOpen) || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("unconfirmed: %v, want a refusal saying how to confirm", err)
	}
	if _, err := Do(TaskMove, path, Agreed(given("number", "1", "state", "done", "first", "true"))); err != nil {
		t.Fatalf("confirmed: %v", err)
	}
	shown, err := Do(TaskShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if state := factsOf(*shown.Task).State; state != "done" {
		t.Errorf("task 1 is %s, want done", state)
	}
	for _, step := range factsOf(*shown.Task).Steps {
		if !step.Done {
			t.Errorf("step %q is not done", step.Title)
		}
	}
	if !TaskMove.Finishes || !TaskEdit.Finishes {
		t.Error("a move or an edit into done does not say it may tick the task's steps")
	}
}

func TestMove_takesTheOthersAlongBehindIt(t *testing.T) {
	path := storeAt(t, "first", "second", "third", "fourth")
	in := given("number", "4", "state", "todo", "first", "true")
	in.Listed = map[string][]string{"also": {"2"}}
	if _, err := Do(TaskMove, path, in); err != nil {
		t.Fatal(err)
	}
	listed, err := Do(TaskList, path, given())
	if err != nil {
		t.Fatal(err)
	}
	var order []int
	for _, one := range listed.Tasks {
		if factsOf(one).State == "todo" {
			order = append(order, factsOf(one).Number)
		}
	}
	if len(order) != 2 || order[0] != 4 || order[1] != 2 {
		t.Errorf("todo holds %v, want [4 2]", order)
	}
}
