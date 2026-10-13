package engine

// What a test kills only once && and || are swapped, or continue and break:
// each case here is a branch the rest of the suite reached and never told
// apart from its opposite.

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/trace"
)

// A sweep that reaches a task already off the board passes over it and goes
// on to the next, rather than stopping there.
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
	if !factsOf(*shown.Task).archived {
		t.Error("the task after one already off the board was left on it")
	}
}

// Named as nothing, the number names no task, and what is left to archive by
// is nothing at all.
func TestArchive_aNumberNamedAsNothingNamesNoTask(t *testing.T) {
	if _, err := Do(TaskArchive, storeAt(t, "one"), given("number", "")); !errors.Is(err, ErrNothingToArchive) {
		t.Errorf("err = %v, want ErrNothingToArchive", err)
	}
}

// A store from a newer cairn is said as itself, never as a store that cannot
// be read — the same as one that is behind.
func TestStoreProblem_passesOnAStoreOfTheWrongAgeAsItIs(t *testing.T) {
	for _, kind := range []error{ErrStoreBehind, ErrNewerStore} {
		if err := storeProblem(fmt.Errorf("opening: %w", kind)); errors.Is(err, ErrStore) {
			t.Errorf("%v was said as a store that cannot be read: %v", kind, err)
		}
	}
}

// Check reads every argument named, past the ones that were not.
func TestCheck_readsPastAnArgumentLeftOut(t *testing.T) {
	if err := TaskArchive.Check(given("due-before", "soon")); !errors.Is(err, domain.ErrNotADay) {
		t.Errorf("Check = %v, want the day refused", err)
	}
}

// A task's creation writes down every attribute it carries, past the ones it
// carries nothing for.
func TestCreation_writesDownWhatComesAfterAnEmptyAttribute(t *testing.T) {
	path := storeAt(t, "Write the spec")
	acts, err := Do(TraceShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range acts.Acts {
		var f actFacts
		a.Tell(&f)
		if f.field == "column" {
			return
		}
	}
	t.Error("the column the task started in was not written down")
}

// A variable that names the store as nothing names no store.
func TestLocate_refusesAVariableNamingNothing(t *testing.T) {
	h := Home{Lookup: func(string) (string, bool) { return "", true }, Missing: errors.New("no home")}
	if _, err := Locate(h); !errors.Is(err, ErrStore) {
		t.Errorf("Locate = %v, want ErrStore", err)
	}
}

// Every word typed is read, past a yes named on its own.
func TestRead_readsPastAYes(t *testing.T) {
	in, err := TaskMove.Read([]string{"--first", "4"})
	if err != nil || in.Given["number"] != "4" || in.Given["first"] != "true" {
		t.Errorf("Read = %+v, %v", in, err)
	}
}

// Every field sent is read, past a list, a null and a yes.
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

// A narrowed listing counts the parts of a task among every task there is,
// not only among the ones it kept.
func TestListing_countsPartsTheNarrowingLeftOut(t *testing.T) {
	path := storeAt(t, "Ship it")
	if _, err := Do(TaskEdit, path, Input{Given: map[string]string{"number": "1"}, Listed: map[string][]string{"label": {"x"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(TaskNew, path, given("title", "a part", "parent", "1")); err != nil {
		t.Fatal(err)
	}
	listed, err := Do(TaskList, path, Input{Listed: map[string][]string{"label": {"x"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := listed.Along[1].SubTasks.Total; got != 1 {
		t.Errorf("task 1 counts %d parts, want the one the narrowing left out", got)
	}
}

// A task deleted since keeps its time and loses its name, and the tasks after
// it keep theirs.
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

// actFacts reads what an act says about its field.
type actFacts struct{ field string }

func (f *actFacts) On(int)            {}
func (f *actFacts) Did(string)        {}
func (f *actFacts) Changing(s string) { f.field = s }
func (f *actFacts) NothingNamed()     {}
func (f *actFacts) Was(string)        {}
func (f *actFacts) WasNothing()       {}
func (f *actFacts) Became(string)     {}
func (f *actFacts) BecameNothing()    {}
func (f *actFacts) Through(string)    {}
func (f *actFacts) ByHand()           {}
func (f *actFacts) At(time.Time)      {}

var _ trace.Told = (*actFacts)(nil)

// A deletion writes down everything the task held, past what it held nothing
// of.
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
	var f deletedFacts
	last.Tell(&f)
	if !strings.Contains(f.was, "column: backlog") {
		t.Errorf("the deletion holds %q, want the column past the empty description", f.was)
	}
}

// deletedFacts reads what an act says a task was.
type deletedFacts struct {
	actFacts
	was string
}

func (d *deletedFacts) Was(s string) { d.was = s }

// Days are counted on the calendar: a moment late yesterday is a day ago, and
// one early this morning is today, however few hours separate them.
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

// A task asked about says since when it has stood where it is; one that is
// not says nothing of it.
func TestStoodText_saysSinceWhenOnlyWhenItKnows(t *testing.T) {
	task := add(t, openEngine(t), "Write the spec")
	if got := stoodText(task, time.Time{}); got != "" {
		t.Errorf("stoodText with no moment = %q, want nothing", got)
	}
	if got := stoodText(task, time.Now()); !strings.Contains(got, "in backlog since") || !strings.HasSuffix(got, "today\n") {
		t.Errorf("stoodText = %q", got)
	}
}

// The day is the one on the person's calendar, not the store's: a moment is
// kept in UTC, and read as it is, a task taken up just after midnight in Paris
// said it had stood there since the day before.
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
