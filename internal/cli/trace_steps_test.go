package cli

// Executable acceptance scenarios for bdd/trace.feature. The steps every
// feature shares live in world_test.go; only what belongs to the trace is
// declared here.

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"

	"cairn/test"
)

// anAct is how a line of a history begins: the moment it happened. Recognising
// an act by its moment rather than by its words is what lets a step count them
// without knowing which acts a scenario put there.
var anAct = regexp.MustCompile(`^\s*\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}`)

// actsIn reads the lines of a history that carry an act.
func actsIn(shown string) []string {
	acts := []string{}
	for _, line := range strings.Split(shown, "\n") {
		if anAct.MatchString(line) {
			acts = append(acts, line)
		}
	}
	return acts
}

// aTaskNumber is how a listing names the task an act was on. Matched by its
// hash and not by being a number: a line carries a date, and counting its
// digits among the tasks answers a different question — measured, it answered
// six.
var aTaskNumber = regexp.MustCompile(`#(\d+)`)

// namesOnly keeps the part of a listing that names tasks, dropping what it
// works out at the end: the summary carries numbers of its own, and counting
// those among the tasks would answer a different question.
func namesOnly(shown string) string {
	kept := []string{}
	for _, line := range strings.Split(shown, "\n") {
		if anAct.MatchString(line) {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

// registerTrace declares the steps about what cairn records.
func (w *world) registerTrace(sc *godog.ScenarioContext) {
	// Looking at something, through the real command: a step that only pretended
	// to look would prove nothing about whether looking is recorded.
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

	// The tasks a listing named, counted by their numbers rather than by its
	// lines: what carries the rule is how many tasks appear, not how many acts
	// each of them has.
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

	sc.Step(`^it says there were (\d+) stretches of work on (\d+) tasks?$`,
		func(stretches, tasks int) error {
			return test.WantContains(w.Out.Stdout,
				fmt.Sprintf("%d stretches of work on %d task", stretches, tasks), "the count")
		})

	sc.Step(`^it says the work switched (\d+) times?$`, func(switches int) error {
		return test.WantContains(w.Out.Stdout,
			fmt.Sprintf("switched %d time", switches), "the count")
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

	// Only that a moment is there, never which one: a scenario cannot name the
	// time it is running at.
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

	// The count is what carries the rule: an absence does not show, so the only
	// way to see that reading was not recorded is that nothing was added.
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
