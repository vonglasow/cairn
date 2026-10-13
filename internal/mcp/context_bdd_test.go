package mcp_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"

	"cairn/test"
)

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

func (w *world) registerContext(sc *godog.ScenarioContext) {
	sc.Step(`^a configuration that says the notes live at "([^"]*)"$`, func(place string) error {
		return w.WriteConfig(fmt.Sprintf("notes = %q\n", place))
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
