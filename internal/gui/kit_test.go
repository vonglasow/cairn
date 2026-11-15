package gui

import (
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"slices"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
)

func TestAKitRecordsWhatIsPutOnTheLayerItIsOn(t *testing.T) {
	var drawn Drawn
	k := NewKit(layout.Context{}, drawingTheme(), design.Dark(), &drawn)
	if k.layer != LayerBoard {
		t.Errorf("a new kit draws on layer %v, not the board", k.layer)
	}
	if on := k.on(LayerPanel); on.layer != LayerPanel || k.layer != LayerBoard {
		t.Errorf("a kit put on a panel is on %v and leaves the first on %v", on.layer, k.layer)
	}
	k.drawn.Put("card", geom.Box{X: 1, W: 2, H: 3}, k.layer)
	k.on(LayerPopover).drawn.Put("menu", geom.Box{X: 5, W: 6, H: 7}, LayerPopover)
	want := []Drawing{{Part: "card", At: geom.Box{X: 1, W: 2, H: 3}, Layer: LayerBoard}, {Part: "menu", At: geom.Box{X: 5, W: 6, H: 7}, Layer: LayerPopover}}
	if got := drawn.All(); !slices.Equal(got, want) {
		t.Errorf("the kit recorded %+v, want %+v", got, want)
	}
}

func TestAPartDrawnInsideAnotherIsRecordedAsItsChild(t *testing.T) {
	var drawn Drawn
	k := NewKit(layout.Context{}, drawingTheme(), design.Dark(), &drawn).on(LayerPanel)
	stripe := geom.Box{X: 1, W: 5, H: 9}
	if got := k.inside("card").put("stripe", stripe); got != stripe {
		t.Errorf("putting a part gives back %+v, not its box %+v", got, stripe)
	}
	k.put("card", geom.Box{W: 40, H: 9})
	k.inside("card").inside("label").put("cross", geom.Box{X: 2})
	want := []Drawing{{Part: "card/stripe", At: stripe, Layer: LayerPanel}, {Part: "card", At: geom.Box{W: 40, H: 9}, Layer: LayerPanel}, {Part: "card/label/cross", At: geom.Box{X: 2}, Layer: LayerPanel}}
	if got := drawn.All(); !slices.Equal(got, want) {
		t.Errorf("the kit recorded %+v, want %+v", got, want)
	}
}

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

func TestAPartDrawnWithinAClipIsRecordedAsFarAsItIsSeen(t *testing.T) {
	var drawn Drawn
	var ops op.Ops
	k := NewKit(layout.Context{Ops: &ops}, drawingTheme(), design.Dark(), &drawn)
	column, clipped := k.within(geom.Box{X: 10, Y: 10, W: 100, H: 100})
	lane, inner := column.within(geom.Box{X: 50, Y: 0, W: 100, H: 50})
	lane.put("card", geom.Box{X: 40, Y: 40, W: 40, H: 40})
	inner.Pop()
	column.put("card", geom.Box{X: 0, Y: 100, W: 40, H: 40})
	clipped.Pop()
	k.put("card", geom.Box{X: 0, Y: 100, W: 40, H: 40})
	want := []Drawing{
		{Part: "card", At: geom.Box{X: 50, Y: 40, W: 30, H: 10}},
		{Part: "card", At: geom.Box{X: 10, Y: 100, W: 30, H: 10}},
		{Part: "card", At: geom.Box{X: 0, Y: 100, W: 40, H: 40}},
	}
	if got := drawn.All(); !slices.Equal(got, want) {
		t.Errorf("the kit recorded %+v, want %+v", got, want)
	}
}
