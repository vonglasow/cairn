package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"cairn/test"
)

func numberOfRow(row string) (int, bool) {
	fields := strings.Fields(row)
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "#") {
		return 0, false
	}
	number, err := strconv.Atoi(strings.TrimPrefix(fields[0], "#"))
	return number, err == nil
}

type change struct {
	Name    string `json:"name"`
	Applied bool   `json:"applied"`
}

func (w *world) changesListed() ([]change, error) {
	before := w.Out
	w.Run(`cairn migration list --json`)
	listed := w.Out
	w.Out = before

	if listed.Code != exitOK {
		return nil, fmt.Errorf("listing the changes failed: exit %d (%s)", listed.Code, listed.Stderr)
	}
	var listing struct {
		Changes []change `json:"changes"`
	}
	if err := json.Unmarshal([]byte(listed.Stdout), &listing); err != nil {
		return nil, fmt.Errorf("the changes are not JSON (%w):\n%s", err, listed.Stdout)
	}
	return listing.Changes, nil
}

func (w *world) rollBackPast(name string) error {
	changes, err := w.changesListed()
	if err != nil {
		return err
	}
	for range changes {
		applied, err := w.isApplied(name)
		if err != nil {
			return err
		}
		if !applied {
			return nil
		}
		if err := w.Seed(`cairn migration rollback --yes`); err != nil {
			return err
		}
	}
	return fmt.Errorf("the change named %q is still applied with nothing left to roll back", name)
}

func (w *world) isApplied(name string) (bool, error) {
	changes, err := w.changesListed()
	if err != nil {
		return false, err
	}
	for _, one := range changes {
		if one.Name == name {
			return one.Applied, nil
		}
	}
	return false, fmt.Errorf("cairn knows no change named %q", name)
}

func (w *world) registerStore(sc *godog.ScenarioContext) {
	sc.Step(`^a store from an older cairn, holding a task titled "([^"]*)"$`, w.AStoreFromAnOlderCairn)
	sc.Step(`^a store from before labels were written one way$`, func() error {
		before, err := os.ReadFile(filepath.Join("testdata", "labels-before-the-rule.db"))
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(w.StorePath), 0o700); err != nil {
			return err
		}
		return os.WriteFile(w.StorePath, before, 0o600)
	})
	sc.Step(`^(task|objective) (\d+) carries exactly the labels "([^"]*)"$`, func(kind string, n int, want string) error {
		w.Run(fmt.Sprintf("cairn %s show %d --json", kind, n))
		var shown struct {
			Labels []string `json:"labels"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &shown); err != nil {
			return fmt.Errorf("reading %s %d: %w\n%s%s", kind, n, err, w.Out.Stdout, w.Out.Stderr)
		}
		if got := strings.Join(shown.Labels, ", "); got != want {
			return fmt.Errorf("%s %d carries %q, want %q", kind, n, got, want)
		}
		return nil
	})

	sc.Step(`^I apply the last change waiting, past the one before it$`, func() error {
		before := w.Out
		w.Run(`cairn migration list`)
		listing := w.Out.Stdout
		w.Out = before
		last := 0
		for _, row := range strings.Split(listing, "\n") {
			if number, ok := numberOfRow(row); ok && strings.Contains(row, "waiting") {
				last = number
			}
		}
		if last == 0 {
			return fmt.Errorf("nothing is waiting:\n%s", listing)
		}
		w.Run(fmt.Sprintf(`cairn migration apply %d`, last))
		return nil
	})

	sc.Step(`^it says an earlier change is still waiting$`, func() error {
		return test.WantContains(w.Out.Stderr, "an earlier change is still waiting", "the reason")
	})

	sc.Step(`^I rolled the last change back$`, func() error {
		return w.Seed(`cairn migration rollback --yes`)
	})

	sc.Step(`^I rolled back the change named "([^"]*)"$`, func(name string) error {
		return w.rollBackPast(name)
	})

	sc.Step(`^I applied what was waiting$`, func() error {
		return w.Seed(`cairn migration apply`)
	})

	sc.Step("^it says to run `([^`]*)`$", func(what string) error {
		return test.WantContains(w.Out.Stderr, what, "the reason")
	})

	sc.Step(`^it says the file written holds (\d+) tasks? and (\d+) objectives?$`,
		func(tasks, objectives int) error {
			said := fmt.Sprintf("%s holds %d task(s) and %d objective(s)", w.file, tasks, objectives)
			return test.WantContains(w.Out.Stdout, said, "the answer")
		})

	sc.Step(`^I made a backup of the store$`, func() error {
		return w.Seed(`cairn backup new`)
	})

	sc.Step(`^it numbers the backup (\d+)$`, func(n int) error {
		return test.WantContains(w.Out.Stdout, "#"+strconv.Itoa(n), "the answer")
	})

	sc.Step(`^it names every change made to the store$`, func() error {
		for _, name := range []string{"the shape a store starts with", "a task can hang under another"} {
			if err := test.WantContains(w.Out.Stdout, name, "the answer"); err != nil {
				return err
			}
		}
		return nil
	})

	sc.Step(`^it says when each was made$`, func() error {
		for _, row := range strings.Split(strings.TrimSpace(w.Out.Stdout), "\n") {
			if strings.Contains(row, "the shape a store starts with") && !strings.Contains(row, "-") {
				return fmt.Errorf("a change is named without saying when: %q", row)
			}
		}
		return nil
	})

	sc.Step(`^it names the change it undid$`, func() error {
		return test.WantContains(w.Out.Stdout, `rolled back "`, "the answer")
	})

	for step, want := range map[string]string{
		`^it says there is no such change$`:          "there is no such change",
		`^it says there is nothing waiting$`:         "nothing waiting",
		`^it says the change is already applied$`:    "already applied",
		`^it says the store is behind$`:              "the store is behind",
		`^it says there is no backup$`:               "there is no backup",
		`^it says there is no such backup$`:          "there is no such backup",
		`^it says there is no backup to restore$`:    "there is no backup to restore",
		`^it says there is nothing left to undo$`:    "nothing left to undo",
		`^it says the store was brought up to date$`: "brought up to date",
	} {
		wanted, where := want, step
		sc.Step(where, func() error {
			if strings.Contains(where, "no backup$") || strings.Contains(where, "nothing waiting") {
				return test.WantContains(w.Out.Stdout, wanted, "the answer")
			}
			return test.WantContains(w.Out.Stderr, wanted, "the reason")
		})
	}
}
