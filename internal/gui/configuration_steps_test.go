package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/config"
)

func (w *world) configurationSteps(sc *godog.ScenarioContext) {
	sc.Given(`^a configuration file naming the store$`, func() error {
		path := config.FileIn(w.home)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(fmt.Sprintf("store = %q\n", w.path)), 0o600)
	})
	sc.Then(`^the configuration reads "([^"]*)" from the file$`, func(name string) error {
		for _, line := range w.w.Configured() {
			if line.Key == name {
				if !strings.HasSuffix(line.Means, "(the file)") {
					return fmt.Errorf("%q reads %q, not from the file", name, line.Means)
				}
				return nil
			}
		}
		return fmt.Errorf("the configuration holds no %q: %v", name, w.w.Configured())
	})
	sc.Then(`^the window asks the system to open the configuration file$`, func() error {
		got, asked := w.w.OpenAsked()
		if !asked {
			return fmt.Errorf("the window asked the system to open nothing")
		}
		if want := config.FileIn(w.home); got != want {
			return fmt.Errorf("the window asks the system to open %q, not %q", got, want)
		}
		return nil
	})
	sc.Then(`^a configuration file exists$`, func() error {
		if _, err := os.Stat(config.FileIn(w.home)); err != nil {
			return fmt.Errorf("no configuration file: %w", err)
		}
		return nil
	})
}
