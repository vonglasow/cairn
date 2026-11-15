package engine

import (
	"cairn/internal/engine/change"
	"errors"
	"testing"
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
	if got, _ := e.show(1); got.Carried().Description != "built on main at def456" {
		t.Errorf("description = %q", got.Carried().Description)
	}
}

func TestTaskEdit_refusesAPieceItCannotReplace(t *testing.T) {
	for _, tc := range []struct {
		in   Input
		want error
	}{
		{given("number", "1", "replace", "develop", "by", "x"), change.ErrReplacedNowhere},
		{given("number", "1", "replace", "main", "by", "x"), change.ErrReplacedMoreThanOnce},
		{given("number", "1", "replace", "main"), errReplaceAlone},
		{given("number", "1", "by", "main"), errReplaceAlone},
		{given("number", "1", "desc", "other", "replace", "main", "by", "x"), errDescriptionBothWays},
	} {
		e := describedTask(t, "main into main")
		if _, err := e.Ask(TaskEdit, tc.in); !errors.Is(err, tc.want) {
			t.Errorf("%v: err = %v, want %v", tc.in.Given, err, tc.want)
		}
		if got, _ := e.show(1); got.Carried().Description != "main into main" {
			t.Errorf("%v: description became %q", tc.in.Given, got.Carried().Description)
		}
	}
}
