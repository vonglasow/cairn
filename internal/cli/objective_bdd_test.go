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

	"cairn/test"
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
		w.Run(fmt.Sprintf(`cairn objective new "%s"`, w.long))
		return nil
	})

	sc.Step(`^I added an objective titled "([^"]*)" with the (description|deadline|priority|state|estimate) "([^"]*)"$`,
		func(title, attribute, value string) error {
			flag := map[string]string{
				"description": "--desc", "deadline": "--deadline",
				"priority": "--priority", "state": "--state", "estimate": "--estimate",
			}[attribute]
			return w.Seed(fmt.Sprintf(`cairn objective new "%s" %s "%s"`, title, flag, value))
		})

	sc.Step(`^I added an objective titled "([^"]*)" with the labels "([^"]*)" and "([^"]*)"$`,
		func(title, first, second string) error {
			return w.Seed(fmt.Sprintf(`cairn objective new "%s" --label "%s" --label "%s"`, title, first, second))
		})

	// Written as one step rather than three hundred lines of Given. The share
	// only bites at a scale nobody types out, and a scenario nobody can read is
	// a scenario nobody checks.
	sc.Step(`^I added (\d+) tasks under it, of which (\d+) (?:is|are) done$`,
		func(total, done int) error {
			for i := 1; i <= total; i++ {
				if err := w.Seed(fmt.Sprintf(`cairn task new "Task %d" --objective 1`, i)); err != nil {
					return err
				}
				if i <= done {
					if err := w.Seed(fmt.Sprintf(`cairn task edit %d --state done`, i)); err != nil {
						return err
					}
				}
			}
			return nil
		})

	// --- what it answered -----------------------------------------------------

	sc.Step(`^it reports an objective titled "([^"]*)"$`, func(title string) error {
		return test.WantContains(w.heading(), title, "the answer")
	})

	sc.Step(`^it numbers the objective (\d+)$`, func(n int) error {
		return test.WantContains(w.heading(), fmt.Sprintf("#%d", n), "the answer")
	})

	sc.Step(`^it lists (\d+) objectives?$`, func(n int) error {
		return test.CountsTasks(w.Out.Stdout, n)
	})

	sc.Step(`^it lists objective (\d+) titled "([^"]*)"$`, func(n int, title string) error {
		return test.ListsTask(w.Out.Stdout, n, title)
	})

	sc.Step("^`([^`]*)` says there is no objective$", func(line string) error {
		w.Run(line)
		return test.WantContains(w.Out.Stdout, "there is no objective", "the answer")
	})

	sc.Step("^`([^`]*)` still reports an objective titled \"([^\"]*)\"$", func(line, title string) error {
		w.Run(line)
		return test.WantContains(w.heading(), title, "the answer")
	})

	sc.Step("^`([^`]*)` still carries the objective \"([^\"]*)\"$", func(line, want string) error {
		w.Run(line)
		return w.wantField("objectives", served(want))
	})

	sc.Step("^`([^`]*)` says nothing is under it$", func(line string) error {
		w.Run(line)
		return test.WantContains(w.Out.Stdout, "nothing is under it", "the answer")
	})

	sc.Step(`^it says nothing is under it$`, func() error {
		return test.WantContains(w.Out.Stdout, "nothing is under it", "the answer")
	})

	// How many tasks an objective is holding, counted off its card. Counted
	// rather than looked for as a number: the card lists them, and a count
	// written on it would be a second thing to keep true.
	sc.Step(`^it says (\d+) tasks? (?:is|are) under it$`, func(n int) error {
		return countedUnder(w.Out.Stdout, n)
	})

	sc.Step("^`([^`]*)` says (\\d+) tasks? (?:is|are) under it$",
		func(line string, n int) error {
			w.Run(line)
			return countedUnder(w.Out.Stdout, n)
		})

	sc.Step("^`([^`]*)` says the objective is (\\d+)% done$",
		func(line string, share int) error {
			w.Run(line)
			return test.WantContains(w.Out.Stdout,
				fmt.Sprintf("the objective is %d%% done", share), "the answer")
		})

	// The task appears indented under the objective's own heading, the way a
	// sub-task appears under its parent — same shape, so it reads the same.
	sc.Step(`^task (\d+) is shown under objective (\d+)$`, func(task, objective int) error {
		if err := test.WantContains(w.heading(), fmt.Sprintf("#%d", objective), "the heading"); err != nil {
			return err
		}
		_, block, found := strings.Cut(w.Out.Stdout, "\ntasks\n")
		if !found {
			return fmt.Errorf("nothing is shown under the objective:\n%s", w.Out.Stdout)
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

	// A task serves a set, so the card lays out "objectives". These two steps
	// keep the wording the frozen scenarios use: what changed is where the
	// answer is read, not what is being asserted.
	sc.Step(`^the task carries the objective "([^"]*)"$`, func(want string) error {
		return w.wantField("objectives", served(want))
	})

	sc.Step(`^the task has no objective$`, func() error {
		if got := w.field("objectives"); got != "" {
			return fmt.Errorf("objectives = %q, want none", got)
		}
		return nil
	})

	// --- machine-readable answers ---------------------------------------------

	sc.Step(`^it carries (\d+) objectives?$`, func(n int) error {
		rows, _, err := listedUnder(w.Out.Stdout, "objectives")
		if err != nil {
			return err
		}
		if len(rows) != n {
			return fmt.Errorf("%d objectives, want %d", len(rows), n)
		}
		return nil
	})

	// What a task serves crosses as a list, named "objectives". Read out of the
	// JSON rather than off the card: the two readings come from one engine, and
	// a step that checked the card twice would prove the wire nothing.
	sc.Step(`^it says the task serves objectives (\d+) and (\d+)$`,
		func(first, second int) error {
			var card struct {
				Objectives []int `json:"objectives"`
			}
			if err := json.Unmarshal([]byte(w.Out.Stdout), &card); err != nil {
				return fmt.Errorf("the card is not JSON (%w):\n%s", err, w.Out.Stdout)
			}
			if len(card.Objectives) != 2 || card.Objectives[0] != first || card.Objectives[1] != second {
				return fmt.Errorf("objectives = %v, want [%d %d]", card.Objectives, first, second)
			}
			return nil
		})

	sc.Step(`^it carries the tasks under it$`, func() error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
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
		sc.Step(step, func() error { return test.WantContains(w.Out.Stderr, said, "the reason") })
	}
}

// served writes an objective number the way a card writes it, so a step can
// compare what it asked for against what was laid out.
func served(number string) string {
	return "#" + number
}

// servesExactly runs a command line and fails unless the card names exactly
// these objectives. Exactly, and not merely among them: a set that quietly
// kept one nobody asked for would read the same as one that did not.
func (w *world) servesExactly(line string, numbers ...int) error {
	w.Run(line)
	if w.Out.Code != exitOK {
		return fmt.Errorf("%s failed: exit %d (%s)", line, w.Out.Code, w.Out.Stderr)
	}
	want := make([]string, 0, len(numbers))
	for _, one := range numbers {
		want = append(want, served(fmt.Sprint(one)))
	}
	if got := w.field("objectives"); got != strings.Join(want, ", ") {
		return fmt.Errorf("it serves %q, want %q", got, strings.Join(want, ", "))
	}
	return nil
}

// countedUnder counts the tasks an objective's card lists under it.
func countedUnder(card string, want int) error {
	_, block, found := strings.Cut(card, "\ntasks\n")
	if !found {
		if want == 0 {
			return nil
		}
		return fmt.Errorf("nothing is under it, want %d:\n%s", want, card)
	}
	count := 0
	for _, row := range strings.Split(block, "\n") {
		if strings.HasPrefix(strings.TrimSpace(row), "#") {
			count++
		}
	}
	if count != want {
		return fmt.Errorf("%d tasks are under it, want %d:\n%s", count, want, card)
	}
	return nil
}
