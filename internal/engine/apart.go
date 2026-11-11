package engine

import "path/filepath"

// Apart is how another window is asked to open on a workspace, a folder or a store: the store it would hold, and the flags naming it to a cairn started for it.
type Apart struct {
	Store string
	Flags []string
}

// OpenApart says what a window of its own on another workspace, folder or store would hold, found from the home of the engine held, which is left holding what it held; what is not there is refused as Switch refuses it.
func OpenApart(held *Engine, to Elsewhere) (Apart, error) {
	h, err := to.from(held.placed())
	if err != nil {
		return Apart{}, err
	}
	at, err := Locate(h)
	if err != nil {
		return Apart{}, err
	}
	if to.Workspace != "" {
		return Apart{Store: at.Store, Flags: []string{workspaceFlagName, to.Workspace}}, nil
	}
	return Apart{Store: at.Store, Flags: []string{storeFlagName, at.Store}}, nil
}

// Store is the store of the workspace: the cairn.db of its folder.
func (w Workspace) Store() string { return filepath.Join(w.Folder, storeFile) }

// Name is how the place a board shows is named to a person, short, never a whole path: the workspace that names its store, the name the store's settings give it, or the folder holding the store.
func (s WorkspaceShown) Name() string {
	switch {
	case s.Workspace != "":
		return s.Workspace
	case s.Own != "":
		return s.Own
	case s.Store == "":
		return ""
	}
	return filepath.Base(filepath.Dir(s.Store))
}

// Cost is what a card says it cost against what it was reckoned to cost, and whether that is over, when the time is shown; nothing when it is hidden or off.
func (p Pages) Cost(c Card) (string, bool) {
	if !p.Shows(PageTime) {
		return "", false
	}
	return c.Cost()
}
