package tui

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"cairn/internal/engine"
)

type roaming struct {
	*fake
	shown    map[string]engine.WorkspaceShown
	spaces   []engine.Workspace
	on       string
	refuse   error
	pages    map[string]engine.Pages
	switched []string
	recent   []engine.RecentFolder
	added    []string
}

func newRoaming(t *testing.T, f *fake) *roaming {
	root := t.TempDir()
	for _, folder := range []string{"work", "home", "client-a"} {
		if err := os.Mkdir(filepath.Join(root, folder), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	r := &roaming{
		fake: f,
		on:   "work",
		shown: map[string]engine.WorkspaceShown{
			"work": {Workspace: "work", Folder: "/c/work", Store: "/c/work/cairn.db"},
			"home": {Workspace: "home", Folder: "/c/home", Store: "/c/home/cairn.db"},
		},
		spaces: []engine.Workspace{
			{Name: "work", Folder: "/c/work", Default: true},
			{Name: "home", Folder: "/c/home"},
			{Name: "client-a", Folder: "/c/client-a"},
		},
		pages: map[string]engine.Pages{},
	}
	for i := range r.spaces {
		r.spaces[i].Folder = filepath.Join(root, r.spaces[i].Name)
	}
	return r
}

func (r *roaming) Ask(op engine.Operation, in engine.Input) (engine.Answer, error) {
	switch op.Thing + " " + op.Verb {
	case "workspace show":
		shown := r.shown[r.on]
		return engine.Answer{WorkspaceShown: &shown}, nil
	case "workspace list":
		return engine.Answer{Workspaces: &engine.Workspaces{Workspaces: r.spaces, Recent: r.recent}}, nil
	case "workspace add":
		name, folder := in.Given["name"], in.Given["folder"]
		for _, one := range r.spaces {
			if one.Name == name {
				return engine.Answer{}, engine.ErrWorkspaceKnown
			}
		}
		r.added = append(r.added, name+" "+folder)
		r.spaces = append(r.spaces, engine.Workspace{Name: name, Folder: folder})
		r.recent = slices.DeleteFunc(r.recent, func(one engine.RecentFolder) bool { return one.Folder == folder })
		r.shown[name] = engine.WorkspaceShown{Workspace: name, Folder: folder, Store: folder + "/cairn.db"}
		return engine.Answer{}, nil
	}
	return r.fake.Ask(op, in)
}

func (r *roaming) Switch(to engine.Elsewhere) (engine.Pages, error) {
	if r.refuse != nil {
		return engine.Pages{}, r.refuse
	}
	name := to.Workspace + to.Folder + to.Store
	r.switched = append(r.switched, name)
	r.on = name
	r.cards = nil
	r.focus = nil
	return r.pages[r.on], nil
}

func roamingBoard(t *testing.T, r *roaming) Board {
	t.Helper()
	b := standing(r.fake)
	b.work = r
	b = b.Switching(r.Switch)
	return run(b, b.placing())
}

func run(b Board, cmd tea.Cmd) Board {
	for depth := 0; cmd != nil && depth < 4; depth++ {
		msg := cmd()
		if msg == nil {
			return b
		}
		model, next := b.Update(msg)
		b, cmd = model.(Board), next
	}
	return b
}

func pressing(b Board, keys ...string) Board {
	for _, k := range keys {
		model, cmd := b.Update(keyPress(k))
		b = run(model.(Board), cmd)
	}
	return b
}

func headerOf(b Board) string { return strings.TrimSpace(strings.SplitN(strip(b.Draw()), "\n", 2)[0]) }

func TestHeader_opensWithTheWorkspace(t *testing.T) {
	b := roamingBoard(t, newRoaming(t, newFake()))
	if got := headerOf(b); !strings.HasPrefix(got, "● work") {
		t.Errorf("header = %q; want it to open with ● work", got)
	}
}

func TestHeader_aStoreNoWorkspaceNamesShowsItsFolder(t *testing.T) {
	r := newRoaming(t, newFake())
	r.shown["work"] = engine.WorkspaceShown{Store: "/scratch/cairn.db"}
	if got := headerOf(roamingBoard(t, r)); !strings.HasPrefix(got, "◌ /scratch ") {
		t.Errorf("header = %q; want it to open with ◌ /scratch", got)
	}
}

func TestHeader_aLongFolderKeepsItsEnd(t *testing.T) {
	r := newRoaming(t, newFake())
	r.shown["work"] = engine.WorkspaceShown{Store: "/" + strings.Repeat("deep/", 20) + "scratch/cairn.db"}
	got := headerOf(roamingBoard(t, r))
	if !strings.HasPrefix(got, "◌ …") || !strings.Contains(got, "/scratch ") {
		t.Errorf("header = %q; want the folder cut at its start, its end kept", got)
	}
}

func TestHeader_saysNothingOfAPlaceItWasNotTold(t *testing.T) {
	b := New(newFake(), GlyphsFor(Unicode), func() time.Time { return today }, 100, 40)
	if got := headerOf(b); strings.ContainsAny(got, "●◌") {
		t.Errorf("header = %q; want no workspace said", got)
	}
}

func TestCtrlO_listsTheWorkspacesTheOneHereMarked(t *testing.T) {
	b := pressing(roamingBoard(t, newRoaming(t, newFake())), "ctrl+o")
	screen := strip(b.Draw())
	if b.screen != OnWorkspaces || !strings.Contains(screen, "OPEN A WORKSPACE") {
		t.Fatalf("ctrl+o shows:\n%s", screen)
	}
	for _, want := range []string{"WORKSPACES", "here · default", "home", "esc back"} {
		if !strings.Contains(screen, want) {
			t.Errorf("the list does not say %q:\n%s", want, screen)
		}
	}
	if got := lineOf(screen, "● home "); !strings.HasSuffix(strings.TrimRight(got, " │"), "1") {
		t.Errorf("home's line = %q; want it numbered 1, the one here left unnumbered", got)
	}
	if got := lineOf(screen, "● client-a "); !strings.HasSuffix(strings.TrimRight(got, " │"), "2") {
		t.Errorf("client-a's line = %q; want it numbered 2", got)
	}
	if got := lineOf(screen, "● home "); !strings.Contains(got, "›") {
		t.Errorf("home's line = %q; want the cursor on the first one not here", got)
	}
}

func lineOf(screen, holding string) string {
	for _, line := range strings.Split(screen, "\n") {
		if strings.Contains(line, holding) {
			return line
		}
	}
	return ""
}

func TestCtrlO_isNotOfferedByABoardThatCannotSwitch(t *testing.T) {
	b := New(newFake(), GlyphsFor(Unicode), func() time.Time { return today }, 160, 40)
	b = pressing(b, "ctrl+o")
	if b.screen != OnTheBoard || strings.Contains(strip(b.Draw()), "workspace") {
		t.Errorf("a board that cannot switch offers it:\n%s", strip(b.Draw()))
	}
	r := roamingBoard(t, newRoaming(t, newFake()))
	r.width = 160
	if !strings.Contains(strip(r.Draw()), "^O workspace") {
		t.Errorf("the bar does not offer ^O:\n%s", strip(r.Draw()))
	}
}

func TestSwitcher_narrowsAsOneTypes(t *testing.T) {
	b := pressing(roamingBoard(t, newRoaming(t, newFake())), "ctrl+o", "h", "o")
	if got := names(b.listed()); !slices.Equal(got, []string{"home"}) {
		t.Errorf("listed = %v; want home alone", got)
	}
	b = pressing(b, "backspace", "backspace")
	if got := names(b.listed()); len(got) != 3 {
		t.Errorf("listed = %v; want every one again", got)
	}
	b = pressing(b, "j", "k")
	if got := names(b.listed()); len(got) != 0 || !strings.Contains(strip(b.Draw()), "jk") {
		t.Errorf("listed = %v; want j and k typed, matching none", got)
	}
}

func names(spaces []engine.Workspace) []string {
	var out []string
	for _, one := range spaces {
		out = append(out, one.Name)
	}
	return out
}

func TestSwitcher_arrowsChooseAndEnterOpens(t *testing.T) {
	r := newRoaming(t, newFake())
	b := pressing(roamingBoard(t, r), "ctrl+o", "down", "down", "up", "down", "enter")
	if !slices.Equal(r.switched, []string{"client-a"}) {
		t.Fatalf("switched = %v; want client-a", r.switched)
	}
	if b.screen != OnTheBoard {
		t.Errorf("screen = %v; want the board", b.screen)
	}
}

func TestSwitcher_enterOnTheOneHereOnlyCloses(t *testing.T) {
	r := newRoaming(t, newFake())
	b := pressing(roamingBoard(t, r), "ctrl+o", "up", "up", "enter")
	if len(r.switched) != 0 || b.screen != OnTheBoard {
		t.Errorf("switched = %v, screen %v; want nothing opened, the board back", r.switched, b.screen)
	}
}

func TestSwitcher_aDigitOpensByPlace(t *testing.T) {
	r := newRoaming(t, newFake())
	pressing(roamingBoard(t, r), "ctrl+o", "2")
	if !slices.Equal(r.switched, []string{"client-a"}) {
		t.Errorf("switched = %v; want client-a, second after the one here", r.switched)
	}
	r = newRoaming(t, newFake())
	pressing(roamingBoard(t, r), "ctrl+o", "7")
	if len(r.switched) != 0 {
		t.Errorf("switched = %v; want nothing at a place nobody holds", r.switched)
	}
}

func TestSwitcher_escLeavesTheBoardAsItWas(t *testing.T) {
	f := newFake()
	f.cards = []engine.Card{{Number: 1, Title: "Rotate", State: "backlog", Priority: "normal"}}
	r := newRoaming(t, f)
	b := pressing(roamingBoard(t, r), "ctrl+o", "h", "esc")
	if b.screen != OnTheBoard || len(r.switched) != 0 || !strings.Contains(strip(b.Draw()), "Rotate") {
		t.Errorf("esc left:\n%s", strip(b.Draw()))
	}
	if b = pressing(b, "ctrl+o"); len(b.listed()) != 3 {
		t.Errorf("opened again, listed = %v; want what was typed forgotten", names(b.listed()))
	}
}

func TestSwitched_replacesTheBoardAndSaysSo(t *testing.T) {
	f := newFake()
	f.cards = []engine.Card{{Number: 152, Title: "Choose the month", State: "current", Priority: "normal", InFocus: true}}
	f.focus = []int{152}
	r := newRoaming(t, f)
	b := pressing(roamingBoard(t, r), "l", "l", "i", "ctrl+o", "1")
	screen := strip(b.Draw())
	if !strings.HasPrefix(headerOf(b), "● home") {
		t.Errorf("header = %q; want home", headerOf(b))
	}
	if !strings.Contains(screen, "opened home in place of work · work's focus on #152 is still held") {
		t.Errorf("the board does not say what it opened:\n%s", screen)
	}
	if strings.Contains(screen, "Choose the month") || b.mode != Normal || b.task != 0 {
		t.Errorf("work's board is still there:\n%s", screen)
	}
	if b = pressing(b, "l"); strings.Contains(strip(b.Draw()), "opened home") {
		t.Errorf("the line stays after a key:\n%s", strip(b.Draw()))
	}
}

func TestSwitched_withNoFocusSaysOnlyWhatItOpened(t *testing.T) {
	b := pressing(roamingBoard(t, newRoaming(t, newFake())), "ctrl+o", "1")
	screen := strip(b.Draw())
	if !strings.Contains(screen, "opened home in place of work") || strings.Contains(screen, "still held") {
		t.Errorf("the board says:\n%s", screen)
	}
}

func TestSwitched_takesThePagesOfTheOneOpened(t *testing.T) {
	r := newRoaming(t, newFake())
	r.pages["home"] = engine.PagesOff(engine.PageTime)
	b := roamingBoard(t, r)
	b = pressing(b, "tab", "tab")
	if b.tab != Time {
		t.Fatalf("tab = %v; want time", b.tab)
	}
	b = pressing(b, "ctrl+o", "1")
	if b.tab != Tasks || slices.Contains(b.shown(), Time) {
		t.Errorf("tabs = %v on %v; want no time tab", b.shown(), b.tab)
	}
}

func TestSwitched_refusedSaysWhyAndStaysOnTheList(t *testing.T) {
	r := newRoaming(t, newFake())
	r.refuse = engine.ErrWorkspaceNotThere
	b := pressing(roamingBoard(t, r), "ctrl+o", "1")
	screen := strip(b.Draw())
	if b.screen != OnWorkspaces || !strings.Contains(screen, engine.ErrWorkspaceNotThere.Error()) {
		t.Errorf("a refusal shows:\n%s", screen)
	}
	if !strings.HasPrefix(headerOf(b), "● work") {
		t.Errorf("header = %q; want work still", headerOf(b))
	}
}

func TestSwitcher_aWorkspaceWhoseFolderIsGoneIsSaidNotFound(t *testing.T) {
	r := newRoaming(t, newFake())
	r.spaces[1].Folder = "/nowhere/at/all"
	r.spaces[0].Folder = t.TempDir()
	screen := strip(pressing(roamingBoard(t, r), "ctrl+o").Draw())
	if got := lineOf(screen, "● home "); !strings.Contains(got, "not found") {
		t.Errorf("home's line = %q; want it said not found", got)
	}
	if got := lineOf(screen, "● work "); strings.Contains(got, "not found") {
		t.Errorf("work's line = %q; want it found", got)
	}
}

func TestSwitcher_aListThatCannotBeReadIsSaid(t *testing.T) {
	r := newRoaming(t, newFake())
	b := roamingBoard(t, r)
	model, _ := b.Update(Listed{Err: errNope})
	if b = model.(Board); b.screen != OnTheBoard || !errors.Is(b.err, errNope) {
		t.Errorf("screen %v, err %v; want the board saying why", b.screen, b.err)
	}
}

func TestSwitcher_aBoardThatCannotSwitchIsRefused(t *testing.T) {
	b := New(newFake(), GlyphsFor(Unicode), func() time.Time { return today }, 100, 40)
	msg := b.switchingTo(engine.Elsewhere{Workspace: "home"})()
	if switched, ok := msg.(Switched); !ok || !errors.Is(switched.Err, errNoSwitching) {
		t.Errorf("switching = %#v; want it refused", msg)
	}
}

func TestSwitcher_fitsANarrowWindow(t *testing.T) {
	r := newRoaming(t, newFake())
	r.spaces[2].Folder = "/" + strings.Repeat("deep/", 30)
	b := roamingBoard(t, r)
	b.width, b.height = Narrowest, Shortest
	b = pressing(b, "ctrl+o")
	if err := fits(b); err != nil {
		t.Error(err)
	}
}

func fits(b Board) error {
	lines := strings.Split(strings.TrimRight(b.Draw(), "\n"), "\n")
	if len(lines) > b.height {
		return errors.New("the list runs past the bottom of the window:\n" + strip(b.Draw()))
	}
	for _, line := range lines {
		if lipgloss.Width(line) > b.width {
			return errors.New("a line runs past the window: " + strip(line))
		}
	}
	return nil
}
