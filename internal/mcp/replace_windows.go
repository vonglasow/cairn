package mcp

import (
	"errors"
	"os"
	"os/exec"
)

func replaceItself(path string, payload []byte) error {
	successor := exec.Command(path, os.Args[1:]...)
	successor.Stdin, successor.Stdout, successor.Stderr = os.Stdin, os.Stdout, os.Stderr
	successor.Env = append(os.Environ(), carriedVariable+"="+string(payload))
	err := successor.Run()
	var exited *exec.ExitError
	if err != nil && !errors.As(err, &exited) {
		return err
	}
	os.Exit(successor.ProcessState.ExitCode())
	return nil
}
