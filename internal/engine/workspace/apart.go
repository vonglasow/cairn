package workspace

import "path/filepath"

// Apart is how another window is asked to open on a workspace, a folder or a store: the store it would hold, and the flags naming it to a cairn started for it.
type Apart struct {
	Store string
	Flags []string
}

// Store is the store of the workspace: the cairn.db of its folder.
func (w Workspace) Store() string { return filepath.Join(w.Folder, StoreFile) }
