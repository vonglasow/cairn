package engine

import (
	"cairn/internal/domain"
	"cairn/internal/task"
	"cairn/internal/trace"
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

func TestFocus_bringsWaitingWorkIntoCurrentAndLeavesTheRest(t *testing.T) {
	for state, lands := range map[string]string{
		"backlog": "current", "todo": "current", "paused": "current",
		"blocked": "blocked", "watching": "watching", "current": "current",
	} {
		path := storeAt(t)
		if _, err := Do(TaskNew, path, given("title", "Already here", "state", "current")); err != nil {
			t.Fatal(err)
		}
		if _, err := Do(TaskNew, path, given("title", "Write the spec", "state", state)); err != nil {
			t.Fatal(err)
		}
		if _, err := Do(TaskFocus, path, given("number", "2")); err != nil {
			t.Fatalf("focusing a task in %s: %v", state, err)
		}
		shown, err := Do(TaskShow, path, given("number", "2"))
		if err != nil {
			t.Fatal(err)
		}
		if got := string(factsOf(*shown.Task).State); got != lands {
			t.Errorf("a task in %s that takes the focus is in %s, want %s", state, got, lands)
		}
		if lands != "current" || state == "current" {
			continue
		}
		listed, err := Do(TaskList, path, given())
		if err != nil {
			t.Fatal(err)
		}
		if len(listed.Tasks) == 0 {
			t.Fatal("nothing is listed")
		}
		if first := factsOf(listed.Tasks[0]).Number; first != 2 {
			t.Errorf("a task in %s brought into current sits behind task %d, want at the head", state, first)
		}
	}
}

func TestFocus_isLetGoOfByATaskMovedIntoBlocked(t *testing.T) {
	path := storeAt(t, "one", "two")
	for _, n := range []string{"1", "2"} {
		if _, err := Do(TaskFocus, path, given("number", n)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Do(TaskEdit, path, given("number", "1", "state", "blocked", "because", "an answer")); err != nil {
		t.Fatal(err)
	}
	if held := heldIn(t, path); len(held) != 1 || held[0] != 2 {
		t.Errorf("in focus after one was blocked: %v, want [2]", held)
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

func numbersIn(tasks []task.Task, state domain.State) []int {
	var out []int
	for _, t := range tasks {
		if factsOf(t).State == state {
			out = append(out, factsOf(t).Number)
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
	if a.InFocus == nil || !slices.Equal(*a.InFocus, []int{1, 3}) || len(a.Held) != 2 || factsOf(a.Held[1]).Number != 3 {
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
	if got := actsDone(t, e, trace.FocusReleased); !slices.Equal(got, []int{2, 1, 3}) {
		t.Errorf("focus released on %v, want 2 then 1 and 3", got)
	}
}

func TestTaskFocus_answersWithTheTaskBroughtIntoCurrentAndTheFocus(t *testing.T) {
	e := openEngine(t)
	add(t, e, "one")
	add(t, e, "two")
	a, err := e.Ask(TaskFocus, given("number", "2"))
	if err != nil {
		t.Fatal(err)
	}
	if a.Task == nil || factsOf(*a.Task).Number != 2 || factsOf(*a.Task).State != domain.Current {
		t.Fatalf("answered with task %+v, want task 2 in current", a.Task)
	}
	if a.InFocus == nil || !slices.Equal(*a.InFocus, []int{2}) || len(a.Sessions) != 1 {
		t.Errorf("in focus %v, sessions %d", a.InFocus, len(a.Sessions))
	}
}

func TestMove_recordsAMoveOnlyWhenTheTaskChangedPlace(t *testing.T) {
	e := openEngine(t)
	add(t, e, "one")
	add(t, e, "two")
	for range 2 {
		if err := e.move(2, First, 0); err != nil {
			t.Fatal(err)
		}
	}
	if got := actsDone(t, e, trace.Moved); !slices.Equal(got, []int{2}) {
		t.Errorf("moves recorded on %v, want one on task 2", got)
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
	if got := numbersIn(a.Tasks, domain.Backlog); !slices.Equal(got, []int{2, 1}) {
		t.Errorf("backlog is %v, want 2 then 1", got)
	}
	if _, err := e.Ask(TaskMove, given("number", "1")); !errors.Is(err, ErrNowhereToMove) {
		t.Errorf("moving nowhere answered %v", err)
	}
	if _, err := e.Ask(TaskMove, given("number", "1", "first", "true", "last", "true", "state", "todo")); !errors.Is(err, ErrMoreThanOnePlace) {
		t.Errorf("moving to two places answered %v", err)
	}
}

func TestTaskMove_takesTheOthersAlongBehindItIntoItsColumn(t *testing.T) {
	e := openEngine(t)
	for _, title := range []string{"one", "two", "three", "four"} {
		add(t, e, title)
	}
	in := given("number", "3", "state", "todo", "first", "true")
	in.Listed = map[string][]string{"also": {"1", "2"}}
	a, err := e.Ask(TaskMove, in)
	if err != nil {
		t.Fatal(err)
	}
	if got := numbersIn(a.Tasks, domain.Todo); !slices.Equal(got, []int{3, 1, 2}) {
		t.Errorf("todo is %v, want 3, 1, 2", got)
	}
}
