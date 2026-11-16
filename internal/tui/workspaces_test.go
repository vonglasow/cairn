package tui

import (
	"cairn/internal/board"
	"cairn/internal/engine/workspace"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
)

type roaming struct {
	*fake
	shown    map[string]workspace.Shown
	spaces   []workspace.Workspace
	on       string
	refuse   error
	pages    map[string]engine.Pages
	switched []string
	recent   []workspace.RecentFolder
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
		shown: map[string]workspace.Shown{
			"work": {Workspace: "work", Folder: "/c/work", Store: "/c/work/cairn.db"},
			"home": {Workspace: "home", Folder: "/c/home", Store: "/c/home/cairn.db"},
		},
		spaces: []workspace.Workspace{
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
		return engine.Answer{Workspaces: &workspace.Workspaces{Workspaces: r.spaces, Recent: r.recent}}, nil
	case "workspace add":
		name, folder := in.Given["name"], in.Given["folder"]
		for _, one := range r.spaces {
			if one.Name == name {
				return engine.Answer{}, workspace.ErrWorkspaceKnown
			}
		}
		r.added = append(r.added, strings.Join(strings.Fields(name+" "+folder+" "+in.Given["notes"]+" "+in.Given["colour"]), " "))
		r.spaces = append(r.spaces, workspace.Workspace{Name: name, Folder: folder})
		r.recent = slices.DeleteFunc(r.recent, func(one workspace.RecentFolder) bool { return one.Folder == folder })
		r.shown[name] = workspace.Shown{Workspace: name, Folder: folder, Store: folder + "/cairn.db"}
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

func lineOf(screen, holding string) string {
	for _, line := range strings.Split(screen, "\n") {
		if strings.Contains(line, holding) {
			return line
		}
	}
	return ""
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

func names(spaces []workspace.Workspace) []string {
	var out []string
	for _, one := range spaces {
		out = append(out, one.Name)
	}
	return out
}

func TestSwitcher_enterOnTheOneHereOnlyCloses(t *testing.T) {
	r := newRoaming(t, newFake())
	b := pressing(roamingBoard(t, r), "ctrl+o", "up", "up", "enter")
	if len(r.switched) != 0 || b.screen != board.OnTheBoard {
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

func TestSwitched_replacesTheBoardAndSaysSo(t *testing.T) {
	f := newFake()
	f.cards = []board.Card{{Number: 152, Title: "Choose the month", State: "current", Priority: "normal", InFocus: true}}
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
	if strings.Contains(screen, "Choose the month") || b.task != 0 {
		t.Errorf("work's board is still there:\n%s", screen)
	}
	if b = pressing(b, "l"); strings.Contains(strip(b.Draw()), "opened home") {
		t.Errorf("the line stays after a key:\n%s", strip(b.Draw()))
	}
}

func TestSwitcher_aListThatCannotBeReadIsSaid(t *testing.T) {
	r := newRoaming(t, newFake())
	b := roamingBoard(t, r)
	model, _ := b.Update(Listed{Err: errNope})
	if b = model.(Board); b.screen != board.OnTheBoard || !errors.Is(b.err, errNope) {
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
