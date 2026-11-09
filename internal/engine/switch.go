package engine

import (
	"os"
	"sync"
)

// Switch holds the store of another workspace in the engine a board holds, in place of the one it held: found from the same home, acting as before, the one it held closed once the other is open, and said with the pages the settings of the store now held show. The store held stays when the other cannot be opened, and the configuration is not written.
func Switch(held *Engine, workspace string) (Pages, error) {
	h := held.placed()
	h.Workspace, h.Store, h.Kept = workspace, "", nil
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

// Found says whether the folder of the workspace is there.
func (w Workspace) Found() bool {
	_, err := os.Stat(w.Folder)
	return err == nil
}
