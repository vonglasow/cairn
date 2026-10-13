package engine

import (
	"errors"
	"strings"
	"testing"
)

func TestParseLevel_acceptsTheThreeLevels(t *testing.T) {
	for name, want := range map[string]Level{
		"read":  Read,
		"write": Write,
		"full":  Full,
	} {
		got, err := ParseLevel(name)
		if err != nil {
			t.Errorf("ParseLevel(%q): %v", name, err)
			continue
		}
		if got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestParseLevel_refusesWhatItCannotMakeSenseOf(t *testing.T) {
	_, err := ParseLevel("sideways")
	if !errors.Is(err, ErrNotALevel) {
		t.Fatalf("ParseLevel(sideways) err = %v, want ErrNotALevel", err)
	}
	got := err.Error()
	for _, want := range []string{`"sideways"`, "read", "write", "full"} {
		if !strings.Contains(got, want) {
			t.Errorf("err = %q, want it to mention %q", got, want)
		}
	}
}
