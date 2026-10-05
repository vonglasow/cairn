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

func initUsage(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		// The shared world runs `cairn` and says whether it succeeded; only
		// what usage itself claims is written here.
		w := test.NewWorld(t, run)
		w.Register(sc)

		sc.Step(`^it shows how to use it$`, func() error {
			if !strings.Contains(w.Out.Stdout, "USAGE") {
				return fmt.Errorf("stdout = %q, want it to show how cairn is used", w.Out.Stdout)
			}
			return nil
		})

		sc.Step(`^it writes nothing to standard error$`, func() error {
			if w.Out.Stderr != "" {
				return fmt.Errorf("stderr = %q, want nothing: this is a result, not a reason", w.Out.Stderr)
			}
			return nil
		})
	}
}

func TestAcceptance_Usage_BareCommandSaysWhatItCanDo(t *testing.T) {
	runAcceptance(t, "usage.feature", initUsage(t), "@usage-bare")
}
