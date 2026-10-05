package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"cairn/internal/engine"
)

// updateSchemas writes the pinned schemas again, after a change that meant
// to change them: UPDATE_SCHEMAS=1 go test ./internal/mcp.
var updateSchemas = os.Getenv("UPDATE_SCHEMAS") != ""

// An agent holds on to the shape of what a tool takes and answers, so every
// tool's two schemas, its annotations and its level are pinned, and a change to either is one somebody meant
// — see docs/mcp.md, on what changes in an answer and what does not.
func TestSchemas_changeOnlyWhenSomebodyMeantThemTo(t *testing.T) {
	said := map[string]map[string]any{}
	for _, op := range engine.All() {
		tool := toolFor(op)
		// What a client decides on is pinned with what it sends and reads:
		// whether to ask a person first, and from which level a tool is
		// offered at all.
		said[tool.Name] = map[string]any{
			"input": tool.InputSchema, "output": tool.OutputSchema,
			"annotations": tool.Annotations, "level": int(op.Level),
		}
	}
	got, err := json.MarshalIndent(said, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	path := filepath.Join("testdata", "schemas.json")
	if updateSchemas {
		if err := os.MkdirAll("testdata", 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v — run with UPDATE_SCHEMAS=1 to write it", path, err)
	}
	if string(got) != string(want) {
		t.Fatalf("a tool's schema changed. If that was meant, it is a change an agent sees: say so in the commit and run with UPDATE_SCHEMAS=1.\n--- pinned ---\n%s\n--- now ---\n%s", want, got)
	}
}
