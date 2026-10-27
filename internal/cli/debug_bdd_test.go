package cli

import (
	"testing"

	"github.com/cucumber/godog"
)

func initDebug(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerDebug(sc)
	}
}

func TestAcceptance_Debug(t *testing.T) {
	runAcceptance(t, "debug.feature", initDebug(t), "")
}
