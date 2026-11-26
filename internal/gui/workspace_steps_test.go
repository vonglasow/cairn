package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/workspace"
	"cairn/internal/gui/geom"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/store/files"
	"cairn/test"
)

type elsewhere struct {
	windows Windows
	claims  []io.Closer
	raised  map[string]*atomic.Int32
	started [][]string
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
	reach := func(to engine.Elsewhere) (workspace.Apart, error) { return engine.OpenApart(held, to) }
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

func (w *world) clickAt(b geom.Box) {
	w.w = w.w.clicked(geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2})
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
		path := files.FileIn(w.spaces.Dir)
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
	sc.Then(`^the dialog "([^"]*)" asks (".*")$`, func(head, list string) error {
		want := quoted(strings.Replace(list, `" and "`, `", "`, 1))
		if !w.w.MakingSpace() || head != NewSpaceHead || !slices.Equal(NewSpaceAsks(), want) {
			return fmt.Errorf("the window shows no dialog %q asking %q: %v", head, want, w.w.Screen)
		}
		return nil
	})
	sc.Then(`^the folder of the new workspace reads "([^"]*)"$`, func(want string) error {
		if got := w.w.NewSpaceFolderShown(); got != want {
			return fmt.Errorf("the folder reads %q, not %q", got, want)
		}
		return nil
	})
	sc.When(`^I click the field "([^"]*)" of the dialog$`, func(field string) error {
		laid := w.w.NewSpaceLaid()
		w.clickAt(map[string]geom.Box{NewSpaceName: laid.Name, NewSpaceFolder: laid.Folder, NewSpaceNotes: laid.Notes}[field])
		return nil
	})
	sc.Then(`^the notes of the new workspace read "([^"]*)"$`, func(want string) error {
		if got := w.w.NewSpaceNotesShown(); got != want {
			return fmt.Errorf("the notes read %q, not %q", got, want)
		}
		return nil
	})
	sc.Then(`^the configuration names the notes of "([^"]*)" in the folder "([^"]*)"$`, func(name, folder string) error {
		h := w.spaces.Home("")
		h.Store = ""
		answer, err := engine.DoFrom(engine.WorkspaceList, h, engine.Input{}, engine.Through("test", h))
		if err != nil {
			return err
		}
		for _, one := range answer.Workspaces.Workspaces {
			if one.Name == name && one.Notes == filepath.Join(w.spaces.Dir, folder) {
				return nil
			}
		}
		return fmt.Errorf("the configuration names no notes of %q in %q: %+v", name, folder, answer.Workspaces.Workspaces)
	})
	sc.When(`^I click the colour "([^"]*)" in the dialog$`, func(colour string) error {
		for _, swatch := range w.w.NewSpaceLaid().Swatches {
			if swatch.Colour == colour {
				w.clickAt(swatch.At)
				return nil
			}
		}
		return fmt.Errorf("the dialog offers no colour %q", colour)
	})
	sc.When(`^I click "([^"]*)" in the dialog$`, func(said string) error {
		laid := w.w.NewSpaceLaid()
		at, ok := map[string]geom.Box{NewSpaceChoose: laid.Choose, NewSpaceSwitch: laid.Switch, NewSpaceCancel: laid.Cancel, NewSpaceCreate: laid.Create}[said]
		if !ok || !w.w.MakingSpace() {
			return fmt.Errorf("the dialog offers no %q", said)
		}
		w.clickAt(at)
		return nil
	})
	sc.When(`^I pick the folder "([^"]*)" in the system's dialog$`, func(folder string) error {
		return w.answerDialog(dialog{open: filepath.Join(w.spaces.Dir, folder)})
	})
	sc.Then(`^the dialog "([^"]*)" says "([^"]*)"$`, func(_, want string) error {
		if got := w.w.NewSpaceRefused(); !w.w.MakingSpace() || !strings.Contains(got, want) {
			return fmt.Errorf("the dialog says %q", got)
		}
		return nil
	})
	sc.Then(`^the configuration does not name the workspace "([^"]*)"$`, func(name string) error {
		h := w.spaces.Home("")
		h.Store = ""
		answer, err := engine.DoFrom(engine.WorkspaceList, h, engine.Input{}, engine.Through("test", h))
		if err != nil {
			return err
		}
		for _, one := range answer.Workspaces.Workspaces {
			if one.Name == name {
				return fmt.Errorf("the configuration names %q", name)
			}
		}
		return nil
	})
	sc.Given(`^the workspace "([^"]*)" holds the notes of a task$`, func(name string) error {
		summary := filepath.Join(w.spaces.FolderOf(name), "notes", "1", "summary.md")
		if err := os.MkdirAll(filepath.Dir(summary), 0o700); err != nil {
			return err
		}
		return os.WriteFile(summary, []byte("where it stands"), 0o600)
	})
	sc.Then(`^the notes of the task are in the folder "([^"]*)"$`, func(folder string) error {
		_, err := os.Stat(filepath.Join(w.spaces.Dir, folder, "1", "summary.md"))
		return err
	})
	sc.When(`^I type the folder "([^"]*)"$`, func(folder string) error {
		w.w = w.w.Types(filepath.Join(w.spaces.Dir, folder))
		return nil
	})
	sc.Then(`^the window is titled "([^"]*)"$`, func(want string) error {
		if got := w.w.Title(); got != want {
			return fmt.Errorf("the window is titled %q, not %q", got, want)
		}
		return nil
	})
	sc.Then(`^the top bar opens with the (workspace|folder) "([^"]*)", before the tabs$`, func(kind, name string) error {
		l := w.w.Layout()
		if w.w.PlaceSays() != name || w.w.Loose() != (kind == "folder") {
			return fmt.Errorf("the top bar opens with %q, loose %v", w.w.PlaceSays(), w.w.Loose())
		}
		if l.Place.W == 0 || l.Place.X+l.Place.W > w.w.TabsAt()[0].X {
			return fmt.Errorf("the pill at %+v does not stand before the tabs at %+v", l.Place, w.w.TabsAt()[0])
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
		if w.w.Screen != board.OnWorkspaces {
			return fmt.Errorf("the window is on %v, not the switcher", w.w.Screen)
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
		said := w.w.Telling
		if err := w.w.Wrong(); err != nil {
			said += " " + err.Error()
		}
		if !strings.Contains(said, want) {
			return fmt.Errorf("the window says %q, not %q", said, want)
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
}
