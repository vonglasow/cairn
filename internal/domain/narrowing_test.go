package domain

import (
	"slices"
	"testing"
)

type heard struct{ said []string }

func (h *heard) InColumn(s string)   { h.said = append(h.said, "column "+s) }
func (h *heard) Carrying(p string)   { h.said = append(h.said, "priority "+p) }
func (h *heard) DueBefore(d string)  { h.said = append(h.said, "due "+d) }
func (h *heard) Serving(n int)       { h.said = append(h.said, "serving "+string(rune('0'+n))) }
func (h *heard) Labelled(l []string) { h.said = append(h.said, "labels "+l[0]) }
func (h *heard) Archived()           { h.said = append(h.said, "archived") }
func (h *heard) WithArchived()       { h.said = append(h.said, "both") }

func TestNarrowing_tellsWhatWasNamedInTheOrderItWasNamed(t *testing.T) {
	var n Narrowing
	n.InColumn("done")
	n.Carrying("high")
	n.DueBefore("2026-10-01")
	n.Serving(3)
	n.Labelled([]string{"spec"})
	n.Archived()
	n.WithArchived()
	h := &heard{}
	n.Tell(h)
	want := []string{"column done", "priority high", "due 2026-10-01", "serving 3", "labels spec", "archived", "both"}
	if !slices.Equal(h.said, want) {
		t.Errorf("told %v, want %v", h.said, want)
	}
}

func TestNarrowing_narrowsOnlyWhenSomethingWasNamed(t *testing.T) {
	var nothing Narrowing
	if nothing.Narrows() {
		t.Error("a narrowing nobody named anything on narrows")
	}
	var both Narrowing
	both.WithArchived()
	if both.Narrows() {
		t.Error("asking for both sides of the board narrows")
	}
	var off Narrowing
	off.Archived()
	if !off.Narrows() {
		t.Error("asking for what was taken off narrows nothing")
	}
	var column Narrowing
	column.InColumn("todo")
	if !column.Narrows() {
		t.Error("a column named narrows nothing")
	}
}
