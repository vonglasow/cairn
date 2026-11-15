package cli

import (
	"fmt"
	"testing"

	"github.com/cucumber/godog"
)

func initTime(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerConfig(sc)
		w.registerTimesheet(sc)
		w.registerHelp(sc)
		w.registerTime(sc)
	}
}

func TestAcceptance_Time(t *testing.T) {
	runAcceptance(t, "time.feature", initTime(t), "")
}

func (w *world) registerTime(sc *godog.ScenarioContext) {
	sc.Step(`^the pages say time = (\S+)$`, func(said string) error {
		return w.WriteConfig(fmt.Sprintf("[pages]\ntime = %s\n", said))
	})
}
