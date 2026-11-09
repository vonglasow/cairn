package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cairn/internal/config"
)

func TestSwitch_aFolderTypedOpensItsStoreAndIsRemembered(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h", "scratch")
	held := heldOn(t, h)
	if _, err := Switch(held, Elsewhere{Folder: "~/scratch"}); err != nil {
		t.Fatalf("Switch = %v", err)
	}
	shown, err := held.Ask(WorkspaceShow, Input{})
	if err != nil || shown.WorkspaceShown.Store != filepath.Join(h.Dir, "scratch", "cairn.db") || shown.WorkspaceShown.Workspace != "" {
		t.Errorf("WorkspaceShow = %+v, %v; want the store of scratch, no workspace", shown.WorkspaceShown, err)
	}
	if recent := recentOf(t, h); len(recent) != 1 || recent[0].Folder != filepath.Join(h.Dir, "scratch") {
		t.Errorf("Recent = %+v; want scratch", recent)
	}
}

func TestSwitch_aFolderNotAbsoluteOrNotThereIsRefusedAndTheStoreHeldStays(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	held := heldOn(t, h)
	if _, err := Switch(held, Elsewhere{Folder: "scratch"}); !errors.Is(err, config.ErrNotAbsolute) {
		t.Errorf("Switch to a relative folder = %v; want %v", err, config.ErrNotAbsolute)
	}
	gone := filepath.Join(h.Dir, "gone")
	if _, err := Switch(held, Elsewhere{Folder: gone}); !errors.Is(err, ErrNotThere) {
		t.Errorf("Switch to a folder not there = %v; want %v", err, ErrNotThere)
	}
	if _, err := os.Stat(gone); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the folder not there was made: %v", err)
	}
	if shown, _ := held.Ask(WorkspaceShow, Input{}); shown.WorkspaceShown.Workspace != "work" {
		t.Errorf("the board holds %+v; want work still", shown.WorkspaceShown)
	}
}

func TestSwitch_aStoreOpenedBeforeOpensAgainAndOneGoneIsRefused(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h", "scratch")
	held := heldOn(t, h)
	store := filepath.Join(h.Dir, "scratch", "cairn.db")
	if _, err := Switch(held, Elsewhere{Folder: "~/scratch"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Switch(held, Elsewhere{Workspace: "work"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Switch(held, Elsewhere{Store: store}); err != nil {
		t.Errorf("Switch to the store opened before = %v", err)
	}
	gone := filepath.Join(h.Dir, "trip", "cairn.db")
	if _, err := Switch(held, Elsewhere{Store: gone}); !errors.Is(err, ErrNotThere) {
		t.Errorf("Switch to a store gone = %v; want %v", err, ErrNotThere)
	}
	if _, err := os.Stat(filepath.Dir(gone)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the folder of the store gone was made: %v", err)
	}
}
