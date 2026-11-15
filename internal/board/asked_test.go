package board

import (
	"cairn/internal/engine"
	"testing"
)

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
