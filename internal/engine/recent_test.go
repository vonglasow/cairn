package engine

import (
	"cairn/internal/config"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func recentOf(t *testing.T, h Home) []RecentFolder {
	t.Helper()
	answer, err := DoFrom(WorkspaceList, h, Input{}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	return answer.Workspaces.Recent
}

func TestHold_aStoreNamedByTheFlagIsRememberedAmongTheRecentFolders(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	h.Store = filepath.Join(h.Dir, "loose", "cairn.db")
	heldOn(t, h)
	h.Store = ""
	recent := recentOf(t, h)
	if len(recent) != 1 || recent[0].Store != filepath.Join(h.Dir, "loose", "cairn.db") || recent[0].Folder != filepath.Join(h.Dir, "loose") {
		t.Fatalf("Recent = %+v; want the loose store alone", recent)
	}
	if time.Since(recent[0].Opened) > time.Minute || recent[0].NotFound {
		t.Errorf("Recent = %+v; want it opened now, and found", recent[0])
	}
}

func TestHold_aStoreTheEnvironmentNamesIsRemembered(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{"CAIRN_STORE": "~/elsewhere/cairn.db"}, "w", "h")
	heldOn(t, h)
	if recent := recentOf(t, h); len(recent) != 1 || recent[0].Folder != filepath.Join(h.Dir, "elsewhere") {
		t.Errorf("Recent = %+v; want the store the environment names", recent)
	}
}

func TestHold_aWorkspaceOrTheDefaultIsNotRemembered(t *testing.T) {
	for _, c := range []struct {
		name string
		with func(Home) Home
	}{
		{"the default", func(h Home) Home { return h }},
		{"--workspace", func(h Home) Home { h.Workspace = "home"; return h }},
		{"--store naming a workspace's store", func(h Home) Home { h.Store = filepath.Join(h.Dir, "h", "cairn.db"); return h }},
	} {
		h := homeHolding(t, twoWorkspaces, nil, "w", "h")
		heldOn(t, c.with(h))
		if recent := recentOf(t, h); len(recent) != 0 {
			t.Errorf("%s: Recent = %+v; want none", c.name, recent)
		}
	}
}

func TestHold_theFilesStoreIsNotRemembered(t *testing.T) {
	h := homeHolding(t, "store = \"~/s/cairn.db\"\n", nil, "s")
	heldOn(t, h)
	if recent := recentOf(t, h); len(recent) != 0 {
		t.Errorf("Recent = %+v; want none for the store the file names", recent)
	}
}

func TestHold_tenAreKeptTheNewestFirstAndOneOpenedAgainMovesToTheHead(t *testing.T) {
	h := homeHolding(t, "", nil)
	for n := range 12 {
		h.Store = filepath.Join(h.Dir, fmt.Sprint(n), "cairn.db")
		heldOn(t, h)
	}
	h.Store = filepath.Join(h.Dir, "5", "cairn.db")
	heldOn(t, h)
	h.Store = ""
	recent := recentOf(t, h)
	if len(recent) != recentKept {
		t.Fatalf("%d recent folders are kept; want %d", len(recent), recentKept)
	}
	var folders []string
	for _, one := range recent {
		folders = append(folders, filepath.Base(one.Folder))
	}
	if got := strings.Join(folders, " "); got != "5 11 10 9 8 7 6 4 3 2" {
		t.Errorf("the recent folders read %s; want 5 11 10 9 8 7 6 4 3 2", got)
	}
}

func TestRecent_aStoreGoneIsSaidNotFound(t *testing.T) {
	h := homeHolding(t, "", nil)
	h.Store = filepath.Join(h.Dir, "trip", "cairn.db")
	_ = heldOn(t, h).Close()
	if err := os.RemoveAll(filepath.Join(h.Dir, "trip")); err != nil {
		t.Fatal(err)
	}
	h.Store = ""
	if recent := recentOf(t, h); len(recent) != 1 || !recent[0].NotFound {
		t.Errorf("Recent = %+v; want the trip said not found", recent)
	}
}

func TestRecent_aStoreNotCalledCairnDbHasNoFolderToKeep(t *testing.T) {
	h := homeHolding(t, "", nil)
	h.Store = filepath.Join(h.Dir, "odd", "tasks.db")
	heldOn(t, h)
	h.Store = ""
	if recent := recentOf(t, h); len(recent) != 1 || recent[0].Folder != "" || recent[0].Shown() != filepath.Join(h.Dir, "odd", "tasks.db") {
		t.Errorf("Recent = %+v; want it shown by its path, with no folder to keep as a workspace", recent)
	}
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
	h := homeHolding(t, "", map[string]string{StateVariable: state})
	h.Store = filepath.Join(h.Dir, "loose", "cairn.db")
	heldOn(t, h)
	if _, err := os.Stat(filepath.Join(state, "cairn", "recent.json")); err != nil {
		t.Errorf("nothing kept under XDG_STATE_HOME: %v", err)
	}
	if _, err := os.Stat(filepath.Join(h.Dir, ".local", "state")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the home's state folder was written too: %v", err)
	}
}

func TestRecent_aRelativeXDGStateHomeIsPassedOver(t *testing.T) {
	h := homeHolding(t, "", map[string]string{StateVariable: "state"})
	h.Store = filepath.Join(h.Dir, "loose", "cairn.db")
	heldOn(t, h)
	if _, err := os.Stat(filepath.Join(h.Dir, ".local", "state", "cairn", "recent.json")); err != nil {
		t.Errorf("nothing kept under the home's state folder: %v", err)
	}
}

func TestRecent_aFileThatCannotBeReadIsNoneAndTheBoardStillOpens(t *testing.T) {
	h := homeHolding(t, "", nil)
	path := filepath.Join(h.Dir, ".local", "state", "cairn", "recent.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if recent := recentOf(t, h); len(recent) != 0 {
		t.Errorf("Recent = %+v; want none from a file that cannot be read", recent)
	}
	h.Store = filepath.Join(h.Dir, "loose", "cairn.db")
	heldOn(t, h)
	h.Store = ""
	if recent := recentOf(t, h); len(recent) != 1 {
		t.Errorf("Recent = %+v; want the loose store, the unreadable file written over", recent)
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
		if got := (RecentFolder{Opened: c.opened, NotFound: c.gone}).Said(now); got != c.want {
			t.Errorf("Said(%v, gone %v) = %q; want %q", c.opened, c.gone, got, c.want)
		}
	}
}

func TestWorkspaceList_saysTheRecentFoldersUnderTheWorkspaces(t *testing.T) {
	said := RenderText(WorkspaceList, Answer{Workspaces: &Workspaces{
		Workspaces: []Workspace{{Name: "work", Folder: "/w", Default: true, Colour: "green"}},
		Recent:     []RecentFolder{{Store: "/l/cairn.db", Folder: "/l", Opened: time.Now()}, {Store: "/t/x.db", NotFound: true}},
	}})
	for _, want := range []string{"work  /w  default  green\n", "\nrecent folders, opened by path:\n", "  /l  opened today\n", "  /t/x.db  not found now\n"} {
		if !strings.Contains(said, want) {
			t.Errorf("workspace list says:\n%s\nwant it saying %q", said, want)
		}
	}
}

func TestRecentFolder_isNamedByItsSettings(t *testing.T) {
	folder := t.TempDir()
	if err := os.WriteFile(filepath.Join(folder, config.SettingsFile), []byte("[workspace]\nname = \"mine\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := (RecentFolder{Store: filepath.Join(folder, "cairn.db"), Folder: folder}).Name(); got != "mine" {
		t.Errorf("named %q", got)
	}
}

func TestRemember_aWorkspaceOpenedIsNotWrittenDown(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	h.Workspace = "home"
	heldOn(t, h)
	if kept := rememberedIn(recentPath(h)); len(kept) != 0 {
		t.Errorf("remembered %+v", kept)
	}
}

func TestRecentIn_listsALooseStoreAfterAWorkspacesOwn(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h", "loose")
	loose := filepath.Join(h.Dir, "loose", "cairn.db")
	if err := remember(h, loose, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := remember(h, filepath.Join(h.Dir, "h", "cairn.db"), time.Now()); err != nil {
		t.Fatal(err)
	}
	if recent := recentOf(t, h); len(recent) != 1 || recent[0].Store != loose {
		t.Errorf("recent %+v, want the loose store", recent)
	}
}

func TestRecentFolder_whoseSettingsNameNothingIsNamedByItsFolder(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "mine")
	if err := os.MkdirAll(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, config.SettingsFile), []byte("[pages]\ntime = \"on\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := (RecentFolder{Store: filepath.Join(folder, "cairn.db"), Folder: folder}).Name(); got != "mine" {
		t.Errorf("named %q", got)
	}
}

func TestRemember_aStoreNamedThatIsAWorkspacesOwnIsNotWrittenDown(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	h.Store = filepath.Join(h.Dir, "h", "cairn.db")
	heldOn(t, h)
	if kept := rememberedIn(recentPath(h)); len(kept) != 0 {
		t.Errorf("remembered %+v", kept)
	}
}
