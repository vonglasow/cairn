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
		w.registerTimesheet(sc)
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

	sc.Step(`^I added an objective titled "([^"]*)" with the (description|deadline|priority|state|estimate) "([^"]*)"$`,
		func(title, attribute, value string) error {
			flag := map[string]string{
				"description": "--desc", "deadline": "--deadline",
				"priority": "--priority", "state": "--state", "estimate": "--estimate",
			}[attribute]
			return w.Seed(fmt.Sprintf(`cairn objective new "%s" %s "%s"`, title, flag, value))
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

	sc.Step(`^the objective carries the (description|deadline|estimate|priority|state|label|time spent|left) "([^"]*)"$`,
		func(attribute, want string) error {
			if attribute == "label" {
				return w.wantLabel(want)
			}
			return w.wantField(attribute, want)
		})

	sc.Step(`^the month "([^"]*)" says "([^"]*)"$`, func(month, want string) error {
		return w.wantField(month, want)
	})
	sc.Step(`^it says nothing of the month "([^"]*)"$`, func(month string) error {
		if strings.Contains(w.Out.Stdout, month) {
			return fmt.Errorf("it speaks of %s:\n%s", month, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the JSON report gives the month "([^"]*)" at "([^"]*)" and task (\d+) done$`, func(month, spent string, n int) error {
		var got struct {
			Objective struct {
				Number int `json:"number"`
			} `json:"objective"`
			Months []struct {
				Month     string `json:"month"`
				TimeSpent string `json:"time_spent"`
			} `json:"months"`
			Done []struct {
				Number int `json:"number"`
			} `json:"done"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &got); err != nil {
			return err
		}
		months := fmt.Sprint(got.Months)
		if got.Objective.Number != 1 || !strings.Contains(months, "{"+month+" "+spent+"}") || len(got.Done) != 1 || got.Done[0].Number != n {
			return fmt.Errorf("the JSON report says %s", w.Out.Stdout)
		}
		return nil
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

	sc.Step(`^the task carries the objective "([^"]*)"$`, func(want string) error {
		return w.wantField("objectives", served(want))
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

	w.registerSubObjective(sc)

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

func (w *world) registerSubObjective(sc *godog.ScenarioContext) {
	sc.Step(`^the objective sits under "([^"]*)"$`, func(want string) error {
		return w.wantField("parent", want)
	})
	sc.Step(`^the refusal says "([^"]*)"$`, func(said string) error {
		return test.WantContains(w.Out.Stderr, said, "the reason")
	})
	sc.Step(`^I edited objective (\d+) with the state "([^"]*)"$`, func(n int, state string) error {
		return w.Seed(fmt.Sprintf(`cairn objective edit %d --state %s`, n, state))
	})
	sc.Step(`^the objective waits on "([^"]*)"$`, func(want string) error {
		return w.wantField("waits on", want)
	})
}
