package gui

import (
	"fmt"
	"testing"
	"time"

	"cairn/internal/engine"
)

func TestAReferenceWritesWhyItIsThereInTheRoomLeftAfterItsNumber(t *testing.T) {
	c := engine.Card{Number: 54, Labels: []string{"store"}}
	m := holding(newFake()).measured()
	row := ReferenceRowOf(c, "outside this view · blocks #6", Box{X: 10, W: 151, H: lineHigh}, m, time.Time{}, engine.Pages{})
	if row.Facts == "" || row.FactsAt.X < row.NumberAt.X+row.NumberAt.W+2*pillGap {
		t.Errorf("the why %q is written at %+v, over the number at %+v", row.Facts, row.FactsAt, row.NumberAt)
	}
	if row.FactsAt.X+row.FactsAt.W > 10+151 {
		t.Errorf("the why %q at %+v runs past the row", row.Facts, row.FactsAt)
	}
	for i, pill := range row.Pills {
		if pill.X+pill.W > row.FactsAt.X {
			t.Errorf("the pill of %s at %+v runs under the why at %+v", row.Shown[i], pill, row.FactsAt)
		}
	}
}

func TestAReferenceInAColumnSayingNothingIsHereStandsBelowWhatItSays(t *testing.T) {
	board := aFixturesBoard(t)
	for _, g := range board.snap.Goals {
		w := board.onPage(Objectives)
		w.chosen, w.chosenGoal = g.Number, true
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

func TestTheProgressBesideAnObjectiveStopsAboveTheFormsFoot(t *testing.T) {
	work := onTheFixtures(t)
	for _, size := range pinnedSizes {
		board := New(work, size.X, size.Y).At(time.Now()).Look()
		for _, g := range board.snap.Goals {
			w := board.onPage(Objectives)
			w.chosen, w.chosenGoal = g.Number, true
			w = w.Edit()
			along := w.offersAlong(w.formPanel())
			if len(along) == 0 {
				continue
			}
			if box := w.ProgressBox(); box.Y+box.H > along[0].Y-between {
				t.Errorf("at %dx%d, the progress of %s ends at %d, under the foot at %d", size.X, size.Y, g.Title, box.Y+box.H, along[0].Y-between)
			}
		}
	}
}
