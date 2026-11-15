package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/test"
)

func run(t *testing.T, args ...string) test.Result {
	t.Helper()
	return runIn(t, home(), args...)
}

func runIn(_ *testing.T, h engine.Home, args ...string) test.Result {
	return runWith(context.Background(), h, args...)
}

func runWith(ctx context.Context, h engine.Home, args ...string) test.Result {
	var out, err bytes.Buffer
	code := MainIn(ctx, h, args, strings.NewReader(""), &out, &err)
	return test.Result{Stdout: out.String(), Stderr: err.String(), Code: code}
}

func mustRun(t *testing.T, args ...string) test.Result {
	t.Helper()
	r := run(t, args...)
	if r.Code != exitOK {
		t.Fatalf("run %v: exit %d (stderr: %s)", args, r.Code, r.Stderr)
	}
	return r
}

func runAcceptance(t *testing.T, feature string, init func(*godog.ScenarioContext), tag string) {
	t.Helper()
	t.Parallel()
	ranTo(t, test.Suite(t, feature, init, tag), tag)
}

func runOneAtATime(t *testing.T, feature string, init func(*godog.ScenarioContext), tag string) {
	t.Helper()
	suite := test.Suite(t, feature, init, tag)
	suite.Options.Concurrency = 1
	ranTo(t, suite, tag)
}

func ranTo(t *testing.T, suite godog.TestSuite, tag string) {
	t.Helper()
	if suite.Run() != 0 {
		t.Fatalf("acceptance %s failed", tag)
	}
}

func storeFor(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "cairn.db")
}

func TestMain_answersARefusalAsJSONHoweverTheFlagWasSpelled(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	for _, flag := range []string{"--json", "--json=true", "--json=1"} {
		r := run(t, "task", "show", "1", flag)
		if r.Code == exitOK || !strings.HasPrefix(strings.TrimSpace(r.Stdout), `{"error":`) {
			t.Errorf("%s: stdout %q, want the refusal as JSON", flag, r.Stdout)
		}
	}
	r := run(t, "task", "show", "1", "--json=false")
	if strings.Contains(r.Stdout, `{"error":`) {
		t.Errorf("--json=false answered as JSON: %q", r.Stdout)
	}
}

func TestMain_saysToWriteALengthUnderZeroAfterTheDoubleDash(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	r := run(t, "timesheet", "add", "1", "-30m")
	if !strings.Contains(r.Stderr, "after --") || strings.Contains(r.Stderr, "task number") {
		t.Errorf("stderr %q, want it to say to write -30m after --", r.Stderr)
	}
	if r := run(t, "task", "show", "-1"); !strings.Contains(r.Stderr, "a task number is expected") {
		t.Errorf("stderr %q, want a negative number still read as one", r.Stderr)
	}
}

func TestMain_readsATextNamedAsADashFromStandardInput(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	var out, errs bytes.Buffer
	stdin := strings.NewReader("first paragraph\n\nsecond paragraph\n")
	if code := Main(context.Background(), []string{"task", "new", "Write the spec", "--desc", "-"}, stdin, &out, &errs); code != exitOK {
		t.Fatalf("exit %d: %s", code, errs.String())
	}
	shown := mustRun(t, "task", "show", "1").Stdout
	if !strings.Contains(shown, "first paragraph") || !strings.Contains(shown, "second paragraph") {
		t.Errorf("task show = %q, want the description read from standard input", shown)
	}
}

func TestMain_refusesTwoTextsNamedAsADash(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	r := run(t, "task", "new", "-", "--desc", "-")
	if r.Code != exitUsage {
		t.Errorf("exit %d, want %d: only one text can come from standard input", r.Code, exitUsage)
	}
}

func TestAskedFor_readsAFlagAnywhereBeforeTheDoubleDash(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want bool
	}{
		{[]string{"task", "list", "--debug"}, true},
		{[]string{"--debug", "task", "list"}, true},
		{[]string{"task", "list", "--debug=true"}, true},
		{[]string{"task", "list", "--debug=false"}, false},
		{[]string{"task", "new", "--", "--debug"}, false},
		{[]string{"task", "list"}, false},
	} {
		if got := askedFor(debugFlag, tc.args); got != tc.want {
			t.Errorf("askedFor(debug, %v) = %v, want %v", tc.args, got, tc.want)
		}
	}
}
