package engine

import (
	"cairn/internal/engine/workspace"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/store/files"
)

func removing(name string) Input { return Input{Given: map[string]string{"name": name}} }

func TestWorkspaceRemove_dropsTheEntryAndTouchesNoFile(t *testing.T) {
	h := homeHolding(t, "# mine\n"+twoWorkspaces+"\n[pages]\nobjectives = false\n", nil, "w", "h")
	store := filepath.Join(h.Dir, "h", "cairn.db")
	if err := os.WriteFile(store, []byte("kept"), 0o600); err != nil {
		t.Fatal(err)
	}
	answer, err := DoFrom(workspaceRemove, h, removing("home"), Acting{})
	if err != nil {
		t.Fatal(err)
	}
	said, _ := os.ReadFile(files.FileIn(h.Dir))
	if want := "# mine\n[workspaces]\ndefault = \"work\"\nwork = \"~/w\"\n\n[pages]\nobjectives = false\n"; string(said) != want {
		t.Errorf("the configuration reads\n%s\nwant\n%s", said, want)
	}
	if kept, err := os.ReadFile(store); err != nil || string(kept) != "kept" {
		t.Errorf("the store of home is %q, %v; want it as it was", kept, err)
	}
	removed := *answer.WorkspaceRemoved
	if removed.Name != "home" || removed.Folder != filepath.Join(h.Dir, "h") || removed.Written != files.FileIn(h.Dir) {
		t.Errorf("WorkspaceRemoved = %+v", removed)
	}
	if text := RenderText(workspaceRemove, answer); !strings.Contains(text, "home") || !strings.Contains(text, removed.Folder+" is left as it was") {
		t.Errorf("it says\n%s", text)
	}
	if got := Of(workspaceRemove, answer).(workspace.Removed); got != removed {
		t.Errorf("Of = %+v", got)
	}
}

func TestWorkspaceRemove_refusesAnUnknownNameAndTheDefault(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	if _, err := DoFrom(workspaceRemove, h, removing("play"), Acting{}); !errors.Is(err, workspace.ErrNoSuchWorkspace) {
		t.Errorf("removing play = %v; want %v", err, workspace.ErrNoSuchWorkspace)
	}
	_, err := DoFrom(workspaceRemove, h, removing("work"), Acting{})
	if !errors.Is(err, workspace.ErrRemovingTheDefault) || !strings.Contains(err.Error(), `"work"`) || !strings.Contains(err.Error(), "cairn workspace use") {
		t.Errorf("removing work = %v; want %v, naming it and how to choose another", err, workspace.ErrRemovingTheDefault)
	}
	if said, _ := os.ReadFile(files.FileIn(h.Dir)); string(said) != twoWorkspaces {
		t.Errorf("the configuration reads %q; want it as it was", said)
	}
}

func TestWorkspaceRemove_isForAPersonOnly(t *testing.T) {
	if _, found := called(workspaceRemove.Thing, workspaceRemove.Verb); found {
		t.Error("workspace remove is among All(), so offered to an agent")
	}
	for _, op := range ForAPersonOnly() {
		if op.Name() == workspaceRemove.Name() {
			if op.Level != write || op.Needs != aConfiguration {
				t.Errorf("workspace remove is at %v, needing %v", op.Level, op.Needs)
			}
			return
		}
	}
	t.Error("workspace remove is not for a person at the command line")
}
