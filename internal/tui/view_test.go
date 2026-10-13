package tui

import (
	"strings"
	"testing"
)

func TestTheFootPutsTheQuestionItIsWaitingOn(t *testing.T) {
	b := hit(standing(twoTasks()), "l", "i", "d")
	drawn := b.Draw()
	if !strings.Contains(drawn, "delete task 1?") {
		t.Fatalf("the foot does not ask about the deletion:\n%s", drawn)
	}
	if !strings.Contains(drawn, "keep") {
		t.Fatalf("the foot does not say how to back out:\n%s", drawn)
	}
}

func TestTheDetailServesTheNavigationItOffers(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i", "enter")
	if b.Where().Task != 1 {
		t.Fatalf("the detail opened on %d, want task 1", b.Where().Task)
	}
	b = hit(b, "j")
	if b.Where().Task != 2 {
		t.Fatalf("j in the detail went to %d, want task 2", b.Where().Task)
	}
}
