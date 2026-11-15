package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/workspace"
	"cairn/internal/gui/geom"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"gioui.org/io/key"

	"cairn/internal/engine"
)

func windowsIn(t *testing.T) Windows {
	t.Helper()
	dir, err := os.MkdirTemp("", "w")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return Windows{Dir: dir}
}

func TestWindows_aStoreClaimedIsShownAndRaisedThroughItsSocket(t *testing.T) {
	ws := windowsIn(t)
	store := filepath.Join(t.TempDir(), "cairn.db")
	if ws.Shows(store) || ws.Raise(store) {
		t.Fatal("a store nobody claimed is shown")
	}
	var raised atomic.Int32
	claim, err := ws.Claim(store, func() { raised.Add(1) })
	if err != nil {
		t.Fatalf("Claim = %v", err)
	}
	if !ws.Shows(store) {
		t.Error("a store claimed is not shown")
	}
	if !ws.Raise(store) || raised.Load() != 1 {
		t.Errorf("Raise reached the window %d times; want once", raised.Load())
	}
	if ws.Shows(filepath.Join(t.TempDir(), "cairn.db")) {
		t.Error("another store is shown")
	}
	_ = claim.Close()
	if ws.Shows(store) {
		t.Error("a store let go of is still shown")
	}
}

func TestWindows_aSecondClaimBringsTheFirstForwardAndIsRefused(t *testing.T) {
	ws := windowsIn(t)
	store := filepath.Join(t.TempDir(), "cairn.db")
	var raised atomic.Int32
	claim, err := ws.Claim(store, func() { raised.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = claim.Close() }()
	if _, err := ws.Claim(store, func() {}); !errors.Is(err, ErrShownElsewhere) {
		t.Errorf("a second Claim = %v; want %v", err, ErrShownElsewhere)
	}
	if raised.Load() != 1 {
		t.Errorf("the first window was raised %d times; want once", raised.Load())
	}
}

func TestWindows_whatIsLeftOfAWindowGoneIsClaimedAgain(t *testing.T) {
	ws := windowsIn(t)
	store := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(ws.socket(store), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	claim, err := ws.Claim(store, func() {})
	if err != nil {
		t.Fatalf("Claim over what a window left = %v", err)
	}
	_ = claim.Close()
}

func TestWindowsHere_isTheRuntimeFolderWhenItIsAbsolute(t *testing.T) {
	given := func(dir string) func(string) (string, bool) {
		return func(name string) (string, bool) { return dir, name == "XDG_RUNTIME_DIR" && dir != "" }
	}
	if got := WindowsHere(given("/run/user/1")).Dir; got != "/run/user/1" {
		t.Errorf("WindowsHere = %q", got)
	}
	if got := WindowsHere(given("relative")).Dir; got != os.TempDir() {
		t.Errorf("WindowsHere with a relative folder = %q, want %q", got, os.TempDir())
	}
}

func TestOpeningApart_bringsForwardTheWindowShowingItOrStartsACairn(t *testing.T) {
	ws := windowsIn(t)
	shown, other := filepath.Join(t.TempDir(), "cairn.db"), filepath.Join(t.TempDir(), "cairn.db")
	var raised atomic.Int32
	claim, err := ws.Claim(shown, func() { raised.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = claim.Close() }()
	var started [][]string
	start := func(args ...string) error { started = append(started, args); return nil }
	reach := func(to engine.Elsewhere) (workspace.Apart, error) {
		if to.Workspace == "gone" {
			return workspace.Apart{}, workspace.ErrWorkspaceNotThere
		}
		if to.Workspace == "dev" {
			return workspace.Apart{Store: shown}, nil
		}
		return workspace.Apart{Store: other, Flags: []string{"--workspace", to.Workspace}}, nil
	}
	if went := OpeningApart(engine.Elsewhere{Workspace: "dev"}, reach, ws, start); !went.Raised || went.Err != nil || len(started) != 0 || raised.Load() != 1 {
		t.Errorf("the one shown: %+v, started %v, raised %d", went, started, raised.Load())
	}
	if went := OpeningApart(engine.Elsewhere{Workspace: "home"}, reach, ws, start); went.Raised || went.Err != nil || len(started) != 1 || !slices.Equal(started[0], []string{"gui", "--workspace", "home"}) {
		t.Errorf("another: %+v, started %v", went, started)
	}
	if went := OpeningApart(engine.Elsewhere{Workspace: "gone"}, reach, ws, start); !errors.Is(went.Err, workspace.ErrWorkspaceNotThere) || len(started) != 1 {
		t.Errorf("one not there: %+v, started %v", went, started)
	}
}

func TestTyped_commandAndShiftTogetherAreSaidTogether(t *testing.T) {
	for name, want := range map[string]string{"O": "cmd+shift+o", "2": "cmd+shift+2", "@": "cmd+shift+2", "(": "cmd+shift+9"} {
		if got := typedKey(name, key.ModCommand|key.ModShift); got != want {
			t.Errorf("⌘⇧%s reads %q, want %q", name, got, want)
		}
	}
	if got := typedKey("O", key.ModCommand); got != "cmd+o" {
		t.Errorf("⌘O reads %q", got)
	}
	if got := Pretty("cmd+shift+o"); got != "⌘⇧O" {
		t.Errorf("Pretty = %q", got)
	}
	if got := macKey("Open a workspace…"); got != "O" {
		t.Errorf("macKey = %q, want O, which AppKit reads as ⌘⇧O", got)
	}
}

func typedKey(name string, with key.Modifiers) string {
	said, _ := keyOf(name, with)
	return said
}

func atPlace(f *fake, shown workspace.Shown, cards ...board.Card) Window {
	f.shown = &shown
	return holding(f, cards...).Look()
}

func TestTheWindowIsTitledAfterItsWorkspace(t *testing.T) {
	if got := atPlace(newFake(), workspace.Shown{Workspace: "work", Store: "/w/cairn.db"}).Title(); got != "cairn — work" {
		t.Errorf("Title = %q", got)
	}
	if got := atPlace(newFake(), workspace.Shown{Store: "/scratch/cairn.db"}).Title(); got != "cairn — scratch" {
		t.Errorf("Title of a loose store = %q", got)
	}
	if got := holding(newFake()).Title(); got != "cairn" {
		t.Errorf("Title of a window that knows no place = %q", got)
	}
}

func switching(f *fake, seeing func(string) bool) Window {
	f.listed = &workspace.Workspaces{
		Workspaces: []workspace.Workspace{
			{Name: "work", Folder: "/w", Default: true},
			{Name: "home", Folder: "/h"},
			{Name: "dev", Folder: "/d"},
		},
		Recent: []workspace.RecentFolder{{Store: "/s/cairn.db", Folder: "/s", Opened: time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)}},
	}
	return atPlace(f, workspace.Shown{Workspace: "work", Store: "/w/cairn.db"}).At(time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)).Seeing(seeing).Key("cmd+shift+o")
}

func TestTheSwitcherSaysWhatChoosingEachDoes(t *testing.T) {
	w := switching(newFake(), func(store string) bool { return store == "/d/cairn.db" })
	if w.Screen != board.OnWorkspaces {
		t.Fatalf("⌘⇧O leaves the window on %v", w.Screen)
	}
	want := map[string][]string{
		"work": {"this window · default", "⌘⇧1"},
		"home": {"opens a new window", "⌘⇧2"},
		"dev":  {"open · brings it forward", "⌘⇧3"},
	}
	for _, line := range w.SwitcherLines() {
		if line.Recent {
			if line.Tag != "opened yesterday" {
				t.Errorf("the recent folder is tagged %q", line.Tag)
			}
			continue
		}
		if tags := want[line.Name]; line.Tag != tags[0] || line.Key != tags[1] {
			t.Errorf("%s is tagged %q %q, want %q", line.Name, line.Tag, line.Key, tags)
		}
	}
	if got := w.Types("ho").SwitcherLines(); len(got) != 1 || got[0].Name != "home" {
		t.Errorf("typing ho lists %+v", got)
	}
}

func TestChoosingAnotherAsksForAWindowOfItsOwn(t *testing.T) {
	w := switching(newFake(), func(string) bool { return false }).Key("down").Key("enter")
	to, asked := w.ApartAsked()
	if !asked || to.Workspace != "home" {
		t.Fatalf("ApartAsked = %+v, %v; want home", to, asked)
	}
	if w = w.ApartTaken(); w.Went(Went{}).Screen != board.OnTheBoard {
		t.Errorf("the switcher stays open on %v once the window is open", w.Went(Went{}).Screen)
	}
	if _, asked := w.ApartAsked(); asked {
		t.Error("the window asks twice")
	}
	if said := w.Went(Went{}).Telling; said != "home opens in a window of its own" {
		t.Errorf("the window says %q", said)
	}
	if said := w.Went(Went{Raised: true}).Telling; said != "home is brought forward" {
		t.Errorf("the window says %q", said)
	}
	if got := w.Went(Went{Err: workspace.ErrWorkspaceNotThere}).Wrong(); !errors.Is(got, workspace.ErrWorkspaceNotThere) {
		t.Errorf("a refusal is said as %v", got)
	}
	here := switching(newFake(), nil).Key("enter")
	if _, asked := here.ApartAsked(); asked || here.Screen != board.OnTheBoard {
		t.Errorf("choosing this window asks for one and leaves the window on %v", here.Screen)
	}
	if to, asked := switching(newFake(), nil).Key("esc").Key("cmd+shift+3").ApartAsked(); !asked || to.Workspace != "dev" {
		t.Errorf("⌘⇧3 asks for %+v, %v", to, asked)
	}
}

func TestTheMenuBarOffersTheSwitcher(t *testing.T) {
	w := atPlace(newFake(), workspace.Shown{Workspace: "work", Store: "/w/cairn.db"}).Picked("Open a workspace…")
	if w.Screen != board.OnWorkspaces {
		t.Errorf("Open a workspace… leaves the window on %v", w.Screen)
	}
}

func TestARecentFolderIsKeptAsAWorkspaceThroughWorkspaceAdd(t *testing.T) {
	f := newFake()
	w := switching(f, nil)
	if acts := SwitcherActs(w); slices.ContainsFunc(acts, func(a SwitcherAct) bool { return a.Said == KeepAsWorkspaceSays }) {
		t.Errorf("on a workspace, the switcher offers %+v", acts)
	}
	keeping := func(w Window) Window {
		t.Helper()
		w = w.Key("down").Key("down").Key("down")
		for _, act := range SwitcherActs(w) {
			if act.Said == KeepAsWorkspaceSays {
				return w.clicked(geom.Point{X: act.At.X + act.At.W/2, Y: act.At.Y + act.At.H/2})
			}
		}
		t.Fatalf("on a recent folder, the switcher offers %+v", SwitcherActs(w))
		return w
	}
	w = keeping(w)
	if asks := w.SwitcherAsks(); asks != "Keep /s as" {
		t.Errorf("the switcher asks %q", asks)
	}
	w = w.Types("scratch").Key("enter")
	in, asked := f.gave("workspace", "add")
	if !asked || in.Given["name"] != "scratch" || in.Given["folder"] != "/s" {
		t.Fatalf("workspace add given %+v, %v", in, asked)
	}
	if w.Screen != board.OnWorkspaces || w.SwitcherAsks() != "" {
		t.Errorf("once kept, the window is on %v asking %q", w.Screen, w.SwitcherAsks())
	}
	if _, apart := w.ApartAsked(); apart {
		t.Error("keeping a folder asks for a window")
	}
	f.failOp = errRefuse
	refused := keeping(switching(f, nil)).Types("x").Key("enter")
	if !errors.Is(refused.Wrong(), errRefuse) {
		t.Errorf("a refusal is said as %v", refused.Wrong())
	}
}

func TestAStoreIsShownByAShortNameItsPathUnderThePointer(t *testing.T) {
	w := switching(newFake(), nil)
	lines := w.SwitcherLines()
	recent := slices.IndexFunc(lines, func(l SwitcherLine) bool { return l.Recent })
	home := slices.IndexFunc(lines, func(l SwitcherLine) bool { return l.Name == "home" })
	if recent < 0 || home < 0 || lines[recent].Name != "s" {
		t.Fatalf("the switcher reads %+v, want the recent folder named s", lines)
	}
	for at, path := range map[int]string{recent: "/s", home: "/h"} {
		box := lines[at].At
		if said := w.Hovering(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2}).Bubble(); said != path {
			t.Errorf("the pointer on %q says %q, want %q", lines[at].Name, said, path)
		}
	}
	loose := atPlace(newFake(), workspace.Shown{Store: "/Users/me/projects/proj/cairn.db"})
	if said := loose.PlaceShown(); said != "proj" {
		t.Errorf("the pill of a store no workspace names reads %q, want proj", said)
	}
}
