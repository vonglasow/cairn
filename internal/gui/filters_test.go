package gui

import (
	"slices"
	"strings"
	"testing"

	"cairn/internal/engine"
)

func TestAnObjectiveTooLongForTheFilterIsCutAndShownWholeUnderThePointer(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Spec", State: "todo", Objectives: []int{1}})
	long := "#1 Objective: late, its deadline past and nobody left to say why"
	w.snap.Goals = []engine.Goal{{Number: 1, Title: strings.TrimPrefix(long, "#1 ")}}
	w = w.Key("shift+f").Key("o")
	columns, _ := w.FilterColumns()
	for _, column := range columns {
		if column.Box.W != filterColumnWide {
			t.Errorf("column %d is %d wide, want every column %d", column.At, column.Box.W, filterColumnWide)
		}
	}
	values, open := w.FilterList()
	at := slices.Index(w.Filters()[w.FilterOn()].Shown, long)
	if !open || at < 0 {
		t.Fatalf("the objective is not offered: %v", w.Filters()[w.FilterOn()].Shown)
	}
	named, count := FilterValueParts(values[at])
	if named.X+named.W > count.X {
		t.Errorf("the name runs to %d, over its count at %d", named.X+named.W, count.X)
	}
	w = w.Hovering(Point{X: named.X + named.W/2, Y: named.Y + named.H/2})
	if w.Bubble() != long {
		t.Errorf("the pointer on the cut title shows %q", w.Bubble())
	}
}
