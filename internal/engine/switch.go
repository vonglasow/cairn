package engine

import (
	"cairn/internal/engine/workspace"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"cairn/internal/store/files"
)

// Elsewhere is where a board is asked to go in place of the store it holds: a workspace by its name, a folder typed, or a store a board opened before.
type Elsewhere struct {
	Workspace string
	Folder    string
	Store     string
}

func (to Elsewhere) from(h Home) (Home, error) {
	h.Workspace, h.Store, h.Kept = to.Workspace, to.Store, nil
	if to.Folder != "" {
		at, err := files.Absolute(to.Folder, h.Dir)
		if err != nil {
			return h, err
		}
		if info, err := os.Stat(at); err != nil || !info.IsDir() {
			return h, fmt.Errorf("%s %w", at, workspace.ErrNotThere)
		}
		h.Store = filepath.Join(at, workspace.StoreFile)
	}
	if _, err := os.Stat(h.Store); to.Store != "" && err != nil {
		return h, fmt.Errorf("%s %w", to.Store, workspace.ErrNotThere)
	}
	return h, nil
}

// Switch holds the store of another workspace, folder or store in the engine a board holds, in place of the one it held: found from the same home, acting as before, the one it held closed once the other is open, and said with the pages the settings of the store now held show. The store held stays when the other cannot be opened, and the configuration is not written.
func Switch(held *Engine, to Elsewhere) (Pages, error) {
	h, err := to.from(held.placed())
	if err != nil {
		return Pages{}, err
	}
	next, err := Hold(h, held.acting)
	if err != nil {
		return Pages{}, err
	}
	unlock := held.taking()
	was := Engine{tasks: held.tasks, refused: held.refused, turn: &sync.Mutex{}}
	held.tasks, held.places, held.history, held.objectives, held.store = next.tasks, next.places, next.history, next.objectives, next.store
	held.notes, held.home, held.path, held.followed, held.sweepOwed, held.refused = next.notes, next.home, next.path, next.followed, next.sweepOwed, next.refused
	unlock()
	return PagesFrom(h), was.Close()
}

func (e *Engine) placed() Home {
	defer e.taking()()
	return e.home
}
