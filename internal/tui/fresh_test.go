package tui

import (
	"cairn/internal/board"
	"testing"

	"cairn/internal/engine"
	"cairn/test"
)

func TestAFormOpenTakesWhatChangedElsewhereButWhatWasTyped(t *testing.T) {
	f := newFake()
	f.cards = []board.Card{{Number: 1, Title: "Ship it", State: "todo"}}
	b := hit(standing(f), "ctrl+l", "e")
	if b.screen != board.OnForm {
		t.Fatalf("e left the board on %v", b.screen)
	}
	b = hit(b, "space", "n", "o", "w")
	focused := b.form.asked.GetFocusedField().GetKey()
	changed := board.Card{Number: 1, Title: "Ship it today", Description: "Two columns", State: "todo",
		Steps: []engine.Part{{Number: 1, Title: "Draft"}}}
	model, _ := b.Update(Looked{Snapshot: test.Searchable(board.Snapshot{Cards: []board.Card{changed}, Version: 2})})
	b = model.(Board)
	if b.screen != board.OnForm || b.form.asked == nil {
		t.Fatalf("a reading made elsewhere closed the form: on %v", b.screen)
	}
	for name, want := range map[string]string{"title": "Ship it now", "description": "Two columns", "steps": "#1 [ ] Draft"} {
		if got := b.form.valueOf(name); got != want {
			t.Errorf("the field %q reads %q, want %q", name, got, want)
		}
	}
	if got := b.form.asked.GetFocusedField().GetKey(); got != focused {
		t.Errorf("the keyboard moved from %q to %q", focused, got)
	}
	if !b.form.Changed() {
		t.Error("what was typed no longer counts as a change")
	}
}
