package gui

import (
	"cairn/internal/engine"
	"slices"
	"testing"
)

func TestAFormsValuesSayWhatWasStruck(t *testing.T) {
	plain := Form{fields: []engine.Field{{Name: "title", Value: "Ship it"}}}
	struck := plain
	struck.struck = []int{3}
	if slices.Equal(plain.values(), struck.values()) {
		t.Errorf("a struck part changes nothing a beat can see: %q", struck.values())
	}
	if got := struck.values(); !slices.Equal(got, []string{"Ship it", "struck 3"}) {
		t.Errorf("the form holds %q", got)
	}
}
