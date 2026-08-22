package cli

// Executable acceptance scenarios for bdd/store.feature. The steps every
// feature shares live in world_test.go; only what belongs to the store and its
// backups is declared here.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// olderStore turns the store into one an older cairn would have written, by
// undoing the last change to its shape through the real command. Nothing here
// reaches for SQL: a step that built the older shape by hand would be asserting
// against a shape cairn never actually wrote.
func (w *world) olderStore(title string) error {
	if err := w.seed(fmt.Sprintf(`cairn task add "%s"`, title)); err != nil {
		return err
	}
	if err := w.seed(`cairn migration rollback --yes`); err != nil {
		return err
	}
	// Undoing set a backup aside, which is right for somebody undoing on purpose
	// and wrong as a starting point: the scenarios count the backups made from
	// here on.
	return w.forgetBackups()
}

// forgetBackups removes every backup, so a scenario counts only what it caused.
func (w *world) forgetBackups() error {
	err := os.RemoveAll(filepath.Join(filepath.Dir(w.storePath), "backups"))
	if err != nil {
		return fmt.Errorf("clearing the backups: %w", err)
	}
	return nil
}

// numberOfRow reads the number a listing row leads with.
func numberOfRow(row string) (int, bool) {
	fields := strings.Fields(row)
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "#") {
		return 0, false
	}
	number, err := strconv.Atoi(strings.TrimPrefix(fields[0], "#"))
	return number, err == nil
}

// backupsListed counts the backups in whatever the last command printed.
func backupsListed(out string) int {
	return strings.Count(out, "#")
}

// registerStore declares the steps about the store itself — its backups and
// its shape. Both bdd/backup.feature and bdd/migration.feature draw on them:
// a scenario about applying a change checks the backup it set aside, and a
// scenario about backups needs a store older than this cairn to backup.
func (w *world) registerStore(sc *godog.ScenarioContext) {
	sc.Step(`^a store from an older cairn, holding a task titled "([^"]*)"$`, w.olderStore)

	// The number is read from the store, never written here: a scenario that
	// spelled out "apply 2" would be describing how many migrations cairn has
	// rather than what applying one does, and the next migration would break it.
	sc.Step(`^I apply the change that is waiting, by its number$`, func() error {
		before := w.out
		w.run(`cairn migration list`)
		listing := w.out.stdout
		w.out = before

		for _, row := range strings.Split(listing, "\n") {
			if !strings.Contains(row, "waiting") {
				continue
			}
			number, ok := numberOfRow(row)
			if !ok {
				return fmt.Errorf("cannot read a number from %q", row)
			}
			w.run(fmt.Sprintf(`cairn migration apply %d`, number))
			return nil
		}
		return fmt.Errorf("nothing is waiting:\n%s", listing)
	})

	sc.Step(`^I rolled the last change back$`, func() error {
		return w.seed(`cairn migration rollback --yes`)
	})

	sc.Step(`^I applied what was waiting$`, func() error {
		return w.seed(`cairn migration apply`)
	})

	sc.Step(`^I ran a command that was refused$`, func() error {
		// Deliberately not seeded: the point is that it failed and changed
		// nothing, so its failure must not stop the scenario.
		before := w.out
		w.run(`cairn task list`)
		w.out = before
		return nil
	})

	// The shape of the answer, not a word from it: "there is nothing waiting to
	// be applied" contains "applied", so a step looking for that word alone
	// passes on a command that applied nothing.
	sc.Step(`^it names the change it applied$`, func() error {
		return wantContains(w.out.stdout, `applied "`, "the answer")
	})

	sc.Step(`^it says what rolling the change back loses$`, func() error {
		return wantContains(w.out.stdout, "which task each sub-task hangs under", "the answer")
	})

	sc.Step(`^it says a backup can put it back$`, func() error {
		return wantContains(w.out.stdout, "cairn backup restore", "the answer")
	})

	// Named by shape too, and for a sharper reason: a rollback unmakes whichever
	// change happens to be last. Asserting its wording here would freeze how
	// many migrations cairn has, and the next one would turn this red for a
	// reason that has nothing to do with rolling back.
	sc.Step(`^it names the change it rolled back$`, func() error {
		return wantContains(w.out.stdout, `rolled back "`, "the answer")
	})

	// A machine reads a field, not a sentence. Checked separately from the text
	// form so neither can quietly satisfy the other's assertion.
	sc.Step(`^it carries the change it rolled back$`, func() error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.out.stdout), &doc); err != nil {
			return fmt.Errorf("stdout is not JSON (%w):\n%s", err, w.out.stdout)
		}
		if named, ok := doc["rolled_back"].(string); !ok || named == "" {
			return fmt.Errorf("no change named in %s", w.out.stdout)
		}
		return nil
	})

	sc.Step(`^it says what it lost$`, func() error {
		lost := "lost: "
		at := strings.Index(flatten(w.out.stdout), lost)
		if at < 0 {
			return fmt.Errorf("the answer does not say what it lost:\n%s", w.out.stdout)
		}
		if strings.TrimSpace(flatten(w.out.stdout)[at+len(lost):]) == "" {
			return fmt.Errorf("the answer says it lost nothing in particular:\n%s", w.out.stdout)
		}
		return nil
	})

	sc.Step(`^it carries the number (\d+)$`, func(n int) error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.out.stdout), &doc); err != nil {
			return fmt.Errorf("stdout is not JSON (%w):\n%s", err, w.out.stdout)
		}
		for _, held := range doc {
			if number, ok := held.(float64); ok && int(number) == n {
				return nil
			}
		}
		return fmt.Errorf("no number %d in %s", n, w.out.stdout)
	})

	sc.Step(`^it names the change$`, func() error {
		return wantContains(w.out.stdout, "a task can hang under another", "the answer")
	})

	sc.Step(`^it shows what the change does$`, func() error {
		if err := wantContains(w.out.stdout, "what it does", "the answer"); err != nil {
			return err
		}
		return wantContains(w.out.stdout, "ALTER TABLE tasks ADD COLUMN parent", "the answer")
	})

	sc.Step(`^it shows what undoing it does$`, func() error {
		if err := wantContains(w.out.stdout, "what undoing it does", "the answer"); err != nil {
			return err
		}
		return wantContains(w.out.stdout, "ALTER TABLE tasks DROP COLUMN parent", "the answer")
	})

	sc.Step(`^it says (\d+) changes? (?:is|are) waiting$`, func(n int) error {
		return wantContains(w.out.stdout, fmt.Sprintf("%d change(s) waiting", n), "the answer")
	})

	sc.Step("^it says to run `([^`]*)`$", func(what string) error {
		return wantContains(w.out.stderr, what, "the reason")
	})

	sc.Step(`^I made a backup of the store$`, func() error {
		return w.seed(`cairn backup add`)
	})

	sc.Step(`^I deleted backup (\d+)$`, func(n int) error {
		return w.seed(fmt.Sprintf(`cairn backup delete %d --yes`, n))
	})

	sc.Step(`^I deleted every backup$`, func() error {
		return w.forgetBackups()
	})

	sc.Step(`^it numbers the backup (\d+)$`, func(n int) error {
		return wantContains(w.out.stdout, "#"+strconv.Itoa(n), "the answer")
	})

	sc.Step(`^it lists (\d+) backups?$`, func(n int) error {
		if got := backupsListed(w.out.stdout); got != n {
			return fmt.Errorf("%d backups listed, want %d:\n%s", got, n, w.out.stdout)
		}
		return nil
	})

	sc.Step(`^it lists backup (\d+)$`, func(n int) error {
		return wantContains(w.out.stdout, "#"+strconv.Itoa(n), "the listing")
	})

	sc.Step("^`([^`]*)` lists (\\d+) backups?$", func(line string, n int) error {
		w.run(line)
		if got := backupsListed(w.out.stdout); got != n {
			return fmt.Errorf("%d backups listed, want %d:\n%s", got, n, w.out.stdout)
		}
		return nil
	})

	sc.Step("^`([^`]*)` says there is no backup$", func(line string) error {
		w.run(line)
		return wantContains(w.out.stdout, "there is no backup", "the answer")
	})

	// Both counts in one step: a backup that named its tasks and stayed quiet
	// about its objectives is exactly the answer this is here to catch.
	held := func(tasks, objectives int) string {
		return fmt.Sprintf("holds %d task(s) and %d objective(s)", tasks, objectives)
	}

	sc.Step(`^it says the backup holds (\d+) tasks? and (\d+) objectives?$`,
		func(tasks, objectives int) error {
			return wantContains(w.out.stdout, held(tasks, objectives), "the answer")
		})

	sc.Step("^`([^`]*)` says the backup holds (\\d+) tasks? and (\\d+) objectives?$",
		func(line string, tasks, objectives int) error {
			w.run(line)
			return wantContains(w.out.stdout, held(tasks, objectives), "the answer")
		})

	sc.Step("^`([^`]*)` says the store holds (\\d+) tasks? and (\\d+) objectives?$",
		func(line string, tasks, objectives int) error {
			w.run(line)
			return wantContains(w.out.stdout, held(tasks, objectives), "the answer")
		})

	sc.Step(`^it carries (\d+) tasks? and (\d+) objectives? held$`,
		func(tasks, objectives int) error {
			var doc map[string]any
			if err := json.Unmarshal([]byte(w.out.stdout), &doc); err != nil {
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
		return wantContains(w.out.stdout, "made ", "the answer")
	})

	sc.Step(`^it names every change made to the store$`, func() error {
		for _, name := range []string{"the shape a store starts with", "a task can hang under another"} {
			if err := wantContains(w.out.stdout, name, "the answer"); err != nil {
				return err
			}
		}
		return nil
	})

	sc.Step(`^it says when each was made$`, func() error {
		for _, row := range strings.Split(strings.TrimSpace(w.out.stdout), "\n") {
			if strings.Contains(row, "the shape a store starts with") && !strings.Contains(row, "-") {
				return fmt.Errorf("a change is named without saying when: %q", row)
			}
		}
		return nil
	})

	sc.Step(`^it names the change it undid$`, func() error {
		return wantContains(w.out.stdout, `rolled back "`, "the answer")
	})

	for step, want := range map[string]string{
		`^it says there is no such change$`:          "there is no such change",
		`^it says there is nothing waiting$`:         "nothing waiting",
		`^it says the change is already applied$`:    "already applied",
		`^it says the store is behind$`:              "the store is behind",
		`^it says there is no backup$`:               "there is no backup",
		`^it says there is no such backup$`:          "there is no such backup",
		`^it says there is no backup to put back$`:   "there is no backup to put back",
		`^it says there is nothing left to undo$`:    "nothing left to undo",
		`^it says the store was brought up to date$`: "brought up to date",
	} {
		wanted, where := want, step
		sc.Step(where, func() error {
			// A result is on stdout, a reason on stderr. Which one a step reads
			// is part of what it asserts.
			if strings.Contains(where, "no backup$") || strings.Contains(where, "nothing waiting") {
				return wantContains(w.out.stdout, wanted, "the answer")
			}
			return wantContains(w.out.stderr, wanted, "the reason")
		})
	}
}
