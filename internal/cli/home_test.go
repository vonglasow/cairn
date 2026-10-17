package cli

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "cairn-cli-home-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, name := range []string{"HOME", "USERPROFILE"} {
		_ = os.Setenv(name, home)
	}
	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}
