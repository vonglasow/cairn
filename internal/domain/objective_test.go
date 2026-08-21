package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewObjective_refusesAnEmptyTitle(t *testing.T) {
	if _, err := NewObjective(""); !errors.Is(err, ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestNewObjective_refusesATitleOfSpaces(t *testing.T) {
	if _, err := NewObjective("   "); !errors.Is(err, ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired: spaces are not a title", err)
	}
}

// The sentinel is shared with a task, the wording is not: a person told "a
// task needs a title" after typing `objective add` learns nothing.
func TestNewObjective_saysItIsAnObjectiveThatNeedsATitle(t *testing.T) {
	_, err := NewObjective("")
	if got := err.Error(); !strings.Contains(got, "an objective needs a title") {
		t.Errorf("err = %q, want it to name the objective", got)
	}
	if _, err := NewTask(""); !strings.Contains(err.Error(), "a task needs a title") {
		t.Errorf("err = %q, want a task to still name itself", err.Error())
	}
}

func TestNewObjective_trimsWhatSurroundsTheTitle(t *testing.T) {
	got, err := NewObjective("  Ship cairn 1.0  ")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	if got.Title != "Ship cairn 1.0" {
		t.Errorf("Title = %q, want it trimmed", got.Title)
	}
}

func TestNewObjective_keepsTheSpacesInsideTheTitle(t *testing.T) {
	got, err := NewObjective("Ship  cairn   1.0")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	if got.Title != "Ship  cairn   1.0" {
		t.Errorf("Title = %q, want the inner spacing untouched", got.Title)
	}
}

func TestNewObjective_startsActiveAtNormalPriority(t *testing.T) {
	got, err := NewObjective("Ship cairn 1.0")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	if got.State != Active {
		t.Errorf("State = %q, want %q", got.State, Active)
	}
	if got.Priority != Normal {
		t.Errorf("Priority = %q, want %q", got.Priority, Normal)
	}
}

func TestNewObjective_carriesNothingElse(t *testing.T) {
	got, err := NewObjective("Ship cairn 1.0")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	if got.Number != 0 || got.Description != "" || len(got.Labels) != 0 {
		t.Errorf("%+v carries more than it was given", got)
	}
	if !got.Deadline.IsZero() || got.Estimate != 0 {
		t.Errorf("%+v carries a deadline or an estimate it was not given", got)
	}
}

func TestParseObjectiveState_readsEveryOne(t *testing.T) {
	for _, want := range []ObjectiveState{Active, OnHold, Reached, Dropped} {
		got, err := ParseObjectiveState(string(want))
		if err != nil {
			t.Errorf("ParseObjectiveState(%q): %v", want, err)
			continue
		}
		if got != want {
			t.Errorf("ParseObjectiveState(%q) = %q", want, got)
		}
	}
}

func TestParseObjectiveState_readsNothingAsActive(t *testing.T) {
	got, err := ParseObjectiveState("")
	if err != nil {
		t.Fatalf("ParseObjectiveState(%q): %v", "", err)
	}
	if got != Active {
		t.Errorf("ParseObjectiveState(%q) = %q, want %q", "", got, Active)
	}
}

// An objective has no column, so it cannot be put in one. A task's state given
// to an objective is the mistake most likely to be made, and the one a shared
// parser would have accepted in silence.
func TestParseObjectiveState_refusesATasksState(t *testing.T) {
	for _, name := range []string{"backlog", "todo", "current", "done", "nope"} {
		if _, err := ParseObjectiveState(name); !errors.Is(err, ErrNotAnObjectiveState) {
			t.Errorf("ParseObjectiveState(%q) err = %v, want ErrNotAnObjectiveState", name, err)
		}
	}
}

func TestParseObjectiveState_saysWhatItWasGiven(t *testing.T) {
	_, err := ParseObjectiveState("todo")
	if got := err.Error(); !strings.Contains(got, `"todo"`) {
		t.Errorf("err = %q, want it to quote what it was given", got)
	}
}

// The other way round: a column is still a column, and the states an
// objective can be in are not among them.
func TestParseState_refusesAnObjectivesState(t *testing.T) {
	for _, name := range []string{"active", "on-hold", "reached", "dropped"} {
		if _, err := ParseState(name); !errors.Is(err, ErrNotAState) {
			t.Errorf("ParseState(%q) err = %v, want ErrNotAState", name, err)
		}
	}
}

func TestNewTask_servesNoObjectiveUnlessItIsTold(t *testing.T) {
	got, err := NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if got.Objective != NoObjective {
		t.Errorf("Objective = %d, want none", got.Objective)
	}
}
