package tui

import (
	"cairn/internal/board"
	"cairn/internal/engine/workspace"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"
)

func withRecent(t *testing.T) *roaming {
	r := newRoaming(t, newFake())
	r.recent = []workspace.RecentFolder{
		{Store: "/s/scratch/cairn.db", Folder: "/s/scratch", Opened: today.Add(-24 * time.Hour)},
		{Store: "/v/trip/cairn.db", Folder: "/v/trip", Opened: today.Add(-48 * time.Hour), NotFound: true},
		{Store: "/o/odd/tasks.db", Opened: today},
	}
	r.shown["/s/scratch/cairn.db"] = workspace.Shown{Store: "/s/scratch/cairn.db"}
	return r
}

func TestSwitcher_theRecentFolderHereOnlyCloses(t *testing.T) {
	r := withRecent(t)
	r.on = "/s/scratch/cairn.db"
	b := roamingBoard(t, r)
	screen := strip(pressing(b, "ctrl+o").Draw())
	if got := lineOf(screen, "│   ◌ /s/scratch "); !strings.Contains(got, "here") {
		t.Errorf("scratch's line = %q; want it marked here", got)
	}
	b = pressing(b, "ctrl+o", "down", "down", "down", "enter")
	if len(r.switched) != 0 || b.screen != board.OnTheBoard {
		t.Errorf("switched = %v, screen %v; want nothing opened", r.switched, b.screen)
	}
}

func TestSwitcher_ctrlOOpensAFolderTyped(t *testing.T) {
	r := withRecent(t)
	b := pressing(roamingBoard(t, r), "ctrl+o", "ctrl+o")
	if screen := strip(b.Draw()); !strings.Contains(screen, "folder ›") || !strings.Contains(screen, "absolute, or under ~") {
		t.Fatalf("ctrl+o in the list asks for no folder:\n%s", screen)
	}
	b = pressing(b, "~", "/", "x", "y", "backspace", "enter")
	if !slices.Equal(r.switched, []string{"~/x"}) || b.screen != board.OnTheBoard {
		t.Errorf("switched = %v, screen %v; want ~/x opened", r.switched, b.screen)
	}
}

func TestSwitcher_aFolderRefusedSaysWhyAndKeepsWhatWasTyped(t *testing.T) {
	r := withRecent(t)
	r.refuse = fmt.Errorf("%q: must be absolute", "x")
	b := pressing(roamingBoard(t, r), "ctrl+o", "ctrl+o", "x", "enter")
	screen := strip(b.Draw())
	if b.screen != board.OnWorkspaces || !strings.Contains(screen, "must be absolute") || !strings.Contains(screen, "folder › x") {
		t.Errorf("a refused folder shows:\n%s", screen)
	}
	b = pressing(b, "esc")
	if screen := strip(b.Draw()); b.screen != board.OnWorkspaces || strings.Contains(screen, "folder ›") {
		t.Errorf("esc leaves:\n%s", screen)
	}
	if b = pressing(b, "esc"); b.screen != board.OnTheBoard {
		t.Errorf("a second esc leaves the screen at %v", b.screen)
	}
}

func TestSwitcher_ctrlNMakesAWorkspaceAndOpensIt(t *testing.T) {
	r := withRecent(t)
	b := pressing(roamingBoard(t, r), "ctrl+o", "ctrl+n", "g", "a", "r", "d", "e", "n")
	if screen := strip(b.Draw()); !strings.Contains(screen, "new workspace › garden") {
		t.Fatalf("ctrl+n asks for no name:\n%s", screen)
	}
	b = pressing(b, "enter")
	if screen := strip(b.Draw()); !strings.Contains(screen, "garden's folder ›") {
		t.Fatalf("enter asks for no folder:\n%s", screen)
	}
	b = pressing(b, "/", "g", "enter")
	if screen := strip(b.Draw()); !strings.Contains(screen, "garden's colour ›") {
		t.Fatalf("enter asks for no colour:\n%s", screen)
	}
	b = pressing(b, "t", "e", "a", "l", "enter")
	if !slices.Equal(r.added, []string{"garden /g teal"}) || !slices.Equal(r.switched, []string{"garden"}) {
		t.Errorf("added %v, switched %v; want garden made at /g in teal, then opened", r.added, r.switched)
	}
	if !strings.HasPrefix(headerOf(b), "● garden") {
		t.Errorf("header = %q; want garden", headerOf(b))
	}
}

func TestSwitcher_aNewWorkspaceRefusedOpensNothing(t *testing.T) {
	r := withRecent(t)
	b := pressing(roamingBoard(t, r), "ctrl+o", "ctrl+n", "w", "o", "r", "k", "enter", "/", "w", "enter", "enter")
	if len(r.switched) != 0 || !strings.Contains(strip(b.Draw()), workspace.ErrWorkspaceKnown.Error()) {
		t.Errorf("switched = %v; want work refused, saying why:\n%s", r.switched, strip(b.Draw()))
	}
}

func TestSwitcher_ctrlKKeepsARecentFolderAsAWorkspace(t *testing.T) {
	r := withRecent(t)
	b := pressing(roamingBoard(t, r), "ctrl+o", "down", "down")
	if screen := flat(strip(b.Draw())); !strings.Contains(screen, "^K keep as a workspace") {
		t.Fatalf("a recent folder chosen offers no ^K:\n%s", screen)
	}
	b = pressing(b, "ctrl+k")
	if screen := flat(strip(b.Draw())); !strings.Contains(screen, "keep as ›") || !strings.Contains(screen, "⏎ keep /s/scratch") {
		t.Fatalf("ctrl+k asks for no name:\n%s", screen)
	}
	b = pressing(b, "s", "c", "enter")
	if !slices.Equal(r.added, []string{"sc /s/scratch"}) || len(r.switched) != 0 {
		t.Errorf("added %v, switched %v; want scratch kept as sc, nothing opened", r.added, r.switched)
	}
	if b.screen != board.OnWorkspaces || !slices.Contains(names(b.listed()), "sc") || len(b.recentListed()) != 2 {
		t.Errorf("screen %v, listed %v, recent %+v; want sc among the workspaces", b.screen, names(b.listed()), b.recentListed())
	}
}
