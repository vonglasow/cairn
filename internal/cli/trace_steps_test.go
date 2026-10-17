package cli

import (
	"encoding/json"
	"fmt"
	"os/user"
	"regexp"
	"strings"

	"github.com/cucumber/godog"

	"cairn/test"
)

var anAct = regexp.MustCompile(`^\s*\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}`)

func actsIn(shown string) []string {
	acts := []string{}
	for _, line := range strings.Split(shown, "\n") {
		if anAct.MatchString(line) {
			acts = append(acts, line)
		}
	}
	return acts
}

var aTaskNumber = regexp.MustCompile(`#(\d+)`)

func namesOnly(shown string) string {
	kept := []string{}
	for _, line := range strings.Split(shown, "\n") {
		if anAct.MatchString(line) {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

func (w *world) registerTrace(sc *godog.ScenarioContext) {
	sc.Step(`^the author is "([^"]*)"$`, func(name string) error {
		w.Setenv("CAIRN_AUTHOR", name)
		return nil
	})
	everyActSays := func(said string) error {
		acts := 0
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			if line == "" || strings.HasPrefix(line, " ") {
				continue
			}
			acts++
			if !strings.Contains(line, said) {
				return fmt.Errorf("the act %q does not say %q", line, said)
			}
		}
		if acts == 0 {
			return fmt.Errorf("no act is written:\n%s", w.Out.Stdout)
		}
		return nil
	}
	sc.Step(`^every act says it was done by "([^"]*)" through "([^"]*)"$`, func(by, through string) error {
		return everyActSays(fmt.Sprintf("(by %s, through %s)", by, through))
	})
	sc.Step(`^every act says it was done by the one logged in, through "([^"]*)"$`, func(through string) error {
		me, err := user.Current()
		if err != nil {
			return err
		}
		return everyActSays(fmt.Sprintf("(by %s, through %s)", me.Username, through))
	})
	sc.Step(`^every act carries the author "([^"]*)" and the surface "([^"]*)"$`, func(by, through string) error {
		var said struct {
			Acts []struct {
				By      string `json:"by"`
				Through string `json:"through"`
			} `json:"acts"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &said); err != nil {
			return err
		}
		if len(said.Acts) == 0 {
			return fmt.Errorf("no act is carried:\n%s", w.Out.Stdout)
		}
		for _, act := range said.Acts {
			if act.By != by || act.Through != through {
				return fmt.Errorf("an act carries the author %q and the surface %q", act.By, act.Through)
			}
		}
		return nil
	})
	sc.Step(`^I looked at task (\d+)$`, func(n int) error {
		return w.Seed(fmt.Sprintf(`cairn task show %d`, n))
	})

	sc.Step(`^I listed the tasks$`, func() error {
		return w.Seed(`cairn task list`)
	})

	sc.Step(`^I looked for "([^"]*)"$`, func(what string) error {
		return w.Seed(fmt.Sprintf(`cairn task search %s`, what))
	})

	sc.Step(`^I let the focus go$`, func() error {
		return w.Seed(`cairn task unfocus`)
	})

	sc.Step(`^I let task (\d+) go$`, func(n int) error {
		return w.Seed(fmt.Sprintf(`cairn task unfocus %d`, n))
	})

	sc.Step(`^I moved task (\d+) to the (first|last) place$`, func(n int, where string) error {
		return w.Seed(fmt.Sprintf(`cairn task move %d --%s`, n, where))
	})

	sc.Step(`^it says the move went to the (first|last) place$`, func(where string) error {
		return test.WantContains(w.Out.Stdout, where, "the move")
	})

	sc.Step(`^it lists the acts of (\d+) tasks?$`, func(want int) error {
		named := map[string]bool{}
		for _, one := range aTaskNumber.FindAllStringSubmatch(namesOnly(w.Out.Stdout), -1) {
			named[one[1]] = true
		}
		if len(named) != want {
			return fmt.Errorf("%d tasks named %v, want %d:\n%s",
				len(named), named, want, w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^it says nothing was done$`, func() error {
		return test.WantContains(w.Out.Stdout, "nothing was done", "the answer")
	})

	sc.Step(`^it says the task was created$`, func() error {
		return test.WantContains(w.Out.Stdout, "created", "the history")
	})

	sc.Step(`^it says the deletion took "([^"]*)" with it$`, func(said string) error {
		return test.WantContains(w.Out.Stdout, said, "the history")
	})

	sc.Step(`^it says it started in the (\w+) at (\w+) priority$`,
		func(column, priority string) error {
			for _, said := range []string{column, priority} {
				if err := test.WantContains(w.Out.Stdout, said, "the creation"); err != nil {
					return err
				}
			}
			return nil
		})

	sc.Step(`^it says when$`, func() error {
		if len(actsIn(w.Out.Stdout)) == 0 {
			return fmt.Errorf("no act carries a moment:\n%s", w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^it says the (\w+) went from "([^"]*)" to "([^"]*)"$`,
		func(field, was, became string) error {
			for _, said := range []string{field, was, became} {
				if err := test.WantContains(w.Out.Stdout, said, "the change"); err != nil {
					return err
				}
			}
			return nil
		})

	sc.Step(`^it carries (\d+) acts?$`, func(want int) error {
		if got := actsIn(w.Out.Stdout); len(got) != want {
			return fmt.Errorf("%d acts, want %d:\n%s", len(got), want, w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^it lists the acts in the order (.+)$`, func(list string) error {
		acts := actsIn(w.Out.Stdout)
		named := titlesNamed(list)
		if len(acts) != len(named) {
			return fmt.Errorf("%d acts, want the %d named %v:\n%s",
				len(acts), len(named), named, w.Out.Stdout)
		}
		for i, name := range named {
			if !strings.Contains(acts[i], name) {
				return fmt.Errorf("act %d is %q, want %q", i+1, strings.TrimSpace(acts[i]), name)
			}
		}
		return nil
	})
}
