package engine

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/config"
)

func TestLocate_theWorkspaceTheEnvironmentNamesGivesBothPlaces(t *testing.T) {
	h := homeHolding(t, "store = \"~/s/cairn.db\"\n"+twoWorkspaces, map[string]string{WorkspaceVariable: "home"}, "w", "h")
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "h", "cairn.db") || at.Notes != filepath.Join(h.Dir, "h", "notes") {
		t.Errorf("Locate = %+v, %v; want home's store and notes, before the default and the file", at, err)
	}
}

func TestLocate_anEmptyWorkspaceVariableSaysNothing(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{WorkspaceVariable: ""}, "w", "h")
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "w", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want the default's store", at, err)
	}
}

func TestLocate_theEnvironmentsStoreComesBeforeItsWorkspaceAndKeepsItsOwnNotes(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{WorkspaceVariable: "home", config.StoreVariable: "~/e/cairn.db"}, "w", "h")
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "e", "cairn.db") || at.Notes != filepath.Join(h.Dir, ".cairn", "notes") {
		t.Errorf("Locate = %+v, %v; want CAIRN_STORE's store, its notes never home's", at, err)
	}
}

func TestLocate_theEnvironmentsNotesComeBeforeItsWorkspacesForTheNotesAlone(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{WorkspaceVariable: "home", config.NotesVariable: "~/n"}, "w", "h")
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "h", "cairn.db") || at.Notes != filepath.Join(h.Dir, "n") {
		t.Errorf("Locate = %+v, %v; want home's store and CAIRN_NOTES's notes", at, err)
	}
}

func TestLocate_aFlagComesBeforeTheWorkspaceTheEnvironmentNames(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{WorkspaceVariable: "home"}, "w", "h")
	h.Workspace = "work"
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "w", "cairn.db") {
		t.Errorf("--workspace: Locate = %+v, %v; want work's store", at, err)
	}
	h.Workspace, h.Store = "", "~/f/cairn.db"
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "f", "cairn.db") || at.Notes != filepath.Join(h.Dir, ".cairn", "notes") {
		t.Errorf("--store: Locate = %+v, %v; want the flag's store and the notes as before workspaces", at, err)
	}
}

func TestLocate_anUnknownWorkspaceInTheEnvironmentIsRefusedNamingTheVariable(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{WorkspaceVariable: "play"}, "w", "h")
	_, err := Locate(h)
	if !errors.Is(err, ErrNoSuchWorkspace) || !strings.Contains(err.Error(), WorkspaceVariable) || !strings.Contains(err.Error(), `"play"`) || !strings.Contains(err.Error(), "home, work") {
		t.Errorf("Locate = %v, want CAIRN_WORKSPACE play refused, naming home and work", err)
	}
}

func TestLocate_aWorkspaceInTheEnvironmentWhoseFolderIsGoneIsRefusedNamingIt(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{WorkspaceVariable: "home"}, "w")
	_, err := Locate(h)
	if !errors.Is(err, ErrWorkspaceNotThere) || !strings.Contains(err.Error(), WorkspaceVariable) || !strings.Contains(err.Error(), filepath.Join(h.Dir, "h")) {
		t.Errorf("Locate = %v, want home refused, naming CAIRN_WORKSPACE and its folder", err)
	}
}

func TestKeep_aServerKeepsTheWorkspaceTheEnvironmentNamed(t *testing.T) {
	h := Keep(homeHolding(t, twoWorkspaces, map[string]string{WorkspaceVariable: "home"}, "w", "h"))
	if h.Kept == nil || h.Kept.Workspace != "home" || h.Kept.Why.Rung != WorkspaceVariable || !h.Kept.Why.Started {
		t.Errorf("kept %+v, want home, named by CAIRN_WORKSPACE when the server started", h.Kept)
	}
}

func TestRecent_aWorkspaceTheEnvironmentNamesIsNotRemembered(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{WorkspaceVariable: "home"}, "w", "h")
	if h.openedByPath() {
		t.Error("a workspace CAIRN_WORKSPACE names is taken for a store opened by path")
	}
}
