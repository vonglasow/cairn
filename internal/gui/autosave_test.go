package gui

import (
	"slices"
	"testing"

	"cairn/internal/reading"
)

// A part struck out is a change a beat has to see: two forms holding the same
// text, one with a part struck, hold different things.
func TestAFormsValuesSayWhatWasStruck(t *testing.T) {
	plain := Form{fields: []reading.Field{{Name: "title", Value: "Ship it"}}}
	struck := plain
	struck.struck = []int{3}
	if slices.Equal(plain.values(), struck.values()) {
		t.Errorf("a struck part changes nothing a beat can see: %q", struck.values())
	}
	if got := struck.values(); !slices.Equal(got, []string{"Ship it", "struck 3"}) {
		t.Errorf("the form holds %q", got)
	}
}
