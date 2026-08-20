package cli

// Executable acceptance scenarios for bdd/task.feature. The Gherkin is the
// contract; the steps live in world_test.go, shared with the other features.

import (
	"testing"

	"github.com/cucumber/godog"
)

func initTask(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		newWorld(t).register(sc)
	}
}

func TestAcceptance_Task(t *testing.T) {
	runAcceptance(t, "task.feature", initTask(t), "")
}
