package gui

import (
	"bytes"
	"image/png"
	"testing"
)

func TestTheLogoIsDrawnLargeWithItsCornersClear(t *testing.T) {
	drawn, err := png.Decode(bytes.NewReader(Logo()))
	if err != nil {
		t.Fatal(err)
	}
	if b := drawn.Bounds(); b.Dx() != 1024 || b.Dy() != 1024 {
		t.Errorf("the logo is %v", b)
	}
	if _, _, _, a := drawn.At(0, 0).RGBA(); a != 0 {
		t.Error("the logo's corner is not clear")
	}
	if Known != "cairn" {
		t.Errorf("the window is known as %q", Known)
	}
}
