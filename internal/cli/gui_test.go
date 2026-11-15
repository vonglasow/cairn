package cli

import (
	"cairn/internal/board"
	"cairn/internal/debug"
	"cairn/internal/engine"
	"cairn/internal/gui"
	"cairn/internal/gui/design"
	"path/filepath"
	"testing"
)

func TestGuiTellsTheWindowItsStoreAndHowToReachAnotherWorkspace(t *testing.T) {
	store := storeFor(t)
	t.Setenv("CAIRN_STORE", store)
	home := t.TempDir()
	t.Setenv("HOME", home)
	if got := run(t, "workspace", "add", "garden", filepath.Join(home, "garden")); got.Code != 0 {
		t.Fatalf("workspace add exited %d: %s", got.Code, got.Stderr)
	}
	var here gui.Here
	was := openGui
	openGui = func(_ board.Work, _ engine.Pages, _ design.Skin, _ *debug.Log, _, _ int, given gui.Here) error {
		here = given
		return nil
	}
	t.Cleanup(func() { openGui = was })
	if got := run(t, "gui"); got.Code != 0 {
		t.Fatalf("gui exited %d: %s", got.Code, got.Stderr)
	}
	if here.Store != store {
		t.Errorf("the window was told it shows %q, not %q", here.Store, store)
	}
	if here.Reach == nil {
		t.Fatal("the window was given no way to reach another workspace")
	}
	apart, err := here.Reach(engine.Elsewhere{Workspace: "garden"})
	if err != nil || apart.Store != filepath.Join(home, "garden", "cairn.db") {
		t.Errorf("reaching garden = %+v, %v", apart, err)
	}
}
