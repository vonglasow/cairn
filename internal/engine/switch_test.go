package engine

import (
	"cairn/internal/engine/workspace"
	"errors"
	"testing"
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

func TestSwitch_aFolderGoneIsRefusedAndTheStoreHeldStays(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w")
	held := heldOn(t, h)
	if _, err := held.Ask(TaskNew, Input{Given: map[string]string{"title": "at work"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Switch(held, Elsewhere{Workspace: "home"}); !errors.Is(err, workspace.ErrWorkspaceNotThere) {
		t.Fatalf("Switch = %v; want %v", err, workspace.ErrWorkspaceNotThere)
	}
	listed, err := held.Ask(TaskList, Input{})
	if err != nil || len(listed.Tasks) != 1 {
		t.Errorf("work lists %d tasks, %v; want its one still", len(listed.Tasks), err)
	}
}
