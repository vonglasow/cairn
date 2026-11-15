package engine

import (
	"slices"
	"testing"

	"cairn/internal/engine/act"
)

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
	if labels := (*shown.Task).Carried().Labels; !slices.Contains(labels, "sprint") {
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
	released, deleted := slices.Index(did, act.FocusReleased), slices.Index(did, act.Deleted)
	if released < 0 || deleted < released {
		t.Errorf("acts = %v, want the focus released then the task deleted", did)
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
	if back.Task == nil || (*back.Task).Carried().Number != 1 || (*back.Task).Carried().Archived {
		t.Errorf("unarchive answered %v, want task 1 back on the board", back.Task)
	}
}

func TestTask_stepsGivenAreReadAndRefusedWhenNoText(t *testing.T) {
	path := storeAt(t)
	made, err := Do(TaskNew, path, given("title", "Write the spec", "steps", "[x] one\n[ ] two"))
	if err != nil {
		t.Fatal(err)
	}
	if steps := (*made.Task).Carried().Steps; len(steps) != 2 || !steps[0].Done {
		t.Errorf("steps = %+v, want one done and one not", steps)
	}
	if _, err := Do(TaskNew, path, given("title", "Ship it", "steps", "[ ] bell\x07")); err == nil {
		t.Error("a step that is no text was taken")
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
