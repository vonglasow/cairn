package gui

import (
	"cairn/internal/board"
	"image/color"
	"testing"
)

func TestAboutCairnPaintsTheLogo(t *testing.T) {
	w := holding(newFake())
	w.Screen, w.Question = board.OnAsking, board.AboutCairn
	shot := drawn(t, w)
	if got := painted(shot, w.AboutLogo(), color.NRGBA{R: 0x7B, G: 0x93, B: 0xFF, A: 255}); got == 0 {
		t.Error("About cairn paints no logo")
	}
}
