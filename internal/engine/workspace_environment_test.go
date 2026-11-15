package engine

import (
	"cairn/internal/engine/workspace"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocate_theWorkspaceTheEnvironmentNamesGivesBothPlaces(t *testing.T) {
	h := homeHolding(t, "store = \"~/s/cairn.db\"\n"+twoWorkspaces, map[string]string{workspace.WorkspaceVariable: "home"}, "w", "h")
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "h", "cairn.db") || at.Notes != filepath.Join(h.Dir, "h", "notes") {
		t.Errorf("Locate = %+v, %v; want home's store and notes, before the default and the file", at, err)
	}
}

func TestLocate_anUnknownWorkspaceInTheEnvironmentIsRefusedNamingTheVariable(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{workspace.WorkspaceVariable: "play"}, "w", "h")
	_, err := Locate(h)
	if !errors.Is(err, workspace.ErrNoSuchWorkspace) || !strings.Contains(err.Error(), workspace.WorkspaceVariable) || !strings.Contains(err.Error(), `"play"`) || !strings.Contains(err.Error(), "home, work") {
		t.Errorf("Locate = %v, want CAIRN_WORKSPACE play refused, naming home and work", err)
	}
}
