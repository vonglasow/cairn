package engine

import (
	"testing"
)

func TestAskedNamesEveryFieldByItsEngineArgument(t *testing.T) {
	fields := TaskFields(Card{
		Title:       "One",
		Description: "why it matters",
		State:       "todo",
		Priority:    "high",
		Labels:      []string{"api", "urgent"},
	})

	in := Asked(fields, TaskEdit)
	if in.Given["desc"] != "why it matters" {
		t.Errorf("desc = %q, want the description", in.Given["desc"])
	}
	if in.Given["state"] != "todo" {
		t.Errorf("state = %q, want todo", in.Given["state"])
	}
	if got := in.Listed["label"]; len(got) != 2 || got[0] != "api" || got[1] != "urgent" {
		t.Errorf("label = %v, want the two labels, listed", got)
	}
	for _, screen := range []string{"description", "labels", "column", "time spent"} {
		if _, named := in.Given[screen]; named {
			t.Errorf("%q travelled: that is the screen's word, not the engine's", screen)
		}
	}
}

func TestAskedLeavesANumberNobodyCanReadUnnamed(t *testing.T) {
	if in := Asked([]Field{{Name: "parent", Arg: "parent", Value: "nine"}}, TaskEdit); in.Named("parent") {
		t.Fatal("a parent nobody can read was passed to the engine anyway")
	}
}

func TestAskedNamesAnEmptyNumber(t *testing.T) {
	if in := Asked([]Field{{Name: "parent", Arg: "parent", Value: "  "}}, TaskEdit); !in.Named("parent") {
		t.Fatal("emptying the parent said nothing at all")
	}
}
