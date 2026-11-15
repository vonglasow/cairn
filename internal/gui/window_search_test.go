package gui

import (
	"cairn/internal/board"
	"testing"
)

func TestBackspaceInASearchTakesBackOneCharacter(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "réunion", State: "backlog"})

	w = w.Key("/").Types("réu")
	if got := w.Key("backspace").Looking.Reads(); got != "ré" {
		t.Fatalf("taking one back left %q", got)
	}
	if got := w.Key("backspace").Key("backspace").Looking.Reads(); got != "r" {
		t.Errorf("taking two back left %q", got)
	}
	empty := w.Key("esc").Key("/")
	for range 3 {
		empty = empty.Key("backspace")
	}
	if got := empty.Looking.Reads(); got != "" {
		t.Errorf("an empty search reads %q", got)
	}
	if got := w.Key("esc").Key("backspace").Looking.Reads(); got != "" {
		t.Errorf("backspace on the board left a search of %q", got)
	}
}
