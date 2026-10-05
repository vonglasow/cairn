package cli

import (
	"strings"
	"testing"
)

// TestMCP_groupNeedsAVerb checks that `mcp` on its own refuses rather than
// printing its help and exiting zero, and that the refusal names the group
// and says how it is used — exactly as `task` does.
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

// TestMCPServe_refusesAnAccessItCannotMakeSenseOf checks that `mcp serve
// --access sideways` fails, names what was typed, and exits as a request that
// was understood and refused — not as one nobody could parse.
func TestMCPServe_refusesAnAccessItCannotMakeSenseOf(t *testing.T) {
	r := run(t, "mcp", "serve", "--access", "sideways")
	if r.Code != exitRefused {
		t.Fatalf("mcp serve --access sideways exit = %d, want %d (%s)", r.Code, exitRefused, r.Stderr)
	}
	if !strings.Contains(strings.ToLower(r.Stderr), "sideways") {
		t.Errorf("stderr = %q, want it to name \"sideways\"", r.Stderr)
	}
}

// An agent reaches only as far as it was trusted to on purpose: nothing given
// is looking and nothing else.
func TestMCPServe_looksOnlyUnlessToldOtherwise(t *testing.T) {
	serve, _, err := NewRootCmd().Find([]string{"mcp", "serve"})
	if err != nil {
		t.Fatal(err)
	}
	if got := serve.Flags().Lookup("access").DefValue; got != "read" {
		t.Errorf("--access defaults to %q, want read", got)
	}
}

// `mcp` on its own was not understood, exactly as `task` on its own is not.
func TestMCP_groupWithoutAVerbIsNotUnderstood(t *testing.T) {
	if r := run(t, "mcp"); r.Code != exitUsage {
		t.Errorf("exit = %d, want %d", r.Code, exitUsage)
	}
}
