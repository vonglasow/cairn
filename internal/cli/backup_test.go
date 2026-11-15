package cli

import (
	"testing"
)

func TestMCPServe_startsWhereTheStoreCannotBeLocated(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("CAIRN_STORE", "")
	if got := run(t, "mcp", "serve").Code; got != exitOK {
		t.Errorf("mcp serve: exit %d, want %d: the store is located at each call", got, exitOK)
	}
}
