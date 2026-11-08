package engine

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/config"
	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

func TestSessionC_asksTheMinuteItsLengthIsReached(t *testing.T) {
	taken := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	acts := []trace.Act{actAt(4, trace.FocusTaken, taken)}
	if got := sessionOf(4, acts, taken.Add(time.Hour), time.Hour); got.Ask == "" {
		t.Error("a session exactly as long as its length asks nothing")
	}
}

func TestSessionC_refusesALengthOfNothingOrLess(t *testing.T) {
	for _, said := range []string{"0s", "-1h"} {
		e := openEngine(t)
		e.home = homeLasting(t, "", said, 0)
		if _, err := e.sessionLength(); !errors.Is(err, config.ErrFocus) {
			t.Errorf("focus.session %q: err = %v, want config.ErrFocus", said, err)
		}
	}
}

func TestSessionC_sessionsAreTheOnesOfTheTasksInFocus(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	if _, err := e.Ask(TaskFocus, given("number", "1")); err != nil {
		t.Fatal(err)
	}
	got, err := e.sessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Task != 1 {
		t.Errorf("sessions = %+v, want the one of task 1", got)
	}
}

func TestBlockingC_aBlockerFurtherAlongTheChainMakesACycle(t *testing.T) {
	further := map[int][]int{2: {3}, 3: {1}}
	err := mayBlock(1, []int{2}, func(one int) ([]int, error) { return further[one], nil })
	if !errors.Is(err, ErrBlockingCycle) {
		t.Errorf("err = %v, want ErrBlockingCycle", err)
	}
}

func TestBlockingC_anEditClosingACycleIsRefused(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	add(t, e, "Ship it")
	if _, err := e.edit(2, Changes{Blocker: of([]string{"1"})}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.edit(1, Changes{Blocker: of([]string{"2"})}); !errors.Is(err, ErrBlockingCycle) {
		t.Errorf("err = %v, want ErrBlockingCycle", err)
	}
}

func TestBlockingC_aBlockedTaskLosingItsLastBlockerIsFreed(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	add(t, e, "Ship it")
	if _, err := e.edit(2, Changes{State: of("blocked"), Blocker: of([]string{"1"})}); err != nil {
		t.Fatal(err)
	}
	freed, err := e.edit(2, Changes{Blocker: of([]string{"-1"})})
	if err != nil {
		t.Fatal(err)
	}
	if f := factsOf(freed); f.Number != 2 || f.State != domain.Todo {
		t.Errorf("freed = task %d in %q, want task 2 in todo", f.Number, f.State)
	}
}

func TestArchiveC_aSweepCountsWhatItArchived(t *testing.T) {
	path := storeAt(t, "Write the spec", "Ship it", "Unrelated")
	for _, n := range []string{"1", "2"} {
		if _, err := Do(TaskEdit, path, given("number", n, "state", "done")); err != nil {
			t.Fatal(err)
		}
	}
	answer, err := Do(TaskArchive, path, confirmed(given("state", "done")))
	if err != nil {
		t.Fatal(err)
	}
	if answer.Archived == nil || *answer.Archived != 2 {
		t.Errorf("Archived = %v, want 2", answer.Archived)
	}
}

func TestLabelC_putsTheLabelOnEveryTaskReached(t *testing.T) {
	path := storeAt(t, "Write the spec", "Ship it")
	answer, err := Do(TaskLabel, path, given("put", "sprint", "state", "backlog"))
	if err != nil {
		t.Fatal(err)
	}
	if answer.Labelled == nil || *answer.Labelled != 2 {
		t.Errorf("Labelled = %v, want 2", answer.Labelled)
	}
	shown, err := Do(TaskShow, path, given("number", "2"))
	if err != nil {
		t.Fatal(err)
	}
	if labels := factsOf(*shown.Task).Labels; !slices.Contains(labels, "sprint") {
		t.Errorf("task 2 carries %v, want sprint among them", labels)
	}
}

func TestLabelC_aLabelThatIsNoTextStopsTheSweep(t *testing.T) {
	path := storeAt(t, "Write the spec")
	if _, err := Do(TaskLabel, path, given("put", "bell\x07", "state", "backlog")); err == nil {
		t.Error("a label that is no text was put")
	}
}

func TestDeleteC_aTaskInFocusIsReleasedThenDeleted(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	if _, err := e.Ask(TaskFocus, given("number", "1")); err != nil {
		t.Fatal(err)
	}
	if err := e.delete(1); err != nil {
		t.Fatal(err)
	}
	acts, err := e.history.ActsOn(1)
	if err != nil {
		t.Fatal(err)
	}
	var did []string
	for _, a := range acts {
		did = append(did, a.Carried().Did)
	}
	released, deleted := slices.Index(did, trace.FocusReleased), slices.Index(did, trace.Deleted)
	if released < 0 || deleted < released {
		t.Errorf("acts = %v, want the focus released then the task deleted", did)
	}
}

func TestObjectiveC_aTaskServingNoObjectiveThatExistsIsRefused(t *testing.T) {
	e := openEngine(t)
	if _, err := e.add("Write the spec", Changes{Objective: serves(7)}); err == nil {
		t.Error("a task serving an objective that does not exist was added")
	}
	addObjective(t, e, "Ship cairn")
	if _, err := e.add("Ship it", Changes{Objective: serves(1, 7)}); err == nil {
		t.Error("a task serving one objective that exists and one that does not was added")
	}
}

func TestObjectiveC_progressCountsWhatServesIt(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship cairn")
	serving(t, e, "Write the spec", 1)
	got, err := e.objectiveProgress(1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tasks.Total != 1 {
		t.Errorf("progress = %+v, want one task counted", got)
	}
}

func TestObjectiveC_stepsAreTalliedPastATaskWithNone(t *testing.T) {
	e := openEngine(t)
	add(t, e, "No steps")
	steps := []task.Step{{Title: "one", Done: true}, {Title: "two"}}
	if _, err := e.add("Two steps", Changes{Steps: &steps}); err != nil {
		t.Fatal(err)
	}
	tasks, _, err := e.alongEvery(nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := tasks[2].Steps; got.Total != 2 || got.Done != 1 {
		t.Errorf("task 2 tallied %+v, want 1 of 2", got)
	}
	p, err := e.taskProgress(2)
	if err != nil {
		t.Fatal(err)
	}
	if p.Steps.Total != 2 || p.Steps.Done != 1 {
		t.Errorf("task 2 progress %+v, want 1 of 2", p.Steps)
	}
}

func TestObjectiveC_newAndListAnswerTheObjectives(t *testing.T) {
	path := storeAt(t)
	made, err := Do(ObjectiveNew, path, given("title", "Ship cairn"))
	if err != nil {
		t.Fatal(err)
	}
	if made.Objective == nil {
		t.Fatal("a new objective answers none")
	}
	listed, err := Do(ObjectiveList, path, given())
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Objectives) != 1 {
		t.Errorf("listed %d objectives, want 1", len(listed.Objectives))
	}
}

func TestObjectiveC_anEditToABlankTitleSaysWhatWasTyped(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(ObjectiveNew, path, given("title", "Ship cairn")); err != nil {
		t.Fatal(err)
	}
	_, err := Do(ObjectiveEdit, path, given("number", "1", "title", "   "))
	if !errors.Is(err, domain.ErrTitleRequired) || !strings.HasSuffix(err.Error(), `, got "   "`) {
		t.Errorf("err = %v, want the title typed said", err)
	}
}

func TestObjectiveC_searchFindsOnlyWhatCarriesIt(t *testing.T) {
	path := storeAt(t)
	for _, title := range []string{"Ship cairn", "Hire someone"} {
		if _, err := Do(ObjectiveNew, path, given("title", title)); err != nil {
			t.Fatal(err)
		}
	}
	found, err := Do(ObjectiveSearch, path, Input{Listed: map[string][]string{"named": {"ship"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(found.MatchedObjectives) != 1 || objective.OfObjective(found.MatchedObjectives[0]).Title != "Ship cairn" {
		t.Errorf("found %v, want Ship cairn alone", found.MatchedObjectives)
	}
}

func TestOperationC_waitingNamesThePositionalArgumentsOnly(t *testing.T) {
	op := Operation{Thing: "task", Level: Full, Takes: []Argument{
		{Name: "number", Positional: true},
		{Name: "state"},
	}}
	err := op.waiting(given("number", "3", "state", "done"))
	if want := "task 3 " + ErrNotConfirmed.Error(); err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
}

func TestOperationC_waitingSaysASweepOnlyWhenNothingIsNamed(t *testing.T) {
	op := Operation{Thing: "task", Sweeps: func(Input) bool { return true }, Takes: []Argument{{Name: "number", Positional: true}}}
	if err := op.waiting(given()); err == nil || err.Error() != "every task it reaches "+ErrNotConfirmed.Error() {
		t.Errorf("nothing named: err = %v", err)
	}
	if err := op.waiting(given("number", "3")); err == nil || err.Error() != "task 3 "+ErrNotConfirmed.Error() {
		t.Errorf("a number named: err = %v", err)
	}
}

func TestAuthorC_aNameSetEmptyIsTheOneLoggedIn(t *testing.T) {
	me, err := user.Current()
	if err != nil {
		t.Skip("no user to compare with:", err)
	}
	empty := Home{Lookup: func(k string) (string, bool) { return "", k == AuthorVariable }}
	if got := Author(empty); got != me.Username {
		t.Errorf("named empty, the author is %q, want %q", got, me.Username)
	}
}

func archivingHome(t *testing.T) (string, Home) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	add(t, e, "Migration Jenkins")
	if _, err := e.Ask(TaskEdit, given("number", "1", "state", "done")); err != nil {
		t.Fatal(err)
	}
	ago := time.Now().AddDate(0, 0, -10)
	if err := e.history.Record(trace.Made(trace.Carried{Task: 1, Did: trace.Changed, Field: "column", Was: "todo", Became: "done", At: ago})); err != nil {
		t.Fatal(err)
	}
	_ = e.Close()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	start := time.Now().AddDate(0, 0, -20).Format(time.DateOnly)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), fmt.Appendf(nil, "[archive]\nstart = %s\n", start), 0o600); err != nil {
		t.Fatal(err)
	}
	return path, Home{Dir: home, Lookup: func(k string) (string, bool) { return path, k == "CAIRN_STORE" }}
}

func TestArchivingC_onlyTheListingOfTasksSweeps(t *testing.T) {
	path, home := archivingHome(t)
	at := Places{Store: path, Home: home}
	for op, in := range map[*Operation]Input{&TaskShow: given("number", "1"), &ObjectiveList: given()} {
		if _, err := DoAt(*op, at, in, Acting{Through: "cli", By: "sam"}); err != nil {
			t.Fatal(err)
		}
		shown, err := Do(TaskShow, path, given("number", "1"))
		if err != nil {
			t.Fatal(err)
		}
		if factsOf(*shown.Task).Archived {
			t.Fatalf("%s archived on its own", op.Line())
		}
	}
	if _, err := DoAt(TaskList, at, given(), Acting{Through: "cli", By: "sam"}); err != nil {
		t.Fatal(err)
	}
	shown, err := Do(TaskShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if !factsOf(*shown.Task).Archived {
		t.Error("task list left a task done before the period under way on the board")
	}
}

func TestTaskC_showSaysWhereItsNotesAre(t *testing.T) {
	path := storeAt(t, "Write the spec")
	notes := t.TempDir()
	shown, err := DoAt(TaskShow, Places{Store: path, Notes: notes}, given("number", "1"), Acting{})
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(notes, "1"); shown.Context != want {
		t.Errorf("Context = %q, want %q", shown.Context, want)
	}
}

func TestTaskC_unarchiveAnswersTheTask(t *testing.T) {
	path := storeAt(t, "Write the spec")
	if _, err := Do(TaskArchive, path, given("number", "1")); err != nil {
		t.Fatal(err)
	}
	back, err := Do(TaskUnarchive, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if back.Task == nil || factsOf(*back.Task).Number != 1 || factsOf(*back.Task).Archived {
		t.Errorf("unarchive answered %v, want task 1 back on the board", back.Task)
	}
}

func TestTaskC_sweepsByNarrowingOnlyWithoutANumber(t *testing.T) {
	if !sweepsByNarrowing(given("number", "0", "state", "done")) {
		t.Error("number 0 and a state does not sweep")
	}
	if sweepsByNarrowing(given("number", "5", "state", "done")) {
		t.Error("number 5 and a state sweeps")
	}
}

func TestTaskC_stepsGivenAreReadAndRefusedWhenNoText(t *testing.T) {
	path := storeAt(t)
	made, err := Do(TaskNew, path, given("title", "Write the spec", "steps", "[x] one\n[ ] two"))
	if err != nil {
		t.Fatal(err)
	}
	if steps := factsOf(*made.Task).Steps; len(steps) != 2 || !steps[0].Done {
		t.Errorf("steps = %+v, want one done and one not", steps)
	}
	if _, err := Do(TaskNew, path, given("title", "Ship it", "steps", "[ ] bell\x07")); err == nil {
		t.Error("a step that is no text was taken")
	}
}

func TestTaskC_aNarrowedListingTalliesTheStepsOfEveryTask(t *testing.T) {
	path := storeAt(t, "Write the spec")
	if _, err := Do(TaskNew, path, given("title", "Ship it", "state", "todo", "steps", "[x] one\n[ ] two")); err != nil {
		t.Fatal(err)
	}
	listed, err := Do(TaskList, path, given("state", "backlog"))
	if err != nil {
		t.Fatal(err)
	}
	if got := listed.Along[2].Steps; got.Total != 2 || got.Done != 1 {
		t.Errorf("Along[2] = %+v, want 1 of 2 although task 2 is not listed", got)
	}
}
