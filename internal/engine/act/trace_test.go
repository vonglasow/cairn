package act

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
