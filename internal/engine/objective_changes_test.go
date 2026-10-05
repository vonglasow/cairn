package engine

import (
	"errors"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

func TestEngine_numbersObjectivesApartFromTasks(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	if got := factsOfObjective(addObjective(t, e, "Ship cairn 1.0")).number; got != 1 {
		t.Errorf("Number = %d, want 1", got)
	}
}

func TestEngine_refusesAnObjectiveWithNoTitle(t *testing.T) {
	if _, err := openEngine(t).addObjective("  ", ObjectiveChanges{}); !errors.Is(err, domain.ErrTitleRequired) {
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
	facts := factsOfObjective(o)
	if facts.description != "the one that unblocks the rest" {
		t.Errorf("Description = %q", facts.description)
	}
	if facts.state != objective.OnHold || facts.priority != domain.High {
		t.Errorf("State/Priority = %q/%q, want on-hold/high", facts.state, facts.priority)
	}
	if len(facts.labels) != 2 || facts.labels[0] != "perso" || facts.labels[1] != "spec" {
		t.Errorf("Labels = %v, want [perso spec] lower-cased and sorted", facts.labels)
	}
	if facts.deadline != (time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Deadline = %v", facts.deadline)
	}
	if facts.estimate != 90*time.Minute {
		t.Errorf("Estimate = %v, want 1h30m", facts.estimate)
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
	if !errors.Is(err, ErrLabelSignAtCreation) {
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
	if len(all) != 2 || factsOfObjective(all[0]).title != "Ship cairn 1.0" {
		t.Errorf("got %v, want them in the order they were added", all)
	}
}

func TestEngine_showsAnObjectiveByItsNumber(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	got, err := e.showObjective(factsOfObjective(o).number)
	if err != nil {
		t.Fatalf("ShowObjective: %v", err)
	}
	if facts := factsOfObjective(got); facts.title != "Ship cairn 1.0" {
		t.Errorf("Title = %q", facts.title)
	}
}

func TestEngine_saysWhenTheObjectiveShownDoesNotExist(t *testing.T) {
	if _, err := openEngine(t).showObjective(99); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestEngine_changesWhatWasNamedOnAnObjective(t *testing.T) {
	e := openEngine(t)
	o, err := e.addObjective("Ship cairn 1.0", ObjectiveChanges{Priority: of("high")})
	if err != nil {
		t.Fatalf("AddObjective: %v", err)
	}
	changed, err := e.editObjective(factsOfObjective(o).number, ObjectiveChanges{State: of("reached")})
	if err != nil {
		t.Fatalf("EditObjective: %v", err)
	}
	facts := factsOfObjective(changed)
	if facts.state != objective.Reached {
		t.Errorf("State = %q, want reached", facts.state)
	}
	if facts.priority != domain.High {
		t.Errorf("Priority = %q, want the high it already carried", facts.priority)
	}
	if facts.title != "Ship cairn 1.0" {
		t.Errorf("Title = %q, want it left alone", facts.title)
	}
}

func TestEngine_refusesAnObjectiveEditThatNamesNothing(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	if _, err := e.editObjective(factsOfObjective(o).number, ObjectiveChanges{}); !errors.Is(err, ErrNothingToChange) {
		t.Errorf("err = %v, want ErrNothingToChange", err)
	}
}

// The label signs work on an objective the way they work on a task, because
// they are the same code: a set is moved, not replaced.
func TestEngine_movesAnObjectivesLabelsWithSigns(t *testing.T) {
	e := openEngine(t)
	o, err := e.addObjective("Ship cairn 1.0", ObjectiveChanges{Labels: of([]string{"perso", "spec"})})
	if err != nil {
		t.Fatalf("AddObjective: %v", err)
	}
	moved, err := e.editObjective(factsOfObjective(o).number, ObjectiveChanges{Labels: of([]string{"+urgent", "-spec"})})
	if err != nil {
		t.Fatalf("EditObjective: %v", err)
	}
	if labels := factsOfObjective(moved).labels; len(labels) != 2 || labels[0] != "perso" || labels[1] != "urgent" {
		t.Errorf("Labels = %v, want [perso urgent]", labels)
	}
}

func TestEngine_deletesAnObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	if err := e.deleteObjective(factsOfObjective(o).number); err != nil {
		t.Fatalf("DeleteObjective: %v", err)
	}
	if _, err := e.showObjective(factsOfObjective(o).number); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

// Deleting an objective that tasks were put under would leave them pointing at
// a number nothing carries, so it is refused rather than cascaded.
func TestEngine_refusesToDeleteAnObjectiveCarryingTasks(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	task := serving(t, e, "Write the spec", factsOfObjective(o).number)

	if err := e.deleteObjective(factsOfObjective(o).number); !errors.Is(err, ErrObjectiveCarriesTasks) {
		t.Errorf("err = %v, want ErrObjectiveCarriesTasks", err)
	}
	if _, err := e.showObjective(factsOfObjective(o).number); err != nil {
		t.Errorf("the objective went away anyway: %v", err)
	}
	if _, err := e.show(factsOf(task).number); err != nil {
		t.Errorf("the task went away with it: %v", err)
	}
}

func TestEngine_saysWhenTheObjectiveDeletedDoesNotExist(t *testing.T) {
	if err := openEngine(t).deleteObjective(99); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestEngine_refusesToEmptyAnObjectivesTitle(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	if _, err := e.editObjective(factsOfObjective(o).number, ObjectiveChanges{Title: of("  ")}); !errors.Is(err, domain.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
	if still, err := e.showObjective(factsOfObjective(o).number); err != nil {
		t.Fatalf("ShowObjective: %v", err)
	} else if facts := factsOfObjective(still); facts.title != "Ship cairn 1.0" {
		t.Errorf("Title = %q, want it untouched", facts.title)
	}
}
