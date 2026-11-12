package cli

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/spf13/cobra"

	"cairn/internal/build"
	"cairn/internal/config"
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

func wantErr(t *testing.T, args ...string) {
	t.Helper()
	if r := run(t, args...); r.Code == exitOK {
		t.Errorf("run %v should have failed, got: %s", args, r.Stdout)
	}
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

func TestMain_rendersErrorsTheWayTheBinaryDoes(t *testing.T) {
	if got := run(t, "nope").Stderr; !strings.Contains(got, "ERROR") {
		t.Errorf("stderr = %q, want fang's rendering of the error", got)
	}
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

func TestEveryThingHasAGroupSentence(t *testing.T) {
	for _, op := range engine.All() {
		if groupShort[op.Thing] == "" {
			t.Errorf("%q has no sentence in groupShort", op.Thing)
		}
	}
}

func TestTheAppGroupSaysWhereAnApplicationIsInstalled(t *testing.T) {
	r := run(t, "--help")
	if !strings.Contains(r.Stdout, "on a Mac or a Linux desktop") {
		t.Errorf("cairn --help says %q", r.Stdout)
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

func TestNoArgumentSpeaksOfTheCommandLine(t *testing.T) {
	for _, op := range engine.All() {
		for _, arg := range op.Takes {
			if strings.Contains(arg.Means, "command line") {
				t.Errorf("%s %q: %q", op.Name(), arg.Name, arg.Means)
			}
		}
	}
}

func TestLeadingMinus_readsNothingOutOfASentence(t *testing.T) {
	if typed, ok := leadingMinus(errors.New("unknown shorthand flag: '3' in -30m")); ok {
		t.Errorf("leadingMinus read %q out of a sentence, want only the parser's own error read", typed)
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

func TestMain_answersAJSONFlagItCannotReadAsJSON(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	r := run(t, "task", "list", "--json=maybe")
	if r.Code == exitOK || !strings.HasPrefix(strings.TrimSpace(r.Stdout), `{"error":`) {
		t.Errorf("stdout %q, stderr %q: want the refusal as JSON", r.Stdout, r.Stderr)
	}
}

func TestMain_readsJSONGivenAsAValueAsTheValue(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	r := run(t, "task", "edit", "1", "--title", "--json")
	if strings.Contains(r.Stdout, `{"error":`) || r.Stderr == "" {
		t.Errorf("stdout %q, stderr %q: want the refusal said to a person", r.Stdout, r.Stderr)
	}
}

func TestHelp_saysATextCanComeFromStandardInput(t *testing.T) {
	if got := mustRun(t, "task", "edit", "--help").Stdout; !strings.Contains(got, "standard input") {
		t.Errorf("task edit --help does not say a text can come from standard input:\n%s", got)
	}
}

func TestMain_quotesWhatWasTypedAsItWasTyped(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	for _, args := range [][]string{
		{"task", "edit", "abc"},
		{"task", "list", "--state", "bogus"},
	} {
		r := run(t, args...)
		typed := args[2]
		if args[2] == "--state" {
			typed = args[3]
		}
		if !strings.Contains(r.Stderr, `"`+typed+`"`) {
			t.Errorf("%v: stderr %q, want %q quoted as typed", args, r.Stderr, typed)
		}
	}
}

func TestMain_neverHandsAnEscapeToTheTerminal(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	r := run(t, "task", "list", "--state", "x\x1b]52;c;aGk=\x07y")
	if strings.ContainsAny(r.Stderr, "\x1b\x07") {
		t.Errorf("stderr %q hands an escape to the terminal", r.Stderr)
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

func homeWithStore(t *testing.T, store string) engine.Home {
	t.Helper()
	return engine.Home{Dir: t.TempDir(), Lookup: func(key string) (string, bool) {
		if key == config.StoreVariable {
			return store, true
		}
		return "", false
	}}
}

func TestMainIn_runsInTheHomeItIsGiven(t *testing.T) {
	mine, theirs := storeFor(t), storeFor(t)
	t.Setenv(config.StoreVariable, theirs)
	var out, errs bytes.Buffer
	if code := MainIn(context.Background(), homeWithStore(t, mine), []string{"task", "new", "Write the spec"}, strings.NewReader(""), &out, &errs); code != exitOK {
		t.Fatalf("task new: exit %d (%s)", code, errs.String())
	}
	if listed := mustRun(t, "task", "list", "--store", mine).Stdout; !strings.Contains(listed, "Write the spec") {
		t.Errorf("the store the home names holds nothing:\n%s", listed)
	}
	if listed := mustRun(t, "task", "list").Stdout; strings.Contains(listed, "Write the spec") {
		t.Errorf("the task went to the store the process names:\n%s", listed)
	}
}

func TestHomeOf_isTheHomeMainWasHanded(t *testing.T) {
	h := homeWithStore(t, storeFor(t))
	cmd := &cobra.Command{}
	cmd.Flags().String(workspaceFlag, "", "")
	cmd.Flags().String(storeFlag, "", "")
	if err := cmd.Flags().Set(workspaceFlag, "work"); err != nil {
		t.Fatal(err)
	}
	cmd.SetContext(withHome(context.Background(), h))
	got := homeOf(cmd)
	if got.Dir != h.Dir || got.Workspace != "work" {
		t.Errorf("homeOf = %q in workspace %q, want %q in workspace work", got.Dir, got.Workspace, h.Dir)
	}
	if store, _ := got.Lookup(config.StoreVariable); store == "" {
		t.Error("homeOf reads the process's environment, not the home's")
	}
}

func TestMainIn_saysTheVersionItsRunWasHanded(t *testing.T) {
	ctx := withStamps(context.Background(), func() build.Info { return build.Stamped("9.9.9", "", "", "") })
	var out, errs bytes.Buffer
	if code := MainIn(ctx, home(), []string{"--version"}, strings.NewReader(""), &out, &errs); code != exitOK {
		t.Fatalf("--version exited %d: %s", code, errs.String())
	}
	if !strings.Contains(out.String(), "9.9.9") {
		t.Errorf("--version says %q, not the version its run was handed", out.String())
	}
}
