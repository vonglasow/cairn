package engine

import "testing"

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
