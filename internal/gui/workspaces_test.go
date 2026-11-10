package gui

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
	reach := func(to engine.Elsewhere) (engine.Apart, error) {
		if to.Workspace == "gone" {
			return engine.Apart{}, engine.ErrWorkspaceNotThere
		}
		if to.Workspace == "dev" {
			return engine.Apart{Store: shown}, nil
		}
		return engine.Apart{Store: other, Flags: []string{"--workspace", to.Workspace}}, nil
	}
	if went := OpeningApart(engine.Elsewhere{Workspace: "dev"}, reach, ws, start); !went.Raised || went.Err != nil || len(started) != 0 || raised.Load() != 1 {
		t.Errorf("the one shown: %+v, started %v, raised %d", went, started, raised.Load())
	}
	if went := OpeningApart(engine.Elsewhere{Workspace: "home"}, reach, ws, start); went.Raised || went.Err != nil || len(started) != 1 || !slices.Equal(started[0], []string{"gui", "--workspace", "home"}) {
		t.Errorf("another: %+v, started %v", went, started)
	}
	if went := OpeningApart(engine.Elsewhere{Workspace: "gone"}, reach, ws, start); !errors.Is(went.Err, engine.ErrWorkspaceNotThere) || len(started) != 1 {
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

func atPlace(f *fake, shown engine.WorkspaceShown, cards ...engine.Card) Window {
	f.shown = &shown
	return holding(f, cards...).Look()
}

func TestTheWindowIsTitledAfterItsWorkspace(t *testing.T) {
	if got := atPlace(newFake(), engine.WorkspaceShown{Workspace: "work", Store: "/w/cairn.db"}).Title(); got != "cairn — work" {
		t.Errorf("Title = %q", got)
	}
	if got := atPlace(newFake(), engine.WorkspaceShown{Store: "/scratch/cairn.db"}).Title(); got != "cairn — /scratch" {
		t.Errorf("Title of a loose store = %q", got)
	}
	if got := holding(newFake()).Title(); got != "cairn" {
		t.Errorf("Title of a window that knows no place = %q", got)
	}
}

func TestThePillOpensTheTopBarAndTheTabsComeAfterIt(t *testing.T) {
	w := atPlace(newFake(), engine.WorkspaceShown{Workspace: "work", Store: "/w/cairn.db", Colour: "teal"})
	l := w.Layout()
	if l.Place.W == 0 || l.Place.X != margin {
		t.Fatalf("the pill is at %+v", l.Place)
	}
	if first := TabAt(l, 0); first.X < l.Place.X+l.Place.W {
		t.Errorf("the first tab at %d stands on the pill ending at %d", first.X, l.Place.X+l.Place.W)
	}
	if end := l.Tabs.X + l.Tabs.W; end != LayoutFor(1200, 800).Tabs.X+LayoutFor(1200, 800).Tabs.W {
		t.Errorf("the top bar ends at %d, not where it ended", end)
	}
	if bare := holding(newFake()).Layout(); bare.Place.W != 0 || bare.Tabs != LayoutFor(1200, 800).Tabs {
		t.Errorf("a window that knows no place moved its tabs: %+v", bare.Tabs)
	}
	clicked := w.clicked(Point{X: l.Place.X + l.Place.W/2, Y: l.Place.Y + l.Place.H/2})
	if clicked.Screen() != OnWorkspaces {
		t.Errorf("a click on the pill leaves the window on %v", clicked.Screen())
	}
}

func switching(f *fake, seeing func(string) bool) Window {
	f.listed = &engine.Workspaces{
		Workspaces: []engine.Workspace{
			{Name: "work", Folder: "/w", Default: true},
			{Name: "home", Folder: "/h"},
			{Name: "dev", Folder: "/d"},
		},
		Recent: []engine.RecentFolder{{Store: "/s/cairn.db", Folder: "/s", Opened: time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)}},
	}
	return atPlace(f, engine.WorkspaceShown{Workspace: "work", Store: "/w/cairn.db"}).At(time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)).Seeing(seeing).Key("cmd+shift+o")
}

func TestTheSwitcherSaysWhatChoosingEachDoes(t *testing.T) {
	w := switching(newFake(), func(store string) bool { return store == "/d/cairn.db" })
	if w.Screen() != OnWorkspaces {
		t.Fatalf("⌘⇧O leaves the window on %v", w.Screen())
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
	if w = w.ApartTaken(); w.Went(Went{}).Screen() != OnTheBoard {
		t.Errorf("the switcher stays open on %v once the window is open", w.Went(Went{}).Screen())
	}
	if _, asked := w.ApartAsked(); asked {
		t.Error("the window asks twice")
	}
	if said := w.Went(Went{}).told; said != "home opens in a window of its own" {
		t.Errorf("the window says %q", said)
	}
	if said := w.Went(Went{Raised: true}).told; said != "home is brought forward" {
		t.Errorf("the window says %q", said)
	}
	if got := w.Went(Went{Err: engine.ErrWorkspaceNotThere}).Wrong(); !errors.Is(got, engine.ErrWorkspaceNotThere) {
		t.Errorf("a refusal is said as %v", got)
	}
	here := switching(newFake(), nil).Key("enter")
	if _, asked := here.ApartAsked(); asked || here.Screen() != OnTheBoard {
		t.Errorf("choosing this window asks for one and leaves the window on %v", here.Screen())
	}
	if to, asked := switching(newFake(), nil).Key("esc").Key("cmd+shift+3").ApartAsked(); !asked || to.Workspace != "dev" {
		t.Errorf("⌘⇧3 asks for %+v, %v", to, asked)
	}
}

func TestTheMenuBarOffersTheSwitcher(t *testing.T) {
	w := atPlace(newFake(), engine.WorkspaceShown{Workspace: "work", Store: "/w/cairn.db"}).Picked("Open a workspace…")
	if w.Screen() != OnWorkspaces {
		t.Errorf("Open a workspace… leaves the window on %v", w.Screen())
	}
}

func TestTheSettingsOfAWorkspaceSayWhereEachKeyComesFrom(t *testing.T) {
	f := newFake()
	f.config = &engine.Config{
		File: "/c/config.toml", Settings: "/h/settings.toml", Store: engine.Emplacement{At: "/h/cairn.db"},
		Pages: &engine.PagesConfig{Tasks: true, Objectives: true, Time: "off"}, Archive: &engine.Archive{Days: 30},
		Focus: &engine.FocusConfig{}, Time: &engine.TimeConfig{Deleted: "kept", Weekends: "shown"}, Window: &engine.WindowConfig{},
		Written: []string{"archive.days", "pages.time"}, WrittenHere: []string{"pages.time"},
	}
	w := atPlace(f, engine.WorkspaceShown{Workspace: "home", Store: "/h/cairn.db"}).Key("cmd+,")
	if got := Heads(w); got != "Settings of home" {
		t.Errorf("the panel is headed %q", got)
	}
	from := map[string]string{}
	for _, row := range w.SettingRows() {
		from[row.Key.Name] = row.Comes
	}
	if from["pages.time"] != SetHere || from["archive.days"] != FromGlobal || from["time.weekends"] != ByDefault {
		t.Errorf("the keys come from %v", from)
	}
	if said := w.SettingsPath(); !strings.Contains(said, "/h/settings.toml") || !strings.Contains(said, "over /c/config.toml") {
		t.Errorf("the scope says %q", said)
	}
	if !slices.Contains(w.Offers(), OpenSettings) {
		t.Errorf("the panel offers %v", w.Offers())
	}
	if got, _ := w.Choose(OpenSettings).OpenAsked(); got != "/h/settings.toml" {
		t.Errorf("Open settings.toml asks the system to open %q", got)
	}
}

func TestACardSaysNothingOfItsTimeWhenTheTimeIsNotShown(t *testing.T) {
	now := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	c := engine.Card{Estimate: "2h", Spent: "1h", Objectives: []int{4}, Serves: []string{"Ops"}}
	if said, _ := CardFacts(c, now, engine.PagesOff(engine.PageTime)); strings.Contains(said, "1h") || !strings.Contains(said, "Ops") {
		t.Errorf("with the time off a card says %q", said)
	}
	if said, _ := CardFacts(c, now, engine.Pages{}); !strings.Contains(said, "1h/2h") {
		t.Errorf("with the time on a card says %q", said)
	}
}

func TestEveryWorkspaceIsListedWhenNoFolderIsRecent(t *testing.T) {
	f := newFake()
	w := switching(f, nil)
	f.listed.Recent = nil
	w = w.Key("esc").Key("cmd+shift+o")
	if got := w.SwitcherLines(); len(got) != 3 {
		t.Errorf("the switcher lists %d of 3 workspaces: %+v", len(got), got)
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
				return w.clicked(Point{X: act.At.X + act.At.W/2, Y: act.At.Y + act.At.H/2})
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
	if w.Screen() != OnWorkspaces || w.SwitcherAsks() != "" {
		t.Errorf("once kept, the window is on %v asking %q", w.Screen(), w.SwitcherAsks())
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

func TestTheSettingsOfAWorkspaceSayItsLabelItsColourAndTheGlobalValueOfAKeySetHere(t *testing.T) {
	f := newFake()
	f.config = &engine.Config{
		File: "/c/config.toml", Settings: "/h/settings.toml", Store: engine.Emplacement{At: "/h/cairn.db"},
		Pages: &engine.PagesConfig{Tasks: true, Objectives: true, Time: "off"}, Archive: &engine.Archive{Days: 30},
		Focus: &engine.FocusConfig{}, Approach: &engine.ApproachConfig{Label: "discuss"}, Time: &engine.TimeConfig{Deleted: "kept", Weekends: "shown"}, Window: &engine.WindowConfig{},
		Workspace: &engine.WorkspaceConfig{Colour: "teal"},
		Written:   []string{"archive.days", "pages.time"}, WrittenHere: []string{"pages.time", "workspace.colour"},
		Global: map[string]string{"pages.time": "on"},
	}
	w := atPlace(f, engine.WorkspaceShown{Workspace: "home", Store: "/h/cairn.db"}).Key("cmd+,")
	rows := map[string]SettingRow{}
	for _, row := range w.SettingRows() {
		rows[row.Key.Name] = row
	}
	if got := rows["pages.time"]; got.Aside != "← global: on" || rows["archive.days"].Aside != "" {
		t.Errorf("pages.time says %q beside it, archive.days %q", got.Aside, rows["archive.days"].Aside)
	}
	if got := rows["approach.label"]; got.Value != "discuss" || got.Comes != ByDefault {
		t.Errorf("approach.label reads %+v", got)
	}
	if colour := rows["workspace.colour"]; colour.Value != "teal" || colour.Comes != SetHere || colour.Aside != "" {
		t.Errorf("the colour reads %+v", colour)
	}
	f.config.Workspace, f.config.Approach.Label = nil, ""
	w = atPlace(f, engine.WorkspaceShown{Store: "/s/cairn.db"}).Key("cmd+,")
	for _, row := range w.SettingRows() {
		if row.Key.Name == "approach.label" && row.Value != "" {
			t.Errorf("a store saying no label reads %+v", row)
		}
	}
}
