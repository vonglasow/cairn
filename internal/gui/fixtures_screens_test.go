package gui

import (
	"cairn/internal/board"
	"fmt"
	"testing"
)

func TestAReferenceInAColumnSayingNothingIsHereStandsBelowWhatItSays(t *testing.T) {
	shown := aFixturesBoard(t)
	for _, g := range shown.Snap.Goals {
		w := shown.onPage(board.Objectives)
		w.Chosen, w.ChosenGoal = g.Number, true
		w = w.Key("t")
		for range max(1, len(w.Narrowing().Segments)) {
			l := w.Layout()
			for _, r := range w.References() {
				for i, column := range l.Column {
					if r.Within != column || w.EmptySays(i) == "" {
						continue
					}
					if empty := EmptyAt(l.CardAt(i, 0, w.scroll[i])); cross(r.Slot, empty) {
						t.Errorf("%s: the reference %d at %+v lies over %q at %+v", fmt.Sprint(g.Title, " ", w.Narrowing().On), r.Card.Number, r.Slot, w.EmptySays(i), empty)
					}
				}
			}
			w = w.Key("]")
		}
	}
}
