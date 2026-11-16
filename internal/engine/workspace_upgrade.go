package engine

import (
	"cairn/internal/engine/workspace"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"cairn/internal/store/files"
)

// UpgradeConfiguration turns a configuration of v0.11.0, which names its store and notes at its top, into workspaces, keeping the file as it was beside it, and says what it did; nothing is said when there was nothing to turn.
func UpgradeConfiguration(h Home) (string, error) {
	if h.Dir == "" {
		return "", nil
	}
	path := files.FileIn(h.Dir)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	up, err := files.Upgraded(data, h.Dir)
	if err != nil || up.Data == nil {
		return "", err
	}
	if err := os.MkdirAll(up.Folder, 0o700); err != nil {
		return "", err
	}
	kept := path + workspace.OldConfiguration
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

// NotesToMove says whether the notes of the workspace answering were last found elsewhere than where the configuration now names them, and remembers where they are when there is nothing to move.
func NotesToMove(h Home) (workspace.NotesMove, bool, error) {
	_, a, err := h.settled()
	if err != nil {
		return workspace.NotesMove{}, false, err
	}
	if a.Folder == "" || a.Notes.By == workspace.ByTheEnvironment || a.Why.Rung == workspace.RungEnvironment || a.Why.Rung == workspace.RungStoreFlag {
		return workspace.NotesMove{}, false, nil
	}
	move, left, was := notesLeft(a)
	if left {
		return move, true, nil
	}
	if was != a.Notes.At {
		workspace.KeepNotesFound(workspace.NotesRecord(a.Store.At), a.Notes.At)
	}
	return workspace.NotesMove{}, false, nil
}

func notesLeft(a workspace.Answering) (workspace.NotesMove, bool, string) {
	var was workspace.NotesFound
	if data, err := os.ReadFile(workspace.NotesRecord(a.Store.At)); err == nil {
		_ = json.Unmarshal(data, &was)
	}
	if was.Notes == a.Notes.At || !workspace.HoldsSomething(was.Notes) {
		return workspace.NotesMove{}, false, was.Notes
	}
	tasks := 0
	if entries, err := os.ReadDir(was.Notes); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				tasks++
			}
		}
	}
	return workspace.NotesMove{Workspace: a.Workspace, From: was.Notes, To: a.Notes.At, Blocked: workspace.HoldsSomething(a.Notes.At), Tasks: tasks}, true, was.Notes
}

// MoveNotes moves a workspace's notes where the configuration now names them, and remembers them there; a folder already holding something is refused.
func MoveNotes(h Home, move workspace.NotesMove) error {
	if workspace.HoldsSomething(move.To) {
		return fmt.Errorf("%s %w", move.To, workspace.ErrNotesNotMoved)
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
	workspace.KeepNotesFound(workspace.NotesRecord(at.Store), move.To)
	return nil
}

// NotesMove moves the notes of the workspace in use to the folder its configuration now names, when they were left where they were.
var NotesMove = Operation{
	Thing:   "notes",
	Verb:    "move",
	Short:   "Move a workspace's notes to the folder it now names",
	Does:    "Move the notes of the workspace in use from where they were to the folder its configuration now names, the one a vault may hold; a folder already holding something is refused, and nothing is moved.",
	Example: "cairn notes move",
	Level:   write,
	Needs:   aConfiguration,
	Again:   ChangesNothingMore,
	Answers: workspace.NotesMove{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		_, a, err := on.Home.settled()
		if err != nil {
			return Answer{}, err
		}
		move, left, _ := notesLeft(a)
		if !left || a.Folder == "" {
			return Answer{}, fmt.Errorf("the notes of %s %w", a.Notes.At, workspace.ErrNoNotesLeft)
		}
		if err := MoveNotes(on.Home, move); err != nil {
			return Answer{}, err
		}
		return Answer{NotesMoved: &move}, nil
	},
}
