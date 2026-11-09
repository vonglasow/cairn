package tui

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

func withRecent(t *testing.T) *roaming {
	r := newRoaming(t, newFake())
	r.recent = []engine.RecentFolder{
		{Store: "/s/scratch/cairn.db", Folder: "/s/scratch", Opened: today.Add(-24 * time.Hour)},
		{Store: "/v/trip/cairn.db", Folder: "/v/trip", Opened: today.Add(-48 * time.Hour), NotFound: true},
		{Store: "/o/odd/tasks.db", Opened: today},
	}
	r.shown["/s/scratch/cairn.db"] = engine.WorkspaceShown{Store: "/s/scratch/cairn.db"}
	return r
}

func TestSwitcher_listsTheRecentFoldersUnderTheWorkspaces(t *testing.T) {
	screen := strip(pressing(roamingBoard(t, withRecent(t)), "ctrl+o").Draw())
	if !strings.Contains(screen, "RECENT FOLDERS") || strings.Index(screen, "RECENT FOLDERS") < strings.Index(screen, "client-a") {
		t.Fatalf("the recent folders are not listed under the workspaces:\n%s", screen)
	}
	for folder, mark := range map[string]string{"◌ /s/scratch ": "opened yesterday", "◌ /v/trip ": "not found now", "◌ /o/odd/tasks.db ": "opened today"} {
		if got := lineOf(screen, folder); !strings.Contains(got, mark) {
			t.Errorf("%s's line = %q; want it marked %q", folder, got, mark)
		}
	}
	if screen := strip(pressing(roamingBoard(t, newRoaming(t, newFake())), "ctrl+o").Draw()); strings.Contains(screen, "RECENT FOLDERS") {
		t.Errorf("with none, the recent folders are still listed:\n%s", screen)
	}
}

func TestSwitcher_aRecentFolderChosenOpensItsStore(t *testing.T) {
	r := withRecent(t)
	b := pressing(roamingBoard(t, r), "ctrl+o", "down", "down", "enter")
	if !slices.Equal(r.switched, []string{"/s/scratch/cairn.db"}) || b.screen != OnTheBoard {
		t.Errorf("switched = %v, screen %v; want the store of scratch opened", r.switched, b.screen)
	}
	r = withRecent(t)
	b = pressing(roamingBoard(t, r), "ctrl+o", "down", "down", "down", "down", "down", "down")
	if b.spaceAt != 5 {
		t.Errorf("the cursor is at %d; want it held on the last recent folder, 5", b.spaceAt)
	}
}

func TestSwitcher_theRecentFoldersNarrowAsOneTypes(t *testing.T) {
	b := pressing(roamingBoard(t, withRecent(t)), "ctrl+o", "t", "r", "i")
	if got := b.recentListed(); len(got) != 1 || got[0].Folder != "/v/trip" {
		t.Errorf("recent listed = %+v; want trip alone", got)
	}
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
	if len(r.switched) != 0 || b.screen != OnTheBoard {
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
	if !slices.Equal(r.switched, []string{"~/x"}) || b.screen != OnTheBoard {
		t.Errorf("switched = %v, screen %v; want ~/x opened", r.switched, b.screen)
	}
}

func TestSwitcher_aFolderRefusedSaysWhyAndKeepsWhatWasTyped(t *testing.T) {
	r := withRecent(t)
	r.refuse = fmt.Errorf("%q: must be absolute", "x")
	b := pressing(roamingBoard(t, r), "ctrl+o", "ctrl+o", "x", "enter")
	screen := strip(b.Draw())
	if b.screen != OnWorkspaces || !strings.Contains(screen, "must be absolute") || !strings.Contains(screen, "folder › x") {
		t.Errorf("a refused folder shows:\n%s", screen)
	}
	b = pressing(b, "esc")
	if screen := strip(b.Draw()); b.screen != OnWorkspaces || strings.Contains(screen, "folder ›") {
		t.Errorf("esc leaves:\n%s", screen)
	}
	if b = pressing(b, "esc"); b.screen != OnTheBoard {
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
	if !slices.Equal(r.added, []string{"garden /g"}) || !slices.Equal(r.switched, []string{"garden"}) {
		t.Errorf("added %v, switched %v; want garden made at /g, then opened", r.added, r.switched)
	}
	if !strings.HasPrefix(headerOf(b), "● garden") {
		t.Errorf("header = %q; want garden", headerOf(b))
	}
}

func TestSwitcher_aNewWorkspaceRefusedOpensNothing(t *testing.T) {
	r := withRecent(t)
	b := pressing(roamingBoard(t, r), "ctrl+o", "ctrl+n", "w", "o", "r", "k", "enter", "/", "w", "enter")
	if len(r.switched) != 0 || !strings.Contains(strip(b.Draw()), engine.ErrWorkspaceKnown.Error()) {
		t.Errorf("switched = %v; want work refused, saying why:\n%s", r.switched, strip(b.Draw()))
	}
}

func TestSwitcher_ctrlKKeepsARecentFolderAsAWorkspace(t *testing.T) {
	r := withRecent(t)
	b := pressing(roamingBoard(t, r), "ctrl+o", "down", "down")
	if screen := strip(b.Draw()); !strings.Contains(screen, "^K keep as a workspace") {
		t.Fatalf("a recent folder chosen offers no ^K:\n%s", screen)
	}
	b = pressing(b, "ctrl+k")
	if screen := strip(b.Draw()); !strings.Contains(screen, "keep as ›") || !strings.Contains(screen, "⏎ keep /s/scratch") {
		t.Fatalf("ctrl+k asks for no name:\n%s", screen)
	}
	b = pressing(b, "s", "c", "enter")
	if !slices.Equal(r.added, []string{"sc /s/scratch"}) || len(r.switched) != 0 {
		t.Errorf("added %v, switched %v; want scratch kept as sc, nothing opened", r.added, r.switched)
	}
	if b.screen != OnWorkspaces || !slices.Contains(names(b.listed()), "sc") || len(b.recentListed()) != 2 {
		t.Errorf("screen %v, listed %v, recent %+v; want sc among the workspaces", b.screen, names(b.listed()), b.recentListed())
	}
}

func TestSwitcher_ctrlKOnWhatCannotBeKeptDoesNothing(t *testing.T) {
	for _, keys := range [][]string{{"ctrl+o", "ctrl+k"}, {"ctrl+o", "down", "down", "down", "down", "ctrl+k"}} {
		b := pressing(roamingBoard(t, withRecent(t)), keys...)
		if screen := strip(b.Draw()); strings.Contains(screen, "keep") || strings.Contains(screen, "^K") {
			t.Errorf("%v offers to keep:\n%s", keys, screen)
		}
	}
}

func TestSwitcher_theFootOffersToOpenAFolderAndMakeAWorkspace(t *testing.T) {
	screen := strip(pressing(roamingBoard(t, withRecent(t)), "ctrl+o").Draw())
	for _, want := range []string{"^O open a folder", "^N new workspace"} {
		if !strings.Contains(screen, want) {
			t.Errorf("the list does not offer %q:\n%s", want, screen)
		}
	}
}

func inked(colour string) string {
	ink := domain.LabelInks[colour]
	return fmt.Sprintf("38;2;%d;%d;%dm● ", ink[0], ink[1], ink[2])
}

func TestHeader_theDotWearsTheColourOfTheWorkspace(t *testing.T) {
	r := newRoaming(t, newFake())
	r.shown["work"] = engine.WorkspaceShown{Workspace: "work", Store: "/c/work/cairn.db", Colour: "green"}
	if header := strings.SplitN(roamingBoard(t, r).Draw(), "\n", 2)[0]; !strings.Contains(header, inked("green")) {
		t.Errorf("the header's dot does not wear green: %q", header)
	}
}

func TestSwitcher_eachDotWearsTheColourOfItsWorkspace(t *testing.T) {
	r := newRoaming(t, newFake())
	r.spaces[1].Colour = "peach"
	r.spaces[2].Colour = "sky"
	drawn := pressing(roamingBoard(t, r), "ctrl+o").Draw()
	for colour, name := range map[string]string{"peach": "home", "sky": "client-a"} {
		at := strings.Index(drawn, inked(colour))
		if at < 0 || !strings.HasPrefix(strip(drawn[at+len(inked(colour))-len("● "):]), "● "+name) {
			t.Errorf("the list does not draw the dot of %s in %s", name, colour)
		}
	}
}

func TestSwitcher_withRecentFoldersFitsANarrowWindow(t *testing.T) {
	r := withRecent(t)
	r.recent[0].Folder = "/" + strings.Repeat("deep/", 30)
	for _, keys := range [][]string{{"ctrl+o"}, {"ctrl+o", "ctrl+n"}, {"ctrl+o", "down", "down"}, {"ctrl+o", "down", "down", "ctrl+k"}} {
		b := roamingBoard(t, r)
		b.width, b.height = Narrowest, Shortest
		if err := fits(pressing(b, keys...)); err != nil {
			t.Errorf("%v: %v", keys, err)
		}
	}
}
