package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var (
	builtOnce sync.Once
	builtAt   string
	builtFail error
	goHome    = os.Getenv("HOME")
)

// WithGoCaches runs a command that builds Go in the HOME the test run began with, so a package that swapped its HOME still finds Go's caches warm.
func WithGoCaches(run *exec.Cmd) *exec.Cmd {
	if run.Env == nil {
		run.Env = os.Environ()
	}
	run.Env = append(run.Env, "HOME="+goHome)
	return run
}

// Built answers the path of a cairn binary built from this tree, once for the whole test run, with the Go caches of the HOME the run began with.
func Built(t *testing.T) string {
	t.Helper()
	builtOnce.Do(func() {
		dir, err := os.MkdirTemp("", "cairn-built-")
		if err != nil {
			builtFail = err
			return
		}
		builtAt = filepath.Join(dir, "cairn")
		build := WithGoCaches(exec.Command("go", "build", "-o", builtAt, "cairn/cmd/cairn"))
		if out, err := build.CombinedOutput(); err != nil {
			builtFail = fmt.Errorf("building cairn: %w\n%s", err, out)
		}
	})
	if builtFail != nil {
		t.Fatal(builtFail)
	}
	return builtAt
}
