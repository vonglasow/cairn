package board

import (
	"cairn/internal/engine"
	"slices"
	"testing"

	"cairn/internal/engine/task"
)

func TestAdded_isTheNumberOfTheTaskAdded(t *testing.T) {
	told := task.From(aTask("Ship it"))
	told.Numbered(7)
	added := told.Task()
	if got := Added(engine.Answer{Task: &added}); got != 7 {
		t.Errorf("an add of task 7 answered %d", got)
	}
	if got := Added(engine.Answer{}); got != 0 {
		t.Errorf("an answer with no task answered %d", got)
	}
}

func TestANewTaskFormCanComeLinked(t *testing.T) {
	fields := LinkedTaskFields(4)
	if state, _ := valueOf(fields, "column"); state != "todo" {
		t.Errorf("a linked task starts in %q, want todo", state)
	}
	if links, _ := valueOf(fields, "links"); links != "4" {
		t.Errorf("a linked task comes linked to %q, want 4", links)
	}
	op, in := Saved(fields, true, 0, engine.TaskNew, engine.TaskEdit)
	if op.Name() != "task_new" || !slices.Equal(in.Listed["link"], []string{"=", "4"}) || in.Given["state"] != "todo" {
		t.Errorf("saving it asked %s %v", op.Name(), in.Given)
	}
}

func TestANewTaskFormCanComeServingAnObjective(t *testing.T) {
	fields := ServingTaskFields(3)
	if goals, _ := valueOf(fields, "objectives"); goals != "3" {
		t.Errorf("a task made from an objective serves %q, want 3", goals)
	}
	op, in := Saved(fields, true, 0, engine.TaskNew, engine.TaskEdit)
	if op.Name() != "task_new" || !slices.Equal(in.Listed["objective"], []string{"=", "3"}) {
		t.Errorf("saving it asked %s %v %v", op.Name(), in.Given, in.Listed)
	}
}

func TestANewObjectiveFormCanComeUnderAParent(t *testing.T) {
	fields := NewSubGoalFields(4)
	if parent, _ := valueOf(fields, "parent"); parent != "4" {
		t.Errorf("a sub-objective made under 4 sits under %q", parent)
	}
	op, in := Saved(fields, true, 0, engine.ObjectiveNew, engine.ObjectiveEdit)
	if op.Name() != "objective_new" || in.Given["parent"] != "4" {
		t.Errorf("saving it asked %s %v", op.Name(), in.Given)
	}
}
