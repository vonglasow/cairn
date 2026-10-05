package domain

import (
	"slices"
	"testing"
	"time"
)

// An absence is a fact that gets told: a value nobody gave is said to be
// missing, never left for whoever listens to notice. The same few lines were
// written out wherever a cell, or the store building one, said what it held.
func TestTellOrNot_saysAValueOrThatThereIsNone(t *testing.T) {
	var said []string
	TellOrNot("spec", func(s string) { said = append(said, "told "+s) }, func() { said = append(said, "none") })
	TellOrNot("", func(s string) { said = append(said, "told "+s) }, func() { said = append(said, "none") })
	TellOrNot(time.Duration(0), func(time.Duration) { said = append(said, "told a length") }, func() { said = append(said, "no length") })
	if want := []string{"told spec", "none", "no length"}; !slices.Equal(said, want) {
		t.Errorf("said %v, want %v", said, want)
	}
}

// A set is told one member at a time, said to be empty when it is, and closed
// either way, so whoever listens knows when it is complete.
func TestTellAll_tellsEachMemberThenClosesTheSet(t *testing.T) {
	var said []string
	each := func(s string) { said = append(said, s) }
	none := func() { said = append(said, "none") }
	closed := func() { said = append(said, "closed") }
	TellAll([]string{"a", "b"}, each, none, closed)
	TellAll(nil, each, none, closed)
	if want := []string{"a", "b", "closed", "none", "closed"}; !slices.Equal(said, want) {
		t.Errorf("said %v, want %v", said, want)
	}
}
