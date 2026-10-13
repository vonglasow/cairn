package cli

import (
	"testing"

	"github.com/cucumber/godog"
)

func initTimesheet(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerTimesheet(sc)
	}
}

func TestAcceptance_Timesheet(t *testing.T) {
	runAcceptance(t, "timesheet.feature", initTimesheet(t), "")
}
