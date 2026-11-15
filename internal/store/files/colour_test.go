package files_test

import (
	"errors"
	"testing"

	"cairn/internal/store/files"
)

func TestWithoutWorkspace_refusesWhatItCannotChange(t *testing.T) {
	for _, was := range []string{
		"",
		"[pages]\nobjectives = false\n",
		"workspaces = { home = \"~/h\" }\n",
		"[workspaces]\n\"home\" = \"~/h\"\n",
		"[workspaces]\nwork = \"~/w\"\n",
	} {
		if _, err := files.WithoutWorkspace([]byte(was), "home"); !errors.Is(err, files.ErrNotWritten) {
			t.Errorf("WithoutWorkspace(%q) = %v; want %v", was, err, files.ErrNotWritten)
		}
	}
}
