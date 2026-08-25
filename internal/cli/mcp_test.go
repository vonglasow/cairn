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
