package engine

import (
	"cairn/internal/engine/item"
	"errors"
	"testing"
)

func TestEngine_refusesAnObjectiveWithNoTitle(t *testing.T) {
	if _, err := openEngine(t).addObjective("  ", ObjectiveChanges{}); !errors.Is(err, item.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestEngine_refusesAnObjectiveEditThatNamesNothing(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	if _, err := e.editObjective(o.Carried().Number, ObjectiveChanges{}); !errors.Is(err, errNothingToChange) {
		t.Errorf("err = %v, want ErrNothingToChange", err)
	}
}
