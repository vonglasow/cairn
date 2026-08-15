package cli

// Executable acceptance scenarios for bdd/version.feature. The Gherkin is the
// contract; this file is only the wiring between it and the real CLI.
//
// A new feature adds a file like this one — a world, an initializer, and one Go
// test per scenario selected by tag — and calls runAcceptance from cli_test.go.
// The suite itself is defined once, there.

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// versionWorld carries what one scenario needs between its steps.
type versionWorld struct {
	t        *testing.T
	out      result
	restores []func()
}

// stamp sets a linker-injected variable for the length of the scenario.
func (w *versionWorld) stamp(target *string, value string) {
	prev := *target
	w.restores = append(w.restores, func() { *target = prev })
	*target = value
}

// endScenario puts back every stamp the scenario set, most recent first.
//
// t.Cleanup would be the obvious home for this and is the wrong one: it fires
// when the Go test ends, not when the scenario does, so the next scenario in
// the same run would start on this one's leftovers.
func (w *versionWorld) endScenario() {
	for i := len(w.restores) - 1; i >= 0; i-- {
		w.restores[i]()
	}
	w.restores = nil
}

// line returns the value printed for key, or "" when the key is absent.
func (w *versionWorld) line(key string) string {
	for _, l := range strings.Split(w.out.stdout, "\n") {
		if rest, ok := strings.CutPrefix(l, key); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

func initVersion(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := &versionWorld{t: t}

		sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
			w.endScenario()
			return ctx, err
		})

		sc.Step(`^a binary built with no version stamps$`, func() error {
			w.stamp(&version, "")
			w.stamp(&commit, "")
			w.stamp(&built, "")
			return nil
		})

		sc.Step(`^a binary stamped with version "([^"]*)"$`, func(v string) error {
			w.stamp(&version, v)
			return nil
		})

		sc.Step(`^stamped with commit "([^"]*)"$`, func(v string) error {
			w.stamp(&commit, v)
			return nil
		})

		sc.Step(`^stamped with build time "([^"]*)"$`, func(v string) error {
			w.stamp(&built, v)
			return nil
		})

		sc.Step("^I run `cairn --version`$", func() error {
			w.out = run(w.t, "--version")
			if w.out.code != exitOK {
				return fmt.Errorf("exit %d (stderr: %s)", w.out.code, w.out.stderr)
			}
			return nil
		})

		// A step returns its error; it never calls t.Fatalf. A Goexit from
		// inside godog wrecks its subtest bookkeeping and surfaces as an
		// unreadable "test executed panic(nil) or runtime.Goexit".
		sc.Step(`^it reports the version as "([^"]*)"$`, func(want string) error {
			first := strings.SplitN(w.out.stdout, "\n", 2)[0]
			if got := strings.TrimSpace(first); got != "cairn "+want {
				return fmt.Errorf("first line = %q, want %q", got, "cairn "+want)
			}
			return nil
		})

		sc.Step(`^it reports the commit as "([^"]*)"$`, func(want string) error {
			return wantLine(w, "commit", want)
		})

		sc.Step(`^it reports the build time as "([^"]*)"$`, func(want string) error {
			return wantLine(w, "built", want)
		})

		sc.Step(`^it reports a real Go toolchain$`, func() error {
			if got := w.line("go"); got != runtime.Version() {
				return fmt.Errorf("go = %q, want the running toolchain %q", got, runtime.Version())
			}
			return nil
		})

		sc.Step(`^it reports a real platform$`, func() error {
			want := runtime.GOOS + "/" + runtime.GOARCH
			if got := w.line("platform"); got != want {
				return fmt.Errorf("platform = %q, want %q", got, want)
			}
			return nil
		})
	}
}

func wantLine(w *versionWorld, key, want string) error {
	if got := w.line(key); got != want {
		return fmt.Errorf("%s = %q, want %q", key, got, want)
	}
	return nil
}

// The stamps are package-level variables, so a scenario that sets one has to
// put it back before the next scenario starts — not at the end of the Go test
// that happened to run it. Every scenario in version.feature currently sets
// every stamp it reads, which is what keeps the leak invisible; the first
// Scenario Outline whose examples don't will start on the previous one's
// leftovers. This test watches the restoring itself, since no scenario can.
func TestVersionWorld_putsAStampBackWhenTheScenarioEnds(t *testing.T) {
	stampFor(t, &version, "what the scenario found")
	w := &versionWorld{t: t}
	w.stamp(&version, "what the scenario set")
	w.endScenario()
	if version != "what the scenario found" {
		t.Errorf("version = %q, want it put back for the next scenario", version)
	}
}

func TestAcceptance_Version_DegradesCleanly(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-degraded")
}

func TestAcceptance_Version_ReportsItsStamps(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-stamped")
}
