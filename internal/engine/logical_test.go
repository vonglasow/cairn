package engine

import (
	"cairn/internal/engine/task"
	"errors"
	"testing"
	"time"
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

func TestDaysSince_countsTheDaysOnTheCalendar(t *testing.T) {
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.Local)
	for _, c := range []struct {
		since time.Time
		want  string
	}{
		{time.Date(2026, 10, 5, 0, 30, 0, 0, time.Local), "today"},
		{time.Date(2026, 10, 4, 23, 30, 0, 0, time.Local), "1 day"},
		{time.Date(2026, 10, 3, 9, 30, 0, 0, time.Local), "2 days"},
		{time.Date(2026, 9, 26, 12, 0, 0, 0, time.Local), "9 days"},
		{time.Date(2026, 3, 20, 12, 0, 0, 0, time.Local), "199 days"},
	} {
		if got := daysSince(c.since, now); got != c.want {
			t.Errorf("daysSince(%v) = %q, want %q", c.since, got, c.want)
		}
	}
}
