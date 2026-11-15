package engine

import (
	"cairn/internal/engine/workspace"
	"cairn/internal/store/files"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func recentOf(t *testing.T, h Home) []workspace.RecentFolder {
	t.Helper()
	answer, err := DoFrom(WorkspaceList, h, Input{}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	return answer.Workspaces.Recent
}

func TestRecent_aStoreAWorkspaceNowNamesIsLeftOut(t *testing.T) {
	h := homeHolding(t, "", nil)
	h.Store = filepath.Join(h.Dir, "loose", "cairn.db")
	heldOn(t, h)
	h.Store = ""
	if _, err := DoFrom(WorkspaceAdd, h, Input{Given: map[string]string{"name": "loose", "folder": "~/loose"}}, Acting{}); err != nil {
		t.Fatal(err)
	}
	if recent := recentOf(t, h); len(recent) != 0 {
		t.Errorf("Recent = %+v; want the store loose names left out", recent)
	}
}

func TestRecent_isKeptUnderXDGStateHomeWhenItIsAbsolute(t *testing.T) {
	state := t.TempDir()
	h := homeHolding(t, "", map[string]string{workspace.StateVariable: state})
	h.Store = filepath.Join(h.Dir, "loose", "cairn.db")
	heldOn(t, h)
	if _, err := os.Stat(filepath.Join(state, "cairn", "recent.json")); err != nil {
		t.Errorf("nothing kept under XDG_STATE_HOME: %v", err)
	}
	if _, err := os.Stat(filepath.Join(h.Dir, ".local", "state")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the home's state folder was written too: %v", err)
	}
}

func TestRecent_withNoHomeNothingIsKept(t *testing.T) {
	if path := recentPath(Home{}); path != "" {
		t.Errorf("recentPath = %q; want none without a home or a state folder", path)
	}
}

func TestRecentFolder_saysWhenItWasOpened(t *testing.T) {
	now := time.Date(2026, 11, 9, 10, 0, 0, 0, time.Local)
	for _, c := range []struct {
		opened time.Time
		gone   bool
		want   string
	}{
		{now.Add(-time.Hour), false, "opened today"},
		{now.Add(time.Hour * 30), false, "opened today"},
		{time.Date(2026, 11, 8, 23, 0, 0, 0, time.Local), false, "opened yesterday"},
		{time.Date(2026, 11, 3, 9, 0, 0, 0, time.Local), false, "opened 3 Nov"},
		{time.Date(2025, 11, 3, 9, 0, 0, 0, time.Local), false, "opened 3 Nov 2025"},
		{now, true, "not found now"},
	} {
		if got := (workspace.RecentFolder{Opened: c.opened, NotFound: c.gone}).Said(now); got != c.want {
			t.Errorf("Said(%v, gone %v) = %q; want %q", c.opened, c.gone, got, c.want)
		}
	}
}

func TestRecentFolder_isNamedByItsSettings(t *testing.T) {
	folder := t.TempDir()
	if err := os.WriteFile(filepath.Join(folder, files.SettingsFile), []byte("[workspace]\nname = \"mine\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := (workspace.RecentFolder{Store: filepath.Join(folder, "cairn.db"), Folder: folder}).Name(); got != "mine" {
		t.Errorf("named %q", got)
	}
}

func TestRecentFolder_whoseSettingsNameNothingIsNamedByItsFolder(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "mine")
	if err := os.MkdirAll(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, files.SettingsFile), []byte("[pages]\ntime = \"on\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := (workspace.RecentFolder{Store: filepath.Join(folder, "cairn.db"), Folder: folder}).Name(); got != "mine" {
		t.Errorf("named %q", got)
	}
}
