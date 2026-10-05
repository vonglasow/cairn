package cli

// Executable acceptance scenarios for bdd/config.feature: where cairn finds
// the file that says its two places, and what happens when it says nothing,
// says something wrong, or is not there at all.
//
// The whole file runs, and not a list of its tags. A list is what let half of
// it go unrun: a scenario added to the feature without its line here did not
// fail, it simply never executed. What is written but not built yet carries
// @later in the feature itself, where whoever reads the scenario sees it.

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
