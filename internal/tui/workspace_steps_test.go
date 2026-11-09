package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
)

type spaces struct {
	dir   string
	loose string
}

func (w *world) registerWorkspaces(sc *godog.ScenarioContext) {
	sc.Given(`^the workspace "([^"]*)"(, the default)?$`, w.theWorkspace)
	sc.Given(`^a store no workspace names, in the folder "([^"]*)"$`, w.aLooseStore)
	sc.Given(`^the workspace "([^"]*)" holds a task titled "([^"]*)"$`, w.workspaceHolds)
	sc.Given(`^the folder of the workspace "([^"]*)" is gone$`, func(name string) error {
		return os.RemoveAll(w.folderOf(name))
	})
	sc.Given(`^the settings of the workspace "([^"]*)" turn the time off$`, func(name string) error {
		return os.WriteFile(filepath.Join(w.folderOf(name), "settings.toml"), []byte("[pages]\ntime = \"off\"\n"), 0o600)
	})
	sc.Then(`^the header opens with "([^"]*)"$`, w.headerOpensWith)
	sc.Then(`^the header opens with "([^"]*)" and the folder "([^"]*)"$`, w.headerOpensWithFolder)
	sc.Then(`^the workspaces listed read (".*")$`, w.workspacesListedRead)
	sc.Then(`^the workspace "([^"]*)" is marked "([^"]*)"$`, w.workspaceMarked)
	sc.Then(`^the default workspace is still "([^"]*)"$`, w.defaultStill)
	sc.Then(`^task (\d+) of the workspace "([^"]*)" is still in focus$`, w.stillInFocus)
}

func (w *world) homeOf(workspace string) engine.Home {
	return engine.Home{Dir: w.spaces.dir, Workspace: workspace, Store: w.spaces.loose}
}

func (w *world) folderOf(name string) string { return filepath.Join(w.spaces.dir, "cairn", name) }

func (w *world) settingSpaces() {
	if w.spaces.dir == "" {
		w.spaces.dir = w.t.TempDir()
	}
}

func (w *world) theWorkspace(name, isDefault string) error {
	w.settingSpaces()
	h := w.homeOf("")
	if _, err := engine.DoFrom(engine.WorkspaceAdd, h, engine.Input{Given: map[string]string{"name": name, "folder": w.folderOf(name)}}, engine.Through("test", h)); err != nil {
		return err
	}
	if isDefault == "" {
		return nil
	}
	if _, err := engine.DoFrom(engine.WorkspaceUse, h, engine.Input{Given: map[string]string{"name": name}}, engine.Through("test", h)); err != nil {
		return err
	}
	return w.storeAt(filepath.Join(w.folderOf(name), "cairn.db"))
}

func (w *world) aLooseStore(folder string) error {
	w.settingSpaces()
	at := filepath.Join(w.spaces.dir, folder, "cairn.db")
	if err := os.MkdirAll(filepath.Dir(at), 0o700); err != nil {
		return err
	}
	w.spaces.loose = at
	return w.storeAt(at)
}

func (w *world) storeAt(path string) error {
	_ = w.other.Close()
	other, err := engine.Open(path)
	if err != nil {
		return err
	}
	w.path, w.other = path, other
	return nil
}

func (w *world) holdThroughHome() error {
	h := w.homeOf("")
	held, err := engine.Hold(h, engine.Through("tui", h))
	if err != nil {
		return err
	}
	_ = w.held.Close()
	w.held, w.pages = held, engine.PagesFrom(h)
	return nil
}

func (w *world) workspaceHolds(name, title string) error {
	h := w.homeOf(name)
	_, err := engine.DoFrom(engine.TaskNew, h, engine.Input{Given: map[string]string{"title": title}}, engine.Through("test", h))
	return err
}

func (w *world) header() string {
	return strings.TrimSpace(strings.SplitN(w.screen(), "\n", 2)[0])
}

func (w *world) headerOpensWith(start string) error {
	if !strings.HasPrefix(w.header(), start) {
		return fmt.Errorf("the header does not open with %q:\n%s", start, w.header())
	}
	return nil
}

func (w *world) headerOpensWithFolder(start, folder string) error {
	if err := w.headerOpensWith(start); err != nil {
		return err
	}
	if !strings.Contains(w.header(), string(filepath.Separator)+folder+" ") {
		return fmt.Errorf("the header does not name the folder %q:\n%s", folder, w.header())
	}
	return nil
}

func (w *world) workspacesListedRead(list string) error {
	var want []string
	for _, one := range strings.Split(list, ", ") {
		want = append(want, strings.Trim(one, `"`))
	}
	var got []string
	for _, one := range w.b.listed() {
		got = append(got, one.Name)
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("the workspaces listed read %q, not %q:\n%s", got, want, w.screen())
	}
	return nil
}

func (w *world) workspaceMarked(name, mark string) error {
	for _, line := range strings.Split(w.screen(), "\n") {
		if strings.Contains(line, "│") && strings.Contains(line, "● "+name+" ") {
			if strings.Contains(line, mark) {
				return nil
			}
			return fmt.Errorf("the workspace %q is not marked %q:\n%s", name, mark, line)
		}
	}
	return fmt.Errorf("the workspace %q is not listed:\n%s", name, w.screen())
}

func (w *world) defaultStill(name string) error {
	h := w.homeOf("")
	answer, err := engine.DoFrom(engine.WorkspaceList, h, engine.Input{}, engine.Through("test", h))
	if err != nil {
		return err
	}
	for _, one := range answer.Workspaces.Workspaces {
		if one.Default && one.Name == name {
			return nil
		}
	}
	return fmt.Errorf("the default workspace is no longer %q: %+v", name, answer.Workspaces.Workspaces)
}

func (w *world) stillInFocus(number int, name string) error {
	h := w.homeOf(name)
	answer, err := engine.DoFrom(engine.TaskList, h, engine.Input{}, engine.Through("test", h))
	if err != nil {
		return err
	}
	if answer.InFocus == nil || !slices.Contains(*answer.InFocus, number) {
		return errors.New("task " + strconv.Itoa(number) + " of " + name + " is no longer in focus")
	}
	return nil
}
