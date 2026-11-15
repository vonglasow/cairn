package workspace

import (
	"errors"
	"os"
)

// ErrNotThere is returned when a board is asked to open a folder, or a store, that is not there: opening it would start an empty board where nobody looks.
var ErrNotThere = errors.New("is not there")

// Found says whether the folder of the workspace is there.
func (w Workspace) Found() bool {
	_, err := os.Stat(w.Folder)
	return err == nil
}
