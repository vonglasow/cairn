package cli

// Executable acceptance scenarios for bdd/migration.feature. The Gherkin is the
// contract; the steps live in world_test.go and store_steps_test.go, shared
// with the other features.

import (
	"testing"

	"github.com/cucumber/godog"
)

func initMigration(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerStore(sc)
	}
}

func TestAcceptance_Migration(t *testing.T) {
	runAcceptance(t, "migration.feature", initMigration(t), "")
}
