package workspace

import (
	"cairn/internal/store/files"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// StateVariable is the environment variable naming where a person's state is kept, the recent folders among it; when it is not absolute, the home's .local/state is.
const StateVariable = "XDG_STATE_HOME"

// RecentKept is how many recent workspaces are kept.
const RecentKept = 10

// RecentFolder is a store a board opened by path, that no workspace names: the store, the folder holding it when it is that folder's cairn.db, when a board last opened it, and whether it is gone since.
type RecentFolder struct {
	Store    string    `json:"store"`
	Folder   string    `json:"folder,omitempty" jsonschema:"the folder holding the store, when the store is its cairn.db: what keeping it as a workspace would name"`
	Opened   time.Time `json:"opened"`
	NotFound bool      `json:"not_found,omitempty"`
}

// Remembered is a store opened, and when.
type Remembered struct {
	Store  string    `json:"store"`
	Opened time.Time `json:"opened"`
}

// RecentFile is what the file of the recent workspaces holds.
type RecentFile struct {
	Recent []Remembered `json:"recent"`
}

// RememberedIn reads the recent workspaces kept in a file.
func RememberedIn(path string) []Remembered {
	data, err := files.Kept(path).Read()
	if err != nil {
		return nil
	}
	var read RecentFile
	if json.Unmarshal(data, &read) != nil {
		return nil
	}
	return read.Recent
}

// Name is how a recent folder is named to a person, short: the name its store's settings give it, or its folder's own name.
func (r RecentFolder) Name() string {
	if own, err := files.NameIn(files.SettingsBeside(r.Store)); err == nil && own != "" {
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
	days := int(day(today).Sub(day(opened)).Hours() / 24)
	if days <= 0 {
		return "opened today"
	}
	if days == 1 {
		return "opened yesterday"
	}
	if opened.Year() == today.Year() {
		return "opened " + opened.Format("2 Jan")
	}
	return "opened " + opened.Format("2 Jan 2006")
}

// RecentText is the recent workspaces for a person.
func RecentText(recent []RecentFolder) string {
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
