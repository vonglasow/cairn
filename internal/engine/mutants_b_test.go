package engine

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

func importInto(t *testing.T, e *Engine, file string) (domain.Held, error) {
	t.Helper()
	from := filepath.Join(t.TempDir(), "held.json")
	if err := os.WriteFile(from, []byte(file), 0o600); err != nil {
		t.Fatal(err)
	}
	read, err := readExport(from)
	if err != nil {
		return domain.Held{}, err
	}
	return e.importing(read)
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

func TestReadExport_refusesAFileOfAnotherKind(t *testing.T) {
	_, err := importInto(t, openEngine(t), `{"cairn": "notes", "version": 4, "tasks": []}`)
	if !errors.Is(err, ErrNotAnExport) {
		t.Errorf("a file of notes answered %v", err)
	}
}

func TestTaskImport_filesEveryLabelPastOneHeldOrOneWithoutColour(t *testing.T) {
	e := openEngine(t)
	if _, err := e.Ask(LabelNew, given("label", "held")); err != nil {
		t.Fatal(err)
	}
	if _, err := importInto(t, e, `{"cairn": "tasks", "version": 4, "tasks": [],
		"labels": [{"label": "held"}, {"label": "idle"}, {"label": "release", "colour": "teal"}]}`); err != nil {
		t.Fatal(err)
	}
	labels, err := e.labelsCarried()
	if err != nil {
		t.Fatal(err)
	}
	var said []string
	for _, l := range labels {
		if l.Label == "release" {
			said = append(said, l.Label+":"+l.Colour)
			continue
		}
		said = append(said, l.Label)
	}
	slices.Sort(said)
	if want := []string{"held", "idle", "release:teal"}; !slices.Equal(said, want) {
		t.Errorf("labels %v, want %v", said, want)
	}
}

func TestTaskImport_countsWhatItAddedAndKeepsWhatEachOneSaid(t *testing.T) {
	e := openEngine(t)
	held, err := importInto(t, e, `{"cairn": "tasks", "version": 4,
		"objectives": [{"number": 1, "title": "Release", "state": "active", "priority": "normal", "labels": ["release"], "estimate": "40h"}],
		"tasks": [{"number": 1, "title": "Ship", "state": "todo", "priority": "normal", "focused": true, "objectives": [1]},
		  {"number": 2, "title": "Watch", "state": "watching", "priority": "normal"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if held != (domain.Held{Tasks: 2, Objectives: 1}) {
		t.Errorf("added %+v, want 2 tasks and 1 objective", held)
	}
	o, err := e.showObjective(1)
	if err != nil {
		t.Fatal(err)
	}
	if shape := objective.OfObjective(o); !slices.Equal(shape.Labels, []string{"release"}) || shape.Estimate != "40h" {
		t.Errorf("objective labels %v, estimate %q", shape.Labels, shape.Estimate)
	}
	watched, err := e.show(2)
	if err != nil {
		t.Fatal(err)
	}
	if factsOf(watched).State != domain.Watching {
		t.Errorf("task 2 is in %s, want watching", factsOf(watched).State)
	}
	if held, _ := e.focus(); !slices.Equal(held, []int{1}) {
		t.Errorf("in focus %v, want task 1", held)
	}
}

func TestTaskImport_refusesABadObjectiveOrTaskBeforeAddingAnything(t *testing.T) {
	const head = `{"cairn": "tasks", "version": 4, "labels": [{"label": "release"}],`
	for _, c := range []struct{ said, file string }{
		{"objective 1:", head + `"objectives": [{"number": 1, "title": "Release", "state": "nowhere", "priority": "normal"}], "tasks": []}`},
		{"task 1:", head + `"objectives": [], "tasks": [{"number": 1, "title": "Ship", "state": "todo", "priority": "pressing"}]}`},
		{"task 1:", head + `"objectives": [], "tasks": [{"number": 1, "title": "Ship", "state": "todo", "priority": "normal", "spent": {"someday": "1h"}}]}`},
	} {
		e := openEngine(t)
		_, err := importInto(t, e, c.file)
		if err == nil || !strings.HasPrefix(err.Error(), c.said) {
			t.Errorf("importing %s answered %v, want it to start with %q", c.file, err, c.said)
		}
		if labels, _ := e.labelsCarried(); len(labels) != 0 {
			t.Errorf("importing %s filed labels %v before refusing", c.file, labels)
		}
	}
}

func TestTaskExport_writesTimeAddedOnADayPastUnderThatDay(t *testing.T) {
	e := openEngine(t)
	add(t, e, "one")
	day := domain.FormatDay(time.Now().AddDate(0, 0, -10))
	if _, err := e.Ask(TimesheetAdd, given("number", "1", "time", "2h", "on", day)); err != nil {
		t.Fatal(err)
	}
	out, err := e.exporting()
	if err != nil {
		t.Fatal(err)
	}
	if got := out.Tasks[0].Spent; got[day] != "2h" || len(got) != 1 {
		t.Errorf("spent %v, want 2h on %s", got, day)
	}
}

func TestObjectiveReport_reviewsTheTimeAndTheTasksDoneInThePeriod(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Release")
	serving(t, e, "Ship", 1)
	today := domain.FormatDay(time.Now())
	if _, err := e.Ask(TimesheetAdd, given("number", "1", "time", "1h", "on", today)); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Ask(TaskEdit, given("number", "1", "state", "done")); err != nil {
		t.Fatal(err)
	}
	a, err := e.Ask(ObjectiveReport, given("number", "1", "since", domain.FormatDay(time.Now().AddDate(0, 0, -7)), "until", today))
	if err != nil {
		t.Fatal(err)
	}
	if a.Period == nil || a.Period.Spent() != time.Hour || len(a.Period.Done) != 1 || a.Period.Done[0].Number != 1 {
		t.Errorf("period %+v, want 1h and task 1 done", a.Period)
	}
}

func TestMonthsOf_countsADayPastOneWhereNothingServedWasDone(t *testing.T) {
	first := time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
	sheet := trace.Sheet{Days: []trace.Day{
		{Day: first, Tasks: []trace.Worked{{Task: 2, Took: time.Hour}}},
		{Day: first.AddDate(0, 0, 1), Tasks: []trace.Worked{{Task: 1, Took: 30 * time.Minute}}},
	}}
	got := monthsOf(sheet, map[int]bool{1: true})
	if len(got) != 1 || !got[0].Month.Equal(first) || got[0].Spent != 30*time.Minute {
		t.Errorf("months %+v, want 30m in October", got)
	}
}

func TestFinishedIn_readsDoneOnlyOffAColumnMoved(t *testing.T) {
	e := openEngine(t)
	done, err := e.add("Ship", Changes{State: of("done")})
	if err != nil {
		t.Fatal(err)
	}
	first := time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
	renamed := []trace.Act{act("changed", "title", "Ship", "done", "", first.AddDate(0, 0, 2))}
	if got := finishedIn(renamed, []task.Task{done}, first, first.AddDate(0, 0, 9)); len(got) != 0 {
		t.Errorf("a title that became done was read as the task finished: %+v", got)
	}
}

func TestFinishedIn_leavesOutATaskPutInDoneBeforeThePeriod(t *testing.T) {
	e := openEngine(t)
	done, err := e.add("Ship", Changes{State: of("done")})
	if err != nil {
		t.Fatal(err)
	}
	first := time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
	last := first.AddDate(0, 0, 9)
	before := []trace.Act{act("changed", "column", "todo", "done", "", first.AddDate(0, 0, -1))}
	if got := finishedIn(before, []task.Task{done}, first, last); len(got) != 0 {
		t.Errorf("done before the period, found %+v", got)
	}
	within := []trace.Act{act("changed", "column", "todo", "done", "", first.AddDate(0, 0, 2))}
	if got := finishedIn(within, []task.Task{done}, first, last); len(got) != 1 || got[0].Number != 1 {
		t.Errorf("done in the period, found %+v", got)
	}
}

func TestTaskImport_keepsALinkMadeFromTheOtherSide(t *testing.T) {
	e := openEngine(t)
	if _, err := importInto(t, e, `{"cairn": "tasks", "version": 4, "objectives": [],
		"tasks": [{"number": 2, "title": "Two", "state": "backlog", "priority": "normal", "blocked_by": [], "linked": []},
		  {"number": 1, "title": "One", "state": "backlog", "priority": "normal", "linked": [2]}]}`); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{1, 2} {
		got, err := e.show(n)
		if err != nil {
			t.Fatal(err)
		}
		if linked := factsOf(got).Linked; len(linked) != 1 {
			t.Errorf("task %d is linked to %v, want the other one", n, linked)
		}
	}
}

func TestSpentByDay_reachesBackToTheDayTimeWasAddedFor(t *testing.T) {
	e := openEngine(t)
	day := domain.FormatDay(time.Now().AddDate(0, 0, -10))
	if err := e.record(1, trace.TimeAdded, "", day, "2h"); err != nil {
		t.Fatal(err)
	}
	spent, err := e.spentByDay()
	if err != nil {
		t.Fatal(err)
	}
	if got := spent[1]; got[day] != "2h" || len(got) != 1 {
		t.Errorf("spent %v, want 2h on %s", got, day)
	}
}

func TestTaskImport_keepsWhatBlocksATask(t *testing.T) {
	e := openEngine(t)
	if _, err := importInto(t, e, `{"cairn": "tasks", "version": 4, "objectives": [],
		"tasks": [{"number": 1, "title": "One", "state": "todo", "priority": "normal"},
		  {"number": 2, "title": "Two", "state": "blocked", "priority": "normal", "blocked_by": [1]}]}`); err != nil {
		t.Fatal(err)
	}
	got, err := e.show(2)
	if err != nil {
		t.Fatal(err)
	}
	if f := factsOf(got); !slices.Equal(f.Blockers, []int{1}) || f.State != domain.Blocked {
		t.Errorf("task 2 is in %s blocked by %v, want blocked by task 1", f.State, f.Blockers)
	}
}
