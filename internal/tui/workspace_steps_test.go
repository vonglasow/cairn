package tui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"cairn/internal/domain"
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
	sc.Given(`^the settings of the workspace "([^"]*)" say its colour is "([^"]*)"$`, func(name, colour string) error {
		return os.WriteFile(filepath.Join(w.folderOf(name), "settings.toml"), []byte("[workspace]\ncolour = \""+colour+"\"\n"), 0o600)
	})
	sc.Given(`^the recent folder "([^"]*)", (opened yesterday|whose store is gone)$`, w.aRecentFolder)
	sc.Given(`^the folder "([^"]*)" holds a task titled "([^"]*)"$`, w.folderHolds)
	sc.When(`^I type the folder "([^"]*)"$`, func(folder string) error {
		return w.typeIn(filepath.Join(w.spaces.dir, folder))
	})
	sc.Then(`^the dot of the header wears "([^"]*)"$`, w.dotWears)
	sc.Then(`^the recent folder "([^"]*)" is marked "([^"]*)"$`, w.recentMarked)
	sc.Then(`^the configuration names the workspace "([^"]*)"$`, w.configurationNames)
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

func (w *world) aRecentFolder(folder, how string) error {
	w.settingSpaces()
	store := filepath.Join(w.spaces.dir, folder, "cairn.db")
	if how == "opened yesterday" {
		if err := os.MkdirAll(filepath.Dir(store), 0o700); err != nil {
			return err
		}
		if _, err := os.Stat(store); errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(store, nil, 0o600); err != nil {
				return err
			}
		}
	}
	path := filepath.Join(w.spaces.dir, ".local", "state", "cairn", "recent.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	var kept struct {
		Recent []map[string]any `json:"recent"`
	}
	if read, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(read, &kept); err != nil {
			return err
		}
	}
	kept.Recent = append(kept.Recent, map[string]any{"store": store, "opened": stood.Add(-24 * time.Hour)})
	said, err := json.Marshal(kept)
	if err != nil {
		return err
	}
	return os.WriteFile(path, said, 0o600)
}

func (w *world) folderHolds(folder, title string) error {
	w.settingSpaces()
	h := engine.Home{Dir: w.spaces.dir, Store: filepath.Join(w.spaces.dir, folder, "cairn.db")}
	_, err := engine.DoFrom(engine.TaskNew, h, engine.Input{Given: map[string]string{"title": title}}, engine.Through("test", h))
	return err
}

func (w *world) dotWears(colour string) error {
	ink, ok := domain.LabelInks[colour]
	if !ok {
		return fmt.Errorf("%q is none of the label colours", colour)
	}
	worn := fmt.Sprintf("38;2;%d;%d;%dm● ", ink[0], ink[1], ink[2])
	if header := strings.SplitN(w.b.Draw(), "\n", 2)[0]; !strings.Contains(header, worn) {
		return fmt.Errorf("the dot of the header does not wear %s (%q):\n%q", colour, worn, header)
	}
	return nil
}

func (w *world) recentMarked(folder, mark string) error {
	for _, line := range strings.Split(w.screen(), "\n") {
		if strings.Contains(line, "│") && strings.Contains(line, "◌ ") && strings.Contains(line, string(filepath.Separator)+folder+" ") {
			if strings.Contains(line, mark) {
				return nil
			}
			return fmt.Errorf("the recent folder %q is not marked %q:\n%s", folder, mark, line)
		}
	}
	return fmt.Errorf("the recent folder %q is not listed:\n%s", folder, w.screen())
}

func (w *world) configurationNames(name string) error {
	h := w.homeOf("")
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
	return fmt.Errorf("the configuration does not name %q: %+v", name, answer.Workspaces.Workspaces)
}
