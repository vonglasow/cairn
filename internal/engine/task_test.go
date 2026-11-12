package engine

import (
	"path/filepath"
	"slices"
	"testing"

	"cairn/internal/trace"
)

func TestTaskShow_saysSinceWhenTheTaskStands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if _, err := DoAt(TaskNew, Places{Store: path}, Input{Given: map[string]string{"title": "Ship it"}}, Acting{}); err != nil {
		t.Fatal(err)
	}
	shown, err := DoAt(TaskShow, Places{Store: path}, Input{Given: map[string]string{"number": "1"}}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	if shown.StoodSince.IsZero() {
		t.Error("a task shown says nothing of since when it stands")
	}
}

func TestLabel_putsTheLabelOnEveryTaskReached(t *testing.T) {
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

func TestLabel_aLabelThatIsNoTextStopsTheSweep(t *testing.T) {
	path := storeAt(t, "Write the spec")
	if _, err := Do(TaskLabel, path, given("put", "bell\x07", "state", "backlog")); err == nil {
		t.Error("a label that is no text was put")
	}
}

func TestDelete_aTaskInFocusIsReleasedThenDeleted(t *testing.T) {
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

func TestTask_showSaysWhereItsNotesAre(t *testing.T) {
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

func TestTask_unarchiveAnswersTheTask(t *testing.T) {
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

func TestTask_sweepsByNarrowingOnlyWithoutANumber(t *testing.T) {
	if !sweepsByNarrowing(given("number", "0", "state", "done")) {
		t.Error("number 0 and a state does not sweep")
	}
	if sweepsByNarrowing(given("number", "5", "state", "done")) {
		t.Error("number 5 and a state sweeps")
	}
}

func TestTask_stepsGivenAreReadAndRefusedWhenNoText(t *testing.T) {
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

func TestTask_aNarrowedListingTalliesTheStepsOfEveryTask(t *testing.T) {
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

func TestTaskEdit_aTitleAndARuleGivenAtOnceAreBothKept(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Water"))
	answer := doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "title", "Watered", "every", "every day"))
	if said := Of(TaskEdit, answer).(Task); said.Title != "Watered" || said.Series == nil {
		t.Errorf("the task reads %+v, series %+v", said.Facts, said.Series)
	}
}

func TestTaskEdit_withSeriesAnswersTheTaskEdited(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday"))
	if answer := doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "title", "Daily", "series", "true")); answer.Task == nil || answer.Task.Carried().Title != "Daily" {
		t.Errorf("answered %+v", answer.Task)
	}
}
