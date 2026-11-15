package tui

import (
	"strings"
	"testing"
)

func TestTheFootPutsTheQuestionItIsWaitingOn(t *testing.T) {
	b := hit(standing(twoTasks()), "ctrl+l", "D")
	drawn := b.Draw()
	if !strings.Contains(drawn, "delete task 1?") {
		t.Fatalf("the foot does not ask about the deletion:\n%s", drawn)
	}
	if !strings.Contains(drawn, "keep") {
		t.Fatalf("the foot does not say how to back out:\n%s", drawn)
	}
}
