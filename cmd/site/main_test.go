package main

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBinary_handsItsExitCodeToTheShell(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "site")
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("building the site's command: %v\n%s", err, out)
	}
	for _, tc := range []struct {
		name string
		args []string
		want int
	}{
		{"told nowhere to write", nil, 2},
		{"a repository with no pages", []string{"--out", t.TempDir(), "--repo", t.TempDir()}, 1},
	} {
		cmd := exec.Command(binary, tc.args...)
		_ = cmd.Run()
		if got := cmd.ProcessState.ExitCode(); got != tc.want {
			t.Errorf("%s: exit code = %d, want %d", tc.name, got, tc.want)
		}
	}
}
