package reading

import (
	"testing"

	"cairn/internal/engine"
)

// Turning a form into what an operation takes belongs here and not to a
// surface: it reads each field's own declaration of which argument it fills,
// and the operation's declaration of what kind that argument is. Two boards
// ask it the same question.
func TestAskedNamesEveryFieldByItsEngineArgument(t *testing.T) {
	fields := TaskFields(Card{
		Title:       "One",
		Description: "why it matters",
		State:       "todo",
		Priority:    "high",
		Labels:      []string{"api", "urgent"},
	})

	in := Asked(fields, engine.TaskEdit)
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

// A number nobody can read is not named at all, rather than sent for the
// engine to refuse: one unreadable field would otherwise lose the whole form.
func TestAskedLeavesANumberNobodyCanReadUnnamed(t *testing.T) {
	if in := Asked([]Field{{Name: "parent", Arg: "parent", Value: "nine"}}, engine.TaskEdit); in.Named("parent") {
		t.Fatal("a parent nobody can read was passed to the engine anyway")
	}
}

// Empty is not unreadable. It reads as zero, which is how such a field is
// cleared, and the engine reads it back the same way.
func TestAskedNamesAnEmptyNumber(t *testing.T) {
	if in := Asked([]Field{{Name: "parent", Arg: "parent", Value: "  "}}, engine.TaskEdit); !in.Named("parent") {
		t.Fatal("emptying the parent said nothing at all")
	}
}
