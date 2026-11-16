package gui

import (
	"testing"

	"cairn/internal/board"
	"cairn/internal/engine/workspace"
)

func TestTheNewWorkspaceDialogDrawsNoPartOverAnother(t *testing.T) {
	f := configuredWith(workspace.Config{})
	f.shown = &workspace.Shown{Workspace: "fixtures", Store: "/h/.cairn/cairn.db", Colour: "peach"}
	for _, size := range [][2]int{{720, 560}, {1200, 800}} {
		w := holding(f, card(1, "Fix the garden gate", "backlog")).Resized(size[0], size[1]).Look()
		w.Screen = board.OnWorkspaces
		w = w.openingNewSpace().Types("garden")
		w.switcher.making.said = `"My Garden": a workspace is named in lowercase letters, digits and dashes, and is not called default`
		drawn := Drawn{all: partsDrawn(w)}
		for _, c := range drawn.Overlaps() {
			t.Errorf("at %v, %s %+v over %s %+v", size, c[0].Part, c[0].At, c[1].Part, c[1].At)
		}
		for _, c := range drawn.Strays() {
			t.Errorf("at %v, %s %+v runs out of %s %+v", size, c[0].Part, c[0].At, c[1].Part, c[1].At)
		}
	}
}
