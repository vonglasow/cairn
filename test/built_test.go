package test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestBuilt_answersACairnThatRuns(t *testing.T) {
	out, err := exec.Command(Built(t), "--version").Output()
	if err != nil {
		t.Fatalf("running the built cairn: %v", err)
	}
	if !strings.HasPrefix(string(out), "cairn ") {
		t.Errorf("--version said %q, want it to start with %q", out, "cairn ")
	}
}

func TestBuilt_buildsOnlyOnce(t *testing.T) {
	if first, second := Built(t), Built(t); first != second {
		t.Errorf("built at %q then %q, want one binary for the run", first, second)
	}
}
