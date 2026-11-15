package gui

import (
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"slices"
	"testing"
)

func TestAListItemIsIndentedByItsDepthAndJoinedToItsParent(t *testing.T) {
	slot := geom.Box{X: 10, Y: 10, W: 400, H: ItemHigh}
	_, top := drawnAlone(t, 440, 80, func(k Kit) {
		ListItem{Number: "#4", Title: "Home automation", Standing: "active", Counts: "1/2 · 50%", Done: 1, Total: 2}.Draw(k, slot)
	})
	if got := parts(top); !slices.Equal(got, []string{"item", "item/stripe", "item/standing", "item/progress"}) {
		t.Errorf("a level-one item recorded %v", got)
	}
	if top[0].At != ItemOf(slot, 0) {
		t.Errorf("a level-one item stands at %+v, want %+v", top[0].At, ItemOf(slot, 0))
	}
	_, under := drawnAlone(t, 440, 80, func(k Kit) {
		ListItem{Number: "#5", Title: "Enclosure", Depth: 1, Last: true, Chip: "project"}.Draw(k, slot)
	})
	if got := parts(under); !slices.Equal(got, []string{"connector", "item", "item/stripe", "item/label"}) {
		t.Errorf("a sub-objective's item recorded %v", got)
	}
	if item := ItemOf(slot, 1); item.X != slot.X+indent || under[1].At != item {
		t.Errorf("a sub-objective's item stands at %+v, want %d in", under[1].At, indent)
	}
	if connector := under[0].At; connector.X+connector.W != under[1].At.X {
		t.Errorf("the connector %+v does not reach the item %+v", connector, under[1].At)
	}
}

func TestACardInTileLayoutSaysItsCountsItsNotesAndItsWarning(t *testing.T) {
	s := design.Dark()
	box := geom.Box{X: 10, Y: 10, W: 220, H: TileHigh}
	tile := &GoalTile{Number: "#7", Title: "Features", State: "on-hold", Chip: "project", Done: 1, Total: 3, Counts: "1 of 3 · 33%",
		Notes: []string{"#44 also in Firmware"}, Warn: "waits on Firmware"}
	shot, drawn := drawnAlone(t, 250, 150, func(k Kit) { Card{Goal: tile, Chosen: true}.Draw(k, box) })
	if lettered(shot, box, s.Colour.State.Late) == 0 {
		t.Error("the tile's warning is not in the late ink")
	}
	want := []string{"card", "card/ring", "card/stripe", "card/standing", "card/label", "card/progress"}
	if got := parts(drawn); !slices.Equal(got, want) {
		t.Errorf("a chosen tile recorded %v, want %v", got, want)
	}
}
