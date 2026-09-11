package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The acceptance suite calls cli.Main in process, which is everything the
// binary does except the one line below main(): handing the returned code to
// os.Exit. That line cannot be reached in process, so it is checked the only
// way it can be — by running the real program and reading what the shell gets.
//
// This test contributes nothing to coverage or mutation: the code runs in a
// separate process, which Go does not instrument. It is here for the one thing
// nothing else can prove.
func TestBinary_handsItsExitCodeToTheShell(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "cairn")
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("building the binary: %v\n%s", err, out)
	}

	// The subprocess inherits os.Environ() below, which still carries this
	// machine's real HOME. Left alone, the binary it starts would read
	// <real HOME>/.config/cairn/config.toml — a file this test has no
	// business touching, and whose shape it does not control. home is a
	// fresh directory this run owns, spliced into the environment the
	// subprocess actually sees.
	home := t.TempDir()
	env := []string{"HOME=" + home}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "HOME=") {
			env = append(env, kv)
		}
	}

	for _, tc := range []struct {
		name string
		args []string
		want int
	}{
		{"a command that worked", []string{"--version"}, 0},
		{"a command it did not understand", []string{"nope"}, 2},
		{"a request it understood and refused", []string{"task", "add", ""}, 1},
	} {
		// Not `go run`: it collapses every non-zero code to 1, which is precisely
		// the thing this test exists to check.
		//
		// semgrep flags a non-literal program name here, and it is right to look:
		// the rule asks whether unverified user data can reach the call. It
		// cannot. `binary` is the path this test just built under t.TempDir(),
		// and the arguments are the literals in the table above — nothing here
		// comes from outside the test.
		// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
		cmd := exec.Command(binary, tc.args...)
		cmd.Env = append(env, "CAIRN_STORE="+filepath.Join(t.TempDir(), "cairn.db"))
		_ = cmd.Run()
		if got := cmd.ProcessState.ExitCode(); got != tc.want {
			t.Errorf("%s: exit code = %d, want %d", tc.name, got, tc.want)
		}
	}
}
