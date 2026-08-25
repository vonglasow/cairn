package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/test"
)

// run executes the real command line in-process, through the same Main the
// binary runs, and returns what it produced. Every test in this package goes
// through it, so what a test observes is what a user observes — fang's error
// rendering included, no mocked command layer, no spawned binary.
//
// An empty reader stands in for standard input: no command exercised here
// reads it.
func run(t *testing.T, args ...string) test.Result {
	t.Helper()
	var out, err bytes.Buffer
	code := Main(context.Background(), args, strings.NewReader(""), &out, &err)
	return test.Result{Stdout: out.String(), Stderr: err.String(), Code: code}
}

// mustRun fails the test if the command did not succeed, and returns what it
// produced.
func mustRun(t *testing.T, args ...string) test.Result {
	t.Helper()
	r := run(t, args...)
	if r.Code != exitOK {
		t.Fatalf("run %v: exit %d (stderr: %s)", args, r.Code, r.Stderr)
	}
	return r
}

// wantErr fails the test if the command unexpectedly succeeds. Error paths are
// behaviour too, and they are where coverage quietly rots.
func wantErr(t *testing.T, args ...string) {
	t.Helper()
	if r := run(t, args...); r.Code == exitOK {
		t.Errorf("run %v should have failed, got: %s", args, r.Stdout)
	}
}

// runAcceptance runs the scenarios carrying tag from the named feature file,
// and fails t on any of them.
//
// The suite itself is described once, in cairn/test — this is only who fails
// when it does not pass, which is each surface's own business.
func runAcceptance(t *testing.T, feature string, init func(*godog.ScenarioContext), tag string) {
	t.Helper()
	if test.Suite(t, feature, init, tag).Run() != 0 {
		t.Fatalf("acceptance %s failed", tag)
	}
}

// storeFor gives a path to a store of this test's own, somewhere nothing else
// will look.
func storeFor(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "cairn.db")
}

func TestVersionFlag_printsEveryField(t *testing.T) {
	out := mustRun(t, "--version").Stdout
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
	if got := run(t, "--version").Code; got != exitOK {
		t.Errorf("exit code = %d, want %d for a command that worked", got, exitOK)
	}
}

func TestMain_reportsACommandItDidNotUnderstand(t *testing.T) {
	if got := run(t, "nope").Code; got != exitUsage {
		t.Errorf("exit code = %d, want %d for a command it could not make sense of", got, exitUsage)
	}
}

// A reason is written for a person to read, so it goes to stderr and nowhere
// else. Nothing enforces that but a test: cobra and fang each default to a
// stream, and a single misrouted Fprintln is invisible until someone pipes the
// output somewhere.
func TestMain_writesAResultToStandardOutputOnly(t *testing.T) {
	r := mustRun(t, "--version")
	if r.Stdout == "" {
		t.Error("stdout is empty, want the result a user asked for")
	}
	if r.Stderr != "" {
		t.Errorf("stderr = %q, want nothing: what worked has no reason to explain", r.Stderr)
	}
}

func TestMain_writesAReasonToStandardErrorOnly(t *testing.T) {
	r := run(t, "nope")
	if r.Stderr == "" {
		t.Error("stderr is empty, want the reason a user needs to read")
	}
	if r.Stdout != "" {
		t.Errorf("stdout = %q, want nothing: a refusal is not a result", r.Stdout)
	}
}

// The binary renders its errors through fang, not through cobra. A harness
// that called cobra directly would assert a message no user is ever shown.
func TestMain_rendersErrorsTheWayTheBinaryDoes(t *testing.T) {
	if got := run(t, "nope").Stderr; !strings.Contains(got, "ERROR") {
		t.Errorf("stderr = %q, want fang's rendering of the error", got)
	}
}
