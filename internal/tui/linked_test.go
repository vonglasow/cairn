package tui

import (
	"reflect"
	"testing"

	"cairn/internal/engine"
)

func TestOOnACardOpensALinkedFormInTodo(t *testing.T) {
	f := newFake()
	f.cards = []engine.Card{{Number: 7, Title: "Ship it", State: "done", Priority: "normal"}}
	f.goals = []engine.Goal{{Number: 3, Title: "Ship it", State: "active"}}
	b := hit(standing(f), "l", "l", "l", "i", "o")
	if b.screen != OnForm || !b.form.Adding {
		t.Fatalf("o on a card left the board on %v, adding %v", b.screen, b.form.Adding)
	}
	if !reflect.DeepEqual(b.form.fields, engine.LinkedTaskFields(7)) {
		t.Fatalf("the form asks %+v, want the fields of a task linked to 7", b.form.fields)
	}
	if got := b.form.valueOf("column"); got != "todo" {
		t.Errorf("the column reads %q, want todo", got)
	}
	if got := hit(standing(f), "o"); got.screen != OnTheBoard {
		t.Errorf("o in normal mode opened %v", got.screen)
	}
	if got := hit(standing(f), "tab", "i", "o"); got.screen != OnTheBoard {
		t.Errorf("o on the objectives opened %v", got.screen)
	}
	if got := hit(standing(newFake()), "i", "o"); got.screen != OnTheBoard {
		t.Errorf("o with nothing held opened %v", got.screen)
	}
}
