package test

import (
	"cairn/internal/board"
	"testing"
)

func TestSearchable_saysWhatTheEngineSays(t *testing.T) {
	s := Searchable(board.Snapshot{
		Cards: []board.Card{{Number: 1, Title: "Write the spec", State: "todo", Labels: []string{"backend"}, Estimate: "90m"}},
		Goals: []board.Goal{{Number: 2, Title: "Ship cairn", State: "on-hold", Priority: "high"}},
	})
	for _, typed := range []string{"spec", "todo", "backend", "1h30m"} {
		if !board.ReachesCard(s.Cards[0], typed) {
			t.Errorf("%q does not reach the card", typed)
		}
	}
	for _, typed := range []string{"cairn", "on-hold", "high", "2"} {
		if !board.ReachesGoal(s.Goals[0], typed) {
			t.Errorf("%q does not reach the objective", typed)
		}
	}
	if board.ReachesCard(s.Cards[0], "lamp") {
		t.Error("a word the card does not carry reached it")
	}
}
