package cli

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
