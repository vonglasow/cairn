package engine

import (
	"cairn/internal/engine/item"
	"errors"
	"testing"
	"time"

	"cairn/internal/engine/objective"
)

func TestEngine_numbersObjectivesApartFromTasks(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	if got := addObjective(t, e, "Ship cairn 1.0").Carried().Number; got != 1 {
		t.Errorf("Number = %d, want 1", got)
	}
}

func TestEngine_refusesAnObjectiveWithNoTitle(t *testing.T) {
	if _, err := openEngine(t).addObjective("  ", ObjectiveChanges{}); !errors.Is(err, item.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestEngine_givesAnObjectiveWhatItWasNamed(t *testing.T) {
	e := openEngine(t)
	o, err := e.addObjective("Ship cairn 1.0", ObjectiveChanges{
		Description: of("the one that unblocks the rest"),
		State:       of("on-hold"),
		Priority:    of("high"),
		Labels:      of([]string{"Spec", "perso"}),
		Deadline:    of("2026-12-31"),
		Estimate:    of("90m"),
	})
	if err != nil {
		t.Fatalf("AddObjective: %v", err)
	}
	facts := o.Carried()
	if facts.Description != "the one that unblocks the rest" {
		t.Errorf("Description = %q", facts.Description)
	}
	if facts.State != objective.OnHold || facts.Priority != item.High {
		t.Errorf("State/Priority = %q/%q, want on-hold/high", facts.State, facts.Priority)
	}
	if len(facts.Labels) != 2 || facts.Labels[0] != "perso" || facts.Labels[1] != "spec" {
		t.Errorf("Labels = %v, want [perso spec] lower-cased and sorted", facts.Labels)
	}
	if facts.Deadline != (time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Deadline = %v", facts.Deadline)
	}
	if facts.Estimate != 90*time.Minute {
		t.Errorf("Estimate = %v, want 1h30m", facts.Estimate)
	}
}

func TestEngine_refusesAnObjectiveStateItCannotMakeSenseOf(t *testing.T) {
	e := openEngine(t)
	if _, err := e.addObjective("Ship cairn 1.0", ObjectiveChanges{State: of("todo")}); !errors.Is(err, objective.ErrNotAnObjectiveState) {
		t.Errorf("err = %v, want ErrNotAnObjectiveState", err)
	}
}

func TestEngine_refusesALabelSignWhenCreatingAnObjective(t *testing.T) {
	e := openEngine(t)
	_, err := e.addObjective("Ship cairn 1.0", ObjectiveChanges{Labels: of([]string{"+urgent"})})
	if !errors.Is(err, errLabelSignAtCreation) {
		t.Errorf("err = %v, want ErrLabelSignAtCreation", err)
	}
}

func TestEngine_listsObjectivesInTheOrderTheyWereAdded(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship cairn 1.0")
	addObjective(t, e, "Learn Go properly")

	all, err := e.listObjectives()
	if err != nil {
		t.Fatalf("ListObjectives: %v", err)
	}
	if len(all) != 2 || all[0].Carried().Title != "Ship cairn 1.0" {
		t.Errorf("got %v, want them in the order they were added", all)
	}
}

func TestEngine_showsAnObjectiveByItsNumber(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	got, err := e.showObjective(o.Carried().Number)
	if err != nil {
		t.Fatalf("ShowObjective: %v", err)
	}
	if facts := got.Carried(); facts.Title != "Ship cairn 1.0" {
		t.Errorf("Title = %q", facts.Title)
	}
}

func TestEngine_saysWhenTheObjectiveShownDoesNotExist(t *testing.T) {
	if _, err := openEngine(t).showObjective(99); !errors.Is(err, objective.ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestEngine_changesWhatWasNamedOnAnObjective(t *testing.T) {
	e := openEngine(t)
	o, err := e.addObjective("Ship cairn 1.0", ObjectiveChanges{Priority: of("high")})
	if err != nil {
		t.Fatalf("AddObjective: %v", err)
	}
	changed, err := e.editObjective(o.Carried().Number, ObjectiveChanges{State: of("reached")})
	if err != nil {
		t.Fatalf("EditObjective: %v", err)
	}
	facts := changed.Carried()
	if facts.State != objective.Reached {
		t.Errorf("State = %q, want reached", facts.State)
	}
	if facts.Priority != item.High {
		t.Errorf("Priority = %q, want the high it already carried", facts.Priority)
	}
	if facts.Title != "Ship cairn 1.0" {
		t.Errorf("Title = %q, want it left alone", facts.Title)
	}
}

func TestEngine_refusesAnObjectiveEditThatNamesNothing(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	if _, err := e.editObjective(o.Carried().Number, ObjectiveChanges{}); !errors.Is(err, errNothingToChange) {
		t.Errorf("err = %v, want ErrNothingToChange", err)
	}
}

func TestEngine_movesAnObjectivesLabelsWithSigns(t *testing.T) {
	e := openEngine(t)
	o, err := e.addObjective("Ship cairn 1.0", ObjectiveChanges{Labels: of([]string{"perso", "spec"})})
	if err != nil {
		t.Fatalf("AddObjective: %v", err)
	}
	moved, err := e.editObjective(o.Carried().Number, ObjectiveChanges{Labels: of([]string{"+urgent", "-spec"})})
	if err != nil {
		t.Fatalf("EditObjective: %v", err)
	}
	if labels := moved.Carried().Labels; len(labels) != 2 || labels[0] != "perso" || labels[1] != "urgent" {
		t.Errorf("Labels = %v, want [perso urgent]", labels)
	}
}

func TestEngine_deletesAnObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	if _, err := e.deleteObjectives([]int{o.Carried().Number}, objective.Deletion{}); err != nil {
		t.Fatalf("DeleteObjective: %v", err)
	}
	if _, err := e.showObjective(o.Carried().Number); !errors.Is(err, objective.ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestEngine_refusesToDeleteAnObjectiveCarryingTasks(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	task := serving(t, e, "Write the spec", o.Carried().Number)

	if _, err := e.deleteObjectives([]int{o.Carried().Number}, objective.Deletion{}); !errors.Is(err, objective.ErrObjectiveCarriesTasks) {
		t.Errorf("err = %v, want ErrObjectiveCarriesTasks", err)
	}
	if _, err := e.showObjective(o.Carried().Number); err != nil {
		t.Errorf("the objective went away anyway: %v", err)
	}
	if _, err := e.show(task.Carried().Number); err != nil {
		t.Errorf("the task went away with it: %v", err)
	}
}

func TestEngine_saysWhenTheObjectiveDeletedDoesNotExist(t *testing.T) {
	if _, err := openEngine(t).deleteObjectives([]int{99}, objective.Deletion{}); !errors.Is(err, objective.ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestEngine_refusesToEmptyAnObjectivesTitle(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	if _, err := e.editObjective(o.Carried().Number, ObjectiveChanges{Title: of("  ")}); !errors.Is(err, item.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
	if still, err := e.showObjective(o.Carried().Number); err != nil {
		t.Fatalf("ShowObjective: %v", err)
	} else if facts := still.Carried(); facts.Title != "Ship cairn 1.0" {
		t.Errorf("Title = %q, want it untouched", facts.Title)
	}
}
