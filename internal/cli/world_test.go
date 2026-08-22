package cli

// The world every feature file shares, and the steps they have in common.
//
// A step vocabulary is worth more than the feature that first needed it:
// `I added a task titled "..."` reads the same whether the scenario is about
// tasks or about the store that holds them. Feature-specific steps live beside
// their feature; everything here is claimed by more than one.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
)

// world carries what one scenario needs between its steps.
type world struct {
	t         *testing.T
	storePath string
	defaults  string // the default place, moved somewhere a test may look at
	out       result
	long      string
	restores  []func()
}

// setenv sets an environment variable for the length of the scenario. Not
// t.Setenv: that restores when the Go test ends, so the next scenario in the
// same run would start on this one's leftovers.
func (w *world) setenv(key, value string) {
	previous, had := os.LookupEnv(key)
	w.restores = append(w.restores, func() {
		if had {
			_ = os.Setenv(key, previous)
			return
		}
		_ = os.Unsetenv(key)
	})
	_ = os.Setenv(key, value)
}

func (w *world) endScenario() {
	for i := len(w.restores) - 1; i >= 0; i-- {
		w.restores[i]()
	}
	w.restores = nil
}

// run executes a command line the way a shell would have handed it over,
// dropping the leading `cairn` the scenario writes for the reader's benefit.
func (w *world) run(line string) {
	args := splitArgs(line)
	w.out = run(w.t, args[1:]...)
}

// field reads a value out of the card a command printed.
func (w *world) field(key string) string {
	for _, l := range strings.Split(w.out.stdout, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(l), key); ok {
			value := strings.TrimSpace(rest)
			if value == absent {
				return ""
			}
			return value
		}
	}
	return ""
}

// listedRow finds the line a task occupies in a listing, indentation included.
func (w *world) listedRow(number int) (string, error) {
	want := "#" + strconv.Itoa(number) + " "
	for _, row := range strings.Split(w.out.stdout, "\n") {
		if strings.HasPrefix(strings.TrimSpace(row), want) {
			return row, nil
		}
	}
	return "", fmt.Errorf("task %d is not in the listing:\n%s", number, w.out.stdout)
}

// heading is the first line of a card: its number and its title.
func (w *world) heading() string {
	return strings.TrimSpace(strings.SplitN(w.out.stdout, "\n", 2)[0])
}

// seed adds a task through the real command, so that what a precondition
// writes is what the command under test would have written. Going around the
// command would leave nothing proving `task add` writes at all.
func (w *world) seed(line string) error {
	before := w.out
	w.run(line)
	if w.out.code != exitOK {
		got := w.out
		w.out = before
		return fmt.Errorf("precondition %q failed: exit %d (%s)", line, got.code, got.stderr)
	}
	w.out = before
	return nil
}

// flatten makes a rendered message comparable: fang capitalises words and
// wraps long lines to the terminal width, so a step comparing exactly would be
// asserting fang's typography rather than cairn's answer.
func flatten(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

// wantContains looks for what inside where, past that typography.
func wantContains(where, what, label string) error {
	if !strings.Contains(flatten(where), flatten(what)) {
		return fmt.Errorf("%s does not mention %q:\n%s", label, what, where)
	}
	return nil
}

// progress reads the progress out of the JSON a command printed.
func (w *world) progress() (map[string]any, error) {
	var doc map[string]any
	if err := json.Unmarshal([]byte(w.out.stdout), &doc); err != nil {
		return nil, err
	}
	progress, ok := doc["progress"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("no progress in %s", w.out.stdout)
	}
	return progress, nil
}

// tally reads one of the two counts out of the progress.
func (w *world) tally(key string) (map[string]any, error) {
	progress, err := w.progress()
	if err != nil {
		return nil, err
	}
	tally, ok := progress[key].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("no %s in %v", key, progress)
	}
	return tally, nil
}

// newWorld builds the state one scenario works against.
func newWorld(t *testing.T) *world { return &world{t: t} }

// register declares the steps every feature shares.
func (w *world) register(sc *godog.ScenarioContext) {
	{
		t := w.t

		sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
			w.endScenario()
			return ctx, err
		})

		// --- the common context -------------------------------------------------

		sc.Step(`^a store of my own, somewhere that is not the default place$`, func() error {
			mine := t.TempDir()
			w.storePath = filepath.Join(mine, "mine", "cairn.db")
			// The default place is moved somewhere else entirely, and somewhere a
			// step may look at. Pointing it at the store's own folder would make
			// the two the same place and empty the scenario of its meaning.
			w.defaults = t.TempDir()
			w.setenv(engine.StoreVariable, w.storePath)
			w.setenv("HOME", w.defaults)
			return nil
		})

		sc.Step(`^nothing in it yet$`, func() error {
			if _, err := os.Stat(w.storePath); err == nil {
				return fmt.Errorf("the store at %s already exists", w.storePath)
			}
			return nil
		})

		// --- doing things -------------------------------------------------------

		sc.Step("^I run `([^`]*)`$", func(line string) error {
			w.run(line)
			return nil
		})

		sc.Step(`^a title of (\d+) characters$`, func(n int) error {
			w.long = strings.Repeat("a", n)
			return nil
		})

		sc.Step(`^a title with a (line break|tab) in it$`, func(what string) error {
			w.long = "two" + map[string]string{"line break": "\n", "tab": "\t"}[what] + "lines"
			return nil
		})

		sc.Step(`^I add(?:ed)? a task with it$`, func() error {
			w.run(fmt.Sprintf(`cairn task add "%s"`, w.long))
			return nil
		})

		// Shared rather than kept with the objectives: a backup counts what it
		// holds of both, so its scenarios have to be able to put an objective
		// in the store.
		sc.Step(`^I added an objective titled "([^"]*)"$`, func(title string) error {
			return w.seed(fmt.Sprintf(`cairn objective add "%s"`, title))
		})

		sc.Step(`^I added a task titled "([^"]*)"$`, func(title string) error {
			return w.seed(fmt.Sprintf(`cairn task add "%s"`, title))
		})

		sc.Step(`^I added a task titled "([^"]*)" with the (description|deadline|priority|state|parent|time spent) "([^"]*)"$`,
			func(title, attribute, value string) error {
				flag := map[string]string{
					"description": "--desc", "deadline": "--deadline",
					"priority": "--priority", "state": "--state", "parent": "--parent",
					"time spent": "--spent",
				}[attribute]
				return w.seed(fmt.Sprintf(`cairn task add "%s" %s "%s"`, title, flag, value))
			})

		sc.Step(`^I added a task titled "([^"]*)" with the labels "([^"]*)" and "([^"]*)"$`,
			func(title, first, second string) error {
				return w.seed(fmt.Sprintf(`cairn task add "%s" --label "%s" --label "%s"`, title, first, second))
			})

		sc.Step(`^I edited task (\d+) with the title "([^"]*)"$`, func(n int, title string) error {
			return w.seed(fmt.Sprintf(`cairn task edit %d --title "%s"`, n, title))
		})

		sc.Step(`^I deleted task (\d+)$`, func(n int) error {
			return w.seed(fmt.Sprintf(`cairn task delete %d --yes`, n))
		})

		sc.Step(`^cairn has exited$`, func() error {
			// Nothing to undo: every invocation opens the store in its RunE and
			// closes it on the way out, so no handle and no cache can outlive a
			// command. What makes the scenario bite is that the task was written
			// by one invocation and is read by another.
			return nil
		})

		// --- what it answered ---------------------------------------------------

		sc.Step(`^it succeeds$`, func() error {
			if w.out.code != exitOK {
				return fmt.Errorf("exit %d, want %d (%s)", w.out.code, exitOK, w.out.stderr)
			}
			return nil
		})

		sc.Step(`^it fails$`, func() error {
			if w.out.code == exitOK {
				return fmt.Errorf("it succeeded, want a failure:\n%s", w.out.stdout)
			}
			return nil
		})

		sc.Step(`^it does not crash$`, func() error {
			if strings.Contains(w.out.stderr, "panic:") {
				return fmt.Errorf("it panicked:\n%s", w.out.stderr)
			}
			return nil
		})

		sc.Step(`^it exits with code (\d+)$`, func(want int) error {
			if w.out.code != want {
				return fmt.Errorf("exit %d, want %d (%s)", w.out.code, want, w.out.stderr)
			}
			return nil
		})

		sc.Step(`^it writes nothing to standard error$`, func() error {
			if w.out.stderr != "" {
				return fmt.Errorf("stderr = %q, want nothing", w.out.stderr)
			}
			return nil
		})

		sc.Step(`^it writes nothing to standard output$`, func() error {
			if w.out.stdout != "" {
				return fmt.Errorf("stdout = %q, want nothing", w.out.stdout)
			}
			return nil
		})

		sc.Step(`^it writes the reason to standard error$`, func() error {
			if w.out.stderr == "" {
				return fmt.Errorf("stderr is empty, want the reason")
			}
			return nil
		})

		sc.Step(`^it writes the task to standard output$`, func() error {
			if w.out.stdout == "" {
				return fmt.Errorf("stdout is empty, want the task")
			}
			return nil
		})

		// --- what it said -------------------------------------------------------

		for step, want := range map[string]string{
			`^it says a title is one line$`:                                      "a title is one line",
			`^it says a task needs a title$`:                                     "a task needs a title",
			`^it says a label needs a name$`:                                     "a label needs a name",
			`^it says there is no such task$`:                                    "there is no such task",
			`^it says a task number is expected$`:                                "a task number is expected",
			`^it says how to confirm$`:                                           "--yes",
			`^it says the flag is unknown$`:                                      "unknown flag",
			`^it says the store cannot be read$`:                                 "the store cannot be read",
			`^it says what it can change$`:                                       "--title",
			`^it says the labels are either replaced or changed, not both$`:      "either replaced or changed",
			`^it says a label is added or taken away only when changing a task$`: "only when changing a task",
			`^it says a task cannot hang under itself$`:                          "cannot hang under itself",
			`^it says a sub-task cannot carry sub-tasks$`:                        "sub-task cannot carry sub-tasks",
			`^it says the task carries sub-tasks$`:                               "carries sub-tasks",
		} {
			wanted := want
			sc.Step(step, func() error {
				return wantContains(w.out.stderr, wanted, "the reason")
			})
		}

		sc.Step(`^it says how to use it$`, func() error {
			said := strings.ToLower(w.out.stderr + w.out.stdout)
			for _, hint := range []string{"usage", "try --help"} {
				if strings.Contains(said, hint) {
					return nil
				}
			}
			return fmt.Errorf("it does not say how to use it:\n%s", said)
		})

		// Quoting what was turned down is the whole of "nothing is guessed": a
		// refusal that does not name the value leaves somebody who passed three
		// flags with no idea which one cairn could not read.
		sc.Step(`^it says what it expects instead of "([^"]*)"$`, func(value string) error {
			if w.out.stderr == "" {
				return fmt.Errorf("stderr is empty, want what it expects instead")
			}
			return wantContains(w.out.stderr, value, "the reason")
		})

		sc.Step(`^it shows the title I typed$`, func() error {
			return wantContains(w.out.stderr, "got ", "the reason")
		})

		sc.Step(`^it says there is no task$`, func() error {
			return wantContains(w.out.stdout, "there is no task", "the answer")
		})

		sc.Step(`^it says task (\d+) was deleted$`, func(n int) error {
			return wantContains(w.out.stdout, fmt.Sprintf("task %d was deleted", n), "the answer")
		})

		// --- what the task carries ----------------------------------------------

		sc.Step(`^it reports a task titled "([^"]*)"$`, func(title string) error {
			return wantContains(w.heading(), title, "the heading")
		})

		sc.Step(`^it reports a task with that exact title$`, func() error {
			return wantContains(w.heading(), w.long, "the heading")
		})

		sc.Step(`^it numbers the task (\d+)$`, func(n int) error {
			return wantContains(w.heading(), "#"+strconv.Itoa(n), "the heading")
		})

		sc.Step(`^the task carries the (description|deadline|estimate|priority|column|parent|time spent) "([^"]*)"$`,
			func(attribute, want string) error {
				key := map[string]string{
					"description": "description", "deadline": "deadline",
					"estimate": "estimate", "priority": "priority",
					"column": "state", "parent": "parent", "time spent": "time spent",
				}[attribute]
				if got := w.field(key); got != want {
					return fmt.Errorf("%s = %q, want %q", key, got, want)
				}
				return nil
			})

		// Carrying a label says it is among the ones the task carries, not that
		// it is the only one. The count is asserted separately, by the scenarios
		// that care how many there are.
		sc.Step(`^the task carries the label "([^"]*)"$`, func(want string) error {
			for _, carried := range strings.Split(w.field("labels"), ", ") {
				if carried == want {
					return nil
				}
			}
			return fmt.Errorf("labels = %q, want one of them to be %q", w.field("labels"), want)
		})

		sc.Step(`^the task carries the labels "([^"]*)" and "([^"]*)"$`, func(first, second string) error {
			want := first + ", " + second
			if got := w.field("labels"); got != want {
				return fmt.Errorf("labels = %q, want %q", got, want)
			}
			return nil
		})

		sc.Step(`^the task carries exactly (\d+) labels?$`, func(n int) error {
			labels := w.field("labels")
			got := 0
			if labels != "" {
				got = len(strings.Split(labels, ", "))
			}
			if got != n {
				return fmt.Errorf("%d labels (%q), want %d", got, labels, n)
			}
			return nil
		})

		sc.Step(`^the task is in the "([^"]*)" column$`, func(want string) error {
			if got := w.field("state"); got != want {
				return fmt.Errorf("state = %q, want %q", got, want)
			}
			return nil
		})

		sc.Step(`^the task has no (description|labels|deadline|estimate|time spent|parent)$`, func(attribute string) error {
			key := map[string]string{
				"description": "description", "labels": "labels", "deadline": "deadline",
				"estimate": "estimate", "time spent": "time spent", "parent": "parent",
			}[attribute]
			if got := w.field(key); got != "" {
				return fmt.Errorf("%s = %q, want none", key, got)
			}
			return nil
		})

		// --- listing --------------------------------------------------------------

		sc.Step("^`([^`]*)` lists (\\d+) tasks?$", func(line string, n int) error {
			w.run(line)
			return countsTasks(w.out.stdout, n)
		})

		sc.Step(`^it lists (\d+) tasks?$`, func(n int) error {
			return countsTasks(w.out.stdout, n)
		})

		sc.Step("^`([^`]*)` lists task (\\d+) titled \"([^\"]*)\"$", func(line string, n int, title string) error {
			w.run(line)
			return listsTask(w.out.stdout, n, title)
		})

		sc.Step(`^it lists task (\d+) titled "([^"]*)"$`, func(n int, title string) error {
			return listsTask(w.out.stdout, n, title)
		})

		sc.Step("^`([^`]*)` says there is no task$", func(line string) error {
			w.run(line)
			return wantContains(w.out.stdout, "there is no task", "the answer")
		})

		sc.Step("^`([^`]*)` still reports a task titled \"([^\"]*)\"$", func(line, title string) error {
			w.run(line)
			return wantContains(w.heading(), title, "the heading")
		})

		// The indentation is what says a task hangs under another, so these read
		// it rather than the store: a listing that lost the link would still be
		// green if the step went and asked the engine again.
		sc.Step(`^task (\d+) is listed under task (\d+)$`, func(hung, under int) error {
			row, err := w.listedRow(hung)
			if err != nil {
				return err
			}
			if !strings.HasPrefix(row, " ") {
				return fmt.Errorf("task %d is not listed under anything: %q", hung, row)
			}
			at := func(n int) int { return strings.Index(w.out.stdout, "#"+strconv.Itoa(n)+" ") }
			if at(under) > at(hung) {
				return fmt.Errorf("task %d is listed before task %d it hangs under:\n%s", hung, under, w.out.stdout)
			}
			return nil
		})

		sc.Step(`^task (\d+) hangs under nothing in the listing$`, func(n int) error {
			row, err := w.listedRow(n)
			if err != nil {
				return err
			}
			if strings.HasPrefix(row, " ") {
				return fmt.Errorf("task %d is listed under something: %q", n, row)
			}
			return nil
		})

		sc.Step(`^task (\d+) is listed before task (\d+)$`, func(first, second int) error {
			at := func(n int) int {
				return strings.Index(w.out.stdout, "#"+strconv.Itoa(n)+" ")
			}
			if at(first) < 0 || at(second) < 0 || at(first) > at(second) {
				return fmt.Errorf("task %d is not listed before task %d:\n%s", first, second, w.out.stdout)
			}
			return nil
		})

		// --- the store ------------------------------------------------------------

		sc.Step(`^the store exists$`, func() error {
			if _, err := os.Stat(w.storePath); err != nil {
				return fmt.Errorf("the store was not created at %s: %w", w.storePath, err)
			}
			return nil
		})

		sc.Step(`^the store is created where I said$`, func() error {
			if _, err := os.Stat(w.storePath); err != nil {
				return fmt.Errorf("the store was not created at %s: %w", w.storePath, err)
			}
			return nil
		})

		sc.Step(`^nothing is written to the default place$`, func() error {
			// The whole folder, not just cairn.db: SQLite leaves -wal and -shm
			// beside it, and a check that missed them would pass on a store that
			// was very much written to.
			left, err := os.ReadDir(filepath.Join(w.defaults, engine.DefaultFolder))
			if err != nil {
				return nil // the folder was never created, which is the point
			}
			if len(left) > 0 {
				return fmt.Errorf("the default place holds %d entries, want none", len(left))
			}
			return nil
		})

		sc.Step(`^the store holds something that is not a cairn store$`, func() error {
			if err := os.MkdirAll(filepath.Dir(w.storePath), 0o700); err != nil {
				return err
			}
			return os.WriteFile(w.storePath, []byte("this is not a cairn store"), 0o600)
		})

		sc.Step(`^it names the store$`, func() error {
			return wantContains(w.out.stderr, w.storePath, "the reason")
		})

		sc.Step(`^the store is left as it was$`, func() error {
			held, err := os.ReadFile(w.storePath)
			if err != nil {
				return fmt.Errorf("the store is gone: %w", err)
			}
			if string(held) != "this is not a cairn store" {
				return fmt.Errorf("the store was replaced: %q", held)
			}
			return nil
		})

		// --- machine-readable answers ---------------------------------------------

		sc.Step(`^it writes JSON to standard output$`, func() error {
			var any any
			if err := json.Unmarshal([]byte(w.out.stdout), &any); err != nil {
				return fmt.Errorf("stdout is not JSON (%w):\n%s", err, w.out.stdout)
			}
			return nil
		})

		sc.Step(`^it carries the title "([^"]*)"$`, func(title string) error {
			var doc map[string]any
			if err := json.Unmarshal([]byte(w.out.stdout), &doc); err != nil {
				return err
			}
			if doc["title"] != title {
				return fmt.Errorf("title = %v, want %q", doc["title"], title)
			}
			return nil
		})

		sc.Step(`^it carries the parent (\d+)$`, func(n int) error {
			var doc map[string]any
			if err := json.Unmarshal([]byte(w.out.stdout), &doc); err != nil {
				return err
			}
			if got, ok := doc["parent"].(float64); !ok || int(got) != n {
				return fmt.Errorf("parent = %v, want %d", doc["parent"], n)
			}
			return nil
		})

		sc.Step(`^it carries an identifier$`, func() error {
			var doc map[string]any
			if err := json.Unmarshal([]byte(w.out.stdout), &doc); err != nil {
				return err
			}
			if doc["number"] == nil {
				return fmt.Errorf("no identifier in %s", w.out.stdout)
			}
			return nil
		})

		sc.Step(`^it carries (\d+) tasks?$`, func(n int) error {
			var docs []map[string]any
			if err := json.Unmarshal([]byte(w.out.stdout), &docs); err != nil {
				return err
			}
			if len(docs) != n {
				return fmt.Errorf("%d tasks, want %d", len(docs), n)
			}
			return nil
		})

		// --- how far along something is -------------------------------------------
		//
		// Shared by tasks and objectives: both are asked the same question and
		// answer it in the same words, so one vocabulary serves both features.

		sc.Step(`^it says (\d+) of (\d+) (tasks|sub-tasks) are done$`,
			func(done, total int, kind string) error {
				return wantContains(w.out.stdout,
					fmt.Sprintf("%d of %d %s are done", done, total, kind), "the answer")
			})

		sc.Step(`^it says (\d+) of (\d+) tasks and (\d+) of (\d+) sub-tasks are done$`,
			func(tasksDone, tasksTotal, subDone, subTotal int) error {
				return wantContains(w.out.stdout,
					fmt.Sprintf("%d of %d tasks and %d of %d sub-tasks are done",
						tasksDone, tasksTotal, subDone, subTotal), "the answer")
			})

		sc.Step(`^it says the (task|objective) is (\d+)% done$`, func(what string, share int) error {
			return wantContains(w.out.stdout, fmt.Sprintf("the %s is %d%% done", what, share), "the answer")
		})

		// Looked for by shape rather than by the word "done", which a state of
		// its own puts on the card: `state done` would match either way.
		sc.Step(`^it does not say how much is done$`, func() error {
			if strings.Contains(w.out.stdout, "% done") {
				return fmt.Errorf("the answer works out a share it has nothing to work out:\n%s", w.out.stdout)
			}
			return nil
		})

		sc.Step(`^it carries (\d+) of (\d+) (tasks|sub-tasks) done$`,
			func(done, total int, kind string) error {
				key := map[string]string{"tasks": "tasks", "sub-tasks": "sub_tasks"}[kind]
				tally, err := w.tally(key)
				if err != nil {
					return err
				}
				if tally["total"] != float64(total) || tally["done"] != float64(done) {
					return fmt.Errorf("%s = %v, want %d of %d", key, tally, done, total)
				}
				return nil
			})

		sc.Step(`^it carries the share (\d+)$`, func(want int) error {
			progress, err := w.progress()
			if err != nil {
				return err
			}
			if progress["percent"] != float64(want) {
				return fmt.Errorf("percent = %v, want %d", progress["percent"], want)
			}
			return nil
		})

		// A task has sub-tasks under it and nothing else, so a count of tasks
		// would name a kind of thing it cannot hold — zeroes included.
		sc.Step(`^it carries no count of tasks$`, func() error {
			progress, err := w.progress()
			if err != nil {
				return err
			}
			if _, present := progress["tasks"]; present {
				return fmt.Errorf("progress = %v, want no count of tasks", progress)
			}
			return nil
		})

		// The key has to be there and be null. A document missing it answers the
		// same to a caller reading doc["progress"], right up to the day the
		// caller starts checking whether the field exists at all.
		sc.Step(`^it carries no progress$`, func() error {
			var doc map[string]any
			if err := json.Unmarshal([]byte(w.out.stdout), &doc); err != nil {
				return err
			}
			if _, present := doc["progress"]; !present {
				return fmt.Errorf("no progress key at all in %s", w.out.stdout)
			}
			if doc["progress"] != nil {
				return fmt.Errorf("progress = %v, want null", doc["progress"])
			}
			return nil
		})

		sc.Step(`^the JSON carries the reason$`, func() error {
			var doc map[string]string
			if err := json.Unmarshal([]byte(w.out.stdout), &doc); err != nil {
				return err
			}
			if doc["error"] == "" {
				return fmt.Errorf("no reason in %s", w.out.stdout)
			}
			return nil
		})
	}
}

func countsTasks(out string, n int) error {
	got := strings.Count(out, "#")
	if got != n {
		return fmt.Errorf("%d tasks listed, want %d:\n%s", got, n, out)
	}
	return nil
}

func listsTask(out string, n int, title string) error {
	want := fmt.Sprintf("#%d  %s", n, title)
	if !strings.Contains(out, want) {
		return fmt.Errorf("the listing does not carry %q:\n%s", want, out)
	}
	return nil
}
