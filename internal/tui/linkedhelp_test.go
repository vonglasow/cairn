package tui

import (
	"strings"
	"testing"

	"cairn/internal/engine"
)

func TestTheBarAndTheHelpNameO(t *testing.T) {
	f := newFake()
	f.cards = []engine.Card{{Number: 7, Title: "Ship it", State: "todo", Priority: "normal"}}
	b := hit(standing(f), "l", "i")
	if bar := strings.Join(b.offers(), " "); !strings.Contains(bar, "linked") {
		t.Errorf("the insert bar reads %q, and does not offer o", bar)
	}
	if help := strings.Join(b.helpKeys(), "\n"); !strings.Contains(help, "a task linked to it") {
		t.Errorf("the help does not name o: %q", help)
	}
}
