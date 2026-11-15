package engine

import (
	"os"
	"path/filepath"
	"testing"

	"cairn/internal/engine/workspace"
	"cairn/internal/store/files"
)

func TestKeep_whatCannotBeSettledIsFoundAtEachCall(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "h")
	if kept := Keep(h); kept.Kept != nil {
		t.Errorf("kept %+v from a default whose folder is not there", kept.Kept)
	}
	if kept := Keep(Home{Missing: errStore}); kept.Kept != nil {
		t.Errorf("kept %+v with no home", kept.Kept)
	}
}

func TestWorkspaceShow_refusesSettingsItCannotRead(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	if err := os.WriteFile(filepath.Join(h.Dir, "w", files.SettingsFile), []byte("not toml ["), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := WorkspaceShow.Run(workedOn{Home: h}, Input{}); err == nil {
		t.Error("workspace show answered over settings that are not TOML")
	}
}

func TestWorkspaceShow_aServerSaysSwitchingIsNotDoneThere(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	h.Workspace = "home"
	answer, err := WorkspaceShow.Run(workedOn{Home: Keep(h)}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	said := Of(WorkspaceShow, answer).(workspace.Shown)
	if said.Workspace != "home" || !said.Why.Started || said.Why.Rung != "--workspace" || said.Switch == "" {
		t.Errorf("it answers %+v", said)
	}
}
