package cli

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

func initPages(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerConfig(sc)
		sc.Step(`^a configuration that turns the page "([^"]*)" off$`, func(page string) error {
			return w.WriteConfig(fmt.Sprintf("[pages]\n%s = false\n", page))
		})
		sc.Step(`^a configuration that turns the pages "([^"]*)" and "([^"]*)" off$`, func(one, other string) error {
			return w.WriteConfig(fmt.Sprintf("[pages]\n%s = false\n%s = false\n", one, other))
		})
		sc.Step(`^a configuration that turns every page off$`, func() error {
			return w.WriteConfig("[pages]\ntasks = false\nobjectives = false\ntime = false\n")
		})
		offers := func(command string) bool {
			return regexp.MustCompile(`(?m)^\s+` + regexp.QuoteMeta(command) + `\s`).MatchString(w.Out.Stdout)
		}
		sc.Step(`^the help offers "([^"]*)"$`, func(command string) error {
			if !offers(command) {
				return fmt.Errorf("the help does not offer %q:\n%s", command, w.Out.Stdout)
			}
			return nil
		})
		sc.Step(`^the help does not offer "([^"]*)"$`, func(command string) error {
			if offers(command) {
				return fmt.Errorf("the help offers %q:\n%s", command, w.Out.Stdout)
			}
			return nil
		})
		sc.Step(`^it says the pages turned off are "([^"]*)"$`, func(pages string) error {
			off := strings.Split(pages, ", ")
			for _, page := range []string{"tasks", "objectives", "time"} {
				want := fmt.Sprintf("\n%s = %t\n", page, !slices.Contains(off, page))
				_, under, _ := strings.Cut(w.Out.Stdout, "\n[pages]")
				under, _, _ = strings.Cut(under+"\n[", "\n[")
				if !strings.Contains(under+"\n", want) {
					return fmt.Errorf("it does not say the pages %q are off:\n%s", pages, w.Out.Stdout)
				}
			}
			return nil
		})
	}
}

func TestAcceptance_Pages(t *testing.T) {
	runAcceptance(t, "pages.feature", initPages(t), "")
}
