package cli

import (
	"testing"

	"github.com/cucumber/godog"
)

func initTask(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerBlocking(sc)
		w.registerLinking(sc)
		w.registerTimesheet(sc)
	}
}

func TestAcceptance_Task(t *testing.T) {
	runAcceptance(t, "task.feature", initTask(t), "")
}
