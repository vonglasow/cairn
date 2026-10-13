package cli

import (
	"testing"

	"github.com/cucumber/godog"
)

func initConfig(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerConfig(sc)
		w.registerContext(sc)
	}
}

func TestAcceptance_Config(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "~@later")
}
