package gui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cucumber/godog"

	"cairn/internal/config"
	"cairn/internal/engine"
	"cairn/test"
)

type elsewhere struct {
	windows Windows
	claims  []io.Closer
	raised  map[string]*atomic.Int32
	started [][]string
	second  error
}

func (w *world) storeWritten(path string) error {
	_ = w.other.Close()
	other, err := engine.Open(path)
	if err != nil {
		return err
	}
	w.path, w.other = path, other
	return nil
}

func (w *world) settingElsewhere() error {
	dir, err := os.MkdirTemp("", "w")
	if err != nil {
		return err
	}
	w.apart = elsewhere{windows: Windows{Dir: dir}, raised: map[string]*atomic.Int32{}}
	*w.spaces = test.Spaces{Made: w.t.TempDir, Writing: w.storeWritten, Now: func() time.Time { return w.today }}
	return nil
}

func (w *world) leavingElsewhere() {
	for _, claim := range w.apart.claims {
		_ = claim.Close()
	}
	_ = os.RemoveAll(w.apart.windows.Dir)
}

func (w *world) storeOf(name string) string {
	return filepath.Join(w.spaces.FolderOf(name), "cairn.db")
}

func (w *world) openOn(h engine.Home) error {
	held, err := engine.Hold(h, engine.Through("gui", h))
	if err != nil {
		return err
	}
	_ = w.held.Close()
	w.held, w.work, w.pages = held, &refusing{held: held}, engine.PagesFrom(h)
	if err := w.open(); err != nil {
		return err
	}
	w.w = w.w.Seeing(w.apart.windows.Shows)
	return nil
}

func (w *world) handingApart() {
	to, asked := w.w.ApartAsked()
	if !asked {
		return
	}
	w.w = w.w.ApartTaken()
	held := w.held
	reach := func(to engine.Elsewhere) (engine.Apart, error) { return engine.OpenApart(held, to) }
	start := func(args ...string) error {
		w.apart.started = append(w.apart.started, args)
		return nil
	}
	w.w = w.w.Went(OpeningApart(to, reach, w.apart.windows, start))
}

func (w *world) claiming(store string) (io.Closer, error) {
	raised := w.apart.raised[store]
	if raised == nil {
		raised = &atomic.Int32{}
		w.apart.raised[store] = raised
	}
	return w.apart.windows.Claim(store, func() { raised.Add(1) })
}

func (w *world) switcherLine(name string, recent bool) (SwitcherLine, error) {
	for _, line := range w.w.SwitcherLines() {
		if line.Recent != recent {
			continue
		}
		if line.Name == name || recent && strings.HasSuffix(line.Name, string(filepath.Separator)+name) {
			return line, nil
		}
	}
	return SwitcherLine{}, fmt.Errorf("the switcher does not list %q: %+v", name, w.w.SwitcherLines())
}

func (w *world) clickAt(b Box) {
	w.w = w.w.clicked(Point{X: b.X + b.W/2, Y: b.Y + b.H/2})
	w.handingApart()
}

func (w *world) startedWith(want []string) error {
	for _, args := range w.apart.started {
		if slices.Equal(args, want) {
			return nil
		}
	}
	return fmt.Errorf("no cairn was started with %q: %q", want, w.apart.started)
}

func (w *world) workspaceSteps(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		return ctx, w.settingElsewhere()
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		w.leavingElsewhere()
		return ctx, err
	})
	w.spaces.Register(sc)
	sc.Given(`^a window is already open on the workspace "([^"]*)"$`, func(name string) error {
		claim, err := w.claiming(w.storeOf(name))
		if err != nil {
			return err
		}
		w.apart.claims = append(w.apart.claims, claim)
		return nil
	})
	sc.Given(`^the global configuration says the archive lasts (\d+) days$`, func(days int) error {
		path := config.FileIn(w.spaces.Dir)
		said, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(path, append(said, []byte(fmt.Sprintf("\n[archive]\ndays = %d\n", days))...), 0o600)
	})
	sc.When(`^I open the window on the workspace "([^"]*)"$`, func(name string) error {
		h := w.spaces.Home(name)
		h.Store = ""
		return w.openOn(h)
	})
	sc.When(`^I open the window on that store$`, func() error { return w.openOn(w.spaces.Home("")) })
	sc.When(`^I click the pill of the workspace$`, func() error {
		w.clickAt(w.w.Layout().Place)
		return nil
	})
	sc.When(`^I click the workspace "([^"]*)" in the switcher$`, func(name string) error {
		line, err := w.switcherLine(name, false)
		if err != nil {
			return err
		}
		w.clickAt(line.At)
		return nil
	})
	sc.When(`^I click "([^"]*)" in the switcher$`, func(said string) error {
		for _, act := range SwitcherActs(w.w) {
			if act.Said == said {
				w.clickAt(act.At)
				return nil
			}
		}
		return fmt.Errorf("the switcher offers no %q: %+v", said, SwitcherActs(w.w))
	})
	sc.When(`^I type the folder "([^"]*)"$`, func(folder string) error {
		w.w = w.w.Types(filepath.Join(w.spaces.Dir, folder))
		return nil
	})
	sc.When(`^a second window is opened on the workspace "([^"]*)"$`, func(name string) error {
		claim, err := w.apart.windows.Claim(w.storeOf(name), func() {})
		if claim != nil {
			w.apart.claims = append(w.apart.claims, claim)
		}
		w.apart.second = err
		return nil
	})
	sc.Then(`^the second window does not open$`, func() error {
		if !errors.Is(w.apart.second, ErrShownElsewhere) {
			return fmt.Errorf("the second window opened: %w", w.apart.second)
		}
		return nil
	})
	sc.Then(`^the window is titled "([^"]*)"$`, func(want string) error {
		if got := w.w.Title(); got != want {
			return fmt.Errorf("the window is titled %q, not %q", got, want)
		}
		return nil
	})
	sc.Then(`^the window is titled after the folder "([^"]*)"$`, func(folder string) error {
		if got, want := w.w.Title(), "cairn — "+folder; got != want {
			return fmt.Errorf("the window is titled %q, not %q", got, want)
		}
		return nil
	})
	sc.Then(`^the top bar opens with the (workspace|folder) "([^"]*)", before the tabs$`, func(kind, name string) error {
		l := w.w.Layout()
		if w.w.PlaceSays() != name || w.w.Loose() != (kind == "folder") {
			return fmt.Errorf("the top bar opens with %q, loose %v", w.w.PlaceSays(), w.w.Loose())
		}
		if l.Place.W == 0 || l.Place.X+l.Place.W > TabAt(l, 0).X {
			return fmt.Errorf("the pill at %+v does not stand before the tabs at %+v", l.Place, TabAt(l, 0))
		}
		return nil
	})
	sc.Then(`^the dot of the pill wears "([^"]*)"$`, func(colour string) error {
		if got := w.w.Place().Colour; got != colour {
			return fmt.Errorf("the dot of the pill wears %q, not %q", got, colour)
		}
		return nil
	})
	sc.Then(`^the switcher lists the workspaces (".*")$`, func(list string) error {
		if w.w.Screen() != OnWorkspaces {
			return fmt.Errorf("the window is on %v, not the switcher", w.w.Screen())
		}
		var got []string
		for _, line := range w.w.SwitcherLines() {
			if !line.Recent {
				got = append(got, line.Name)
			}
		}
		if want := quoted(list); !slices.Equal(got, want) {
			return fmt.Errorf("the switcher lists %q, not %q", got, want)
		}
		return nil
	})
	sc.Then(`^the workspace "([^"]*)" is tagged "([^"]*)" and "([^"]*)"$`, func(name, tag, key string) error {
		line, err := w.switcherLine(name, false)
		if err != nil {
			return err
		}
		if line.Tag != tag || line.Key != w.w.keyboard.Spelled(key) {
			return fmt.Errorf("%q is tagged %q and %q", name, line.Tag, line.Key)
		}
		return nil
	})
	sc.Then(`^the recent folder "([^"]*)" is tagged "([^"]*)"$`, func(folder, tag string) error {
		line, err := w.switcherLine(folder, true)
		if err != nil {
			return err
		}
		if line.Tag != tag {
			return fmt.Errorf("the recent folder %q is tagged %q", folder, line.Tag)
		}
		return nil
	})
	sc.Then(`^the switcher does not end on "([^"]*)"$`, func(said string) error {
		for _, act := range SwitcherActs(w.w) {
			if act.Said == said {
				return fmt.Errorf("the switcher ends on %q", said)
			}
		}
		return nil
	})
	sc.Then(`^the switcher ends on (".*")$`, func(list string) error {
		var got []string
		for _, act := range SwitcherActs(w.w) {
			got = append(got, act.Said)
		}
		if want := quoted(list); !slices.Equal(got, want) {
			return fmt.Errorf("the switcher ends on %q, not %q", got, want)
		}
		return nil
	})
	sc.Then(`^a cairn is started with "([^"]*)"$`, func(said string) error {
		return w.startedWith(strings.Fields(said))
	})
	sc.Then(`^a cairn is started with "([^"]*)" and the store of the folder "([^"]*)"$`, func(said, folder string) error {
		return w.startedWith(append(strings.Fields(said), filepath.Join(w.spaces.Dir, folder, "cairn.db")))
	})
	sc.Then(`^no cairn is started$`, func() error {
		if len(w.apart.started) > 0 {
			return fmt.Errorf("a cairn was started: %q", w.apart.started)
		}
		return nil
	})
	sc.Then(`^the window says "([^"]*)"$`, func(want string) error {
		said := w.w.told
		if err := w.w.Wrong(); err != nil {
			said += " " + err.Error()
		}
		if !strings.Contains(said, want) {
			return fmt.Errorf("the window says %q, not %q", said, want)
		}
		return nil
	})
	sc.Then(`^the window open on "([^"]*)" is brought forward$`, func(name string) error {
		if raised := w.apart.raised[w.storeOf(name)]; raised == nil || raised.Load() == 0 {
			return fmt.Errorf("the window open on %q was not brought forward", name)
		}
		return nil
	})
	sc.Then(`^the configuration names the workspace "([^"]*)"$`, func(name string) error {
		h := w.spaces.Home("")
		h.Store = ""
		answer, err := engine.DoFrom(engine.WorkspaceList, h, engine.Input{}, engine.Through("test", h))
		if err != nil {
			return err
		}
		for _, one := range answer.Workspaces.Workspaces {
			if one.Name == name {
				return nil
			}
		}
		return fmt.Errorf("the configuration does not name %q", name)
	})
	sc.Then(`^the window asks the system to open the settings of "([^"]*)"$`, func(name string) error {
		got, asked := w.w.OpenAsked()
		if want := filepath.Join(w.spaces.FolderOf(name), config.SettingsFile); !asked || got != want {
			return fmt.Errorf("the window asks the system to open %q, not %q", got, want)
		}
		return nil
	})
}
