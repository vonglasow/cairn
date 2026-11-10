package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
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

	sc.Step("^`cairn objective list` lists them in the order ([0-9, ]+)$", func(order string) error {
		w.Run("cairn objective list --json")
		var listed struct {
			Objectives []struct {
				Number int `json:"number"`
			} `json:"objectives"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &listed); err != nil {
			return err
		}
		got := []string{}
		for _, o := range listed.Objectives {
			got = append(got, strconv.Itoa(o.Number))
		}
		if strings.Join(got, ", ") != order {
			return fmt.Errorf("listed %s, want %s", strings.Join(got, ", "), order)
		}
		return nil
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

	sc.Step(`^the month "([^"]*)" says "([^"]*)"$`, func(month, want string) error {
		return w.wantField(month, want)
	})
	sc.Step(`^it says nothing of the month "([^"]*)"$`, func(month string) error {
		if strings.Contains(w.Out.Stdout, month) {
			return fmt.Errorf("it speaks of %s:\n%s", month, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the report names task (\d+) as done in the period$`, func(n int) error {
		_, done, found := strings.Cut(w.Out.Stdout, "done in the period\n")
		if !found || !strings.Contains(done, fmt.Sprintf("#%d ", n)) {
			return fmt.Errorf("task %d is not named as done in the period:\n%s", n, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the report names nothing as done in the period$`, func() error {
		if !strings.Contains(w.Out.Stdout, "nothing was done in the period") {
			return fmt.Errorf("the report reads:\n%s", w.Out.Stdout)
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
	sc.Step(`^the row for objective (\d+) says "([^"]*)"$`, func(n int, want string) error {
		row, err := w.objectiveRow(n)
		if err != nil {
			return err
		}
		if !strings.Contains(row, want) {
			return fmt.Errorf("the row for objective %d is %q, and does not say %q", n, row, want)
		}
		return nil
	})
	sc.Step(`^the row for objective (\d+) says nothing of how far along it is$`, func(n int) error {
		row, err := w.objectiveRow(n)
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

func (w *world) objectiveRow(number int) (string, error) {
	rows := strings.Split(w.Out.Stdout, "\n")
	want := "#" + strconv.Itoa(number) + " "
	for i, row := range rows {
		if !strings.HasPrefix(strings.TrimLeft(row, " ├└│"), want) {
			continue
		}
		if i+1 < len(rows) && !strings.Contains(rows[i+1], "#") {
			return row + "\n" + rows[i+1], nil
		}
		return row, nil
	}
	return "", fmt.Errorf("objective %d is not in the listing:\n%s", number, w.Out.Stdout)
}

func blockUnder(out, heading string) (string, error) {
	_, block, found := strings.Cut(out, "\n"+heading+"\n")
	if !found {
		return "", fmt.Errorf("nothing is said under %q:\n%s", heading, out)
	}
	block, _, _ = strings.Cut(block, "\n\n")
	return block, nil
}

func (w *world) namesUnder(heading, want string) error {
	block, err := blockUnder(w.Out.Stdout, heading)
	if err != nil {
		return err
	}
	for _, row := range strings.Split(block, "\n") {
		if strings.HasPrefix(strings.TrimSpace(row), want) {
			return nil
		}
	}
	return fmt.Errorf("%q is not under %q:\n%s", want, heading, w.Out.Stdout)
}

func (w *world) jsonFamily() (*int, []int, error) {
	var got struct {
		Parent   *int  `json:"parent"`
		Children []int `json:"children"`
	}
	if err := json.Unmarshal([]byte(w.Out.Stdout), &got); err != nil {
		return nil, nil, fmt.Errorf("the answer is not JSON (%w):\n%s", err, w.Out.Stdout)
	}
	return got.Parent, got.Children, nil
}

func (w *world) drawnUnder(child, parent int) error {
	above := 0
	for _, row := range strings.Split(w.Out.Stdout, "\n") {
		if n, ok := numberOfRow(row); ok && strings.HasPrefix(row, "#") {
			above = n
			continue
		}
		trimmed := strings.TrimLeft(row, " │")
		if strings.HasPrefix(trimmed, fmt.Sprintf("├ #%d ", child)) || strings.HasPrefix(trimmed, fmt.Sprintf("└ #%d ", child)) {
			if above != parent {
				return fmt.Errorf("objective %d is drawn under objective %d:\n%s", child, above, w.Out.Stdout)
			}
			return nil
		}
	}
	return fmt.Errorf("objective %d is not drawn under another:\n%s", child, w.Out.Stdout)
}

func (w *world) registerSubObjective(sc *godog.ScenarioContext) {
	sc.Step(`^the objective sits under "([^"]*)"$`, func(want string) error {
		return w.wantField("parent", want)
	})
	sc.Step(`^the objective sits under nothing$`, func() error {
		if !strings.Contains(w.Out.Stdout, "\nparent") {
			return fmt.Errorf("the answer says nothing of a parent:\n%s", w.Out.Stdout)
		}
		return w.wantField("parent", "")
	})
	sc.Step(`^the refusal says "([^"]*)"$`, func(said string) error {
		return test.WantContains(w.Out.Stderr, said, "the reason")
	})
	sc.Step(`^I edited objective (\d+) with the state "([^"]*)"$`, func(n int, state string) error {
		return w.Seed(fmt.Sprintf(`cairn objective edit %d --state %s`, n, state))
	})
	sc.Step(`^it names the sub-objective (\d+) titled "([^"]*)"$`, func(n int, title string) error {
		return w.namesUnder("sub-objectives", fmt.Sprintf("#%d  %s", n, title))
	})
	sc.Step(`^it names task (\d+) on the objective itself$`, func(n int) error {
		return w.namesUnder("tasks on it directly", fmt.Sprintf("#%d ", n))
	})
	sc.Step(`^it names task (\d+) in its sub-objectives$`, func(n int) error {
		return w.namesUnder("tasks in its sub-objectives", fmt.Sprintf("#%d ", n))
	})
	sc.Step(`^objective (\d+) is drawn under objective (\d+)$`, w.drawnUnder)
	sc.Step(`^objective (\d+) is drawn on its own$`, func(n int) error {
		if !strings.Contains("\n"+w.Out.Stdout, fmt.Sprintf("\n#%d ", n)) {
			return fmt.Errorf("objective %d is not drawn on its own:\n%s", n, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the objective waits on "([^"]*)"$`, func(want string) error {
		return w.wantField("waits on", want)
	})
	sc.Step(`^the objective waits on nothing$`, func() error {
		if !strings.Contains(w.Out.Stdout, "\nwaits on") {
			return fmt.Errorf("the answer says nothing of what it waits on:\n%s", w.Out.Stdout)
		}
		return w.wantField("waits on", "")
	})
	sc.Step(`^the JSON objective waits on "([^"]*)" and is waited on by "([^"]*)"$`, func(on, by string) error {
		var said struct {
			WaitsOn    []int `json:"waits_on"`
			WaitedOnBy []int `json:"waited_on_by"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &said); err != nil {
			return err
		}
		if said.WaitsOn == nil || said.WaitedOnBy == nil || fmt.Sprint(said.WaitsOn) != "["+on+"]" || fmt.Sprint(said.WaitedOnBy) != "["+by+"]" {
			return fmt.Errorf("the JSON says it waits on %v and is waited on by %v:\n%s", said.WaitsOn, said.WaitedOnBy, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the JSON objective sits under nothing and holds "([^"]*)"$`, func(holds string) error {
		parent, children, err := w.jsonFamily()
		if err != nil {
			return err
		}
		if parent != nil || fmt.Sprint(children) != "["+holds+"]" {
			return fmt.Errorf("the JSON says parent %v, children %v:\n%s", parent, children, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the JSON objective sits under (\d+) and holds nothing$`, func(under int) error {
		parent, children, err := w.jsonFamily()
		if err != nil {
			return err
		}
		if parent == nil || *parent != under || children == nil || len(children) != 0 {
			return fmt.Errorf("the JSON says parent %v, children %v:\n%s", parent, children, w.Out.Stdout)
		}
		return nil
	})
}
