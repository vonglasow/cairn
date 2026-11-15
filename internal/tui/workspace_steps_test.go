package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
)

func (w *world) registerWorkspaces(sc *godog.ScenarioContext) {
	w.spaces.Register(sc)
	sc.When(`^I type the folder "([^"]*)"$`, func(folder string) error {
		return w.typeIn(filepath.Join(w.spaces.Dir, folder))
	})
	sc.Then(`^the configuration names the workspace "([^"]*)"$`, w.configurationNames)
	sc.Then(`^the header opens with "([^"]*)"$`, w.headerOpensWith)
	sc.Then(`^the header opens with "([^"]*)" and the folder "([^"]*)"$`, w.headerOpensWithFolder)
	sc.Then(`^the workspace "([^"]*)" is marked "([^"]*)"$`, w.workspaceMarked)
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
