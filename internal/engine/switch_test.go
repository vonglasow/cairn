package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cairn/internal/config"
)

func heldOn(t *testing.T, h Home) *Engine {
	t.Helper()
	held, err := Hold(h, Through("tui", h))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = held.Close() })
	return held
}

func TestSwitch_holdsTheOtherWorkspacesStoreInPlace(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	held := heldOn(t, h)
	if _, err := held.Ask(TaskNew, Input{Given: map[string]string{"title": "at work"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Switch(held, "home"); err != nil {
		t.Fatalf("Switch = %v", err)
	}
	listed, err := held.Ask(TaskList, Input{})
	if err != nil || len(listed.Tasks) != 0 {
		t.Errorf("home lists %d tasks, %v; want none of work's", len(listed.Tasks), err)
	}
	shown, err := held.Ask(WorkspaceShow, Input{})
	if err != nil || shown.WorkspaceShown.Workspace != "home" {
		t.Errorf("WorkspaceShow = %+v, %v; want home", shown.WorkspaceShown, err)
	}
	if _, err := os.Stat(filepath.Join(h.Dir, "h", "cairn.db")); err != nil {
		t.Errorf("home's store was not opened: %v", err)
	}
}

func TestSwitch_writesNoConfiguration(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	held := heldOn(t, h)
	if _, err := Switch(held, "home"); err != nil {
		t.Fatal(err)
	}
	said, err := os.ReadFile(config.FileIn(h.Dir))
	if err != nil || string(said) != twoWorkspaces {
		t.Errorf("the configuration reads %q, %v; want it as it was", said, err)
	}
}

func TestSwitch_aFolderGoneIsRefusedAndTheStoreHeldStays(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w")
	held := heldOn(t, h)
	if _, err := held.Ask(TaskNew, Input{Given: map[string]string{"title": "at work"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Switch(held, "home"); !errors.Is(err, ErrWorkspaceNotThere) {
		t.Fatalf("Switch = %v; want %v", err, ErrWorkspaceNotThere)
	}
	listed, err := held.Ask(TaskList, Input{})
	if err != nil || len(listed.Tasks) != 1 {
		t.Errorf("work lists %d tasks, %v; want its one still", len(listed.Tasks), err)
	}
}

func TestSwitch_keepsWhoActsAndThroughWhat(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	held := heldOn(t, h)
	was := held.acting
	if _, err := Switch(held, "home"); err != nil {
		t.Fatal(err)
	}
	if held.acting != was {
		t.Errorf("acting = %+v; want %+v", held.acting, was)
	}
}

func TestSwitch_saysThePagesOfTheStoreNowHeld(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	if err := os.WriteFile(filepath.Join(h.Dir, "h", config.SettingsFile), []byte("[pages]\ntime = \"off\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	held := heldOn(t, h)
	pages, err := Switch(held, "home")
	if err != nil {
		t.Fatal(err)
	}
	if pages.Shows(PageTime) {
		t.Errorf("home shows its time; want it off")
	}
	if pages, err = Switch(held, "work"); err != nil || !pages.Shows(PageTime) {
		t.Errorf("work hides its time, %v; want it shown", err)
	}
}

func TestWorkspace_foundSaysWhetherItsFolderIsThere(t *testing.T) {
	there := Workspace{Name: "work", Folder: t.TempDir()}
	gone := Workspace{Name: "home", Folder: filepath.Join(t.TempDir(), "gone")}
	if !there.Found() || gone.Found() {
		t.Errorf("Found = %v, %v; want true, false", there.Found(), gone.Found())
	}
}
