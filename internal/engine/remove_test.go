package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/config"
)

func removing(name string) Input { return Input{Given: map[string]string{"name": name}} }

func TestWorkspaceRemove_dropsTheEntryAndTouchesNoFile(t *testing.T) {
	h := homeHolding(t, "# mine\n"+twoWorkspaces+"\n[pages]\nobjectives = false\n", nil, "w", "h")
	store := filepath.Join(h.Dir, "h", "cairn.db")
	if err := os.WriteFile(store, []byte("kept"), 0o600); err != nil {
		t.Fatal(err)
	}
	answer, err := DoFrom(WorkspaceRemove, h, removing("home"), Acting{})
	if err != nil {
		t.Fatal(err)
	}
	said, _ := os.ReadFile(config.FileIn(h.Dir))
	if want := "# mine\n[workspaces]\ndefault = \"work\"\nwork = \"~/w\"\n\n[pages]\nobjectives = false\n"; string(said) != want {
		t.Errorf("the configuration reads\n%s\nwant\n%s", said, want)
	}
	if kept, err := os.ReadFile(store); err != nil || string(kept) != "kept" {
		t.Errorf("the store of home is %q, %v; want it as it was", kept, err)
	}
	removed := *answer.WorkspaceRemoved
	if removed.Name != "home" || removed.Folder != filepath.Join(h.Dir, "h") || removed.Written != config.FileIn(h.Dir) {
		t.Errorf("WorkspaceRemoved = %+v", removed)
	}
	if text := RenderText(WorkspaceRemove, answer); !strings.Contains(text, "home") || !strings.Contains(text, removed.Folder+" is left as it was") {
		t.Errorf("it says\n%s", text)
	}
	if got := Of(WorkspaceRemove, answer).(WorkspaceRemoved); got != removed {
		t.Errorf("Of = %+v", got)
	}
}

func TestWorkspaceRemove_refusesAnUnknownNameAndTheDefault(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	if _, err := DoFrom(WorkspaceRemove, h, removing("play"), Acting{}); !errors.Is(err, ErrNoSuchWorkspace) {
		t.Errorf("removing play = %v; want %v", err, ErrNoSuchWorkspace)
	}
	_, err := DoFrom(WorkspaceRemove, h, removing("work"), Acting{})
	if !errors.Is(err, ErrRemovingTheDefault) || !strings.Contains(err.Error(), `"work"`) || !strings.Contains(err.Error(), "cairn workspace use") {
		t.Errorf("removing work = %v; want %v, naming it and how to choose another", err, ErrRemovingTheDefault)
	}
	if said, _ := os.ReadFile(config.FileIn(h.Dir)); string(said) != twoWorkspaces {
		t.Errorf("the configuration reads %q; want it as it was", said)
	}
}

func TestWorkspaceRemove_isForAPersonOnly(t *testing.T) {
	if _, found := Called(WorkspaceRemove.Thing, WorkspaceRemove.Verb); found {
		t.Error("workspace remove is among All(), so offered to an agent")
	}
	for _, op := range ForAPersonOnly() {
		if op.Name() == WorkspaceRemove.Name() {
			if op.Level != Write || op.Needs != AConfiguration {
				t.Errorf("workspace remove is at %v, needing %v", op.Level, op.Needs)
			}
			return
		}
	}
	t.Error("workspace remove is not for a person at the command line")
}
