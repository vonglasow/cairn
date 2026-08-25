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

	"cairn/internal/build"
	"cairn/test"
)

// versionWorld carries what one scenario needs between its steps, the three
// stamps included: no test is linked, so what a scenario can choose is not what
// the linker wrote but how this surface reads it.
type versionWorld struct {
	t        *testing.T
	out      test.Result
	version  string
	commit   string
	built    string
	restores []func()
}

// newVersionWorld builds the state one scenario works against, and puts the
// surface's way of reading its build back when the scenario ends.
func newVersionWorld(t *testing.T) *versionWorld {
	w := &versionWorld{t: t}
	previous := stamps
	w.restores = append(w.restores, func() { stamps = previous })
	stamps = w.reading
	return w
}

// reading is what the surface answers while this scenario runs.
func (w *versionWorld) reading() build.Info {
	return build.Stamped(w.version, w.commit, w.built)
}

// stamp decides what the binary claims to be for the length of the scenario.
func (w *versionWorld) stamp(target *string, value string) { *target = value }

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
	for _, l := range strings.Split(w.out.Stdout, "\n") {
		if rest, ok := strings.CutPrefix(l, key); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

func initVersion(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newVersionWorld(t)

		sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
			w.endScenario()
			return ctx, err
		})

		sc.Step(`^a binary built with no version stamps$`, func() error {
			w.stamp(&w.version, "")
			w.stamp(&w.commit, "")
			w.stamp(&w.built, "")
			return nil
		})

		sc.Step(`^a binary stamped with version "([^"]*)"$`, func(v string) error {
			w.stamp(&w.version, v)
			return nil
		})

		sc.Step(`^stamped with commit "([^"]*)"$`, func(v string) error {
			w.stamp(&w.commit, v)
			return nil
		})

		sc.Step(`^stamped with build time "([^"]*)"$`, func(v string) error {
			w.stamp(&w.built, v)
			return nil
		})

		sc.Step("^I run `cairn --version`$", func() error {
			w.out = run(w.t, "--version")
			if w.out.Code != exitOK {
				return fmt.Errorf("exit %d (stderr: %s)", w.out.Code, w.out.Stderr)
			}
			return nil
		})

		// A step returns its error; it never calls t.Fatalf. A Goexit from
		// inside godog wrecks its subtest bookkeeping and surfaces as an
		// unreadable "test executed panic(nil) or runtime.Goexit".
		sc.Step(`^it reports the version as "([^"]*)"$`, func(want string) error {
			first := strings.SplitN(w.out.Stdout, "\n", 2)[0]
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

// How the surface reads its build is a package-level variable, so a scenario
// that replaces it has to put it back before the next scenario starts — not at
// the end of the Go test that happened to run it. Every scenario in
// version.feature currently sets every stamp it reads, which is what keeps the
// leak invisible; the first Scenario Outline whose examples don't will start on
// the previous one's leftovers. This test watches the restoring itself, since no
// scenario can.
func TestVersionWorld_putsTheBuildBackWhenTheScenarioEnds(t *testing.T) {
	found := stamps
	t.Cleanup(func() { stamps = found })

	w := newVersionWorld(t)
	w.stamp(&w.version, "what the scenario set")
	if got := stamps().Version; got != "what the scenario set" {
		t.Fatalf("version = %q during the scenario", got)
	}

	w.endScenario()
	if got := stamps().Version; got != found().Version {
		t.Errorf("version = %q, want %q put back for the next scenario", got, found().Version)
	}
}

func TestAcceptance_Version_DegradesCleanly(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-degraded")
}

func TestAcceptance_Version_ReportsItsStamps(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-stamped")
}
