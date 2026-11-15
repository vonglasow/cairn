package engine

import (
	"cairn/internal/engine/workspace"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"time"

	"cairn/internal/store/files"
)

func recentPath(h Home) string {
	if state, ok := h.lookup()(workspace.StateVariable); ok && filepath.IsAbs(state) {
		return filepath.Join(state, "cairn", "recent.json")
	}
	if h.Dir == "" {
		return ""
	}
	return filepath.Join(h.Dir, ".local", "state", "cairn", "recent.json")
}

func remember(h Home, store string, at time.Time) error {
	path := recentPath(h)
	if path == "" {
		return nil
	}
	kept := slices.DeleteFunc(workspace.RememberedIn(path), func(one workspace.Remembered) bool { return one.Store == store })
	kept = slices.Insert(kept, 0, workspace.Remembered{Store: store, Opened: at})
	said, err := json.MarshalIndent(workspace.RecentFile{Recent: kept[:min(len(kept), workspace.RecentKept)]}, "", "  ")
	if err != nil {
		return err
	}
	return files.Kept(path).Keep(append(said, '\n'))
}

func (h Home) openedByPath() bool {
	_, a, err := h.settled()
	return err == nil && a.Workspace == "" && (a.Store.By == workspace.StoreFlagName || a.Store.By == workspace.ByTheEnvironment)
}

func recentIn(h Home, p files.Places) []workspace.RecentFolder {
	var listed []workspace.RecentFolder
	for _, one := range workspace.RememberedIn(recentPath(h)) {
		if slices.ContainsFunc(p.Workspaces(), func(w files.Workspace) bool { return filepath.Join(w.Folder, workspace.StoreFile) == one.Store }) {
			continue
		}
		folder := ""
		if filepath.Base(one.Store) == workspace.StoreFile {
			folder = filepath.Dir(one.Store)
		}
		_, err := os.Stat(one.Store)
		listed = append(listed, workspace.RecentFolder{Store: one.Store, Folder: folder, Opened: one.Opened, NotFound: err != nil})
	}
	return listed
}
