package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"cairn/internal/config"
)

// StateVariable is the environment variable naming where a person's state is kept, the recent folders among it; when it is not absolute, the home's .local/state is.
const StateVariable = "XDG_STATE_HOME"

const recentKept = 10

// RecentFolder is a store a board opened by path, that no workspace names: the store, the folder holding it when it is that folder's cairn.db, when a board last opened it, and whether it is gone since.
type RecentFolder struct {
	Store    string    `json:"store"`
	Folder   string    `json:"folder,omitempty" jsonschema:"the folder holding the store, when the store is its cairn.db: what keeping it as a workspace would name"`
	Opened   time.Time `json:"opened"`
	NotFound bool      `json:"not_found,omitempty"`
}

type remembered struct {
	Store  string    `json:"store"`
	Opened time.Time `json:"opened"`
}

type recentFile struct {
	Recent []remembered `json:"recent"`
}

func recentPath(h Home) string {
	if state, ok := h.lookup()(StateVariable); ok && filepath.IsAbs(state) {
		return filepath.Join(state, "cairn", "recent.json")
	}
	if h.Dir == "" {
		return ""
	}
	return filepath.Join(h.Dir, ".local", "state", "cairn", "recent.json")
}

func rememberedIn(path string) []remembered {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var read recentFile
	if json.Unmarshal(data, &read) != nil {
		return nil
	}
	return read.Recent
}

func remember(h Home, store string, at time.Time) error {
	path := recentPath(h)
	if path == "" {
		return nil
	}
	kept := slices.DeleteFunc(rememberedIn(path), func(one remembered) bool { return one.Store == store })
	kept = slices.Insert(kept, 0, remembered{Store: store, Opened: at})
	said, err := json.MarshalIndent(recentFile{Recent: kept[:min(len(kept), recentKept)]}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, append(said, '\n'), 0o600)
}

func (h Home) openedByPath() bool {
	_, a, err := h.settled()
	return err == nil && a.workspace == "" && (a.store.By == storeFlagName || a.store.By == byTheEnvironment)
}

func recentIn(h Home, p config.Places) []RecentFolder {
	var listed []RecentFolder
	for _, one := range rememberedIn(recentPath(h)) {
		if slices.ContainsFunc(p.Workspaces(), func(w config.Workspace) bool { return filepath.Join(w.Folder, storeFile) == one.Store }) {
			continue
		}
		folder := ""
		if filepath.Base(one.Store) == storeFile {
			folder = filepath.Dir(one.Store)
		}
		_, err := os.Stat(one.Store)
		listed = append(listed, RecentFolder{Store: one.Store, Folder: folder, Opened: one.Opened, NotFound: err != nil})
	}
	return listed
}

// Name is how a recent folder is named to a person, short: the name its store's settings give it, or its folder's own name.
func (r RecentFolder) Name() string {
	if own, err := config.NameIn(config.SettingsBeside(r.Store)); err == nil && own != "" {
		return own
	}
	return filepath.Base(r.Shown())
}

// Shown is the whole path of a recent folder: its folder, or the store itself when the store is not its folder's cairn.db.
func (r RecentFolder) Shown() string {
	if r.Folder != "" {
		return r.Folder
	}
	return r.Store
}

// Said is what a person reads beside a recent folder at the moment given: not found now, or when it was last opened, as a day.
func (r RecentFolder) Said(now time.Time) string {
	if r.NotFound {
		return "not found now"
	}
	opened, today := r.Opened.In(now.Location()), now
	day := func(t time.Time) time.Time {
		y, m, d := t.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	}
	switch days := int(day(today).Sub(day(opened)).Hours() / 24); {
	case days <= 0:
		return "opened today"
	case days == 1:
		return "opened yesterday"
	case opened.Year() == today.Year():
		return "opened " + opened.Format("2 Jan")
	default:
		return "opened " + opened.Format("2 Jan 2006")
	}
}

func recentText(recent []RecentFolder) string {
	if len(recent) == 0 {
		return ""
	}
	var said strings.Builder
	said.WriteString("\nrecent folders, opened by path:\n")
	now := time.Now()
	for _, one := range recent {
		fmt.Fprintf(&said, "  %s  %s\n", one.Shown(), one.Said(now))
	}
	return said.String()
}
