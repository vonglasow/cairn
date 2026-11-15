package logo

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"
)

func TestAt_isTheStencilCardOnTheWindowsGround(t *testing.T) {
	drawn, err := At(1024)
	if err != nil {
		t.Fatal(err)
	}
	read, err := png.Decode(bytes.NewReader(drawn))
	if err != nil {
		t.Fatal(err)
	}
	ground := color.NRGBA{R: 0x0C, G: 0x0D, B: 0x0F, A: 255}
	for what, at := range map[string]struct {
		x, y int
		want color.NRGBA
	}{
		"the window's ground": {112, 512, ground},
		"the spine":           {300, 512, color.NRGBA{R: 0xE7, G: 0xE7, B: 0xEA, A: 255}},
		"the piece lit":       {642, 692, color.NRGBA{R: 0x7B, G: 0x93, B: 0xFF, A: 255}},
		"the counter":         {600, 512, ground},
	} {
		if got := color.NRGBAModel.Convert(read.At(at.x, at.y)).(color.NRGBA); got != at.want {
			t.Errorf("%s is %v, want %v", what, got, at.want)
		}
	}
}
