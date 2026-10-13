package gui

import (
	"path/filepath"
	"testing"
)

func TestTheWindowHoldsNoObjectiveCFile(t *testing.T) {
	files, err := filepath.Glob("*.m")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) > 0 {
		t.Errorf("the window holds %v", files)
	}
}
