package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"testing"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/cucumber/godog"
)

const windowless = "CAIRN_WINDOWLESS"

func offScreen(t *testing.T, wide, high int) *headless.Window {
	t.Helper()
	window, err := headless.NewWindow(wide, high)
	if err == nil {
		return window
	}
	if os.Getenv(windowless) != "" {
		t.Skipf("no headless window here: %v", err)
	}
	t.Fatalf("no headless window here: %v — set %s=1 to skip what draws, on a machine that cannot", err, windowless)
	return nil
}

func drawn(t *testing.T, w Window) *image.RGBA {
	t.Helper()
	shot, err := Shot(w, drawingTheme(), design.Dark())
	if err == nil {
		return shot
	}
	if os.Getenv(windowless) != "" {
		t.Skipf("%v", err)
	}
	t.Fatalf("%v — set %s=1 to skip what draws, on a machine that cannot", err, windowless)
	return nil
}

func drawnInAStep(th *material.Theme, w Window) (*image.RGBA, error) {
	shot, err := Shot(w, th, design.Dark())
	if err != nil && os.Getenv(windowless) != "" {
		return nil, godog.ErrSkip
	}
	if err != nil {
		return nil, fmt.Errorf("%w — set %s=1 to skip what draws, on a machine that cannot", err, windowless)
	}
	return shot, nil
}

func drawsWithoutFault(th *material.Theme, w Window) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("drawing the window panicked on screen %v: %v", w.Screen, r)
		}
	}()
	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Exact(image.Pt(w.width, w.height)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}
	Draw(NewKit(gtx, th, design.Dark(), &Drawn{}), w)
	return nil
}

func ink(shot *image.RGBA, b geom.Box, back color.NRGBA) int {
	count := 0
	for y := b.Y; y < b.Y+b.H; y++ {
		for x := b.X; x < b.X+b.W; x++ {
			got := shot.RGBAAt(x, y)
			if got.R != back.R || got.G != back.G || got.B != back.B {
				count++
			}
		}
	}
	return count
}

// lettered counts the pixels of words in an ink: a face's strokes are anti-aliased, and a thin one never covers a whole pixel, so a pixel within a few steps of the ink counts. Shapes are counted exactly, by painted.
func lettered(shot *image.RGBA, b geom.Box, want color.NRGBA) int {
	close := func(a, b uint8) bool { return a >= b-min(b, 12) && a <= b+min(255-b, 12) }
	count := 0
	for y := b.Y; y < b.Y+b.H; y++ {
		for x := b.X; x < b.X+b.W; x++ {
			if got := shot.RGBAAt(x, y); close(got.R, want.R) && close(got.G, want.G) && close(got.B, want.B) {
				count++
			}
		}
	}
	return count
}

func lit(shot *image.RGBA, b geom.Box, above uint8) int {
	count := 0
	for y := b.Y; y < b.Y+b.H; y++ {
		for x := b.X; x < b.X+b.W; x++ {
			if shot.RGBAAt(x, y).R > above {
				count++
			}
		}
	}
	return count
}

func painted(shot *image.RGBA, b geom.Box, want color.NRGBA) int {
	count := 0
	for y := b.Y; y < b.Y+b.H; y++ {
		for x := b.X; x < b.X+b.W; x++ {
			if got := shot.RGBAAt(x, y); got.R == want.R && got.G == want.G && got.B == want.B {
				count++
			}
		}
	}
	return count
}

func TestWhatWentWrongIsDrawnAlongTheBottom(t *testing.T) {
	f := newFake()
	f.fail = errNope
	w := New(f, 1200, 800).Wonder()

	shot := drawn(t, w)
	if got := ink(shot, WrongBox(w.Layout()), design.Dark().Colour.Surface.Back); got == 0 {
		t.Fatal("nothing was said about a store out of reach")
	}
}

func TestACardChosenInALaneIsDrawnChosen(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Client's answer", State: "blocked"})
	box := w.Layout().LaneCardAt(0, 0, 0)
	edge := design.Dark().Colour.Line.Selected

	plain := painted(drawn(t, w), box, edge)
	chosen := painted(drawn(t, w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})), box, edge)
	if chosen <= plain {
		t.Errorf("a chosen lane card drew %d pixels of outline and an unchosen one %d", chosen, plain)
	}
}

func TestTheCardInMidAirAndItsLandingAreDrawn(t *testing.T) {
	f := newFake()
	w := holding(f,
		board.Card{Number: 1, Title: "First", State: "todo"},
		board.Card{Number: 2, Title: "Second", State: "backlog"})

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(middleOf(w, 1, 0))

	shot := drawn(t, w)
	ghost, _ := w.Ghost()
	if got := painted(shot, ghost, design.Dark().Colour.Surface.Back); got == ghost.W*ghost.H {
		t.Error("the card being dragged is not drawn")
	}
	landing, _ := w.Landing()
	if got := painted(shot, landing, design.Dark().Colour.Accent.Base); got == 0 {
		t.Error("where the card would land is not drawn")
	}
}

func TestAPanelSaysWhatItIsAndDrawsItsKeysOnChips(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	box := Over(w.width, w.height)
	help := w.Key("?")

	shot := drawn(t, help)
	if got := painted(shot, HeadingAt(box), design.Dark().Colour.Ink.Text); got == 0 {
		t.Error("the help does not say what it is")
	}
	name, _ := FieldRow(box, FirstRow(box))
	if got := painted(shot, Chip(name), design.Dark().Colour.Surface.Back); got == 0 {
		t.Error("the first key of the help is not on a chip")
	}
}

func TestTheBarSaysWhatToPress(t *testing.T) {
	if got := Shortcut("New"); got != "cmd+n" {
		t.Errorf("New is offered under %q", got)
	}
	if got := Shortcut("nothing at all"); got != "" {
		t.Errorf("an offer with no key is offered under %q", got)
	}
	if got := Pretty("cmd+n"); got != "⌘N" {
		t.Errorf("the bar writes cmd+n as %q", got)
	}
	if got := Pretty("enter"); got != "⏎" {
		t.Errorf("the bar writes enter as %q", got)
	}
	if got := Pretty("esc"); got != "esc" {
		t.Errorf("the bar writes esc as %q", got)
	}
	if got := Pretty("backspace"); got != "⌫" {
		t.Errorf("the bar writes backspace as %q", got)
	}
	if got := New(newFake(), 1200, 800).OnSystem("darwin").Said("New"); !strings.Contains(got, "⌘N") || !strings.Contains(got, "New") {
		t.Errorf("the bar writes New as %q", got)
	}
	if got := New(newFake(), 1200, 800).Said("nothing at all"); got != "nothing at all" {
		t.Errorf("an offer with no key is written %q, with a gap in front of it", got)
	}
}

func TestTheObjectivesAreCutToTheirRun(t *testing.T) {
	f := newFake()
	var goals []board.Goal
	for i := 1; i <= 12; i++ {
		goals = append(goals, board.Goal{Number: i, Title: "Ship it", State: "live"})
	}
	w := holding(f)
	w.Snap.Goals = goals
	l := w.Layout()
	second := w.TabsAt()[1]
	w = w.clicked(geom.Point{X: second.X + second.W/2, Y: second.Y + second.H/2})
	w = w.Scroll(geom.Point{X: l.Column[0].X + 10, Y: l.Column[0].Y + 10}, 3*geom.CardHeight)

	above := geom.Box{X: l.Column[0].X, Y: l.Tabs.Y + l.Tabs.H, W: l.Column[0].W, H: geom.Gap}
	if got := painted(drawn(t, w), above, design.Dark().Colour.Surface.Raised); got != 0 {
		t.Errorf("%d pixels of objective were drawn above the run they belong in", got)
	}
}

func TestTheWorkUnderAnObjectiveIsDrawnAsTilesInItsBox(t *testing.T) {
	f := newFake()
	var cards []board.Card
	for i := range 30 {
		cards = append(cards, board.Card{
			Number: i + 1, Title: fmt.Sprintf("Task %d", i+1), State: "done",
			Objectives: []int{1},
		})
	}
	w := New(f, 1200, 800)
	w.Snap = board.Snapshot{
		Cards: cards,
		Goals: []board.Goal{{Number: 1, Title: "Ship the desktop board", State: "active", Done: 7, Total: 30}},
	}
	w = w.Key("tab")
	w = w.clicked(goalMiddle(w, 1)).Key("enter")
	if w.Screen != board.OnForm || !w.ChosenGoal {
		t.Fatalf("the window is on %v with a goal chosen: %v", w.Screen, w.ChosenGoal)
	}

	frame := w.workFrame()
	track := w.workBar()
	shot := drawn(t, w)

	check := SubCheck(w.ServingCardAt(0))
	strip := func(b geom.Box) geom.Box { return geom.Box{X: check.X, Y: b.Y, W: check.W, H: b.H} }
	if painted(shot, strip(frame), design.Dark().Colour.Ink.Faint) == 0 {
		t.Error("no tile was drawn in the box beside the form")
	}
	under := geom.Box{X: frame.X, Y: frame.Y + frame.H, W: frame.W, H: geom.SubHeight}
	if got := painted(shot, strip(under), design.Dark().Colour.Ink.Faint); got != 0 {
		t.Errorf("%d pixels of tile ran out of the bottom of the box", got)
	}

	high := geom.Box{X: track.X, Y: track.Y, W: track.W, H: track.H / 4}
	if painted(shot, high, design.Dark().Colour.Line.Edge) == 0 {
		t.Error("the grip is not at the top of a list nobody has scrolled")
	}
	walked := w
	for range w.ServingRows() + 2 {
		walked = walked.Scroll(geom.Point{X: frame.X + 4, Y: frame.Y + 4}, geom.SubHeight)
	}
	scrolled := drawn(t, walked)
	if painted(scrolled, high, design.Dark().Colour.Line.Edge) == painted(shot, high, design.Dark().Colour.Line.Edge) {
		t.Error("the grip did not move when the list was scrolled")
	}
}

func goalMiddle(w Window, number int) geom.Point {
	for at, g := range board.Sorted(w.Showing().Goals) {
		if g.Number == number {
			box := GoalAt(w.Layout(), at, w.goals)
			return geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
		}
	}
	return geom.Point{}
}

func TestAListWithNothingToOfferSaysSo(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.Snap = board.Snapshot{Cards: []board.Card{{Number: 1, Title: "Write the spec", State: "todo"}}}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form.Fields[w.Form.On()].Arg != "objective" {
		w = w.Key("tab")
	}
	w = w.Key("enter")

	box := w.ChoosingBox()
	value := RowIn(box, ChoosingFrom(box), lineHigh)
	if painted(drawn(t, w), value, design.Dark().Colour.Ink.Faint) == 0 {
		t.Error("the list says nothing where it should say there is nothing to serve")
	}
}

func TestAHeldRunIsPaintedBehindTheWords(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.Snap = board.Snapshot{Cards: []board.Card{{Number: 1, Title: "Write the spec", State: "todo"}}}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title")

	box := Over(w.width, w.height)
	fields := w.Form.Fields
	var value geom.Box
	for i, field := range fields {
		if field.Name == "title" {
			value = formValue(box, fields, i)
		}
	}
	inside := Inside(Framed(value, 1))

	loose := ink(drawn(t, w), inside, design.Dark().Colour.Surface.Field)
	held := ink(drawn(t, w.Key("cmd+a")), inside, design.Dark().Colour.Surface.Field)
	if held <= loose {
		t.Errorf("holding the whole field painted %d pixels and letting go %d: "+
			"the run is not drawn", held, loose)
	}
	some := ink(drawn(t, w.Key("shift+left").Key("shift+left")), inside, design.Dark().Colour.Surface.Field)
	if some <= loose || some >= held {
		t.Errorf("a run of two characters painted %d, between %d and %d expected",
			some, loose, held)
	}
}

func TestEveryRunOfCardsSaysHowManyItHolds(t *testing.T) {
	f := newFake()
	w := holding(f,
		board.Card{Number: 1, Title: "Write the spec", State: "blocked"},
		board.Card{Number: 2, Title: "Ship it", State: "blocked"},
		board.Card{Number: 3, Title: "Later", State: "watching"})
	if got := w.LaneSays("blocked"); got != "blocked  2" {
		t.Errorf("the blocked lane says %q", got)
	}
	if got := w.LaneSays("paused"); got != "paused  0" {
		t.Errorf("a lane with nothing in it says %q", got)
	}
	if got := w.LaneSays("todo"); got != "todo  0" {
		t.Errorf("a state with no lane says %q", got)
	}
}

func TestADragDrawsALineInALaneAndMovesNothing(t *testing.T) {
	f := newFake()
	w := holding(f,
		board.Card{Number: 1, Title: "First", State: "blocked"},
		board.Card{Number: 2, Title: "Second", State: "blocked"},
		board.Card{Number: 3, Title: "Third", State: "backlog"})
	l := w.Layout()
	second := l.LaneCardAt(0, 1, 0)
	w = w.PressAt(middleOf(w, 0, 0)).MoveTo(geom.Point{X: second.X + second.W/2, Y: second.Y + second.H/4})

	landing, ok := w.Landing()
	if !ok || landing != LineAbove(second) {
		t.Errorf("the card would land at %+v, want the line above Second", landing)
	}
	slots := map[int]geom.Box{}
	for _, p := range w.Placed() {
		slots[p.Card.Number] = p.Slot
	}
	if slots[1] != l.LaneCardAt(0, 0, 0) || slots[2] != l.LaneCardAt(0, 1, 0) {
		t.Errorf("First is at %+v and Second at %+v: the lane moved", slots[1], slots[2])
	}
}
