package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// UpgradedMain is the name the store and notes a configuration of v0.11.0 named at its top are given as a workspace, unless one is already at that folder.
const UpgradedMain = "main"

// ErrNotAWorkspaceStore is returned when the store a configuration of v0.11.0 names is not called cairn.db, which no workspace's folder could hold.
var ErrNotAWorkspaceStore = errors.New("is not called cairn.db, so no workspace's folder holds it: name its folder in [workspaces] by hand")

// Upgrade is what turning a configuration of v0.11.0 into workspaces wrote: the file, the workspace its places became, ~ expanded, and whether it is now the default; no Data when there was nothing to turn.
type Upgrade struct {
	Data    []byte
	Name    string
	Folder  string
	Notes   string
	Default bool
}

var aTopLevelPlace = regexp.MustCompile(`^\s*(store|notes)\s*=`)

// Upgraded is the file with the store and notes it names at its top turned into a workspace — the one already at that folder, or main — the default unless one is already named, every other byte the file's.
func Upgraded(data []byte, home string) (Upgrade, error) {
	var top struct {
		Store      *string        `toml:"store"`
		Notes      *string        `toml:"notes"`
		Workspaces map[string]any `toml:"workspaces"`
	}
	if toml.Unmarshal(data, &top) != nil {
		return Upgrade{}, ErrNotTOML
	}
	if top.Store == nil && top.Notes == nil {
		return Upgrade{}, nil
	}
	store, notes := "~/"+DefaultFolder+"/"+defaultFile, "~/"+DefaultFolder+"/"+defaultNotesFolder
	if top.Store != nil {
		store = *top.Store
	}
	if top.Notes != nil {
		notes = *top.Notes
	}
	up := Upgrade{Name: UpgradedMain}
	var err error
	if up.Folder, err = Absolute(filepath.Dir(store), home); err != nil {
		return Upgrade{}, fmt.Errorf("store %w", err)
	}
	if filepath.Base(store) != defaultFile {
		return Upgrade{}, fmt.Errorf("store %q %w", store, ErrNotAWorkspaceStore)
	}
	if up.Notes, err = Absolute(notes, home); err != nil {
		return Upgrade{}, fmt.Errorf("notes %w", err)
	}
	named, chosen, err := workspacesOf(top.Workspaces, home)
	if err != nil {
		return Upgrade{}, err
	}
	lines := linesOf(data)
	_, end, _ := regionOf(lines, "")
	said := []byte(strings.Join(append(slices.DeleteFunc(slices.Clone(lines[:end]), aTopLevelPlace.MatchString), lines[end:]...), ""))
	folder := filepath.Dir(store)
	at := slices.IndexFunc(named, func(one Workspace) bool { return one.Folder == up.Folder })
	joined := at >= 0
	if joined {
		up.Name = named[at].Name
		if chosen != "" || named[at].Notes != "" {
			up.Notes = named[at].Notes
			if up.Notes == "" {
				up.Notes = filepath.Join(up.Folder, defaultNotesFolder)
			}
		}
	} else if slices.ContainsFunc(named, func(one Workspace) bool { return one.Name == up.Name }) {
		up.Name = UpgradedMain + "-v0-11"
	}
	ownNotes := up.Notes != filepath.Join(up.Folder, defaultNotesFolder)
	rewritten := !joined || (ownNotes && named[at].Notes == "")
	if joined && rewritten {
		short, ok := top.Workspaces[up.Name].(string)
		if !ok {
			return Upgrade{}, ErrNotWritten
		}
		folder = short
		if said, err = WithoutWorkspace(said, up.Name); err != nil {
			return Upgrade{}, err
		}
	}
	if rewritten && !ownNotes {
		if said, err = WithWorkspace(said, up.Name, folder); err != nil {
			return Upgrade{}, err
		}
	}
	if chosen == "" {
		up.Default = true
		if said, err = withDefaultTable(said, up.Name); err != nil {
			return Upgrade{}, err
		}
	}
	if rewritten && ownNotes {
		said = append(said, fmt.Sprintf("\n[%s.%s]\nfolder = %q\nnotes = %q\n", workspacesTable, up.Name, folder, notes)...)
	}
	var check map[string]any
	if toml.Unmarshal(said, &check) != nil {
		return Upgrade{}, ErrNotWritten
	}
	up.Data = said
	return up, nil
}

func withDefaultTable(data []byte, name string) ([]byte, error) {
	if _, _, found := regionOf(linesOf(data), workspacesTable); !found {
		data = []byte(strings.Join(withTable(linesOf(data), workspacesTable), ""))
	}
	return WithDefault(data, name)
}
