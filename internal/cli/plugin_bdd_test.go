package cli

import (
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/build"
)

func initPlugin(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		sc.Step(`^the cairn I run is (\S+), (?:from a release|built from its source)$`, func(version string) {
			w.stamped = func() build.Info { return build.Stamped(version, "", "", "") }
		})
	}
}

func TestAcceptance_Plugin(t *testing.T) {
	runAcceptance(t, "claude-plugin.feature", initPlugin(t), "@plugin-installed-from-its-tag,@plugin-installed-from-a-build")
}
