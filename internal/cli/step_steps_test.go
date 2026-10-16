package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/task"
	"cairn/test"
)

type stepRead struct {
	Place int    `json:"place"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func (w *world) stepsRead() ([]stepRead, error) {
	var said struct {
		Steps []stepRead `json:"steps"`
	}
	if err := json.Unmarshal([]byte(w.Out.Stdout), &said); err != nil {
		return nil, fmt.Errorf("reading the steps: %w\n%s", err, w.Out.Stdout)
	}
	return said.Steps, nil
}

func (w *world) stepsOfTask(n int) ([]stepRead, error) {
	w.Run(fmt.Sprintf("cairn step list %d --json", n))
	return w.stepsRead()
}

func (w *world) namesTheStep(title, how string) error {
	for _, line := range strings.Split(w.Out.Stdout, "\n") {
		if !strings.HasSuffix(strings.TrimSpace(line), "  "+title) {
			continue
		}
		if got := strings.Contains(line, task.DoneMark); got != (how == "done") {
			return fmt.Errorf("the step %q reads %q, and it is %s", title, line, how)
		}
		return nil
	}
	return fmt.Errorf("no step %q is named:\n%s", title, w.Out.Stdout)
}

func (w *world) namesStepAt(place int, title, how string) error {
	for _, line := range strings.Split(w.Out.Stdout, "\n") {
		if !strings.HasSuffix(strings.TrimSpace(line), fmt.Sprintf("%d  %s", place, title)) {
			continue
		}
		if got := strings.Contains(line, task.DoneMark); got != (how == "done") {
			return fmt.Errorf("step %d reads %q, and it is %s", place, line, how)
		}
		return nil
	}
	return fmt.Errorf("no step %d is named %q:\n%s", place, title, w.Out.Stdout)
}

func (w *world) doesNotMention(what string) error {
	if strings.Contains(w.Out.Stdout, what) {
		return fmt.Errorf("it mentions %q:\n%s", what, w.Out.Stdout)
	}
	return nil
}

func (w *world) aStoreBeforeSteps() error {
	before, err := os.ReadFile(filepath.Join("testdata", "sub-tasks-before-steps.db"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(w.StorePath), 0o700); err != nil {
		return err
	}
	return os.WriteFile(w.StorePath, before, 0o600)
}

func (w *world) registerStep(sc *godog.ScenarioContext) {
	sc.Step(`^I added the step "([^"]*)" to task (\d+)(, done)?$`, func(title string, n int, done string) error {
		line := fmt.Sprintf("cairn step new %d %q", n, title)
		if done != "" {
			line += " --done"
		}
		return w.Seed(line)
	})
	sc.Step(`^I ticked step (\d+) of task (\d+)$`, func(place, n int) error {
		return w.Seed(fmt.Sprintf("cairn step edit %d %d --done", n, place))
	})
	sc.Step(`^it names the step "([^"]*)" as (done|not done)$`, w.namesTheStep)
	sc.Step("^`([^`]*)` names the step \"([^\"]*)\" as (done|not done)$", func(line, title, how string) error {
		w.Run(line)
		return w.namesTheStep(title, how)
	})
	sc.Step("^`([^`]*)` names step (\\d+) \"([^\"]*)\" as (done|not done)$", func(line string, place int, title, how string) error {
		w.Run(line)
		return w.namesStepAt(place, title, how)
	})
	sc.Step(`^it does not mention "([^"]*)"$`, w.doesNotMention)
	sc.Step("^`([^`]*)` does not mention \"([^\"]*)\"$", func(line, what string) error {
		w.Run(line)
		return w.doesNotMention(what)
	})
	sc.Step(`^it carries the step "([^"]*)", (done|not done)$`, func(title, how string) error {
		steps, err := w.stepsRead()
		if err != nil {
			return err
		}
		for _, s := range steps {
			if s.Title == title {
				if s.Done != (how == "done") {
					return fmt.Errorf("the step %q is carried as done: %v, and it is %s", title, s.Done, how)
				}
				return nil
			}
		}
		return fmt.Errorf("no step %q is carried: %+v", title, steps)
	})
	sc.Step(`^every step of task (\d+) is done$`, func(n int) error {
		steps, err := w.stepsOfTask(n)
		if err != nil {
			return err
		}
		if len(steps) == 0 || slices.ContainsFunc(steps, func(s stepRead) bool { return !s.Done }) {
			return fmt.Errorf("the steps of task %d are %+v", n, steps)
		}
		return nil
	})
	sc.Step(`^the steps of task (\d+) read (.+)$`, func(n int, list string) error {
		steps, err := w.stepsOfTask(n)
		if err != nil {
			return err
		}
		var titles []string
		for _, s := range steps {
			titles = append(titles, s.Title)
		}
		if want := titlesNamed(list); !slices.Equal(titles, want) {
			return fmt.Errorf("the steps of task %d read %v, want %v", n, titles, want)
		}
		return nil
	})
	sc.Step(`^it says step (\d+) is in the way$`, func(place int) error {
		return test.WantContains(w.Out.Stderr, fmt.Sprintf("step %d", place), "the refusal")
	})
	for step, want := range map[string]string{
		`^it says a step needs a title$`:                 "a step needs a title",
		`^it says a step is done or not done, not both$`: "done or not done, not both",
		`^it says the task has no such step$`:            "no such step",
	} {
		wanted := want
		sc.Step(step, func() error { return test.WantContains(w.Out.Stderr, wanted, "the reason") })
	}
	sc.Step(`^it shows "([^"]*)"$`, func(what string) error {
		return test.WantContains(w.Out.Stdout, what, "what it shows")
	})
	sc.Step(`^a store from before tasks were split into steps$`, w.aStoreBeforeSteps)
}
