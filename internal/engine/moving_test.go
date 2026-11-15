package engine

import (
	"cairn/internal/engine/act"
	"cairn/internal/engine/item"
	"cairn/internal/engine/task"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
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
	if _, err := Do(TaskFocus, path, given("number", "1")); !errors.Is(err, task.ErrDoneHoldsNoFocus) {
		t.Errorf("focusing a task in done: %v, want %v", err, task.ErrDoneHoldsNoFocus)
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
	if state := (*shown.Task).Carried().State; state != "done" {
		t.Errorf("task 1 is %s, want done", state)
	}
	for _, step := range (*shown.Task).Carried().Steps {
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
		if one.Carried().State == "todo" {
			order = append(order, one.Carried().Number)
		}
	}
	if len(order) != 2 || order[0] != 4 || order[1] != 2 {
		t.Errorf("todo holds %v, want [4 2]", order)
	}
}

func actsDone(t *testing.T, e *Engine, did string) []int {
	t.Helper()
	acts, err := e.since(time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	var on []int
	for _, a := range acts {
		if c := a.Carried(); c.Did == did {
			on = append(on, c.Task)
		}
	}
	return on
}

func numbersIn(tasks []task.Task, state item.State) []int {
	var out []int
	for _, t := range tasks {
		if t.Carried().State == state {
			out = append(out, t.Carried().Number)
		}
	}
	return out
}

func TestTaskUnfocus_letsGoOfOneThenEveryTaskAndSaysWhatIsLeft(t *testing.T) {
	e := openEngine(t)
	for _, title := range []string{"one", "two", "three"} {
		add(t, e, title)
	}
	for _, n := range []string{"1", "2", "3"} {
		if _, err := e.Ask(TaskFocus, given("number", n)); err != nil {
			t.Fatal(err)
		}
	}
	a, err := e.Ask(TaskUnfocus, given("number", "2"))
	if err != nil {
		t.Fatal(err)
	}
	if a.InFocus == nil || !slices.Equal(*a.InFocus, []int{1, 3}) || len(a.Held) != 2 || a.Held[1].Carried().Number != 3 {
		t.Errorf("after unfocusing 2: in focus %v, held %d", a.InFocus, len(a.Held))
	}
	a, err = e.Ask(TaskFocus, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if a.InFocus == nil || !slices.Equal(*a.InFocus, []int{1, 3}) || len(a.Held) != 2 || len(a.Sessions) != 2 {
		t.Errorf("asking what is in focus: %v, held %d, sessions %d", a.InFocus, len(a.Held), len(a.Sessions))
	}
	a, err = e.Ask(TaskUnfocus, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if a.InFocus == nil || len(*a.InFocus) != 0 {
		t.Errorf("after unfocusing all: %v", a.InFocus)
	}
	if got := actsDone(t, e, act.FocusReleased); !slices.Equal(got, []int{2, 1, 3}) {
		t.Errorf("focus released on %v, want 2 then 1 and 3", got)
	}
}

func TestTaskMove_putsATaskInPlaceAndRefusesOneGivenNowhereOrTwice(t *testing.T) {
	e := openEngine(t)
	add(t, e, "one")
	add(t, e, "two")
	a, err := e.Ask(TaskMove, given("number", "2", "first", "true"))
	if err != nil {
		t.Fatal(err)
	}
	if got := numbersIn(a.Tasks, item.Backlog); !slices.Equal(got, []int{2, 1}) {
		t.Errorf("backlog is %v, want 2 then 1", got)
	}
	if _, err := e.Ask(TaskMove, given("number", "1")); !errors.Is(err, task.ErrNowhereToMove) {
		t.Errorf("moving nowhere answered %v", err)
	}
	if _, err := e.Ask(TaskMove, given("number", "1", "first", "true", "last", "true", "state", "todo")); !errors.Is(err, task.ErrMoreThanOnePlace) {
		t.Errorf("moving to two places answered %v", err)
	}
}
