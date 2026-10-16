package cli

import (
	"testing"

	"github.com/cucumber/godog"
)

func initStep(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerTrace(sc)
	}
}

func TestAcceptance_Step(t *testing.T) {
	runAcceptance(t, "step.feature", initStep(t), "")
}
