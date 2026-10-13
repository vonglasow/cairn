package engine

import "testing"

// A board searches what it already holds, with what somebody typed: words
// apart, every one of them carried, and nothing typed reaching nothing.
func TestReaches_readsWhatWasTypedAsWords(t *testing.T) {
	said := []string{"Write the spec", "backend"}
	for typed, want := range map[string]bool{
		"spec backend": true,
		"SPEC":         true,
		"spec lamp":    false,
		"":             false,
		"   ":          false,
	} {
		if got := Reaches(said, typed); got != want {
			t.Errorf("Reaches(%q) = %v, want %v", typed, got, want)
		}
	}
}
