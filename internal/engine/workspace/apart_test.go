package workspace

import (
	"path/filepath"
	"testing"
)

func TestWorkspace_itsStoreIsTheCairnDbOfItsFolder(t *testing.T) {
	if got := (Workspace{Folder: filepath.Join("a", "home")}).Store(); got != filepath.Join("a", "home", "cairn.db") {
		t.Errorf("Store = %q", got)
	}
}

func TestWorkspaceShown_isNamedShortNeverByAWholePath(t *testing.T) {
	cases := map[string]Shown{
		"home":     {Workspace: "home", Own: "kept", Store: filepath.Join("x", "cairn.db")},
		"fixtures": {Own: "fixtures", Store: filepath.Join("tmp", "x", "cairn.db")},
		"loose":    {Store: filepath.Join("a", "loose", "cairn.db")},
		"":         {},
	}
	for want, shown := range cases {
		if got := shown.Name(); got != want {
			t.Errorf("Name of %+v = %q, want %q", shown, got, want)
		}
	}
}
