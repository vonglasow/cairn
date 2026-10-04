package mcp_test

// The one scenario of bdd/context.feature that is about the server: an agent
// showing a task is told where its outside context lives, in the shape it
// reads. The rest of that feature runs in internal/cli.

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"

	"cairn/test"
)

// TestAcceptance_Context_TravelsInTheJSON runs it.
func TestAcceptance_Context_TravelsInTheJSON(t *testing.T) {
	init := func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.Register(sc)
		w.registerContext(sc)
	}
	if test.Suite(t, "context.feature", init, "@context-travels-in-the-json").Run() != 0 {
		t.Fatal("acceptance @context-travels-in-the-json failed")
	}
}

// registerContext declares what that scenario needs beyond the shared steps.
func (w *world) registerContext(sc *godog.ScenarioContext) {
	sc.Step(`^a configuration that says the notes live at "([^"]*)"$`, func(place string) error {
		path := w.ConfigPath()
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return err
		}
		return os.WriteFile(path, fmt.Appendf(nil, "notes = %q\n", place), 0o600)
	})

	sc.Step(`^I ask the server to show task (\d+)$`, func(n int) error {
		if err := w.startedServe("--access", "read"); err != nil {
			return err
		}
		_ = w.callTool("task_show", map[string]any{"number": n})
		return nil
	})

	sc.Step(`^the structured answer names where the context lives$`, func() error {
		if w.callErr != nil {
			return fmt.Errorf("the call failed: %w", w.callErr)
		}
		answer, ok := w.result.StructuredContent.(map[string]any)
		if !ok {
			return fmt.Errorf("the answer is not an object: %#v", w.result.StructuredContent)
		}
		want := filepath.Join(w.Defaults, "vault", "1")
		if answer["context"] != want {
			return fmt.Errorf("the answer names %v as the context, want %q", answer["context"], want)
		}
		return nil
	})
}
