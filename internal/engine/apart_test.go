package engine

import (
	"cairn/internal/engine/workspace"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestOpenApart_aWorkspaceIsNamedToTheCairnStartedForIt(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	held := heldOn(t, h)
	apart, err := OpenApart(held, Elsewhere{Workspace: "home"})
	if err != nil {
		t.Fatalf("OpenApart = %v", err)
	}
	if apart.Store != filepath.Join(h.Dir, "h", "cairn.db") || !slices.Equal(apart.Flags, []string{"--workspace", "home"}) {
		t.Errorf("OpenApart = %+v; want home's store, named by --workspace", apart)
	}
	if shown, _ := held.Ask(WorkspaceShow, Input{}); shown.WorkspaceShown.Workspace != "work" {
		t.Errorf("the engine held moved to %q; it holds work still", shown.WorkspaceShown.Workspace)
	}
}

func TestOpenApart_aFolderOrAStoreIsNamedByItsStore(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h", "scratch")
	held := heldOn(t, h)
	store := filepath.Join(h.Dir, "scratch", "cairn.db")
	for _, to := range []Elsewhere{{Folder: filepath.Join(h.Dir, "scratch")}, {Store: store}} {
		if err := os.WriteFile(store, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		apart, err := OpenApart(held, to)
		if err != nil || apart.Store != store || !slices.Equal(apart.Flags, []string{"--store", store}) {
			t.Errorf("OpenApart(%+v) = %+v, %v; want the store of scratch, named by --store", to, apart, err)
		}
	}
}

func TestOpenApart_whatIsNotThereIsRefused(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w")
	held := heldOn(t, h)
	if _, err := OpenApart(held, Elsewhere{Folder: filepath.Join(h.Dir, "gone")}); !errors.Is(err, workspace.ErrNotThere) {
		t.Errorf("a folder not there = %v; want %v", err, workspace.ErrNotThere)
	}
	if _, err := OpenApart(held, Elsewhere{Workspace: "home"}); !errors.Is(err, workspace.ErrWorkspaceNotThere) {
		t.Errorf("a workspace whose folder is gone = %v; want %v", err, workspace.ErrWorkspaceNotThere)
	}
	if _, err := OpenApart(held, Elsewhere{Workspace: "garden"}); !errors.Is(err, workspace.ErrNoSuchWorkspace) {
		t.Errorf("a workspace nobody named = %v; want %v", err, workspace.ErrNoSuchWorkspace)
	}
}

func TestPages_aCardCostsNothingShownWhenTheTimeIsNot(t *testing.T) {
	c := Card{Spent: "2h", Estimate: "1h", Overrun: true}
	if said, over := (Pages{}).Cost(c); said != "2h/1h" || !over {
		t.Errorf("with the time shown, Cost = %q, %v", said, over)
	}
	if said, over := (Pages{off: []string{PageTime}}).Cost(c); said != "" || over {
		t.Errorf("with the time not shown, Cost = %q, %v; want nothing", said, over)
	}
}
