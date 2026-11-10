package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
)

// Spaces is the workspaces, the loose store and the recent folders a scenario names, all under a home of its own, made the first time one is given; Writing is told of each store a given makes the one somebody else writes to, and Now is the moment a board stands at.
type Spaces struct {
	Dir     string
	Loose   string
	Made    func() string
	Writing func(store string) error
	Now     func() time.Time
}

// Home is the home a scenario's workspaces live under, naming the workspace given, and the loose store when one was given.
func (s *Spaces) Home(workspace string) engine.Home {
	return engine.Home{Dir: s.Dir, Workspace: workspace, Store: s.Loose}
}

// FolderOf is the folder a workspace given by that name holds its store and its notes in.
func (s *Spaces) FolderOf(name string) string { return filepath.Join(s.Dir, "cairn", name) }

func (s *Spaces) settling() {
	if s.Dir == "" {
		s.Dir = s.Made()
	}
}

func (s *Spaces) asked(op engine.Operation, h engine.Home, given map[string]string) error {
	_, err := engine.DoFrom(op, h, engine.Input{Given: given}, engine.Through("test", h))
	return err
}

// Workspace names a workspace in the configuration, made by workspace add; the default once made so, and then the store somebody else writes to.
func (s *Spaces) Workspace(name, isDefault string) error {
	s.settling()
	h := s.Home("")
	if err := s.asked(engine.WorkspaceAdd, h, map[string]string{"name": name, "folder": s.FolderOf(name)}); err != nil {
		return err
	}
	if isDefault == "" {
		return nil
	}
	if err := s.asked(engine.WorkspaceUse, h, map[string]string{"name": name}); err != nil {
		return err
	}
	return s.Writing(filepath.Join(s.FolderOf(name), "cairn.db"))
}

// LooseStore makes the folder of a store no workspace names, which becomes the store somebody else writes to.
func (s *Spaces) LooseStore(folder string) error {
	s.settling()
	at := filepath.Join(s.Dir, folder, "cairn.db")
	if err := os.MkdirAll(filepath.Dir(at), 0o700); err != nil {
		return err
	}
	s.Loose = at
	return s.Writing(at)
}

// RecentFolder remembers a store opened by path a day before Now, there still or gone.
func (s *Spaces) RecentFolder(folder, how string) error {
	s.settling()
	store := filepath.Join(s.Dir, folder, "cairn.db")
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
	path := filepath.Join(s.Dir, ".local", "state", "cairn", "recent.json")
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
	kept.Recent = append(kept.Recent, map[string]any{"store": store, "opened": s.Now().Add(-24 * time.Hour)})
	said, err := json.Marshal(kept)
	if err != nil {
		return err
	}
	return os.WriteFile(path, said, 0o600)
}

func saying(path, line string, wanted bool) error {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if said := slices.Contains(strings.Split(string(data), "\n"), line); said != wanted {
		return fmt.Errorf("%s says, wanting %q %v:\n%s", path, line, wanted, data)
	}
	return nil
}

func (s *Spaces) settings(name, said string) error {
	return os.WriteFile(filepath.Join(s.FolderOf(name), "settings.toml"), []byte(said), 0o600)
}

// Register declares the givens about workspaces both boards share.
func (s *Spaces) Register(sc *godog.ScenarioContext) {
	sc.Given(`^the workspace "([^"]*)"(, the default)?$`, s.Workspace)
	sc.Given(`^a store no workspace names, in the folder "([^"]*)"$`, s.LooseStore)
	sc.Given(`^the workspace "([^"]*)" holds a task titled "([^"]*)"$`, func(name, title string) error {
		return s.asked(engine.TaskNew, s.Home(name), map[string]string{"title": title})
	})
	sc.Given(`^the folder of the workspace "([^"]*)" is gone$`, func(name string) error {
		return os.RemoveAll(s.FolderOf(name))
	})
	sc.Given(`^the settings of the workspace "([^"]*)" turn the time off$`, func(name string) error {
		return s.settings(name, "[pages]\ntime = \"off\"\n")
	})
	sc.Given(`^the settings of the workspace "([^"]*)" say its colour is "([^"]*)"$`, func(name, colour string) error {
		return s.settings(name, "[workspace]\ncolour = \""+colour+"\"\n")
	})
	sc.Step(`^the settings of the workspace "([^"]*)" are written by hand:$`, func(name string, said *godog.DocString) error {
		return s.settings(name, said.Content+"\n")
	})
	sc.Then(`^the settings of the workspace "([^"]*)" (say|do not say) `+"`([^`]*)`"+`$`, func(name, says, line string) error {
		return saying(filepath.Join(s.FolderOf(name), "settings.toml"), line, says == "say")
	})
	sc.Then(`^the global configuration (says|does not say) `+"`([^`]*)`"+`$`, func(says, line string) error {
		return saying(filepath.Join(s.Dir, ".config", "cairn", "config.toml"), line, says == "says")
	})
	sc.Given(`^the recent folder "([^"]*)", (opened yesterday|whose store is gone)$`, s.RecentFolder)
	sc.Given(`^the folder "([^"]*)" holds a task titled "([^"]*)"$`, func(folder, title string) error {
		s.settling()
		h := engine.Home{Dir: s.Dir, Store: filepath.Join(s.Dir, folder, "cairn.db")}
		return s.asked(engine.TaskNew, h, map[string]string{"title": title})
	})
}
