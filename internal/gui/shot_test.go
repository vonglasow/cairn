package gui

import (
	"os"
	"testing"

	"cairn/internal/theme"
)

func TestShot_drawsTheWindowOffScreenInTheSkinGiven(t *testing.T) {
	if os.Getenv(windowless) != "" {
		t.Skipf("%s is set: nothing is drawn here", windowless)
	}
	latte, err := theme.Find("", "catppuccin-latte")
	if err != nil {
		t.Fatal(err)
	}
	skin := Skinned(latte.Colours())
	w := holding(newFake(), card(1, "Write the spec", "todo"))
	shot, err := Shot(w, drawingTheme(), skin)
	if err != nil {
		t.Fatal(err)
	}
	if shot.Bounds().Dx() != w.width || shot.Bounds().Dy() != w.height {
		t.Errorf("the shot is %v, want the window's %dx%d", shot.Bounds(), w.width, w.height)
	}
	if painted(shot, Box{W: 1, H: 1}, skin.Colour.Surface.Back) != 1 {
		t.Error("the shot is not drawn on the back of the skin given")
	}
	if painted(shot, PartsOf(w.Layout().CardAt(1, 0, 0)).Title, skin.Colour.Ink.Text) == 0 {
		t.Error("the card's title is not written in the skin's ink")
	}
}
