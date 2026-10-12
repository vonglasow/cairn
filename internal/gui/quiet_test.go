package gui

import (
	"path/filepath"
	"testing"
)

// The window holds no Objective-C file: cmd/go links -lobjc once for every
// package holding one, Gio's own is already one, and a second makes the
// linker say so on every build. What the window says to AppKit it says from
// a cgo preamble instead. See `make quiet-check`.
func TestTheWindowHoldsNoObjectiveCFile(t *testing.T) {
	files, err := filepath.Glob("*.m")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) > 0 {
		t.Errorf("the window holds %v", files)
	}
}
