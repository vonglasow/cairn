package files

import (
	"strings"
	"testing"
)

func TestRead_aKeyNamedLikeTheWorkspacesUnderAnotherTableSaysItIsATable(t *testing.T) {
	read, err := readThis(t, "[pages]\nworkspaces = true\n")
	said := strings.Join(read.Ignored(), "\n")
	if err != nil || !strings.Contains(said, "workspaces is a table of its own, [workspaces]") {
		t.Errorf("Read = %v, ignoring %q; want it saying workspaces is a table", err, said)
	}
}

func TestWithWorkspace_addsTheTableAfterWhatIsThere(t *testing.T) {
	for _, before := range []string{"notes = \"~/vault\"\n", "notes = \"~/vault\""} {
		got, err := WithWorkspace([]byte(before), "play", "/p")
		if want := "notes = \"~/vault\"\n\n[workspaces]\nplay = \"/p\"\n"; err != nil || string(got) != want {
			t.Errorf("WithWorkspace(%q) = %q, %v; want %q", before, got, err, want)
		}
	}
}
