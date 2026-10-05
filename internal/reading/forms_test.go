package reading

import (
	"testing"

	"cairn/internal/engine"
)

// A new task's form starts in the column it was opened on and at normal
// priority; a new objective's starts active and at normal. The terminal filled
// these in and the window left them empty, so one gesture made two different
// tasks.
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

// Editing a task brings its parts with it, written one line each.
func TestEditFields_bringTheParts(t *testing.T) {
	cards := []Card{{Number: 1, Title: "Ship it"}, {Number: 2, Parent: 1, Title: "Read it", State: "todo"}}
	parts, ok := valueOf(EditFields(cards[0], cards), "parts")
	if !ok || parts != "#2 [ ] Read it" {
		t.Errorf("parts = %q, %v", parts, ok)
	}
}

// What a saved form asks for: adding or editing, the title trimmed, and the
// number when it is an edit.
func TestSaved_isTheOperationAndWhatItCarries(t *testing.T) {
	fields := []Field{{Name: "title", Arg: "title", Value: "  Ship it  "}}
	op, in := Saved(fields, true, 0, engine.TaskAdd, engine.TaskEdit)
	if op.Name() != "task_add" || in.Given["title"] != "Ship it" {
		t.Errorf("adding saved %s %v", op.Name(), in.Given)
	}
	if _, numbered := in.Given["number"]; numbered {
		t.Error("adding named a number")
	}
	op, in = Saved(fields, false, 7, engine.ObjectiveAdd, engine.ObjectiveEdit)
	if op.Name() != "objective_edit" || in.Given["number"] != "7" {
		t.Errorf("editing saved %s %v", op.Name(), in.Given)
	}
}
