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

	"cairn/internal/config"
	"cairn/internal/engine"
	"cairn/internal/migration"
	"cairn/test"
)

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

// change is one line of what `cairn migration list --json` answers.
type change struct {
	Name    string `json:"name"`
	Applied bool   `json:"applied"`
}

// changesListed asks cairn what changes it knows and which are applied. Asked
// through the real command rather than read out of migration.Known: what the
// step needs to know is what this store has had done to it, which only cairn
// can say.
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

// rollBackPast undoes changes, one at a time through the real command, until
// the one named is no longer applied. How many that is comes from what cairn
// itself reports, so nothing here has a number written into it.
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

// isApplied says whether the change named has been applied to this store.
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

// registerStore declares the steps about the store itself — its backups and
// its shape. Both bdd/backup.feature and bdd/migration.feature draw on them:
// a scenario about applying a change checks the backup it set aside, and a
// scenario about backups needs a store older than this cairn to backup.
func (w *world) registerStore(sc *godog.ScenarioContext) {
	sc.Step(`^a store from an older cairn, holding a task titled "([^"]*)"$`, w.AStoreFromAnOlderCairn)
	// A store from before labels were written one way, kept as a file rather
	// than made here: no command writes a label that way any more, which is the
	// rule's whole point, and a scenario never opens a database. It was made
	// with the cairn of the change before, then labelled with sqlite3 — task 1
	// "chantier:proxmox", "ado:77594", "proxmox"; task 2 "obj-modularité/helm",
	// "dette_config", "dette-config", "a:b:c", "ado:", "sp26   20", "--";
	// objective 1 "chantier:doc-infra"; objective 2 "été 2026".
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

	// The number is read from the store, never written here: a scenario that
	// spelled out "apply 2" would be describing how many migrations cairn has
	// rather than what applying one does, and the next migration would break it.
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

	// Named rather than counted, and that is not decoration. A scenario whose
	// counts depend on which change was undone cannot say "the last one": what
	// that means changes the day a change is appended, and this scenario spent
	// a while asserting about the wrong one for exactly that reason.
	sc.Step(`^I rolled back the change named "([^"]*)"$`, func(name string) error {
		return w.rollBackPast(name)
	})

	sc.Step(`^I applied what was waiting$`, func() error {
		return w.Seed(`cairn migration apply`)
	})

	sc.Step(`^I ran a command that was refused$`, func() error {
		// Deliberately not seeded: the point is that it failed and changed
		// nothing, so its failure must not stop the scenario.
		before := w.Out
		w.Run(`cairn task list`)
		w.Out = before
		return nil
	})

	// The shape of the answer, not a word from it: "there is nothing waiting to
	// be applied" contains "applied", so a step looking for that word alone
	// passes on a command that applied nothing.
	sc.Step(`^it names the change it applied$`, func() error {
		return test.WantContains(w.Out.Stdout, `applied "`, "the answer")
	})

	sc.Step(`^it says what rolling the change back loses$`, func() error {
		return test.WantContains(w.Out.Stdout, "which task each sub-task hangs under", "the answer")
	})

	sc.Step(`^it says a backup can restore it$`, func() error {
		return test.WantContains(w.Out.Stdout, "cairn backup restore", "the answer")
	})

	// Named by shape too, and for a sharper reason: a rollback unmakes whichever
	// change happens to be last. Asserting its wording here would freeze how
	// many migrations cairn has, and the next one would turn this red for a
	// reason that has nothing to do with rolling back.
	sc.Step(`^it names the change it rolled back$`, func() error {
		return test.WantContains(w.Out.Stdout, `rolled back "`, "the answer")
	})

	// A machine reads a field, not a sentence. Checked separately from the text
	// form so neither can quietly satisfy the other's assertion.
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

	sc.Step(`^I export the store to a file of my own$`, func() error {
		w.file = filepath.Join(w.T.TempDir(), "elsewhere", "cairn.db")
		w.Run("cairn backup export " + w.file)
		return nil
	})
	sc.Step(`^I export the store to that file$`, func() error {
		w.Run("cairn backup export " + w.file)
		return nil
	})
	sc.Step(`^it says the file written holds (\d+) tasks? and (\d+) objectives?$`,
		func(tasks, objectives int) error {
			said := fmt.Sprintf("%s holds %d task(s) and %d objective(s)", w.file, tasks, objectives)
			return test.WantContains(w.Out.Stdout, said, "the answer")
		})
	// Read by pointing cairn at it, the way the other machine will.
	sc.Step(`^that file is a store holding the task "([^"]*)"$`, func(title string) error {
		mine := w.StorePath
		w.Setenv(config.StoreVariable, w.file)
		defer w.Setenv(config.StoreVariable, mine)
		w.Run("cairn task list")
		return test.WantContains(w.Out.Stdout, title, "the listing of that file")
	})
	sc.Step(`^a file of my own that is not a cairn store$`, func() error {
		w.file, w.was = filepath.Join(w.T.TempDir(), "notes.db"), []byte("not a database\n")
		return os.WriteFile(w.file, w.was, 0o600)
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
	// Made the way the other machine made it: by cairn, against a store of
	// its own.
	sc.Step(`^a file of my own holding a store with a task titled "([^"]*)"$`, func(title string) error {
		mine := w.StorePath
		w.file = filepath.Join(w.T.TempDir(), "other", "cairn.db")
		w.Setenv(config.StoreVariable, w.file)
		defer w.Setenv(config.StoreVariable, mine)
		return w.Seed(fmt.Sprintf(`cairn task new "%s"`, title))
	})
	sc.Step(`^I import that file$`, func() error {
		w.Run("cairn backup import " + w.file + " --yes")
		return nil
	})
	sc.Step(`^I run the import of that file without confirming$`, func() error {
		w.Run("cairn backup import " + w.file)
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

	// Both counts in one step: a backup that named its tasks and stayed quiet
	// about its objectives is exactly the answer this is here to catch.
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

	// What the store holds is said in the text and used to be lost in the JSON:
	// a bare list had nowhere to put it. The envelope has.
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
			// A result is on stdout, a reason on stderr. Which one a step reads
			// is part of what it asserts.
			if strings.Contains(where, "no backup$") || strings.Contains(where, "nothing waiting") {
				return test.WantContains(w.Out.Stdout, wanted, "the answer")
			}
			return test.WantContains(w.Out.Stderr, wanted, "the reason")
		})
	}
}
