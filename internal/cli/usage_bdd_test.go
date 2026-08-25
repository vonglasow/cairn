package cli

// Executable acceptance scenarios for bdd/usage.feature. The Gherkin is the
// contract; this file is only the wiring between it and the real CLI.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/test"
)

// usageWorld carries what one scenario needs between its steps.
type usageWorld struct {
	t   *testing.T
	out test.Result
}

func initUsage(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := &usageWorld{t: t}

		sc.Step("^I run `cairn`$", func() error {
			w.out = run(w.t)
			return nil
		})

		sc.Step(`^it succeeds$`, func() error {
			if w.out.Code != exitOK {
				return fmt.Errorf("exit %d, want %d (stderr: %s)", w.out.Code, exitOK, w.out.Stderr)
			}
			return nil
		})

		sc.Step(`^it shows how to use it$`, func() error {
			if !strings.Contains(w.out.Stdout, "USAGE") {
				return fmt.Errorf("stdout = %q, want it to show how cairn is used", w.out.Stdout)
			}
			return nil
		})

		sc.Step(`^it writes nothing to standard error$`, func() error {
			if w.out.Stderr != "" {
				return fmt.Errorf("stderr = %q, want nothing: this is a result, not a reason", w.out.Stderr)
			}
			return nil
		})
	}
}

func TestAcceptance_Usage_BareCommandSaysWhatItCanDo(t *testing.T) {
	runAcceptance(t, "usage.feature", initUsage(t), "@usage-bare")
}
