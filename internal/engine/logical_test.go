package engine

import (
	"cairn/internal/engine/task"
	"errors"
	"testing"
)

func TestArchive_sweepGoesOnPastWhatIsAlreadyOff(t *testing.T) {
	path := storeAt(t, "first", "second")
	for _, n := range []string{"1", "2"} {
		if _, err := Do(TaskEdit, path, given("number", n, "state", "done")); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Do(TaskArchive, path, given("number", "1")); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(TaskArchive, path, confirmed(given("state", "done", "with-archived", "true"))); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	shown, err := Do(TaskShow, path, given("number", "2"))
	if err != nil {
		t.Fatal(err)
	}
	if !(*shown.Task).Carried().Archived {
		t.Error("the task after one already off the board was left on it")
	}
}

func TestArchive_aNumberNamedAsNothingNamesNoTask(t *testing.T) {
	if _, err := Do(TaskArchive, storeAt(t, "one"), given("number", "")); !errors.Is(err, task.ErrNothingToArchive) {
		t.Errorf("err = %v, want ErrNothingToArchive", err)
	}
}

func TestTimesheet_namesWhatComesAfterATaskDeletedSince(t *testing.T) {
	path := storeAt(t, "gone", "kept")
	for _, n := range []string{"1", "2"} {
		if _, err := Do(TimesheetAdd, path, given("number", n, "time", "1h")); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Do(TaskDelete, path, confirmed(given("number", "1"))); err != nil {
		t.Fatal(err)
	}
	answer, err := Do(TimesheetShow, path, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if answer.Named[2] != "kept" {
		t.Errorf("titles = %v, want the task after the deleted one named", answer.Named)
	}
}
