package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBinary_handsItsExitCodeToTheShell(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "cairn")
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("building the binary: %v\n%s", err, out)
	}

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
		{"a request it understood and refused", []string{"task", "new", ""}, 1},
	} {
		cmd := exec.Command(binary, tc.args...)
		cmd.Env = append(env, "CAIRN_STORE="+filepath.Join(t.TempDir(), "cairn.db"))
		_ = cmd.Run()
		if got := cmd.ProcessState.ExitCode(); got != tc.want {
			t.Errorf("%s: exit code = %d, want %d", tc.name, got, tc.want)
		}
	}
}
