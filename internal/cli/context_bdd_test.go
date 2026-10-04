package cli

// Executable acceptance scenarios for bdd/context.feature: where a task's
// outside context lives, said by showing the task. The one scenario about the
// server runs in internal/mcp, where a session can be opened.

import (
	"testing"

	"github.com/cucumber/godog"
)

func initContext(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerConfig(sc)
		w.registerContext(sc)
	}
}

func TestAcceptance_Context(t *testing.T) {
	runAcceptance(t, "context.feature", initContext(t), "~@context-travels-in-the-json")
}
