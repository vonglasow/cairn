package cli

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"cairn/internal/config"
	"cairn/internal/engine"
	"cairn/test"
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
	sc.Step(`^the pages say time = (\S+), and focus sessions last "([^"]*)"$`, func(said, length string) error {
		return w.WriteConfig(fmt.Sprintf("[pages]\ntime = %s\n\n[focus]\nsession = %q\n", said, length))
	})
	sc.Step(`^the settings beside the store say time = (\S+)$`, func(said string) error {
		return os.WriteFile(config.SettingsBeside(w.StorePath), fmt.Appendf(nil, "[pages]\ntime = %s\n", said), 0o600)
	})
	sc.Step(`^it is (\d{2}):(\d{2}) on (\d{4})-(\d{2})-(\d{2})$`, func(hour, minute, year, month, day int) error {
		at := time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local)
		w.Setenv(engine.NowVariable, at.Format(time.RFC3339))
		return nil
	})
	sc.Step(`^it says on standard error that the time is "on", "hidden" or "off"$`, func() error {
		return test.WantContains(w.Out.Stderr, `"on", "hidden" or "off"`, "standard error")
	})
}
