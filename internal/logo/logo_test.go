package logo

import (
	"bytes"
	"image/png"
	"testing"
)

func TestAt_givesTheLogoAtEverySizeItWasDrawnAt(t *testing.T) {
	for _, size := range []int{16, 32, 48, 64, 128, 256, 512, 1024} {
		drawn, err := At(size)
		if err != nil {
			t.Fatalf("%d: %v", size, err)
		}
		read, err := png.DecodeConfig(bytes.NewReader(drawn))
		if err != nil || read.Width != size || read.Height != size {
			t.Errorf("%d: %dx%d, %v", size, read.Width, read.Height, err)
		}
	}
	if _, err := At(100); err == nil {
		t.Error("a size it was not drawn at was given")
	}
}
