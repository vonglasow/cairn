package cli

// Executable acceptance scenarios for bdd/backup.feature. The Gherkin is the
// contract; the steps live in world_test.go and store_steps_test.go, shared
// with the other features.

import (
	"testing"

	"github.com/cucumber/godog"
)

func initBackup(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerStore(sc)
	}
}

func TestAcceptance_Backup(t *testing.T) {
	runAcceptance(t, "backup.feature", initBackup(t), "")
}
