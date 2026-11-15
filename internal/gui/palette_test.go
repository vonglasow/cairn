package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"testing"
	"time"
)

func paletteOver(cards ...board.Card) (*fake, Window) {
	f := newFake()
	return f, holding(f, cards...)
}

func TestCmdKOpensThePaletteAndEscShutsIt(t *testing.T) {
	_, w := paletteOver(board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	open := w.Key("cmd+k")
	if open.Screen != board.OnPalette {
		t.Fatalf("cmd+k left the window on %v", open.Screen)
	}
	if shut := open.Key("esc"); shut.Screen != board.OnTheBoard {
		t.Errorf("esc left the window on %v", shut.Screen)
	}
	if shut := open.Key("backspace"); shut.Screen != board.OnTheBoard {
		t.Errorf("backspace on nothing typed left the window on %v", shut.Screen)
	}
}

func TestAPriorityAndALabelAreAskedForTheSameWay(t *testing.T) {
	f, w := paletteOver(board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	chosen := w.clicked(middleOf(w, 1, 0))
	chosen.Key("cmd+k").Types("priority").Key("enter").Types("high").Key("enter")
	if got := f.wrote()[0].in.Given["priority"]; got != "high" {
		t.Errorf("the priority asked is %q", got)
	}
	chosen.Key("cmd+k").Types("label").Key("enter").Types("backend").Key("enter")
	if got := f.wrote()[1].in.Given["label"]; got != "+backend" {
		t.Errorf("the label asked is %q", got)
	}
}

func TestAClickOnTheBarsLineOpensThePalette(t *testing.T) {
	_, w := paletteOver(board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	line := CommandLine(w.Layout(), w.BarOffers())
	if got := w.clicked(geom.Point{X: line.X + 40, Y: line.Y + line.H/2}); got.Screen != board.OnPalette {
		t.Errorf("a click on the bar's line left the window on %v", got.Screen)
	}
}

func TestThePaletteRisesAsItOpens(t *testing.T) {
	_, w := paletteOver(board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	opened := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	w = w.At(opened).Key("cmd+k")
	if rise, rising := w.At(opened.Add(paletteRises / 2)).PaletteRise(); !rising || rise <= 0.4 || rise >= 0.95 {
		t.Errorf("halfway through, the palette has risen %v, rising %v", rise, rising)
	}
	if rise, rising := w.At(opened.Add(2 * paletteRises)).PaletteRise(); rising || rise != 1 {
		t.Errorf("long after, the palette has risen %v, rising %v", rise, rising)
	}
	if _, rising := w.Key("esc").At(opened.Add(paletteRises / 2)).PaletteRise(); rising {
		t.Error("a palette shut still rises")
	}
}

func TestARecentLineSaysHowLongAgoItWasDone(t *testing.T) {
	start := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	w := holding(newFake(), card(1, "Write the spec", "todo")).At(start).Key("cmd+k").Types("help").Key("enter").Key("esc")
	w = w.At(start.Add(2 * time.Minute)).Key("cmd+k")
	rows := w.PaletteRows()
	if len(rows) == 0 || rows[len(rows)-1].Group != paletteRecentGroup {
		t.Fatalf("the command bar lists %+v", rows)
	}
	if got := rows[len(rows)-1].Ago(w.Now); got != "2 min ago" {
		t.Errorf("a line done two minutes ago says %q", got)
	}
	if got := rows[len(rows)-1].Ago(start.Add(10 * time.Second)); got != "just now" {
		t.Errorf("a line done ten seconds ago says %q", got)
	}
}
