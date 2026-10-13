package cli

import (
	"strings"
	"testing"
)

func TestMCP_groupNeedsAVerb(t *testing.T) {
	r := run(t, "mcp")
	if r.Code == exitOK {
		t.Fatalf("mcp on its own should fail, got exit %d", r.Code)
	}
	if !strings.Contains(strings.ToLower(r.Stderr), "mcp needs a command") {
		t.Errorf("stderr = %q, want it to name the group and that it needs a command", r.Stderr)
	}
	if !strings.Contains(strings.ToLower(r.Stderr), "usage") {
		t.Errorf("stderr = %q, want it to say how it is used", r.Stderr)
	}
}

func TestMCPServe_refusesAnAccessItCannotMakeSenseOf(t *testing.T) {
	r := run(t, "mcp", "serve", "--access", "sideways")
	if r.Code != exitRefused {
		t.Fatalf("mcp serve --access sideways exit = %d, want %d (%s)", r.Code, exitRefused, r.Stderr)
	}
	if !strings.Contains(strings.ToLower(r.Stderr), "sideways") {
		t.Errorf("stderr = %q, want it to name \"sideways\"", r.Stderr)
	}
}

func TestMCPServe_looksOnlyUnlessToldOtherwise(t *testing.T) {
	serve, _, err := NewRootCmd().Find([]string{"mcp", "serve"})
	if err != nil {
		t.Fatal(err)
	}
	if got := serve.Flags().Lookup("access").DefValue; got != "read" {
		t.Errorf("--access defaults to %q, want read", got)
	}
}

func TestMCP_groupWithoutAVerbIsNotUnderstood(t *testing.T) {
	if r := run(t, "mcp"); r.Code != exitUsage {
		t.Errorf("exit = %d, want %d", r.Code, exitUsage)
	}
}
