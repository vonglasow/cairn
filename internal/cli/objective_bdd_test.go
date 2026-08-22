package cli

// Executable acceptance scenarios for bdd/objective.feature. The Gherkin is
// the contract; the steps an objective does not share with a task live here,
// beside the feature that claims them.

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

func initObjective(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerObjective(sc)
	}
}

func TestAcceptance_Objective(t *testing.T) {
	runAcceptance(t, "objective.feature", initObjective(t), "")
}

// wantField checks one field of the card a command printed.
func (w *world) wantField(key, want string) error {
	if got := w.field(key); got != want {
		return fmt.Errorf("%s = %q, want %q", key, got, want)
	}
	return nil
}

// wantLabel checks that the card carries a label among the ones it lists.
func (w *world) wantLabel(want string) error {
	for _, label := range strings.Split(w.field("labels"), ", ") {
		if label == want {
			return nil
		}
	}
	return fmt.Errorf("labels = %q, want one of them to be %q", w.field("labels"), want)
}

// registerObjective declares the steps only objectives use.
func (w *world) registerObjective(sc *godog.ScenarioContext) {
	// --- the common context ---------------------------------------------------

	sc.Step(`^I add(?:ed)? an objective with it$`, func() error {
		w.run(fmt.Sprintf(`cairn objective add "%s"`, w.long))
		return nil
	})

	sc.Step(`^I added an objective titled "([^"]*)" with the (description|deadline|priority|state|estimate) "([^"]*)"$`,
		func(title, attribute, value string) error {
			flag := map[string]string{
				"description": "--desc", "deadline": "--deadline",
				"priority": "--priority", "state": "--state", "estimate": "--estimate",
			}[attribute]
			return w.seed(fmt.Sprintf(`cairn objective add "%s" %s "%s"`, title, flag, value))
		})

	sc.Step(`^I added an objective titled "([^"]*)" with the labels "([^"]*)" and "([^"]*)"$`,
		func(title, first, second string) error {
			return w.seed(fmt.Sprintf(`cairn objective add "%s" --label "%s" --label "%s"`, title, first, second))
		})

	sc.Step(`^I added a task titled "([^"]*)" with the objective "([^"]*)"$`,
		func(title, objective string) error {
			return w.seed(fmt.Sprintf(`cairn task add "%s" --objective "%s"`, title, objective))
		})

	// Written as one step rather than three hundred lines of Given. The share
	// only bites at a scale nobody types out, and a scenario nobody can read is
	// a scenario nobody checks.
	sc.Step(`^I added (\d+) tasks under it, of which (\d+) (?:is|are) done$`,
		func(total, done int) error {
			for i := 1; i <= total; i++ {
				if err := w.seed(fmt.Sprintf(`cairn task add "Task %d" --objective 1`, i)); err != nil {
					return err
				}
				if i <= done {
					if err := w.seed(fmt.Sprintf(`cairn task edit %d --state done`, i)); err != nil {
						return err
					}
				}
			}
			return nil
		})

	// --- what it answered -----------------------------------------------------

	sc.Step(`^it reports an objective titled "([^"]*)"$`, func(title string) error {
		return wantContains(w.heading(), title, "the answer")
	})

	sc.Step(`^it numbers the objective (\d+)$`, func(n int) error {
		return wantContains(w.heading(), fmt.Sprintf("#%d", n), "the answer")
	})

	sc.Step(`^it lists (\d+) objectives?$`, func(n int) error {
		return countsTasks(w.out.stdout, n)
	})

	sc.Step(`^it lists objective (\d+) titled "([^"]*)"$`, func(n int, title string) error {
		return listsTask(w.out.stdout, n, title)
	})

	sc.Step("^`([^`]*)` says there is no objective$", func(line string) error {
		w.run(line)
		return wantContains(w.out.stdout, "there is no objective", "the answer")
	})

	sc.Step("^`([^`]*)` still reports an objective titled \"([^\"]*)\"$", func(line, title string) error {
		w.run(line)
		return wantContains(w.heading(), title, "the answer")
	})

	sc.Step("^`([^`]*)` still carries the objective \"([^\"]*)\"$", func(line, want string) error {
		w.run(line)
		return w.wantField("objective", want)
	})

	sc.Step("^`([^`]*)` says nothing is under it$", func(line string) error {
		w.run(line)
		return wantContains(w.out.stdout, "nothing is under it", "the answer")
	})

	sc.Step(`^it says nothing is under it$`, func() error {
		return wantContains(w.out.stdout, "nothing is under it", "the answer")
	})

	// The task appears indented under the objective's own heading, the way a
	// sub-task appears under its parent — same shape, so it reads the same.
	sc.Step(`^task (\d+) is shown under objective (\d+)$`, func(task, objective int) error {
		if err := wantContains(w.heading(), fmt.Sprintf("#%d", objective), "the heading"); err != nil {
			return err
		}
		_, block, found := strings.Cut(w.out.stdout, "\ntasks\n")
		if !found {
			return fmt.Errorf("nothing is shown under the objective:\n%s", w.out.stdout)
		}
		want := fmt.Sprintf("#%d ", task)
		for _, row := range strings.Split(block, "\n") {
			if !strings.HasPrefix(strings.TrimSpace(row), want) {
				continue
			}
			if !strings.HasPrefix(row, " ") {
				return fmt.Errorf("task %d is not set under the objective: %q", task, row)
			}
			return nil
		}
		return fmt.Errorf("task %d is not among what is under the objective:\n%s", task, block)
	})

	// --- what it carries ------------------------------------------------------

	sc.Step(`^the objective carries the (description|deadline|estimate|priority|state|label) "([^"]*)"$`,
		func(attribute, want string) error {
			if attribute == "label" {
				return w.wantLabel(want)
			}
			return w.wantField(attribute, want)
		})

	sc.Step(`^the objective is (active|on-hold|reached|dropped)$`, func(want string) error {
		return w.wantField("state", want)
	})

	sc.Step(`^the objective has no (description|labels|deadline|estimate)$`, func(attribute string) error {
		key := strings.TrimSuffix(attribute, "s")
		if attribute == "labels" {
			key = "labels"
		}
		if got := w.field(key); got != "" {
			return fmt.Errorf("%s = %q, want nothing", attribute, got)
		}
		return nil
	})

	sc.Step(`^the objective carries exactly (\d+) labels?$`, func(n int) error {
		got := w.field("labels")
		if got == "" {
			if n == 0 {
				return nil
			}
			return fmt.Errorf("no labels, want %d", n)
		}
		if count := len(strings.Split(got, ", ")); count != n {
			return fmt.Errorf("%d labels (%s), want %d", count, got, n)
		}
		return nil
	})

	sc.Step(`^the task carries the objective "([^"]*)"$`, func(want string) error {
		return w.wantField("objective", want)
	})

	sc.Step(`^the task has no objective$`, func() error {
		if got := w.field("objective"); got != "" {
			return fmt.Errorf("objective = %q, want none", got)
		}
		return nil
	})

	// --- machine-readable answers ---------------------------------------------

	sc.Step(`^it carries (\d+) objectives?$`, func(n int) error {
		var docs []map[string]any
		if err := json.Unmarshal([]byte(w.out.stdout), &docs); err != nil {
			return err
		}
		if len(docs) != n {
			return fmt.Errorf("%d objectives, want %d", len(docs), n)
		}
		return nil
	})

	sc.Step(`^it carries the tasks under it$`, func() error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.out.stdout), &doc); err != nil {
			return err
		}
		under, ok := doc["tasks"].([]any)
		if !ok || len(under) == 0 {
			return fmt.Errorf("tasks = %v, want the numbers put under it", doc["tasks"])
		}
		return nil
	})

	// --- what it refused ------------------------------------------------------

	for step, said := range map[string]string{
		// `it says how to confirm` is not here: a task already claims it, and
		// --yes reads the same whichever thing was about to go away.
		`^it says there is no such objective$`:      "there is no such objective",
		`^it says an objective needs a title$`:      "an objective needs a title",
		`^it says a sub-task serves no objective$`:  "a sub-task serves no objective",
		`^it says the objective carries tasks$`:     "the objective carries tasks",
		`^it says an objective number is expected$`: "an objective number is expected",
	} {
		sc.Step(step, func() error { return wantContains(w.out.stderr, said, "the reason") })
	}
}
