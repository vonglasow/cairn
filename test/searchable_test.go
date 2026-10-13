package test

import (
	"testing"

	"cairn/internal/reading"
)

// A card or an objective built by hand is reached by what it carries once it
// says what the engine says it carries, and by nothing else.
func TestSearchable_saysWhatTheEngineSays(t *testing.T) {
	s := Searchable(reading.Snapshot{
		Cards: []reading.Card{{Number: 1, Title: "Write the spec", State: "todo", Labels: []string{"backend"}, Estimate: "90m"}},
		Goals: []reading.Goal{{Number: 2, Title: "Ship cairn", State: "on-hold", Priority: "high"}},
	})
	for _, typed := range []string{"spec", "todo", "backend", "1h30m"} {
		if !reading.ReachesCard(s.Cards[0], typed) {
			t.Errorf("%q does not reach the card", typed)
		}
	}
	for _, typed := range []string{"cairn", "on-hold", "high", "2"} {
		if !reading.ReachesGoal(s.Goals[0], typed) {
			t.Errorf("%q does not reach the objective", typed)
		}
	}
	if reading.ReachesCard(s.Cards[0], "lamp") {
		t.Error("a word the card does not carry reached it")
	}
}
