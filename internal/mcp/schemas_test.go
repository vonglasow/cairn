package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"cairn/internal/engine"
)

var updateSchemas = os.Getenv("UPDATE_SCHEMAS") != ""

func TestSchemas_changeOnlyWhenSomebodyMeantThemTo(t *testing.T) {
	said := map[string]map[string]any{}
	for _, op := range engine.All() {
		tool := toolFor(op)
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
