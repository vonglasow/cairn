package gui

import (
	"testing"
)

func TestTheHelpIsClosedWhereItsCloseIsDrawn(t *testing.T) {
	middle := func(b Box) Point { return Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	open := New(newFake(), 1600, 2100).Key("?")
	if open.Screen() != OnHelp {
		t.Fatalf("the help did not open: %v", open.Screen())
	}
	box := open.HelpBox()
	if got := open.clicked(middle(open.offersAlong(box)[0])); got.Screen() != OnTheBoard {
		t.Errorf("clicking Close where it is drawn left the window on %v", got.Screen())
	}
	under := Point{X: box.X + box.W/2, Y: box.Y + box.H + apart}
	if got := open.clicked(under); got.Screen() != OnTheBoard {
		t.Errorf("clicking under the help left the window on %v", got.Screen())
	}
}

func TestTheHelpPutsEachKeyOnAChipBeforeWhatItDoes(t *testing.T) {
	open := New(newFake(), 1200, 800).Key("?")
	shot := drawn(t, open)
	keys := 0
	for _, line := range open.HelpLines() {
		if line.Head != "" {
			continue
		}
		keys++
		if line.KeyAt.W != KeyChipWide(line.Key) || line.MeansAt.X <= line.KeyAt.X+line.KeyAt.W {
			t.Errorf("%q: its chip %+v and what it does %+v", line.Key, line.KeyAt, line.MeansAt)
		}
		if painted(shot, Box{X: line.KeyAt.X + line.KeyAt.W/2, Y: line.KeyAt.Y + 1, W: 1, H: 1}, Dark().Pill) != 1 {
			t.Errorf("%q is not on a chip", line.Key)
		}
	}
	if keys != len(open.Helping()) {
		t.Errorf("the help lays out %d keys of its %d", keys, len(open.Helping()))
	}
}

func TestTheHelpBalancesItsColumns(t *testing.T) {
	open := New(newFake(), 1200, 800).Key("?")
	per := map[int]int{}
	for _, line := range open.HelpLines() {
		per[line.At.X]++
	}
	var counts []int
	for _, n := range per {
		counts = append(counts, n)
	}
	if len(counts) != 2 || counts[0]-counts[1] > 3 || counts[1]-counts[0] > 3 {
		t.Errorf("the help's columns hold %v lines", counts)
	}
}
