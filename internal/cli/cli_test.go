package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// result is what one run of the command line produced: the two streams kept
// apart, and the code the process would have exited with.
//
// The streams are separate on purpose. A result goes to stdout and a reason
// goes to stderr; merged into one buffer, a message on the wrong stream reads
// exactly like a message on the right one, and no test could ever tell them
// apart.
type result struct {
	stdout string
	stderr string
	code   int
}

// run executes the real command line in-process, through the same Main the
// binary runs, and returns what it produced. Every test in this package goes
// through it, so what a test observes is what a user observes — fang's error
// rendering included, no mocked command layer, no spawned binary.
func run(t *testing.T, args ...string) result {
	t.Helper()
	var out, err bytes.Buffer
	code := Main(context.Background(), args, &out, &err)
	return result{stdout: out.String(), stderr: err.String(), code: code}
}

// mustRun fails the test if the command did not succeed, and returns what it
// produced.
func mustRun(t *testing.T, args ...string) result {
	t.Helper()
	r := run(t, args...)
	if r.code != exitOK {
		t.Fatalf("run %v: exit %d (stderr: %s)", args, r.code, r.stderr)
	}
	return r
}

// wantErr fails the test if the command unexpectedly succeeds. Error paths are
// behaviour too, and they are where coverage quietly rots.
func wantErr(t *testing.T, args ...string) {
	t.Helper()
	if r := run(t, args...); r.code == exitOK {
		t.Errorf("run %v should have failed, got: %s", args, r.stdout)
	}
}

// runAcceptance runs the scenarios carrying tag from the named feature file,
// and fails t on any of them.
//
// Every feature wires itself here rather than copying the suite: one
// definition, one place to fix. A feature adds its own initializer — steps live
// in the package that observes them — and nothing else.
func runAcceptance(t *testing.T, feature string, init func(*godog.ScenarioContext), tag string) {
	t.Helper()
	suite := godog.TestSuite{
		Name:                "cairn",
		ScenarioInitializer: init,
		Options: &godog.Options{
			Format: "pretty",
			// Relative to the package directory: `go test` runs there.
			Paths:          []string{filepath.Join("..", "..", "bdd", feature)},
			Tags:           tag,
			Strict:         true,
			NoColors:       true,
			TestingT:       t,
			DefaultContext: context.Background(),
		},
	}
	if suite.Run() != 0 {
		t.Fatalf("acceptance %s failed", tag)
	}
}

// stampFor sets a linker-injected variable for the length of the test.
func stampFor(t *testing.T, target *string, value string) {
	t.Helper()
	prev := *target
	t.Cleanup(func() { *target = prev })
	*target = value
}

func TestBuild_truncatesALongCommit(t *testing.T) {
	stampFor(t, &commit, "4c5a15f700dcabcdef0123456789")
	if got := Build().Commit; got != "4c5a15f700dc" {
		t.Errorf("Commit = %q, want it truncated to %d characters", got, shortCommitLen)
	}
}

func TestBuild_keepsAShortCommitWhole(t *testing.T) {
	stampFor(t, &commit, "4c5a15f")
	if got := Build().Commit; got != "4c5a15f" {
		t.Errorf("Commit = %q, want it left alone when already short", got)
	}
}

func TestApplyVCS_fillsAnUnstampedCommit(t *testing.T) {
	got := applyVCS(BuildInfo{}, []debug.BuildSetting{
		{Key: "vcs.revision", Value: "4c5a15f700dcabcdef"},
		{Key: "vcs.time", Value: "2026-08-15T13:59:37Z"},
	})
	if got.Commit != "4c5a15f700dcabcdef" {
		t.Errorf("Commit = %q, want the embedded revision", got.Commit)
	}
	if got.CommitDate != "2026-08-15T13:59:37Z" {
		t.Errorf("CommitDate = %q, want the embedded time", got.CommitDate)
	}
}

func TestApplyVCS_leavesAStampedCommitAlone(t *testing.T) {
	got := applyVCS(BuildInfo{Commit: "from-the-linker"}, []debug.BuildSetting{
		{Key: "vcs.revision", Value: "from-the-working-tree"},
	})
	if got.Commit != "from-the-linker" {
		t.Errorf("Commit = %q, want the linker stamp to win", got.Commit)
	}
}

func TestApplyVCS_ignoresUnrelatedSettings(t *testing.T) {
	got := applyVCS(BuildInfo{}, []debug.BuildSetting{{Key: "GOARCH", Value: "arm64"}})
	if got.Commit != "" || got.CommitDate != "" {
		t.Errorf("unrelated settings must not fill anything, got %+v", got)
	}
}

func TestBuild_readsTheRuntime(t *testing.T) {
	b := Build()
	if b.Go != runtime.Version() {
		t.Errorf("Go = %q, want %q", b.Go, runtime.Version())
	}
	if want := runtime.GOOS + "/" + runtime.GOARCH; b.Platform != want {
		t.Errorf("Platform = %q, want %q", b.Platform, want)
	}
}

func TestVersionFlag_printsEveryField(t *testing.T) {
	out := mustRun(t, "--version").stdout
	for _, key := range []string{"commit", "built", "commit date", "go", "platform"} {
		if !strings.Contains(out, key) {
			t.Errorf("version output is missing %q:\n%s", key, out)
		}
	}
}

func TestRoot_rejectsAnUnknownCommand(t *testing.T) {
	wantErr(t, "nope")
}

func TestMain_reportsSuccessAsZero(t *testing.T) {
	if got := run(t, "--version").code; got != exitOK {
		t.Errorf("exit code = %d, want %d for a command that worked", got, exitOK)
	}
}

func TestMain_reportsACommandItDidNotUnderstand(t *testing.T) {
	if got := run(t, "nope").code; got != exitUsage {
		t.Errorf("exit code = %d, want %d for a command it could not make sense of", got, exitUsage)
	}
}

// A reason is written for a person to read, so it goes to stderr and nowhere
// else. Nothing enforces that but a test: cobra and fang each default to a
// stream, and a single misrouted Fprintln is invisible until someone pipes the
// output somewhere.
func TestMain_writesAResultToStandardOutputOnly(t *testing.T) {
	r := mustRun(t, "--version")
	if r.stdout == "" {
		t.Error("stdout is empty, want the result a user asked for")
	}
	if r.stderr != "" {
		t.Errorf("stderr = %q, want nothing: what worked has no reason to explain", r.stderr)
	}
}

func TestMain_writesAReasonToStandardErrorOnly(t *testing.T) {
	r := run(t, "nope")
	if r.stderr == "" {
		t.Error("stderr is empty, want the reason a user needs to read")
	}
	if r.stdout != "" {
		t.Errorf("stdout = %q, want nothing: a refusal is not a result", r.stdout)
	}
}

// The binary renders its errors through fang, not through cobra. A harness
// that called cobra directly would assert a message no user is ever shown.
func TestMain_rendersErrorsTheWayTheBinaryDoes(t *testing.T) {
	if got := run(t, "nope").stderr; !strings.Contains(got, "ERROR") {
		t.Errorf("stderr = %q, want fang's rendering of the error", got)
	}
}
