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
	"os"
	"os/exec"
	"path/filepath"
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
	t   *testing.T
	out test.Result
	// archive is where a source archive was made, and unpacked where it was
	// opened; repo is a repository a release is tried in.
	archive, unpacked, repo string
	version                 string
	commit                  string
	built                   string
	committed               string
	restores                []func()
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
	return build.Stamped(w.version, w.commit, w.built, w.committed)
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

// stampedThere asks make, where the archive was unpacked, what a build there
// would be stamped with, and compares one of the three words it answers.
func (w *versionWorld) stampedThere(at int, want string) error {
	said, err := exec.Command("make", "-s", "-C", w.unpacked, "stamps").Output()
	if err != nil {
		return fmt.Errorf("make stamps: %w", err)
	}
	if got := strings.Fields(string(said)); len(got) != 3 || got[at] != want {
		return fmt.Errorf("make stamps answered %q, want %q in place %d", said, want, at)
	}
	return nil
}

// root is the repository the tests run in.
func root() string { return filepath.Join("..", "..") }

// command runs a program in a directory and says what it wrote when it fails.
func command(dir, name string, args ...string) error {
	run := exec.Command(name, args...)
	run.Dir = dir
	if out, err := run.CombinedOutput(); err != nil {
		return fmt.Errorf("%s %v: %w\n%s", name, args, err, out)
	}
	return nil
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
			w.stamp(&w.committed, "")
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

		sc.Step(`^stamped with commit date "([^"]*)"$`, func(v string) error {
			w.stamp(&w.committed, v)
			return nil
		})

		sc.Step(`^the source archive of this repository made as version "([^"]*)"$`, func(version string) error {
			w.archive = filepath.Join(w.t.TempDir(), "dist")
			return command(root(), "scripts/dist.sh", version, w.archive)
		})

		sc.Step(`^I unpack it somewhere git has never been$`, func() error {
			found, err := filepath.Glob(filepath.Join(w.archive, "cairn-*-src.tar.gz"))
			if err != nil {
				return err
			}
			if len(found) != 1 {
				return fmt.Errorf("the archive made is %v, want one source archive", found)
			}
			w.unpacked = w.t.TempDir()
			if err := command(w.unpacked, "tar", "-xzf", found[0]); err != nil {
				return err
			}
			inside, err := filepath.Glob(filepath.Join(w.unpacked, "cairn-*"))
			if err != nil {
				return err
			}
			if len(inside) != 1 {
				return fmt.Errorf("the archive unpacks to %v, want one directory", inside)
			}
			w.unpacked = inside[0]
			if _, err := os.Stat(filepath.Join(w.unpacked, ".git")); err == nil {
				return fmt.Errorf("the archive carries a git history")
			}
			return nil
		})

		sc.Step(`^a build there would be stamped with version "([^"]*)"$`, func(want string) error {
			return w.stampedThere(0, want)
		})

		sc.Step(`^a build there would be stamped with the commit the archive was made from$`, func() error {
			head, err := exec.Command("git", "-C", root(), "rev-parse", "HEAD").Output()
			if err != nil {
				return err
			}
			return w.stampedThere(1, strings.TrimSpace(string(head)))
		})

		sc.Step(`^a build there would be stamped with the date of that commit$`, func() error {
			ask := exec.Command("git", "-C", root(), "show", "-s", "--format=%cd",
				"--date=format-local:%Y-%m-%dT%H:%M:%SZ", "HEAD")
			ask.Env = append(os.Environ(), "TZ=UTC")
			when, err := ask.Output()
			if err != nil {
				return err
			}
			return w.stampedThere(2, strings.TrimSpace(string(when)))
		})

		sc.Step(`^the archive says how to build and install it$`, func() error {
			said, err := os.ReadFile(filepath.Join(w.unpacked, "INSTALL.md"))
			if err != nil {
				return err
			}
			for _, must := range []string{"make build", "bin/cairn --version"} {
				if !strings.Contains(string(said), must) {
					return fmt.Errorf("INSTALL.md says nothing of %q", must)
				}
			}
			return nil
		})

		sc.Step(`^a repository on a branch called "([^"]*)"$`, func(branch string) error {
			w.repo = w.t.TempDir()
			if err := os.MkdirAll(filepath.Join(w.repo, "scripts"), 0o755); err != nil {
				return err
			}
			script, err := os.ReadFile(filepath.Join(root(), "scripts", "release.sh"))
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(w.repo, "scripts", "release.sh"), script, 0o755); err != nil {
				return err
			}
			for _, args := range [][]string{
				{"init", "-q", "-b", branch},
				{"-c", "user.name=t", "-c", "user.email=t@t", "add", "."},
				{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "feat: begin"},
			} {
				if err := command(w.repo, "git", args...); err != nil {
					return err
				}
			}
			return nil
		})

		sc.Step(`^a file changed and not committed$`, func() error {
			return os.WriteFile(filepath.Join(w.repo, "notes.txt"), []byte("not committed\n"), 0o644)
		})

		sc.Step(`^I make a release there$`, func() error {
			run := exec.Command("bash", filepath.Join(w.repo, "scripts", "release.sh"))
			run.Dir = w.repo
			out, err := run.CombinedOutput()
			w.out = test.Result{Stderr: string(out), Code: exitOK}
			if err != nil {
				w.out.Code = 1
			}
			return nil
		})

		sc.Step(`^it refuses, saying "([^"]*)"$`, func(want string) error {
			if w.out.Code == exitOK || !strings.Contains(w.out.Stderr, want) {
				return fmt.Errorf("it answered %d, saying %q", w.out.Code, w.out.Stderr)
			}
			return nil
		})

		sc.Step(`^no tag was made$`, func() error {
			tags, err := exec.Command("git", "-C", w.repo, "tag").Output()
			if err != nil {
				return err
			}
			if len(strings.TrimSpace(string(tags))) != 0 {
				return fmt.Errorf("tags were made: %s", tags)
			}
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

		sc.Step(`^it reports the commit date as "([^"]*)"$`, func(want string) error {
			return wantLine(w, "commit date", want)
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

func TestAcceptance_Version_ComesWithASourceArchive(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-from-a-source-archive")
}

func TestAcceptance_Version_IsMadeOnlyFromMainAndClean(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-is-made-on-main,@version-is-made-from-a-clean-tree")
}
