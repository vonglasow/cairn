package cli

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

func (w *world) wantField(key, want string) error {
	if got := w.field(key); got != want {
		return fmt.Errorf("%s = %q, want %q", key, got, want)
	}
	return nil
}

func (w *world) wantLabel(want string) error {
	for _, label := range strings.Split(w.field("labels"), ", ") {
		if label == want {
			return nil
		}
	}
	return fmt.Errorf("labels = %q, want one of them to be %q", w.field("labels"), want)
}

func (w *world) registerObjective(sc *godog.ScenarioContext) {

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

	sc.Step(`^the objective carries the (description|deadline|estimate|priority|state|label|time spent|left) "([^"]*)"$`,
		func(attribute, want string) error {
			if attribute == "label" {
				return w.wantLabel(want)
			}
			return w.wantField(attribute, want)
		})

	sc.Step(`^the row for objective (\d+) says "([^"]*)"$`, func(n int, want string) error {
		row, err := w.listedRow(n)
		if err != nil {
			return err
		}
		if !strings.Contains(row, want) {
			return fmt.Errorf("the row for objective %d is %q, and does not say %q", n, row, want)
		}
		return nil
	})
	sc.Step(`^the row for objective (\d+) says nothing of how far along it is$`, func(n int) error {
		row, err := w.listedRow(n)
		if err != nil {
			return err
		}
		if strings.Contains(row, "done") {
			return fmt.Errorf("the row for objective %d is %q", n, row)
		}
		return nil
	})
	sc.Step(`^the JSON objective carries the time spent "([^"]*)" and leaves the tasks "([^"]*)"$`, func(spent, left string) error {
		var got struct {
			TimeSpent string `json:"time_spent"`
			Left      []int  `json:"left"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &got); err != nil {
			return err
		}
		if got.TimeSpent != spent || fmt.Sprint(got.Left) != "["+left+"]" {
			return fmt.Errorf("the JSON says %+v", got)
		}
		return nil
	})
	sc.Step(`^the JSON list gives objective (\d+) at (\d+)% done$`, func(n, percent int) error {
		var got struct {
			Objectives []struct {
				Number   int `json:"number"`
				Progress *struct {
					Percent int `json:"percent"`
				} `json:"progress"`
			} `json:"objectives"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &got); err != nil {
			return err
		}
		for _, o := range got.Objectives {
			if o.Number == n && o.Progress != nil && o.Progress.Percent == percent {
				return nil
			}
		}
		return fmt.Errorf("the JSON list says %s", w.Out.Stdout)
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
		return w.wantField("objectives", served(want))
	})

	sc.Step(`^the task has no objective$`, func() error {
		if got := w.field("objectives"); got != "" {
			return fmt.Errorf("objectives = %q, want none", got)
		}
		return nil
	})

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

	for step, said := range map[string]string{
		`^it says there is no such objective$`:      "there is no such objective",
		`^it says an objective needs a title$`:      "an objective needs a title",
		`^it says the objective carries tasks$`:     "the objective carries tasks",
		`^it says an objective number is expected$`: "an objective number is expected",
	} {
		sc.Step(step, func() error { return test.WantContains(w.Out.Stderr, said, "the reason") })
	}
}

func served(number string) string {
	return "#" + number
}

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
