package mcp

import (
	"errors"
	"strings"
	"testing"
)

func TestParseAccess_acceptsTheThreeLevels(t *testing.T) {
	for name, want := range map[string]Access{
		"read":  Read,
		"write": Write,
		"full":  Full,
	} {
		got, err := ParseAccess(name)
		if err != nil {
			t.Errorf("ParseAccess(%q): %v", name, err)
			continue
		}
		if got != want {
			t.Errorf("ParseAccess(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestParseAccess_refusesWhatItCannotMakeSenseOf(t *testing.T) {
	_, err := ParseAccess("sideways")
	if !errors.Is(err, ErrNotAnAccess) {
		t.Fatalf("ParseAccess(sideways) err = %v, want ErrNotAnAccess", err)
	}
	got := err.Error()
	for _, want := range []string{`"sideways"`, "read", "write", "full"} {
		if !strings.Contains(got, want) {
			t.Errorf("err = %q, want it to mention %q", got, want)
		}
	}
}
