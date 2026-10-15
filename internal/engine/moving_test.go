package engine

import (
	"errors"
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
