package gui

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/engine"
)

func paletteOver(cards ...engine.Card) (*fake, Window) {
	f := newFake()
	return f, holding(f, cards...)
}

func says(rows []PaletteRow) []string {
	var out []string
	for _, row := range rows {
		out = append(out, row.Group+": "+row.Says)
	}
	return out
}

func TestCmdKOpensThePaletteAndEscShutsIt(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	open := w.Key("cmd+k")
	if open.Screen() != OnPalette {
		t.Fatalf("cmd+k left the window on %v", open.Screen())
	}
	if shut := open.Key("esc"); shut.Screen() != OnTheBoard {
		t.Errorf("esc left the window on %v", shut.Screen())
	}
	if shut := open.Key("backspace"); shut.Screen() != OnTheBoard {
		t.Errorf("backspace on nothing typed left the window on %v", shut.Screen())
	}
}

func TestThePaletteOffersTheChosenTasksVerbsFirst(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	rows := w.clicked(middleOf(w, 1, 0)).Key("cmd+k").PaletteRows()
	if len(rows) == 0 || rows[0].Group != paletteActionsGroup || !strings.Contains(rows[0].Says, "#1") {
		t.Fatalf("the palette opens on %v", says(rows))
	}
	for _, want := range []string{"Move #1 to…", "Priority of #1…", "Label #1 with…", "Edit #1"} {
		if !strings.Contains(strings.Join(says(rows), "|"), want) {
			t.Errorf("the palette does not offer %q: %v", want, says(rows))
		}
	}
}

func TestTypingNarrowsThePaletteWordByWord(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	rows := w.clicked(middleOf(w, 1, 0)).Key("cmd+k").Types("to move").PaletteRows()
	if len(rows) != 1 || rows[0].Says != "Move #1 to…" {
		t.Errorf("'to move' leaves %v", says(rows))
	}
}

func TestAMoveAsksForItsColumnThenAsksTheEngine(t *testing.T) {
	f, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Key("cmd+k").Types("move").Key("enter")
	if w.Screen() != OnPalette || len(w.PaletteRows()) != len(engine.States()) {
		t.Fatalf("a move asks %v", says(w.PaletteRows()))
	}
	w = w.Types("done").Key("enter")
	if wrote := f.wrote(); w.Screen() != OnTheBoard || len(wrote) != 1 || wrote[0].in.Given["state"] != "done" {
		t.Errorf("moving to done left the window on %v having asked %+v", w.Screen(), f.asked)
	}
}

func TestAPriorityAndALabelAreAskedForTheSameWay(t *testing.T) {
	f, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
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

func TestThePaletteFindsATaskByTitleAndOpensIt(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"}, engine.Card{Number: 2, Title: "Ship it", State: "todo"})
	w = w.Key("cmd+k").Types("ship")
	at := -1
	for i, row := range w.PaletteRows() {
		if row.Group == paletteTasksGroup && strings.HasSuffix(row.Says, "Ship it") {
			at = i
		}
	}
	if at < 0 {
		t.Fatalf("'ship' finds %v", says(w.PaletteRows()))
	}
	for range at {
		w = w.Key("down")
	}
	if w = w.Key("enter"); w.Screen() != OnForm || w.Chosen() != 2 {
		t.Errorf("opening it left the window on %v with %d chosen", w.Screen(), w.Chosen())
	}
}

func TestANumberFindsItsTaskFirst(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"}, engine.Card{Number: 22, Title: "Ship it", State: "todo"})
	rows := w.Key("cmd+k").Types("#22").PaletteRows()
	if len(rows) == 0 || rows[0].Group != paletteTasksGroup || !strings.HasSuffix(rows[0].Says, "Ship it") {
		t.Errorf("#22 finds %v", says(rows))
	}
}

func TestWhatThePaletteDidIsOfferedAgain(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Key("cmd+k").Types("help").Key("enter").Key("esc")
	rows := w.Key("cmd+k").PaletteRows()
	last := rows[len(rows)-1]
	if last.Group != paletteRecentGroup || last.Says != "Help" {
		t.Errorf("after Help the palette offers %v", says(rows))
	}
}

func TestAClickOnTheBarsLineOpensThePalette(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	line := CommandLine(w.Layout(), w.BarOffers())
	if got := w.clicked(Point{X: line.X + 40, Y: line.Y + line.H/2}); got.Screen() != OnPalette {
		t.Errorf("a click on the bar's line left the window on %v", got.Screen())
	}
}

func TestThePaletteIsDrawnRisingFromTheBar(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Key("cmd+k")
	w = w.At(w.now.Add(time.Second))
	box := w.PaletteBox()
	bar := w.Layout().Bar
	if box.Y+box.H != bar.Y+bar.H || box.X+box.W/2 != w.width/2 {
		t.Errorf("the palette at %+v does not rise from the bar %+v", box, bar)
	}
	if painted(drawn(t, w), Box{X: box.X + box.W/2, Y: box.Y + 4, W: 1, H: 1}, Dark().Panel) != 1 {
		t.Error("the palette is not drawn")
	}
}

func TestASearchSaysHowManyItKeepsOfHowMany(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"}, engine.Card{Number: 2, Title: "Ship it", State: "todo"}, engine.Card{Number: 3, Title: "Spec the bar", State: "todo"})
	if _, ok := w.SearchCount(); ok {
		t.Error("a board nobody searches counts what it keeps")
	}
	w = w.Key("/").Types("spec")
	said, ok := w.SearchCount()
	if !ok || said != "2 of 3" {
		t.Errorf("searching spec counts %q, %v", said, ok)
	}
	pill, _ := w.SearchCountPill()
	if painted(drawn(t, w), Box{X: pill.X + pill.W/2, Y: pill.Y + 1, W: 1, H: 1}, Dark().Pill) != 1 {
		t.Errorf("the count is not drawn as a pill at %+v", pill)
	}
}

func TestThePaletteRisesAsItOpens(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
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
