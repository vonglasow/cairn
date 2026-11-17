package gui

import (
	"cairn/internal/board"
	"maps"
	"strconv"
	"testing"
)

func TestShiftAndTheArrowsCarryTheCardAndLettingShiftGoDropsIt(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "First", "todo"), card(2, "Second", "todo"), card(3, "Third", "current"))
	w = w.clicked(middleOf(w, 1, 0)).Key("shift+down")
	l := w.Layout()
	if land, ok := w.Landing(); !w.Dragging() || !ok || land != LineAbove(l.CardAt(1, 2, 0)) || len(f.wrote()) != 0 {
		t.Fatalf("⇧↓ carries %v, lands at %v %v, wrote %v", w.Dragging(), land, ok, f.wrote())
	}
	if said := w.CommandSays(); said != "Carrying #1 — let ⇧ go to drop it, esc to put it back" {
		t.Errorf("the bar says %q while a card is carried", said)
	}
	if w = w.Key("shift+right"); len(f.wrote()) != 0 {
		t.Fatalf("⇧→ wrote %v", f.wrote())
	}
	if land, _ := w.Landing(); land != LineAbove(l.CardAt(2, 1, 0)) {
		t.Errorf("⇧→ lands at %v, want under the card of current", land)
	}
	if w = w.ShiftLetGo(); w.Dragging() {
		t.Error("letting ⇧ go left the card carried")
	}
	in, ok := f.gave("task", "move")
	if !ok || in.Given["number"] != "1" || in.Given["state"] != "current" || in.Given["after"] != "3" {
		t.Errorf("the drop asked %+v", in)
	}
	f.asked = nil
	w = w.clicked(middleOf(w, 1, 0)).Key("shift+down").Key("esc")
	if w.Dragging() || len(f.wrote()) != 0 || w.Screen != board.OnTheBoard {
		t.Errorf("esc left it carried %v, wrote %v", w.Dragging(), f.wrote())
	}
	if w = w.ShiftLetGo(); len(f.wrote()) != 0 {
		t.Errorf("letting ⇧ go after esc wrote %v", f.wrote())
	}
}

func TestWhereTheKeyboardDropsACard(t *testing.T) {
	todo := []board.Card{card(1, "First", "todo"), card(2, "Second", "todo")}
	for _, c := range []struct {
		name   string
		cards  []board.Card
		chosen int
		keys   []string
		want   map[string]string
	}{
		{"sideways into the next column", todo, 1, []string{"shift+right"}, map[string]string{"state": "current", "first": "true"}},
		{"sideways back", todo, 1, []string{"shift+left"}, map[string]string{"state": "backlog", "first": "true"}},
		{"off the last column, nowhere", []board.Card{card(9, "Last", "done")}, 9, []string{"shift+right"}, nil},
		{"down behind the next", todo, 1, []string{"shift+down"}, map[string]string{"after": "2"}},
		{"up to the head", todo, 2, []string{"shift+up"}, map[string]string{"first": "true"}},
		{"down past the column into its lane", todo, 2, []string{"shift+down", "shift+down"}, map[string]string{"state": "blocked", "first": "true"}},
		{"out of the band to the left", []board.Card{card(1, "First", "todo"), card(2, "Blocked", "blocked")}, 2, []string{"shift+left"}, map[string]string{"state": "backlog", "first": "true"}},
		{"up and back, nowhere", todo, 1, []string{"shift+down", "shift+up"}, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newFake()
			w := holding(f, c.cards...)
			w.Chosen = c.chosen
			for _, k := range c.keys {
				w = w.Key(k)
			}
			w.ShiftLetGo()
			in, ok := f.gave("task", "move")
			if c.want == nil {
				if ok {
					t.Errorf("asked %+v, want nothing", in)
				}
				return
			}
			c.want["number"] = strconv.Itoa(c.chosen)
			if !ok || !maps.Equal(in.Given, c.want) {
				t.Errorf("asked %+v, want %v", in.Given, c.want)
			}
		})
	}
}

func TestShiftStandsOnACapOfItsOwn(t *testing.T) {
	for key, want := range map[string]int{
		"⇧F":    capWide("⇧") + capGap + capWide("F"),
		"⇧←→↑↓": capWide("⇧") + capGap + capWide("←→↑↓"),
		"x ⇧X":  capWide("x") + capGap + capWide("⇧") + capGap + capWide("X"),
		"⌘K":    capWide("⌘K"),
	} {
		if got := KeyChipWide(key); got != want {
			t.Errorf("%q is %d wide, want %d", key, got, want)
		}
	}
}
