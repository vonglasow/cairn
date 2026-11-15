package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// OldConfiguration is the suffix the configuration of v0.11.0 is kept under, beside the file, once turned into workspaces.
const OldConfiguration = ".v0.11.0"

// ErrNotesNotMoved is returned when a workspace's notes are asked to move into a folder that already holds something.
var ErrNotesNotMoved = errors.New("already holds something: the notes were not moved there")

// NotesMove is a workspace's notes found last in one folder while the configuration now names another, and whether that other already holds something.
type NotesMove struct {
	Workspace string
	From      string
	To        string
	Blocked   bool
}

// NotesFound is the folder a workspace's notes were last found in.
type NotesFound struct {
	Notes string `json:"notes"`
}

// NotesRecord is the file that keeps where a store's notes were found.
func NotesRecord(store string) string { return store + ".notes.json" }

// KeepNotesFound keeps where a store's notes were found.
func KeepNotesFound(record, notes string) {
	if _, err := os.Stat(filepath.Dir(record)); err != nil {
		return
	}
	data, _ := json.Marshal(NotesFound{Notes: notes})
	_ = os.WriteFile(record, data, 0o600)
}

// HoldsSomething says whether a folder holds anything.
func HoldsSomething(folder string) bool {
	entries, err := os.ReadDir(folder)
	return err == nil && len(entries) > 0
}
