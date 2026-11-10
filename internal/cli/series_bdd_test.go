package cli

import (
	"testing"

	"github.com/cucumber/godog"
)

func initSeries(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerStore(sc)
		w.registerTimesheet(sc)
		sc.Step(`^it says on standard error "([^"]*)"$`, w.itSaysOnStandardError)
	}
}

func TestAcceptance_Series(t *testing.T) {
	runAcceptance(t, "series.feature", initSeries(t), "")
}
