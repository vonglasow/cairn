package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/store/files"
	"cairn/test"
)

func initExchange(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerStore(sc)
		w.registerExchange(sc)
		w.registerLabel(sc)
	}
}

func TestAcceptance_Exchange(t *testing.T) {
	runAcceptance(t, "exchange.feature", initExchange(t), "")
}

func (w *world) registerExchange(sc *godog.ScenarioContext) {
	sc.Step(`^I export the tasks to a file of my own$`, func() error {
		w.file = filepath.Join(w.T.TempDir(), "elsewhere", "tasks.json")
		w.Run("cairn task export " + w.file)
		return nil
	})
	sc.Step(`^a file of my own exported from a store where these were run:$`, func(run *godog.DocString) error {
		mine := w.StorePath
		other := filepath.Join(w.T.TempDir(), "other", "cairn.db")
		w.file = filepath.Join(w.T.TempDir(), "other", "tasks.json")
		if settings, err := os.ReadFile(files.SettingsBeside(mine)); err == nil {
			if err := os.MkdirAll(filepath.Dir(other), 0o700); err != nil {
				return err
			}
			if err := os.WriteFile(files.SettingsBeside(other), settings, 0o600); err != nil {
				return err
			}
		}
		w.Setenv(files.StoreVariable, other)
		defer w.Setenv(files.StoreVariable, mine)
		for _, line := range strings.Split(strings.TrimSpace(run.Content), "\n") {
			if err := w.Seed(strings.TrimSpace(line)); err != nil {
				return err
			}
		}
		return w.Seed("cairn task export " + w.file)
	})
	sc.Step(`^I import the tasks of that file$`, func() error {
		w.Run("cairn task import " + w.file)
		return nil
	})
	sc.Step(`^it says (\d+) tasks? and (\d+) objectives? (?:was|were) added$`, func(tasks, objectives int) error {
		said := fmt.Sprintf("%d task(s) and %d objective(s) were added from %s", tasks, objectives, w.file)
		return test.WantContains(w.Out.Stdout, said, "the answer")
	})
}
