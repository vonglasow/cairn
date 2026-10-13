package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/config"
	"cairn/test"
)

func initExchange(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerStore(sc)
		w.registerExchange(sc)
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
	sc.Step(`^I export the tasks to that file$`, func() error {
		w.Run("cairn task export " + w.file)
		return nil
	})
	sc.Step(`^a file of my own that holds no tasks$`, func() error {
		w.file, w.was = filepath.Join(w.T.TempDir(), "notes.json"), []byte("not tasks\n")
		return os.WriteFile(w.file, w.was, 0o600)
	})
	sc.Step(`^a file of my own exported from a store where these were run:$`, func(run *godog.DocString) error {
		mine := w.StorePath
		other := filepath.Join(w.T.TempDir(), "other", "cairn.db")
		w.file = filepath.Join(w.T.TempDir(), "other", "tasks.json")
		w.Setenv(config.StoreVariable, other)
		defer w.Setenv(config.StoreVariable, mine)
		for _, line := range strings.Split(strings.TrimSpace(run.Content), "\n") {
			if err := w.Seed(strings.TrimSpace(line)); err != nil {
				return err
			}
		}
		return w.Seed("cairn task export " + w.file)
	})
	sc.Step(`^a file of my own holding a task that hangs under a task it does not hold$`, func() error {
		w.file = filepath.Join(w.T.TempDir(), "orphan.json")
		return os.WriteFile(w.file, []byte(`{"cairn": "tasks", "version": 1, "objectives": [],
			"tasks": [{"number": 1, "parent": 9, "title": "Orphan", "state": "backlog", "priority": "normal"}]}`), 0o600)
	})
	sc.Step(`^I import the tasks of that file$`, func() error {
		w.Run("cairn task import " + w.file)
		return nil
	})
	sc.Step(`^it says (\d+) tasks? and (\d+) objectives? (?:was|were) added$`, func(tasks, objectives int) error {
		said := fmt.Sprintf("%d task(s) and %d objective(s) were added from %s", tasks, objectives, w.file)
		return test.WantContains(w.Out.Stdout, said, "the answer")
	})
	sc.Step(`^it says task (\d+) is not in the file$`, func(n int) error {
		return test.WantContains(w.Out.Stderr, fmt.Sprintf("task %d is not in the file", n), "the reason")
	})
	sc.Step("^`([^`]*)` says \"([^\"]*)\"$", func(command, want string) error {
		w.Run(command)
		return test.WantContains(w.Out.Stdout, want, "the answer of "+command)
	})
}
