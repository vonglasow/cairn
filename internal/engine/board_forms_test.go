package engine

import (
	"testing"

	"cairn/internal/task"
)

func TestNewForms_startWhereEveryBoardStartsThem(t *testing.T) {
	task := NewTaskFields("todo")
	if state, _ := valueOf(task, "column"); state != "todo" {
		t.Errorf("a new task starts in %q, want todo", state)
	}
	if priority, _ := valueOf(task, "priority"); priority != "normal" {
		t.Errorf("a new task starts at %q, want normal", priority)
	}
	goal := NewGoalFields()
	if state, _ := valueOf(goal, "standing"); state != "active" {
		t.Errorf("a new objective starts %q, want active", state)
	}
	if priority, _ := valueOf(goal, "priority"); priority != "normal" {
		t.Errorf("a new objective starts at %q, want normal", priority)
	}
}

func TestEditFields_bringTheParts(t *testing.T) {
	cards := []Card{{Number: 1, Title: "Ship it"}, {Number: 2, Parent: 1, Title: "Read it", State: "todo"}}
	parts, ok := valueOf(EditFields(cards[0], cards), "parts")
	if !ok || parts != "#2 [ ] Read it" {
		t.Errorf("parts = %q, %v", parts, ok)
	}
}

func TestSaved_isTheOperationAndWhatItCarries(t *testing.T) {
	fields := []Field{{Name: "title", Arg: "title", Value: "  Ship it  "}}
	op, in := Saved(fields, true, 0, TaskNew, TaskEdit)
	if op.Name() != "task_new" || in.Given["title"] != "Ship it" {
		t.Errorf("adding saved %s %v", op.Name(), in.Given)
	}
	if _, numbered := in.Given["number"]; numbered {
		t.Error("adding named a number")
	}
	op, in = Saved(fields, false, 7, ObjectiveNew, ObjectiveEdit)
	if op.Name() != "objective_edit" || in.Given["number"] != "7" {
		t.Errorf("editing saved %s %v", op.Name(), in.Given)
	}
}

func TestAddingFields_carryAnEmptyPartsField(t *testing.T) {
	fields := AddingFields("todo")
	if last := fields[len(fields)-1]; last.Name != "parts" || last.Value != "" || !last.Long {
		t.Errorf("the last field of a new task is %+v", last)
	}
	if state, _ := valueOf(fields, "column"); state != "todo" {
		t.Errorf("a new task starts in %q, want todo", state)
	}
}

func TestAdded_isTheNumberOfTheTaskAdded(t *testing.T) {
	told := task.From(aTask("Ship it"))
	told.Numbered(7)
	added := told.Task()
	if got := Added(Answer{Task: &added}); got != 7 {
		t.Errorf("an add of task 7 answered %d", got)
	}
	if got := Added(Answer{}); got != 0 {
		t.Errorf("an answer with no task answered %d", got)
	}
}
