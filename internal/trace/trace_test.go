package trace

import (
	"testing"
	"time"
)

func built() Act {
	var a AsTold
	a.On(4)
	a.Did(Changed)
	a.Changing("priority")
	a.Was("normal")
	a.Became("high")
	a.Through("mcp")
	a.At(time.Date(2026, 3, 3, 9, 15, 0, 0, time.UTC))
	return a.Act()
}

func TestAsTold_losesNothingOnTheWayThrough(t *testing.T) {
	first := built()
	again := From(first)
	if first.Carried() != again.Act().Carried() {
		t.Errorf("built twice differs:\n  %+v\n  %+v", first.Carried(), again.Act().Carried())
	}
}

func TestLastNamed_isTheLastActThatNamedTheField(t *testing.T) {
	acts := []Act{
		acted(1, Created, "column", "", "backlog", "", 9, 0),
		acted(1, Changed, "column", "backlog", "todo", "", 10, 0),
		acted(1, Changed, "priority", "normal", "high", "", 11, 0),
	}
	if got := LastNamed(acts, "column"); !got.Equal(acts[1].at) {
		t.Errorf("LastNamed = %v, want the 10:00 change", got)
	}
	if got := LastNamed(acts, "deadline"); !got.IsZero() {
		t.Errorf("LastNamed of a field never named = %v, want nothing", got)
	}
}
