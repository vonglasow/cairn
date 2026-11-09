package gui

import (
	"slices"
	"testing"

	"gioui.org/layout"
)

func TestAKitRecordsWhatIsPutOnTheLayerItIsOn(t *testing.T) {
	var drawn Drawn
	k := NewKit(layout.Context{}, Theme(), Dark(), &drawn)
	if k.layer != LayerBoard {
		t.Errorf("a new kit draws on layer %v, not the board", k.layer)
	}
	if on := k.on(LayerPanel); on.layer != LayerPanel || k.layer != LayerBoard {
		t.Errorf("a kit put on a panel is on %v and leaves the first on %v", on.layer, k.layer)
	}
	k.drawn.Put("card", Box{X: 1, W: 2, H: 3}, k.layer)
	k.on(LayerPopover).drawn.Put("menu", Box{X: 5, W: 6, H: 7}, LayerPopover)
	want := []Drawing{{Part: "card", At: Box{X: 1, W: 2, H: 3}, Layer: LayerBoard}, {Part: "menu", At: Box{X: 5, W: 6, H: 7}, Layer: LayerPopover}}
	if got := drawn.All(); !slices.Equal(got, want) {
		t.Errorf("the kit recorded %+v, want %+v", got, want)
	}
}
