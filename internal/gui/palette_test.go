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

func TestThePaletteIsDrawnUnderTheTabs(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Key("cmd+k")
	w = w.At(w.now.Add(time.Second))
	box, l := w.PaletteBox(), w.Layout()
	if box.Y != l.Tabs.Y+l.Tabs.H || box.X+box.W/2 != w.width/2 || box.Y+box.H > l.Bar.Y {
		t.Errorf("the palette at %+v is not under the tabs %+v, above the bar %+v", box, l.Tabs, l.Bar)
	}
	if painted(drawn(t, w), Box{X: box.X + box.W/2, Y: box.Y + 4, W: 1, H: 1}, Dark().Panel) != 1 {
		t.Error("the palette is not drawn")
	}
}

func TestEveryActionOfThePaletteSaysItsKey(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Key("cmd+k")
	keys := map[string]string{}
	for _, row := range w.PaletteRows() {
		if row.Group == paletteActionsGroup {
			keys[row.Says] = row.Key
		}
	}
	for says, key := range map[string]string{"Move #1 to…": "m", "Edit #1": "e", "Delete #1": "cmd+backspace", "New task": "cmd+n", "Search": "/", "Help": "?"} {
		if keys[says] != key {
			t.Errorf("%q says the key %q, not %q", says, keys[says], key)
		}
	}
}

func TestASearchSaysHowManyItKeepsOfHowMany(t *testing.T) {
	_, w := paletteOver(engine.Card{Number: 1, Title: "Write the spec", State: "todo"}, engine.Card{Number: 2, Title: "Ship it", State: "todo"}, engine.Card{Number: 3, Title: "Spec the bar", State: "todo"})
	if got := w.TabCount(Tasks); got != "3" {
		t.Errorf("a board nobody searches counts %q", got)
	}
	w = w.Key("/").Types("spec")
	if got := w.TabCount(Tasks); got != "2/3" {
		t.Errorf("searching spec counts %q", got)
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

func TestTheCommandBarsChosenRowSaysEnterOnAChip(t *testing.T) {
	w := holding(newFake(), card(1, "Write the spec", "todo")).Key("cmd+k")
	w = w.At(w.now.Add(time.Second))
	lines := w.PaletteLines()
	var chosen PaletteLine
	for _, line := range lines {
		if line.Row == 0 {
			chosen = line
		}
	}
	chip := PaletteKeyAt(chosen.Box, w.keyboard.Spelled(Pretty("enter")))
	if chip.X+chip.W > chosen.Box.X+chosen.Box.W || chip.W == 0 {
		t.Fatalf("the chip %+v is not inside the row %+v", chip, chosen.Box)
	}
	if painted(drawn(t, w), Box{X: chip.X + chip.W/2, Y: chip.Y + 1, W: 1, H: 1}, darker(Dark().Pill)) != 1 {
		t.Error("the chosen row draws no chip for enter")
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
	if got := rows[len(rows)-1].Ago(w.now); got != "2 min ago" {
		t.Errorf("a line done two minutes ago says %q", got)
	}
	if got := rows[len(rows)-1].Ago(start.Add(10 * time.Second)); got != "just now" {
		t.Errorf("a line done ten seconds ago says %q", got)
	}
}
