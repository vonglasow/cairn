package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"slices"
	"testing"
)

func TestTheCommandBarIsFramedAndItsHintsLitUnderThePointer(t *testing.T) {
	s := design.Dark()
	bar := geom.Box{X: 10, Y: 10, W: 300, H: 32}
	shot, drawn := drawnAlone(t, 330, 60, func(k Kit) {
		Bar{}.Draw(k, bar)
		Hint{Key: "n", Word: "new", Lit: true}.Draw(k, geom.Box{X: 100, Y: bar.Y, W: 80, H: bar.H})
		Hint{Key: "e", Word: "edit", Faint: true}.Draw(k, geom.Box{X: 200, Y: bar.Y, W: 80, H: bar.H})
	})
	if pixel(shot, bar.X+bar.W/2, bar.Y) != s.Colour.Line.Rule || pixel(shot, bar.X+20, bar.Y+bar.H/2) != s.Colour.Surface.Raised {
		t.Error("the bar is not a raised ground edged with a rule")
	}
	if pixel(shot, 100+2, bar.Y+bar.H/2) != s.Colour.Surface.Hover {
		t.Error("a lit hint is not on the hover's ground")
	}
	if lettered(shot, geom.Box{X: 200, Y: bar.Y, W: 80, H: bar.H}, s.Colour.Ink.Faint) == 0 {
		t.Error("a faint hint is not written faint")
	}
	if want := []string{"bar", "hint", "hint/key", "hint", "hint/key"}; !slices.Equal(parts(drawn), want) {
		t.Errorf("the bar recorded %v", parts(drawn))
	}
}

func TestAlertAt_standsInsideItsTabClearOfTheEdge(t *testing.T) {
	tab := geom.Box{X: 100, Y: 10, W: 120, H: 32}
	if got := AlertAt(tab); got.X+got.W > tab.X+tab.W-tabSlack || got.X < tab.X {
		t.Errorf("the alert stands at %v in the tab %v, closer than %d to its edge", got, tab, tabSlack)
	}
}

func TestATabsCountStandsOnItsNamesBaseline(t *testing.T) {
	w := homeWindow(1200, 800)
	w.Tab = board.Objectives
	shot := drawn(t, w)
	tab := w.TabsAt()[0]
	room := TabName(tab)
	split := room.X + ends(w.measured(), Titled(Pages[board.Tasks]), reads) + countGap/2
	lowest := func(from, to int) int {
		low := -1
		for y := tab.Y; y < tab.Y+tab.H; y++ {
			for x := from; x < to; x++ {
				if c := shot.RGBAAt(x, y); int(c.R)+int(c.G)+int(c.B) > 200 {
					low = y
				}
			}
		}
		return low
	}
	if name, count := lowest(room.X, split), lowest(split, tab.X+tab.W); name != count {
		t.Errorf("the name's foot is at %d and the count's at %d", name, count)
	}
}
