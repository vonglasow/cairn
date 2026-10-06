package cli

// The world every feature file shares, and the steps they have in common.
//
// A step vocabulary is worth more than the feature that first needed it:
// `I added a task titled "..."` reads the same whether the scenario is about
// tasks or about the store that holds them. Feature-specific steps live beside
// their feature; everything here is claimed by more than one.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"cairn/internal/engine"
	"cairn/internal/task"

	"github.com/cucumber/godog"

	"cairn/internal/config"
	"cairn/internal/domain"
	"cairn/test"
)

// world carries what one scenario needs between its steps: the world both
// surfaces share, and the one thing only the command line's scenarios need.
type world struct {
	*test.World
	long string
	// board is an engine held open the way a board holds one, for the
	// scenarios that need somebody else at the store while a command runs.
	board *engine.Engine
}

// newWorld builds the state one scenario works against, reaching the command
// line through run — the same Main the binary calls.
func newWorld(t *testing.T) *world {
	return &world{World: test.NewWorld(t, run)}
}

// field reads a value out of the card a command printed.
// daysToday is today as a step sees it, and as the command it checks may have
// seen it a moment before midnight turned it into yesterday.
func daysToday() []string {
	return []string{domain.FormatDay(time.Now()), domain.FormatDay(time.Now().Add(-time.Minute))}
}

func (w *world) field(key string) string {
	for _, l := range strings.Split(w.Out.Stdout, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(l), key); ok {
			value := strings.TrimSpace(rest)
			if value == domain.Absent {
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
	for _, row := range strings.Split(w.Out.Stdout, "\n") {
		if strings.HasPrefix(strings.TrimSpace(row), want) {
			return row, nil
		}
	}
	return "", fmt.Errorf("task %d is not in the listing:\n%s", number, w.Out.Stdout)
}

// heading is the first line of a card: its number and its title.
func (w *world) heading() string {
	return strings.TrimSpace(strings.SplitN(w.Out.Stdout, "\n", 2)[0])
}

// progress reads the progress out of the JSON a command printed.
func (w *world) progress() (map[string]any, error) {
	var doc map[string]any
	if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
		return nil, err
	}
	progress, ok := doc["progress"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("no progress in %s", w.Out.Stdout)
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

// register declares the steps only the command line's scenarios claim. What
// both surfaces claim is registered by the shared world.
// listedUnder reads the rows of a listing out of the envelope it crosses in.
//
// A listing answers an object naming what it carries, not a bare list, so a step
// asking how many tasks came back asks the envelope for its tasks. Written once
// because five listings cross this way and a second reading of the same envelope
// is a second thing to keep right.
func listedUnder(out, key string) ([]map[string]any, map[string]any, error) {
	var envelope map[string]any
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		return nil, nil, fmt.Errorf("the answer is not one object (%w):\n%s", err, out)
	}
	raw, carried := envelope[key]
	if !carried {
		named := make([]string, 0, len(envelope))
		for k := range envelope {
			named = append(named, k)
		}
		sort.Strings(named)
		return nil, envelope, fmt.Errorf("the answer carries %v, and not %q", named, key)
	}
	items, list := raw.([]any)
	if !list && raw != nil {
		return nil, envelope, fmt.Errorf("%q is not a list but a %T", key, raw)
	}
	rows := make([]map[string]any, 0, len(items))
	for _, one := range items {
		row, _ := one.(map[string]any)
		rows = append(rows, row)
	}
	return rows, envelope, nil
}

// flagFor is the flag an attribute is given by, named once because both the
// step that creates a task with one and the step that edits one read it.
// numbersIn reads every whole number a message names.
var numbersIn = regexp.MustCompile(`\d+`)

func flagFor(attribute string) string {
	return map[string]string{
		"title": "--title", "description": "--desc", "deadline": "--deadline",
		"estimate": "--estimate",
		"priority": "--priority", "state": "--state", "parent": "--parent",
		"time spent": "--spent", "objective": "--objective", "blocker": "--blocker",
		"because": "--because", "label": "--label",
	}[attribute]
}

func (w *world) register(sc *godog.ScenarioContext) {
	w.Register(sc)

	// --- doing things -----------------------------------------------------------

	sc.Step(`^a title of (\d+) characters$`, func(n int) error {
		w.long = strings.Repeat("a", n)
		return nil
	})

	sc.Step(`^a title with a (line break|tab) in it$`, func(what string) error {
		w.long = "two" + map[string]string{"line break": "\n", "tab": "\t"}[what] + "lines"
		return nil
	})

	sc.Step(`^I add(?:ed)? a task with it$`, func() error {
		w.Run(fmt.Sprintf(`cairn task add "%s"`, w.long))
		return nil
	})

	// Shared rather than kept with the objectives: a backup counts what it
	// holds of both, so its scenarios have to be able to put an objective
	// in the store.
	sc.Step(`^I added an objective titled "([^"]*)"$`, func(title string) error {
		return w.Seed(fmt.Sprintf(`cairn objective add "%s"`, title))
	})

	sc.Step(`^I added a task titled "([^"]*)" with the (description|deadline|estimate|priority|state|parent|time spent|objective) "([^"]*)"$`,
		func(title, attribute, value string) error {
			return w.Seed(fmt.Sprintf(`cairn task add "%s" %s "%s"`, title, flagFor(attribute), value))
		})

	sc.Step(`^I edited task (\d+) with the labels "([^"]*)" and "([^"]*)"$`,
		func(n int, first, second string) error {
			return w.Seed(fmt.Sprintf(`cairn task edit %d --label "%s" --label "%s"`, n, first, second))
		})

	sc.Step(`^I added a task titled "([^"]*)" with the objectives "([^"]*)" and "([^"]*)"$`,
		func(title, first, second string) error {
			return w.Seed(fmt.Sprintf(`cairn task add "%s" --objective "%s" --objective "%s"`,
				title, first, second))
		})

	sc.Step(`^I added a task titled "([^"]*)" with the labels "([^"]*)" and "([^"]*)"$`,
		func(title, first, second string) error {
			return w.Seed(fmt.Sprintf(`cairn task add "%s" --label "%s" --label "%s"`, title, first, second))
		})

	sc.Step(`^I edited task (\d+) with the (title|description|deadline|estimate|priority|state|parent|time spent|blocker|because|objective|label) "([^"]*)"$`,
		func(n int, attribute, value string) error {
			return w.Seed(fmt.Sprintf(`cairn task edit %d %s "%s"`, n, flagFor(attribute), value))
		})

	sc.Step(`^I put task (\d+) in focus$`, func(n int) error {
		return w.Seed(fmt.Sprintf(`cairn task focus %d`, n))
	})

	sc.Step(`^I deleted task (\d+)$`, func(n int) error {
		return w.Seed(fmt.Sprintf(`cairn task delete %d --yes`, n))
	})

	sc.Step(`^cairn has exited$`, func() error {
		// Nothing to undo: every invocation opens the store in its RunE and
		// closes it on the way out, so no handle and no cache can outlive a
		// command. What makes the scenario bite is that the task was written
		// by one invocation and is read by another.
		return nil
	})

	// --- what it answered ---------------------------------------------------

	sc.Step(`^it does not crash$`, func() error {
		if strings.Contains(w.Out.Stderr, "panic:") {
			return fmt.Errorf("it panicked:\n%s", w.Out.Stderr)
		}
		return nil
	})

	sc.Step(`^it exits with code (\d+)$`, func(want int) error {
		if w.Out.Code != want {
			return fmt.Errorf("exit %d, want %d (%s)", w.Out.Code, want, w.Out.Stderr)
		}
		return nil
	})

	sc.Step(`^it writes nothing to standard error$`, func() error {
		if w.Out.Stderr != "" {
			return fmt.Errorf("stderr = %q, want nothing", w.Out.Stderr)
		}
		return nil
	})

	sc.Step(`^it writes nothing to standard output$`, func() error {
		if w.Out.Stdout != "" {
			return fmt.Errorf("stdout = %q, want nothing", w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^it writes the reason to standard error$`, func() error {
		if w.Out.Stderr == "" {
			return fmt.Errorf("stderr is empty, want the reason")
		}
		return nil
	})

	sc.Step(`^it writes the task to standard output$`, func() error {
		if w.Out.Stdout == "" {
			return fmt.Errorf("stdout is empty, want the task")
		}
		return nil
	})

	// --- what it said -------------------------------------------------------

	for step, want := range map[string]string{
		`^it says a title is one line$`:                                           "a title is one line",
		`^it says a task needs a title$`:                                          "a task needs a title",
		`^it says a label needs a name$`:                                          "a label needs a name",
		`^it says there is no such task$`:                                         "there is no such task",
		`^it says a task number is expected$`:                                     "a task number is expected",
		`^it says how to confirm$`:                                                "--yes",
		`^it says the flag is unknown$`:                                           "unknown flag",
		`^it says the store cannot be read$`:                                      "the store cannot be read",
		`^it says what it can change$`:                                            "--title",
		`^it says the labels are either replaced or changed, not both$`:           "either replaced or changed",
		`^it says a label is added or taken away only when changing a task$`:      "only when changing a task",
		`^it says a task cannot hang under itself$`:                               "cannot hang under itself",
		`^it says a task cannot block itself$`:                                    "cannot block itself",
		`^it says the objectives are either replaced or changed, not both$`:       "either replaced or changed",
		`^it says an objective is added or taken away only when changing a task$`: "only when changing a task",
		`^it says what is in the way and how to take it away$`:                    "take them away first: task ",
		`^it says nothing would ever be able to start$`:                           "nothing would ever be able to start",
		`^it says a sub-task cannot carry sub-tasks$`:                             "sub-task cannot carry sub-tasks",
		`^it says a task cannot be moved behind itself$`:                          "cannot be moved behind itself",
		`^it says a task is moved inside its own column$`:                         "moved inside its own column",
		`^it says a sub-task is moved among its sisters$`:                         "moved among its sisters",
		`^it says where a task can be moved$`:                                     "behind another",
	} {
		wanted := want
		sc.Step(step, func() error {
			return test.WantContains(w.Out.Stderr, wanted, "the reason")
		})
	}

	// A refusal that does not quote the number leaves somebody who typed two of
	// them with no idea which one cairn could not find, so the number is looked
	// for among those the question actually carried.
	//
	// Shared rather than kept with the order: the trace refuses the same way,
	// and a step two features claim belongs where both can reach it.
	sc.Step(`^it shows the number it could not find$`, func() error {
		asked := numbersIn.FindAllString(w.Line, -1)
		if len(asked) == 0 {
			return fmt.Errorf("the question named no number: %q", w.Line)
		}
		for _, number := range asked {
			if strings.Contains(w.Out.Stderr, number) {
				return nil
			}
		}
		return fmt.Errorf("the reason quotes none of the numbers in %q:\n%s", w.Line, w.Out.Stderr)
	})

	sc.Step(`^it shows the title I typed$`, func() error {
		return test.WantContains(w.Out.Stderr, "got ", "the reason")
	})

	// The refusal names what is in the way, which is the useful thing to be
	// told: "it is not finished" leaves somebody hunting for which one.
	sc.Step(`^it says task (\d+) is in the way$`, func(n int) error {
		return test.WantContains(w.Out.Stderr, fmt.Sprintf("task %d", n), "the refusal")
	})

	// Asked of the store through the real command, not of what the last run
	// printed: what is being checked is where the task ended up, and the run
	// under test may well have refused and printed nothing about it.
	sc.Step(`^task (\d+) is in the column "([^"]*)"$`, func(n int, column string) error {
		w.Run(fmt.Sprintf(`cairn task show %d --json`, n))
		shown := w.Out
		var said struct {
			State string `json:"state"`
		}
		if err := json.Unmarshal([]byte(shown.Stdout), &said); err != nil {
			return fmt.Errorf("reading what task %d carries: %w", n, err)
		}
		if said.State != column {
			return fmt.Errorf("task %d is in %q, not %q", n, said.State, column)
		}
		return nil
	})

	sc.Step(`^it says a sub-task is done or it is not$`, func() error {
		return test.WantContains(w.Out.Stderr, "done or it is not", "the refusal")
	})

	sc.Step(`^it says nothing of task (\d+)$`, func(n int) error {
		if strings.Contains(w.Out.Stdout, fmt.Sprintf("#%d ", n)) {
			return fmt.Errorf("the listing names task %d:\n%s", n, w.Out.Stdout)
		}
		return nil
	})

	// Named as done or not: a section that listed what hangs under a task
	// without saying which of them were finished made somebody open each one
	// to find out, which is the question the section exists to answer.
	sc.Step(`^it names task (\d+) as (done|not done)$`, func(n int, how string) error {
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			if !strings.Contains(line, fmt.Sprintf("#%d ", n)) {
				continue
			}
			if got := strings.Contains(line, task.DoneMark); got != (how == "done") {
				return fmt.Errorf("task %d reads %q, and it is %s", n, line, how)
			}
			return nil
		}
		return fmt.Errorf("task %d is not named at all:\n%s", n, w.Out.Stdout)
	})

	sc.Step(`^it carries task (\d+) under it, (done|not done)$`, func(n int, how string) error {
		var said struct {
			SubTasks []struct {
				Number int    `json:"number"`
				Title  string `json:"title"`
				State  string `json:"state"`
			} `json:"sub_tasks"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &said); err != nil {
			return fmt.Errorf("reading what it carries: %w", err)
		}
		for _, under := range said.SubTasks {
			if under.Number != n {
				continue
			}
			if got := under.State == "done"; got != (how == "done") {
				return fmt.Errorf("task %d is %q, and it is %s", n, under.State, how)
			}
			return nil
		}
		return fmt.Errorf("nothing under it is task %d: %+v", n, said.SubTasks)
	})

	sc.Step(`^I archived task (\d+)$`, func(n int) error {
		return w.Seed(fmt.Sprintf("cairn task archive %d", n))
	})

	sc.Step(`^task (\d+) is archived$`, func(n int) error { return w.archived(n, true) })
	sc.Step(`^task (\d+) is not archived$`, func(n int) error { return w.archived(n, false) })

	sc.Step(`^it says (\d+) tasks? were archived$`, func(n int) error {
		return test.WantContains(w.Out.Stdout, fmt.Sprintf("%d", n), "what it shows")
	})

	sc.Step(`^it says task (\d+) is what would be archived$`, func(n int) error {
		return test.WantContains(w.Out.Stderr, fmt.Sprintf("task %d", n), "the refusal")
	})

	sc.Step(`^it says (\d+) tasks? (?:was|were) labelled$`, func(n int) error {
		return test.WantContains(w.Out.Stdout, fmt.Sprintf("%d", n), "what it shows")
	})

	sc.Step(`^task (\d+) carries the label "([^"]*)"$`, func(n int, label string) error {
		return w.carriesLabel(n, label, true)
	})
	sc.Step(`^task (\d+) does not carry the label "([^"]*)"$`, func(n int, label string) error {
		return w.carriesLabel(n, label, false)
	})

	sc.Step(`^it says what it needs to go on$`, func() error {
		return test.WantContains(w.Out.Stderr, "narrow", "the refusal")
	})

	sc.Step(`^it says the task is archived$`, func() error {
		return test.WantContains(w.Out.Stdout, "archived", "what it shows")
	})

	sc.Step(`^it says the task is on the board$`, func() error {
		return test.WantContains(w.Out.Stderr, "on the board", "the refusal")
	})
	sc.Step(`^it says the task is already archived$`, func() error {
		return test.WantContains(w.Out.Stderr, "already archived", "the refusal")
	})

	sc.Step(`^it says there is no task$`, func() error {
		return test.WantContains(w.Out.Stdout, "there is no task", "the answer")
	})

	sc.Step(`^it says nothing is in focus$`, func() error {
		return test.WantContains(w.Out.Stdout, "nothing is in focus", "the answer")
	})

	sc.Step(`^it says there is no objective$`, func() error {
		return test.WantContains(w.Out.Stdout, "there is no objective", "the answer")
	})

	sc.Step(`^it says task (\d+) was deleted$`, func(n int) error {
		return test.WantContains(w.Out.Stdout, fmt.Sprintf("task %d was deleted", n), "the answer")
	})

	// --- what the task carries ----------------------------------------------

	sc.Step(`^it reports (?:a|the) task titled "([^"]*)"$`, func(title string) error {
		return test.WantContains(w.heading(), title, "the heading")
	})

	sc.Step(`^it reports a task with that exact title$`, func() error {
		return test.WantContains(w.heading(), w.long, "the heading")
	})

	sc.Step(`^it numbers the task (\d+)$`, func(n int) error {
		return test.WantContains(w.heading(), "#"+strconv.Itoa(n), "the heading")
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

	// What blocks a task, read off the card a person reads rather than out of
	// the store: what is asserted is that the card answers "why is this
	// stopped?", and a store the card never showed would answer nothing.
	sc.Step("^`([^`]*)` says task (\\d+) blocks it$", func(line string, n int) error {
		return w.blocksIt(line, fmt.Sprintf("#%d", n))
	})

	sc.Step("^`([^`]*)` says \"([^\"]*)\" blocks it$", func(line, words string) error {
		return w.blocksIt(line, words)
	})

	// What a task serves, shared rather than kept with the objectives: the
	// migration that carries an attachment over has to be able to read it back,
	// and so does anything else that changes what a task carries.
	sc.Step("^`([^`]*)` says it serves objectives (\\d+) and (\\d+)$",
		func(line string, first, second int) error {
			return w.servesExactly(line, first, second)
		})

	sc.Step("^`([^`]*)` says it serves objective (\\d+) alone$",
		func(line string, only int) error {
			return w.servesExactly(line, only)
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

	sc.Step(`^it lists (\d+) tasks?$`, func(n int) error {
		return test.CountsTasks(w.Out.Stdout, n)
	})

	sc.Step("^`([^`]*)` lists task (\\d+) titled \"([^\"]*)\"$", func(line string, n int, title string) error {
		w.Run(line)
		return test.ListsTask(w.Out.Stdout, n, title)
	})

	sc.Step(`^it lists task (\d+) titled "([^"]*)"$`, func(n int, title string) error {
		return test.ListsTask(w.Out.Stdout, n, title)
	})

	sc.Step(`^the line for task (\d+) says "([^"]*)"$`, func(n int, said string) error {
		row, err := w.listedRow(n)
		if err != nil {
			return err
		}
		if !strings.Contains(row, said) {
			return fmt.Errorf("the line for task %d reads %q, and says nothing of %q", n, row, said)
		}
		return nil
	})

	sc.Step("^`([^`]*)` says there is no task$", func(line string) error {
		w.Run(line)
		return test.WantContains(w.Out.Stdout, "there is no task", "the answer")
	})

	sc.Step("^`([^`]*)` still reports a task titled \"([^\"]*)\"$", func(line, title string) error {
		w.Run(line)
		return test.WantContains(w.heading(), title, "the heading")
	})

	sc.Step("^`([^`]*)` says there is no such task$", func(line string) error {
		w.Run(line)
		return test.WantContains(w.Out.Stderr, "there is no such task", "the reason")
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
		at := func(n int) int { return strings.Index(w.Out.Stdout, "#"+strconv.Itoa(n)+" ") }
		if at(under) > at(hung) {
			return fmt.Errorf("task %d is listed before task %d it hangs under:\n%s", hung, under, w.Out.Stdout)
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
			return strings.Index(w.Out.Stdout, "#"+strconv.Itoa(n)+" ")
		}
		if at(first) < 0 || at(second) < 0 || at(first) > at(second) {
			return fmt.Errorf("task %d is not listed before task %d:\n%s", first, second, w.Out.Stdout)
		}
		return nil
	})

	// --- the store ------------------------------------------------------------

	sc.Step(`^the store exists$`, func() error {
		if _, err := os.Stat(w.StorePath); err != nil {
			return fmt.Errorf("the store was not created at %s: %w", w.StorePath, err)
		}
		return nil
	})

	sc.Step(`^the store is created where I said$`, func() error {
		if _, err := os.Stat(w.StorePath); err != nil {
			return fmt.Errorf("the store was not created at %s: %w", w.StorePath, err)
		}
		return nil
	})

	sc.Step(`^nothing is written to the default place$`, func() error {
		// The whole folder, not just cairn.db: SQLite leaves -wal and -shm
		// beside it, and a check that missed them would pass on a store that
		// was very much written to.
		left, err := os.ReadDir(filepath.Join(w.Defaults, config.DefaultFolder))
		if err != nil {
			return nil // the folder was never created, which is the point
		}
		if len(left) > 0 {
			return fmt.Errorf("the default place holds %d entries, want none", len(left))
		}
		return nil
	})

	sc.Step(`^the store holds something that is not a cairn store$`, func() error {
		if err := os.MkdirAll(filepath.Dir(w.StorePath), 0o700); err != nil {
			return err
		}
		return os.WriteFile(w.StorePath, []byte("this is not a cairn store"), 0o600)
	})

	sc.Step(`^it names the store$`, func() error {
		return test.WantContains(w.Out.Stderr, w.StorePath, "the reason")
	})

	sc.Step(`^the store is left as it was$`, func() error {
		held, err := os.ReadFile(w.StorePath)
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
		if err := json.Unmarshal([]byte(w.Out.Stdout), &any); err != nil {
			return fmt.Errorf("stdout is not JSON (%w):\n%s", err, w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^it says in JSON the refusal is "([^"]*)", because "([^"]*)"$`, func(kind, reason string) error {
		var said engine.Refusal
		if err := json.Unmarshal([]byte(w.Out.Stdout), &said); err != nil {
			return fmt.Errorf("stdout is not a refusal in JSON (%w):\n%s", err, w.Out.Stdout)
		}
		if said.Kind != kind || said.Reason != reason {
			return fmt.Errorf("refusal = %+v, want %s because %s", said, kind, reason)
		}
		return nil
	})

	sc.Step(`^it says task (\d+) was waiting on it$`, func(n int) error {
		return test.WantContains(w.Out.Stdout, fmt.Sprintf("task %d was waiting on it", n), "the deletion")
	})

	sc.Step(`^it names task (\d+)$`, func(n int) error {
		return test.WantContains(strings.ToLower(w.Out.Stderr), fmt.Sprintf("task %d", n), "the refusal")
	})

	sc.Step("^I run `([^`]*)` with \"([^\"]*)\" and \"([^\"]*)\" on standard input$", func(line, first, second string) error {
		var out, errs bytes.Buffer
		stdin := strings.NewReader(first + "\n\n" + second + "\n")
		code := Main(context.Background(), test.SplitArgs(line)[1:], stdin, &out, &errs)
		w.Line, w.Out = line, test.Result{Stdout: out.String(), Stderr: errs.String(), Code: code}
		return nil
	})

	sc.Step("^`([^`]*)` shows \"([^\"]*)\"$", func(line, said string) error {
		r := run(w.T, test.SplitArgs(line)[1:]...)
		return test.WantContains(r.Stdout, said, line)
	})

	sc.Step(`^I add a task whose description carries an escape$`, func() error {
		w.Out = run(w.T, "task", "add", "Write the spec", "--desc", "look \x1b]52;c;aGk=\x07 here")
		return nil
	})

	sc.Step(`^it says only text, line breaks and tabs can be written there$`, func() error {
		return test.WantContains(strings.ToLower(w.Out.Stderr), "only text, line breaks and tabs", "the refusal")
	})

	sc.Step(`^it says the task has stood in "([^"]*)" since today$`, func(column string) error {
		for _, day := range daysToday() {
			if strings.Contains(w.Out.Stdout, "in "+column+" since "+day) {
				return nil
			}
		}
		return fmt.Errorf("it does not say the task has stood in %q since today:\n%s", column, w.Out.Stdout)
	})

	sc.Step(`^it says in JSON the task has stood in its column since today$`, func() error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return err
		}
		said, _ := doc["in_column_since"].(string)
		since, err := time.Parse(time.RFC3339Nano, said)
		if err != nil {
			return fmt.Errorf("in_column_since = %v: %w", doc["in_column_since"], err)
		}
		for _, day := range daysToday() {
			if domain.FormatDay(since) == day {
				return nil
			}
		}
		return fmt.Errorf("in_column_since = %s, want today", said)
	})

	sc.Step(`^it carries the title "([^"]*)"$`, func(title string) error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return err
		}
		if doc["title"] != title {
			return fmt.Errorf("title = %v, want %q", doc["title"], title)
		}
		return nil
	})

	sc.Step(`^it carries the parent (\d+)$`, func(n int) error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return err
		}
		if got, ok := doc["parent"].(float64); !ok || int(got) != n {
			return fmt.Errorf("parent = %v, want %d", doc["parent"], n)
		}
		return nil
	})

	sc.Step(`^it carries an identifier$`, func() error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return err
		}
		if doc["number"] == nil {
			return fmt.Errorf("no identifier in %s", w.Out.Stdout)
		}
		return nil
	})

	// The focus is said once, at the top of a listing, and nowhere else. Both
	// halves matter: said twice, the two can disagree, and a reader has to
	// decide which to believe.
	sc.Step(`^it says task (\d+) is the one in focus$`, func(n int) error {
		_, envelope, err := listedUnder(w.Out.Stdout, "tasks")
		if err != nil {
			return err
		}
		held, said := envelope["in_focus"]
		if !said {
			return fmt.Errorf("the listing says nothing about the focus")
		}
		numbers, ok := held.([]any)
		if !ok {
			return fmt.Errorf("in_focus is %T, want a list of numbers", held)
		}
		if got, want := fmt.Sprint(numbers), fmt.Sprint([]any{float64(n)}); got != want {
			return fmt.Errorf("it says %s is in focus, want %s", got, want)
		}
		return nil
	})

	// The flag belongs to the task, so a listing the task is not in has nothing
	// to say about it — narrowing past the tasks in focus says none of these
	// rather than naming a task the answer does not carry.
	sc.Step(`^it says no task in it is in focus$`, func() error {
		_, envelope, err := listedUnder(w.Out.Stdout, "tasks")
		if err != nil {
			return err
		}
		held := envelope["in_focus"]
		if held == nil {
			return fmt.Errorf("the listing says nothing about the focus, and it was asked")
		}
		if numbers, ok := held.([]any); !ok || len(numbers) != 0 {
			return fmt.Errorf("it says %v is in focus, and the answer does not carry it", held)
		}
		return nil
	})

	sc.Step(`^no task in it carries that on its own$`, func() error {
		rows, _, err := listedUnder(w.Out.Stdout, "tasks")
		if err != nil {
			return err
		}
		for _, row := range rows {
			if _, carried := row["in_focus"]; carried {
				return fmt.Errorf("task %v carries the focus on its own line too", row["number"])
			}
		}
		return nil
	})

	// A listing names the work and leaves what each task says to be shown: a
	// board of a hundred tasks handed over with every description came to more
	// than an agent can read in one answer.
	sc.Step(`^no task in it carries its description$`, func() error {
		rows, _, err := listedUnder(w.Out.Stdout, "tasks")
		if err != nil {
			return err
		}
		for _, row := range rows {
			if _, carried := row["description"]; carried {
				return fmt.Errorf("task %v carries its description in the listing", row["number"])
			}
		}
		return nil
	})

	sc.Step(`^it says task (\d+) (has a description|has none)$`, func(n int, said string) error {
		rows, _, err := listedUnder(w.Out.Stdout, "tasks")
		if err != nil {
			return err
		}
		for _, row := range rows {
			if number, _ := row["number"].(float64); int(number) != n {
				continue
			}
			if want := said == "has a description"; row["described"] != want {
				return fmt.Errorf("task %d says described = %v, want %v", n, row["described"], want)
			}
			return nil
		}
		return fmt.Errorf("the listing carries no task %d", n)
	})

	sc.Step(`^it carries the description "([^"]*)"$`, func(said string) error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return err
		}
		if doc["description"] != said {
			return fmt.Errorf("description = %v, want %q", doc["description"], said)
		}
		return nil
	})

	sc.Step(`^it carries (\d+) tasks?$`, func(n int) error {
		rows, _, err := listedUnder(w.Out.Stdout, "tasks")
		if err != nil {
			return err
		}
		if len(rows) != n {
			return fmt.Errorf("%d tasks, want %d", len(rows), n)
		}
		return nil
	})

	// --- how far along something is -------------------------------------------
	//
	// Shared by tasks and objectives: both are asked the same question and
	// answer it in the same words, so one vocabulary serves both features.

	sc.Step(`^it says (\d+) of (\d+) (tasks|sub-tasks) are done$`,
		func(done, total int, kind string) error {
			return test.WantContains(w.Out.Stdout,
				fmt.Sprintf("%d of %d %s are done", done, total, kind), "the answer")
		})

	sc.Step(`^it says (\d+) of (\d+) tasks and (\d+) of (\d+) sub-tasks are done$`,
		func(tasksDone, tasksTotal, subDone, subTotal int) error {
			return test.WantContains(w.Out.Stdout,
				fmt.Sprintf("%d of %d tasks and %d of %d sub-tasks are done",
					tasksDone, tasksTotal, subDone, subTotal), "the answer")
		})

	sc.Step(`^it says the (task|objective) is (\d+)% done$`, func(what string, share int) error {
		return test.WantContains(w.Out.Stdout, fmt.Sprintf("the %s is %d%% done", what, share), "the answer")
	})

	// Looked for by shape rather than by the word "done", which a state of
	// its own puts on the card: `state done` would match either way.
	sc.Step(`^it does not say how much is done$`, func() error {
		if strings.Contains(w.Out.Stdout, "% done") {
			return fmt.Errorf("the answer works out a share it has nothing to work out:\n%s", w.Out.Stdout)
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
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return err
		}
		if _, present := doc["progress"]; !present {
			return fmt.Errorf("no progress key at all in %s", w.Out.Stdout)
		}
		if doc["progress"] != nil {
			return fmt.Errorf("progress = %v, want null", doc["progress"])
		}
		return nil
	})

	sc.Step(`^the JSON carries the reason$`, func() error {
		var doc map[string]string
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return err
		}
		if doc["error"] == "" {
			return fmt.Errorf("no reason in %s", w.Out.Stdout)
		}
		return nil
	})
}

// blocksIt runs a command line and fails unless the card it answered names
// what is said to block the task, under the heading that says so.
//
// The heading is looked for as well as the thing: a title that happened to
// contain "#3" would otherwise pass for a blocker, and what is being asserted
// is that the card says this is what stops the task.
func (w *world) blocksIt(line, blocker string) error {
	w.Run(line)
	if w.Out.Code != exitOK {
		return fmt.Errorf("%s failed: exit %d (%s)", line, w.Out.Code, w.Out.Stderr)
	}
	blocking := w.field("blocked by")
	if blocking == "" {
		return fmt.Errorf("the card says nothing blocks it:\n%s", w.Out.Stdout)
	}
	if !strings.Contains(blocking, blocker) {
		return fmt.Errorf("what blocks it is %q, and does not name %q", blocking, blocker)
	}
	return nil
}

// archived asks the store what a task carries, through the real command: what
// is being checked is what was kept, not what the last run printed.
func (w *world) archived(number int, want bool) error {
	w.Run(fmt.Sprintf("cairn task show %d --json", number))
	var said struct {
		Archived bool `json:"archived"`
	}
	if err := json.Unmarshal([]byte(w.Out.Stdout), &said); err != nil {
		return fmt.Errorf("reading what task %d carries: %w", number, err)
	}
	if said.Archived != want {
		return fmt.Errorf("task %d archived = %v, want %v", number, said.Archived, want)
	}
	return nil
}

// carriesLabel asks the store what a task carries, through the real command.
func (w *world) carriesLabel(number int, label string, want bool) error {
	w.Run(fmt.Sprintf("cairn task show %d --json", number))
	var said struct {
		Labels []string `json:"labels"`
	}
	if err := json.Unmarshal([]byte(w.Out.Stdout), &said); err != nil {
		return fmt.Errorf("reading what task %d carries: %w", number, err)
	}
	if slices.Contains(said.Labels, label) != want {
		return fmt.Errorf("task %d carries %v, want %q to be there: %v", number, said.Labels, label, want)
	}
	return nil
}
