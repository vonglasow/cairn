package cli

// Executable acceptance scenarios for bdd/trace.feature. The Gherkin is the
// contract; the steps every feature shares live in world_test.go, and what
// belongs to the trace is declared in trace_steps_test.go.

import (
	"testing"

	"github.com/cucumber/godog"
)

func initTrace(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerTrace(sc)
	}
}

func TestAcceptance_Trace(t *testing.T) {
	runAcceptance(t, "trace.feature", initTrace(t), "")
}
