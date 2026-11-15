package gui

import (
	"cairn/internal/gui/geom"
	"slices"
	"testing"
)

func TestOverlapsAreSiblingsOnOneLayerWhoseBoxesCross(t *testing.T) {
	var drawn Drawn
	drawn.Put("card", geom.Box{W: 40, H: 20}, LayerBoard)
	drawn.Put("card/number", geom.Box{X: 2, W: 20, H: 10}, LayerBoard)
	drawn.Put("card/facts", geom.Box{X: 15, W: 20, H: 10}, LayerBoard)
	drawn.Put("card/label", geom.Box{X: 22, Y: 10, W: 10, H: 10}, LayerBoard)
	drawn.Put("card/ring", geom.Box{W: 40, H: 20}, LayerBoard)
	drawn.Put("card", geom.Box{Y: 20, W: 40, H: 20}, LayerBoard)
	drawn.Put("card/stripe", geom.Box{Y: 20, W: 5, H: 20}, LayerBoard)
	drawn.Put("card", geom.Box{Y: 30, W: 40, H: 20}, LayerBoard)
	drawn.Put("card/stripe", geom.Box{Y: 30, W: 5, H: 20}, LayerBoard)
	drawn.Put("popover", geom.Box{W: 40, H: 40}, LayerPopover)
	drawn.Put("hint", geom.Box{X: 40, W: 0, H: 20}, LayerBoard)
	got := drawn.Overlaps()
	want := [][2]Drawing{
		{{Part: "card/number", At: geom.Box{X: 2, W: 20, H: 10}}, {Part: "card/facts", At: geom.Box{X: 15, W: 20, H: 10}}},
		{{Part: "card", At: geom.Box{Y: 20, W: 40, H: 20}}, {Part: "card", At: geom.Box{Y: 30, W: 40, H: 20}}},
	}
	if !slices.Equal(got, want) {
		t.Errorf("overlaps %+v, want %+v", got, want)
	}
}
