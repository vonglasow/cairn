package cli

// Executable acceptance scenarios for bdd/ordering.feature. The Gherkin is the
// contract; the steps live in world_test.go, shared with the other features.

import (
	"testing"

	"github.com/cucumber/godog"
)

func initOrdering(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerOrdering(sc)
	}
}

func TestAcceptance_Ordering(t *testing.T) {
	runAcceptance(t, "ordering.feature", initOrdering(t), "")
}
