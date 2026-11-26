package engine

import (
	"os"
	"os/exec"
	"testing"
)

func TestProcessesAlive(t *testing.T) {
	if !(processes{}).Alive(os.Getpid(), "") {
		t.Error("the running test process is said dead")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ended := exec.Command(exe, "-test.run=^$")
	if err := ended.Run(); err != nil {
		t.Fatal(err)
	}
	if (processes{}).Alive(ended.Process.Pid, "") {
		t.Error("an ended process is said alive")
	}
}
