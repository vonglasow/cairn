package board

import (
	"cairn/internal/engine"
	"slices"
	"testing"
)

func TestAskedNamesEveryFieldByItsEngineArgument(t *testing.T) {
	fields := taskFields(Card{
		Title:       "One",
		Description: "why it matters",
		State:       "todo",
		Priority:    "high",
		Labels:      []string{"api", "urgent"},
	})

	in := InputOfFields(fields, engine.TaskEdit)
	if in.Given["desc"] != "why it matters" {
		t.Errorf("desc = %q, want the description", in.Given["desc"])
	}
	if in.Given["state"] != "todo" {
		t.Errorf("state = %q, want todo", in.Given["state"])
	}
	if got := in.Listed["label"]; len(got) != 3 || got[0] != "=" || got[1] != "api" || got[2] != "urgent" {
		t.Errorf("label = %v, want the two labels, listed", got)
	}
	for _, screen := range []string{"description", "labels", "column", "time spent"} {
		if _, named := in.Given[screen]; named {
			t.Errorf("%q travelled: that is the screen's word, not the engine's", screen)
		}
	}
}

func TestAskedLeavesANumberNobodyCanReadUnnamed(t *testing.T) {
	if in := InputOfFields([]Field{{Name: "after", Arg: "after", Value: "nine"}}, engine.TaskMove); in.Named("after") {
		t.Fatal("a number nobody can read was passed to the engine anyway")
	}
}

func TestAskedNamesAnEmptyNumber(t *testing.T) {
	if in := InputOfFields([]Field{{Name: "after", Arg: "after", Value: "  "}}, engine.TaskMove); !in.Named("after") {
		t.Fatal("emptying a number said nothing at all")
	}
}

func TestAsked_passesOverAFieldThatIsNoArgument(t *testing.T) {
	in := InputOfFields([]Field{{Name: "note"}, {Name: "title", Arg: "title", Value: "Ship it"}}, engine.TaskEdit)
	if in.Given["title"] != "Ship it" {
		t.Errorf("a field behind one that is no argument was lost: %+v", in.Given)
	}
}

func TestAsked_replacesWhatAListHeld(t *testing.T) {
	in := InputOfFields([]Field{{Name: "labels", Arg: "label", Value: "api urgent"}}, engine.TaskEdit)
	if got := in.Listed["label"]; !slices.Equal(got, []string{"=", "api", "urgent"}) {
		t.Errorf("label = %v, want the form's labels replacing the ones there", got)
	}
}
