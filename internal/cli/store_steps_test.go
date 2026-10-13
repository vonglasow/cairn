package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/migration"
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

func backupsListed(out string) int {
	return strings.Count(out, "#")
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

	sc.Step(`^I apply the change that is waiting, by its number$`, func() error {
		before := w.Out
		w.Run(`cairn migration list`)
		listing := w.Out.Stdout
		w.Out = before

		for _, row := range strings.Split(listing, "\n") {
			if !strings.Contains(row, "waiting") {
				continue
			}
			number, ok := numberOfRow(row)
			if !ok {
				return fmt.Errorf("cannot read a number from %q", row)
			}
			w.Run(fmt.Sprintf(`cairn migration apply %d`, number))
			return nil
		}
		return fmt.Errorf("nothing is waiting:\n%s", listing)
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

	sc.Step(`^I ran a command that was refused$`, func() error {
		before := w.Out
		w.Run(`cairn task list`)
		w.Out = before
		return nil
	})

	sc.Step(`^it names the change it applied$`, func() error {
		return test.WantContains(w.Out.Stdout, `applied "`, "the answer")
	})

	sc.Step(`^it says what rolling the change back loses$`, func() error {
		return test.WantContains(w.Out.Stdout, "which task each sub-task hangs under", "the answer")
	})

	sc.Step(`^it says a backup can restore it$`, func() error {
		return test.WantContains(w.Out.Stdout, "cairn backup restore", "the answer")
	})

	sc.Step(`^it names the change it rolled back$`, func() error {
		return test.WantContains(w.Out.Stdout, `rolled back "`, "the answer")
	})

	sc.Step(`^it carries the change it rolled back$`, func() error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return fmt.Errorf("stdout is not JSON (%w):\n%s", err, w.Out.Stdout)
		}
		if named, ok := doc["rolled_back"].(string); !ok || named == "" {
			return fmt.Errorf("no change named in %s", w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^it says what it lost$`, func() error {
		lost := "lost: "
		at := strings.Index(test.Flatten(w.Out.Stdout), lost)
		if at < 0 {
			return fmt.Errorf("the answer does not say what it lost:\n%s", w.Out.Stdout)
		}
		if strings.TrimSpace(test.Flatten(w.Out.Stdout)[at+len(lost):]) == "" {
			return fmt.Errorf("the answer says it lost nothing in particular:\n%s", w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^it carries the number (\d+)$`, func(n int) error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return fmt.Errorf("stdout is not JSON (%w):\n%s", err, w.Out.Stdout)
		}
		for _, held := range doc {
			if number, ok := held.(float64); ok && int(number) == n {
				return nil
			}
		}
		return fmt.Errorf("no number %d in %s", n, w.Out.Stdout)
	})

	sc.Step(`^it names the change$`, func() error {
		return test.WantContains(w.Out.Stdout, "a task can hang under another", "the answer")
	})

	sc.Step(`^it shows what the change does$`, func() error {
		if err := test.WantContains(w.Out.Stdout, "what it does", "the answer"); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stdout, "ALTER TABLE tasks ADD COLUMN parent", "the answer")
	})

	sc.Step(`^it shows what undoing it does$`, func() error {
		if err := test.WantContains(w.Out.Stdout, "what undoing it does", "the answer"); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stdout, "ALTER TABLE tasks DROP COLUMN parent", "the answer")
	})

	sc.Step(`^it says (\d+) changes? (?:is|are) waiting$`, func(n int) error {
		return test.WantContains(w.Out.Stdout, fmt.Sprintf("%d change(s) waiting", n), "the answer")
	})

	sc.Step("^it says to run `([^`]*)`$", func(what string) error {
		return test.WantContains(w.Out.Stderr, what, "the reason")
	})

	sc.Step(`^a board is open on the store$`, func() error {
		held, err := engine.Hold(home())
		if err != nil {
			return err
		}
		w.board = held
		w.OnEnd(func() { _ = held.Close() })
		return nil
	})

	sc.Step(`^the board adds a task titled "([^"]*)"$`, func(title string) error {
		_, err := w.board.Ask(engine.TaskNew, engine.Input{Given: map[string]string{"title": title}})
		return err
	})

	sc.Step(`^it says the file written holds (\d+) tasks? and (\d+) objectives?$`,
		func(tasks, objectives int) error {
			said := fmt.Sprintf("%s holds %d task(s) and %d objective(s)", w.file, tasks, objectives)
			return test.WantContains(w.Out.Stdout, said, "the answer")
		})
	sc.Step(`^that file is as it was$`, func() error {
		got, err := os.ReadFile(w.file)
		if err != nil {
			return err
		}
		if string(got) != string(w.was) {
			return fmt.Errorf("the file now holds %q", got)
		}
		return nil
	})

	sc.Step(`^I made a backup of the store$`, func() error {
		return w.Seed(`cairn backup new`)
	})

	sc.Step(`^I deleted backup (\d+)$`, func(n int) error {
		return w.Seed(fmt.Sprintf(`cairn backup delete %d --yes`, n))
	})

	sc.Step(`^I deleted every backup$`, func() error {
		return w.ForgetBackups()
	})

	sc.Step(`^it numbers the backup (\d+)$`, func(n int) error {
		return test.WantContains(w.Out.Stdout, "#"+strconv.Itoa(n), "the answer")
	})

	sc.Step(`^it lists (\d+) backups?$`, func(n int) error {
		if got := backupsListed(w.Out.Stdout); got != n {
			return fmt.Errorf("%d backups listed, want %d:\n%s", got, n, w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^it lists backup (\d+)$`, func(n int) error {
		return test.WantContains(w.Out.Stdout, "#"+strconv.Itoa(n), "the listing")
	})

	sc.Step("^`([^`]*)` lists (\\d+) backups?$", func(line string, n int) error {
		w.Run(line)
		if got := backupsListed(w.Out.Stdout); got != n {
			return fmt.Errorf("%d backups listed, want %d:\n%s", got, n, w.Out.Stdout)
		}
		return nil
	})

	sc.Step("^`([^`]*)` says there is no backup$", func(line string) error {
		w.Run(line)
		return test.WantContains(w.Out.Stdout, "there is no backup", "the answer")
	})

	held := func(tasks, objectives int) string {
		return fmt.Sprintf("holds %d task(s) and %d objective(s)", tasks, objectives)
	}

	sc.Step(`^it says the backup holds (\d+) tasks? and (\d+) objectives?$`,
		func(tasks, objectives int) error {
			return test.WantContains(w.Out.Stdout, held(tasks, objectives), "the answer")
		})

	sc.Step("^`([^`]*)` says the backup holds (\\d+) tasks? and (\\d+) objectives?$",
		func(line string, tasks, objectives int) error {
			w.Run(line)
			return test.WantContains(w.Out.Stdout, held(tasks, objectives), "the answer")
		})

	sc.Step(`^it carries every change$`, func() error {
		rows, _, err := listedUnder(w.Out.Stdout, "changes")
		if err != nil {
			return err
		}
		if len(rows) != len(migration.Known) {
			return fmt.Errorf("%d changes, want the %d cairn knows", len(rows), len(migration.Known))
		}
		return nil
	})

	sc.Step(`^it says the store holds (\d+) tasks? and (\d+) objectives?$`,
		func(tasks, objectives int) error {
			_, envelope, err := listedUnder(w.Out.Stdout, "changes")
			if err != nil {
				return err
			}
			holds, said := envelope["holds"].(map[string]any)
			if !said {
				return fmt.Errorf("the answer says nothing about what the store holds")
			}
			for what, want := range map[string]int{"tasks": tasks, "objectives": objectives} {
				if got := fmt.Sprint(holds[what]); got != fmt.Sprint(want) {
					return fmt.Errorf("it holds %s %s, want %d", got, what, want)
				}
			}
			return nil
		})

	sc.Step("^`([^`]*)` says the store holds (\\d+) tasks? and (\\d+) objectives?$",
		func(line string, tasks, objectives int) error {
			w.Run(line)
			return test.WantContains(w.Out.Stdout, held(tasks, objectives), "the answer")
		})

	sc.Step(`^it carries (\d+) tasks? and (\d+) objectives? held$`,
		func(tasks, objectives int) error {
			var doc map[string]any
			if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
				return err
			}
			holds, ok := doc["holds"].(map[string]any)
			if !ok {
				return fmt.Errorf("holds = %v, want the two counts", doc["holds"])
			}
			if holds["tasks"] != float64(tasks) || holds["objectives"] != float64(objectives) {
				return fmt.Errorf("holds = %v, want %d task(s) and %d objective(s)",
					holds, tasks, objectives)
			}
			return nil
		})

	sc.Step(`^it says when the backup was made$`, func() error {
		return test.WantContains(w.Out.Stdout, "made ", "the answer")
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
