package cli

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
