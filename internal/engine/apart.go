package engine

import (
	"cairn/internal/engine/workspace"
)

// OpenApart says what a window of its own on another workspace, folder or store would hold, found from the home of the engine held, which is left holding what it held; what is not there is refused as Switch refuses it.
func OpenApart(held *Engine, to Elsewhere) (workspace.Apart, error) {
	h, err := to.from(held.placed())
	if err != nil {
		return workspace.Apart{}, err
	}
	at, err := Locate(h)
	if err != nil {
		return workspace.Apart{}, err
	}
	if to.Workspace != "" {
		return workspace.Apart{Store: at.Store, Flags: []string{workspace.WorkspaceFlagName, to.Workspace}}, nil
	}
	return workspace.Apart{Store: at.Store, Flags: []string{workspace.StoreFlagName, at.Store}}, nil
}
