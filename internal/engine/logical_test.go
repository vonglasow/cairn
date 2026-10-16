package engine

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
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
	if !factsOf(*shown.Task).Archived {
		t.Error("the task after one already off the board was left on it")
	}
}

func TestArchive_aNumberNamedAsNothingNamesNoTask(t *testing.T) {
	if _, err := Do(TaskArchive, storeAt(t, "one"), given("number", "")); !errors.Is(err, ErrNothingToArchive) {
		t.Errorf("err = %v, want ErrNothingToArchive", err)
	}
}

func TestStoreProblem_passesOnAStoreOfTheWrongAgeAsItIs(t *testing.T) {
	for _, kind := range []error{ErrStoreBehind, ErrNewerStore} {
		if err := storeProblem(fmt.Errorf("opening: %w", kind)); errors.Is(err, ErrStore) {
			t.Errorf("%v was said as a store that cannot be read: %v", kind, err)
		}
	}
}

func TestCheck_readsPastAnArgumentLeftOut(t *testing.T) {
	if err := TaskArchive.Check(given("due-before", "soon")); !errors.Is(err, domain.ErrNotADay) {
		t.Errorf("Check = %v, want the day refused", err)
	}
}

func TestCreation_writesDownWhatComesAfterAnEmptyAttribute(t *testing.T) {
	path := storeAt(t, "Write the spec")
	acts, err := Do(TraceShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range acts.Acts {
		f := a.Carried()
		if f.Field == "column" {
			return
		}
	}
	t.Error("the column the task started in was not written down")
}

func TestLocate_refusesAVariableNamingNothing(t *testing.T) {
	h := Home{Lookup: func(string) (string, bool) { return "", true }, Missing: errors.New("no home")}
	if _, err := Locate(h); !errors.Is(err, ErrStore) {
		t.Errorf("Locate = %v, want ErrStore", err)
	}
}

func TestRead_readsPastAYes(t *testing.T) {
	in, err := TaskMove.Read([]string{"--first", "4"})
	if err != nil || in.Given["number"] != "4" || in.Given["first"] != "true" {
		t.Errorf("Read = %+v, %v", in, err)
	}
}

func TestFromJSON_readsPastEveryKindOfField(t *testing.T) {
	in, err := TaskNew.FromJSON([]byte(`{"desc": null, "label": ["spec"], "title": "Write the spec"}`))
	if err != nil || in.Given["title"] != "Write the spec" {
		t.Errorf("FromJSON = %+v, %v; want the title read past the list and the null", in, err)
	}
	moved, err := TaskMove.FromJSON([]byte(`{"first": true, "number": 4}`))
	if err != nil || moved.Given["number"] != "4" {
		t.Errorf("FromJSON = %+v, %v; want the number read past the yes", moved, err)
	}
}

func TestListing_countsTheStepsOfWhatItNarrowedTo(t *testing.T) {
	path := storeAt(t, "Ship it")
	if _, err := Do(TaskEdit, path, Input{Given: map[string]string{"number": "1"}, Listed: map[string][]string{"label": {"x"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(StepNew, path, given("number", "1", "title", "a step")); err != nil {
		t.Fatal(err)
	}
	listed, err := Do(TaskList, path, Input{Listed: map[string][]string{"label": {"x"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := listed.Along[1].Steps.Total; got != 1 {
		t.Errorf("task 1 counts %d steps, want its one", got)
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

func TestDeletion_writesDownWhatComesAfterAnEmptyAttribute(t *testing.T) {
	path := storeAt(t, "Write the spec")
	if _, err := Do(TaskDelete, path, confirmed(given("number", "1"))); err != nil {
		t.Fatal(err)
	}
	acts, err := Do(TraceShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	last := acts.Acts[len(acts.Acts)-1]
	if f := last.Carried(); !strings.Contains(f.Was, "column: backlog") {
		t.Errorf("the deletion holds %q, want the column past the empty description", f.Was)
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

func TestStoodText_saysSinceWhenOnlyWhenItKnows(t *testing.T) {
	task := add(t, openEngine(t), "Write the spec")
	if got := stoodText(task, time.Time{}); got != "" {
		t.Errorf("stoodText with no moment = %q, want nothing", got)
	}
	if got := stoodText(task, time.Now()); !strings.Contains(got, "in backlog since") || !strings.HasSuffix(got, "today\n") {
		t.Errorf("stoodText = %q", got)
	}
}

func TestStoodText_saysTheDayOnTheLocalCalendar(t *testing.T) {
	kept := time.Local
	time.Local = time.FixedZone("UTC+2", 2*60*60)
	t.Cleanup(func() { time.Local = kept })
	task := add(t, openEngine(t), "Write the spec")
	justAfterMidnight := time.Date(2026, 10, 11, 22, 30, 0, 0, time.UTC)
	if got := stoodText(task, justAfterMidnight); !strings.Contains(got, "since 2026-10-12") {
		t.Errorf("stoodText = %q, want the local day, 2026-10-12", got)
	}
}
