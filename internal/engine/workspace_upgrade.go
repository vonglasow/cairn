package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"cairn/internal/config"
)

// OldConfiguration is the suffix the configuration of v0.11.0 is kept under, beside the file, once turned into workspaces.
const OldConfiguration = ".v0.11.0"

// ErrNotesNotMoved is returned when a workspace's notes are asked to move into a folder that already holds something.
var ErrNotesNotMoved = errors.New("already holds something: the notes were not moved there")

// UpgradeConfiguration turns a configuration of v0.11.0, which names its store and notes at its top, into workspaces, keeping the file as it was beside it, and says what it did; nothing is said when there was nothing to turn.
func UpgradeConfiguration(h Home) (string, error) {
	if h.Dir == "" {
		return "", nil
	}
	path := config.FileIn(h.Dir)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	up, err := config.Upgraded(data, h.Dir)
	if err != nil || up.Data == nil {
		return "", err
	}
	if err := os.MkdirAll(up.Folder, 0o700); err != nil {
		return "", err
	}
	kept := path + OldConfiguration
	if _, err := os.Stat(kept); errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(kept, data, 0o600); err != nil {
			return "", err
		}
	}
	if err := os.WriteFile(path, up.Data, 0o600); err != nil {
		return "", err
	}
	chosen := "the default"
	if !up.Default {
		chosen = "beside the default, which stays"
	}
	return fmt.Sprintf("the configuration of v0.11.0 is now in workspaces: its store and notes are the workspace %q, its folder %s and its notes %s, %s; nothing was moved; the file as it was is kept at %s", up.Name, up.Folder, up.Notes, chosen, kept), nil
}

// NotesMove is a workspace's notes found last in one folder while the configuration now names another, and whether that other already holds something.
type NotesMove struct {
	Workspace string
	From      string
	To        string
	Blocked   bool
}

type notesFound struct {
	Notes string `json:"notes"`
}

func notesRecord(store string) string { return store + ".notes.json" }

// NotesToMove says whether the notes of the workspace answering were last found elsewhere than where the configuration now names them, and remembers where they are when there is nothing to move.
func NotesToMove(h Home) (NotesMove, bool, error) {
	_, a, err := h.settled()
	if err != nil {
		return NotesMove{}, false, err
	}
	if a.folder == "" || a.notes.By == byTheEnvironment || a.why.Rung == rungEnvironment || a.why.Rung == rungStoreFlag {
		return NotesMove{}, false, nil
	}
	record := notesRecord(a.store.At)
	var was notesFound
	if data, err := os.ReadFile(record); err == nil && json.Unmarshal(data, &was) == nil && was.Notes != a.notes.At && holdsSomething(was.Notes) {
		return NotesMove{Workspace: a.workspace, From: was.Notes, To: a.notes.At, Blocked: holdsSomething(a.notes.At)}, true, nil
	}
	if was.Notes != a.notes.At {
		keepNotesFound(record, a.notes.At)
	}
	return NotesMove{}, false, nil
}

// MoveNotes moves a workspace's notes where the configuration now names them, and remembers them there; a folder already holding something is refused.
func MoveNotes(h Home, move NotesMove) error {
	if holdsSomething(move.To) {
		return fmt.Errorf("%s %w", move.To, ErrNotesNotMoved)
	}
	at, err := Locate(h)
	if err != nil {
		return err
	}
	if err := os.Remove(move.To); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(move.To), 0o700); err != nil {
		return err
	}
	if err := os.Rename(move.From, move.To); err != nil {
		return err
	}
	keepNotesFound(notesRecord(at.Store), move.To)
	return nil
}

func keepNotesFound(record, notes string) {
	if _, err := os.Stat(filepath.Dir(record)); err != nil {
		return
	}
	data, _ := json.Marshal(notesFound{Notes: notes})
	_ = os.WriteFile(record, data, 0o600)
}

func holdsSomething(folder string) bool {
	entries, err := os.ReadDir(folder)
	return err == nil && len(entries) > 0
}
