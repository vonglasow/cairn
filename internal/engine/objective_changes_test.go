package engine

import (
	"errors"
	"testing"
	"time"

	"cairn/internal/domain"
)

func TestEngine_numbersObjectivesApartFromTasks(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	if got := addObjective(t, e, "Ship cairn 1.0").Number; got != 1 {
		t.Errorf("Number = %d, want 1", got)
	}
}

func TestEngine_refusesAnObjectiveWithNoTitle(t *testing.T) {
	if _, err := openEngine(t).AddObjective("  ", ObjectiveChanges{}); !errors.Is(err, domain.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestEngine_givesAnObjectiveWhatItWasNamed(t *testing.T) {
	e := openEngine(t)
	o, err := e.AddObjective("Ship cairn 1.0", ObjectiveChanges{
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
	if o.Description != "the one that unblocks the rest" {
		t.Errorf("Description = %q", o.Description)
	}
	if o.State != domain.OnHold || o.Priority != domain.High {
		t.Errorf("State/Priority = %q/%q, want on-hold/high", o.State, o.Priority)
	}
	if len(o.Labels) != 2 || o.Labels[0] != "perso" || o.Labels[1] != "spec" {
		t.Errorf("Labels = %v, want [perso spec] lower-cased and sorted", o.Labels)
	}
	if o.Deadline != (time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Deadline = %v", o.Deadline)
	}
	if o.Estimate != 90*time.Minute {
		t.Errorf("Estimate = %v, want 1h30m", o.Estimate)
	}
}

func TestEngine_refusesAnObjectiveStateItCannotMakeSenseOf(t *testing.T) {
	e := openEngine(t)
	if _, err := e.AddObjective("Ship cairn 1.0", ObjectiveChanges{State: of("todo")}); !errors.Is(err, domain.ErrNotAnObjectiveState) {
		t.Errorf("err = %v, want ErrNotAnObjectiveState", err)
	}
}

func TestEngine_refusesALabelSignWhenCreatingAnObjective(t *testing.T) {
	e := openEngine(t)
	_, err := e.AddObjective("Ship cairn 1.0", ObjectiveChanges{Labels: of([]string{"+urgent"})})
	if !errors.Is(err, ErrLabelSignAtCreation) {
		t.Errorf("err = %v, want ErrLabelSignAtCreation", err)
	}
}

func TestEngine_listsObjectivesInTheOrderTheyWereAdded(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship cairn 1.0")
	addObjective(t, e, "Learn Go properly")

	all, err := e.ListObjectives()
	if err != nil {
		t.Fatalf("ListObjectives: %v", err)
	}
	if len(all) != 2 || all[0].Title != "Ship cairn 1.0" {
		t.Errorf("got %v, want them in the order they were added", all)
	}
}

func TestEngine_showsAnObjectiveByItsNumber(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	got, err := e.ShowObjective(o.Number)
	if err != nil {
		t.Fatalf("ShowObjective: %v", err)
	}
	if got.Title != "Ship cairn 1.0" {
		t.Errorf("Title = %q", got.Title)
	}
}

func TestEngine_saysWhenTheObjectiveShownDoesNotExist(t *testing.T) {
	if _, err := openEngine(t).ShowObjective(99); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestEngine_changesWhatWasNamedOnAnObjective(t *testing.T) {
	e := openEngine(t)
	o, err := e.AddObjective("Ship cairn 1.0", ObjectiveChanges{Priority: of("high")})
	if err != nil {
		t.Fatalf("AddObjective: %v", err)
	}
	changed, err := e.EditObjective(o.Number, ObjectiveChanges{State: of("reached")})
	if err != nil {
		t.Fatalf("EditObjective: %v", err)
	}
	if changed.State != domain.Reached {
		t.Errorf("State = %q, want reached", changed.State)
	}
	if changed.Priority != domain.High {
		t.Errorf("Priority = %q, want the high it already carried", changed.Priority)
	}
	if changed.Title != "Ship cairn 1.0" {
		t.Errorf("Title = %q, want it left alone", changed.Title)
	}
}

func TestEngine_refusesAnObjectiveEditThatNamesNothing(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	if _, err := e.EditObjective(o.Number, ObjectiveChanges{}); !errors.Is(err, ErrNothingToChange) {
		t.Errorf("err = %v, want ErrNothingToChange", err)
	}
}

// The label signs work on an objective the way they work on a task, because
// they are the same code: a set is moved, not replaced.
func TestEngine_movesAnObjectivesLabelsWithSigns(t *testing.T) {
	e := openEngine(t)
	o, err := e.AddObjective("Ship cairn 1.0", ObjectiveChanges{Labels: of([]string{"perso", "spec"})})
	if err != nil {
		t.Fatalf("AddObjective: %v", err)
	}
	moved, err := e.EditObjective(o.Number, ObjectiveChanges{Labels: of([]string{"+urgent", "-spec"})})
	if err != nil {
		t.Fatalf("EditObjective: %v", err)
	}
	if len(moved.Labels) != 2 || moved.Labels[0] != "perso" || moved.Labels[1] != "urgent" {
		t.Errorf("Labels = %v, want [perso urgent]", moved.Labels)
	}
}

func TestEngine_deletesAnObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	if err := e.DeleteObjective(o.Number); err != nil {
		t.Fatalf("DeleteObjective: %v", err)
	}
	if _, err := e.ShowObjective(o.Number); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

// Deleting an objective that tasks were put under would leave them pointing at
// a number nothing carries, so it is refused rather than cascaded.
func TestEngine_refusesToDeleteAnObjectiveCarryingTasks(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	task := serving(t, e, "Write the spec", o.Number)

	if err := e.DeleteObjective(o.Number); !errors.Is(err, ErrObjectiveCarriesTasks) {
		t.Errorf("err = %v, want ErrObjectiveCarriesTasks", err)
	}
	if _, err := e.ShowObjective(o.Number); err != nil {
		t.Errorf("the objective went away anyway: %v", err)
	}
	if _, err := e.Show(task.Number); err != nil {
		t.Errorf("the task went away with it: %v", err)
	}
}

func TestEngine_saysWhenTheObjectiveDeletedDoesNotExist(t *testing.T) {
	if err := openEngine(t).DeleteObjective(99); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestEngine_refusesToEmptyAnObjectivesTitle(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	if _, err := e.EditObjective(o.Number, ObjectiveChanges{Title: of("  ")}); !errors.Is(err, domain.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
	if still, err := e.ShowObjective(o.Number); err != nil {
		t.Fatalf("ShowObjective: %v", err)
	} else if still.Title != "Ship cairn 1.0" {
		t.Errorf("Title = %q, want it untouched", still.Title)
	}
}
