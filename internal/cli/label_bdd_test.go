package cli

import (
	"testing"

	"github.com/cucumber/godog"
)

func initLabel(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerStore(sc)
		w.registerTimesheet(sc)
		w.registerLabel(sc)
	}
}

func TestAcceptance_Label(t *testing.T) {
	runAcceptance(t, "label.feature", initLabel(t), "")
}
