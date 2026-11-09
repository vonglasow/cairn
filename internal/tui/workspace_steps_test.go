package tui

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

func (w *world) registerWorkspaces(sc *godog.ScenarioContext) {
	w.spaces.Register(sc)
	sc.When(`^I type the folder "([^"]*)"$`, func(folder string) error {
		return w.typeIn(filepath.Join(w.spaces.Dir, folder))
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

func (w *world) homeOf(workspace string) engine.Home { return w.spaces.Home(workspace) }

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
