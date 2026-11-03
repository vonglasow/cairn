package engine

import (
	"errors"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

func describedTask(t *testing.T, said string) *Engine {
	t.Helper()
	e := openEngine(t)
	if _, err := e.add("Ship it", Changes{Description: of(said)}); err != nil {
		t.Fatal(err)
	}
	return e
}

func TestTaskEdit_replacesOnePieceOfTheDescription(t *testing.T) {
	e := describedTask(t, "built on main at abc123")
	if _, err := e.Ask(TaskEdit, given("number", "1", "replace", "abc123", "by", "def456")); err != nil {
		t.Fatal(err)
	}
	if got, _ := e.show(1); factsOf(got).Description != "built on main at def456" {
		t.Errorf("description = %q", factsOf(got).Description)
	}
}

func TestTaskEdit_refusesAPieceItCannotReplace(t *testing.T) {
	for _, tc := range []struct {
		in   Input
		want error
	}{
		{given("number", "1", "replace", "develop", "by", "x"), domain.ErrReplacedNowhere},
		{given("number", "1", "replace", "main", "by", "x"), domain.ErrReplacedMoreThanOnce},
		{given("number", "1", "replace", "main"), ErrReplaceAlone},
		{given("number", "1", "by", "main"), ErrReplaceAlone},
		{given("number", "1", "desc", "other", "replace", "main", "by", "x"), ErrDescriptionBothWays},
	} {
		e := describedTask(t, "main into main")
		if _, err := e.Ask(TaskEdit, tc.in); !errors.Is(err, tc.want) {
			t.Errorf("%v: err = %v, want %v", tc.in.Given, err, tc.want)
		}
		if got, _ := e.show(1); factsOf(got).Description != "main into main" {
			t.Errorf("%v: description became %q", tc.in.Given, factsOf(got).Description)
		}
	}
}

func TestObjectiveEdit_replacesOnePieceOfTheDescription(t *testing.T) {
	e := openEngine(t)
	if _, err := e.addObjective("Ship cairn", ObjectiveChanges{Description: of("for v0.10, on main")}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Ask(ObjectiveEdit, given("number", "1", "replace", "v0.10", "by", "v0.11")); err != nil {
		t.Fatal(err)
	}
	got, _ := e.showObjective(1)
	if said := objective.OfObjective(got).Description; said != "for v0.11, on main" {
		t.Errorf("description = %q", said)
	}
}
