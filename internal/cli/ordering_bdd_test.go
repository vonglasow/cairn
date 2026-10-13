package cli

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
