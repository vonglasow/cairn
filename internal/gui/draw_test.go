package gui

import (
	"cairn/internal/engine"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"testing"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/io/key"
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
	shot, err := rendered(offScreen(t, w.width, w.height), w)
	if err != nil {
		t.Fatal(err)
	}
	return shot
}

func drawnInAStep(w Window) (*image.RGBA, error) {
	window, err := headless.NewWindow(w.width, w.height)
	if err != nil {
		if os.Getenv(windowless) != "" {
			return nil, godog.ErrSkip
		}
		return nil, fmt.Errorf("no headless window here: %w — set %s=1 to skip what draws, on a machine that cannot", err, windowless)
	}
	return rendered(window, w)
}

func drawsWithoutFault(th *material.Theme, w Window) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("drawing the window panicked on screen %v: %v", w.Screen(), r)
		}
	}()
	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Exact(image.Pt(w.width, w.height)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}
	Draw(gtx, th, Dark(), w)
	return nil
}

func rendered(window *headless.Window, w Window) (*image.RGBA, error) {
	defer window.Release()
	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Exact(image.Pt(w.width, w.height)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}
	Draw(gtx, Theme(), Dark(), w)
	if err := window.Frame(gtx.Ops); err != nil {
		return nil, fmt.Errorf("drawing: %w", err)
	}
	shot := image.NewRGBA(image.Rect(0, 0, w.width, w.height))
	if err := window.Screenshot(shot); err != nil {
		return nil, fmt.Errorf("reading the pixels back: %w", err)
	}
	return shot, nil
}

func ink(shot *image.RGBA, b Box, back color.NRGBA) int {
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

func TestTheBoardIsDrawnWhereTheLayoutSaysItIs(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "Write the spec", State: "todo", Priority: "high"},
		engine.Card{Number: 2, Title: "Client's answer", State: "blocked"})

	shot := drawn(t, w)
	back := Dark().Back
	l := w.Layout()

	for what, box := range map[string]Box{
		"the tabs":            l.Tabs,
		"the todo header":     l.Head[1],
		"the first todo card": l.CardAt(1, 0, 0),
		"the blocked lane":    l.Lane[0],
		"the bar":             l.Bar,
	} {
		if got := ink(shot, box, back); got == 0 {
			t.Errorf("%s is blank", what)
		}
	}

	empty := l.CardAt(3, 0, 0)
	if got := ink(shot, empty, back); got != 0 {
		t.Errorf("an empty column drew %d pixels where its first card would be", got)
	}
}

func TestAWindowTooSmallDrawsItsMessageAndNoBoard(t *testing.T) {
	f := newFake()
	w := New(f, 400, 300)

	shot := drawn(t, w)
	back := Dark().Back
	if got := ink(shot, Box{W: 400, H: 300}, back); got == 0 {
		t.Fatal("a window too small drew nothing at all")
	}
	bottom := Box{X: 0, Y: 290, W: 400, H: 10}
	if got := ink(shot, bottom, back); got != 0 {
		t.Errorf("a window too small drew %d pixels of board along its bottom", got)
	}
}

func inked(shot *image.RGBA, b Box, ground color.NRGBA) int {
	count := 0
	for y := b.Y; y < b.Y+b.H; y++ {
		for x := b.X; x < b.X+b.W; x++ {
			if got := shot.RGBAAt(x, y); got.R != ground.R || got.G != ground.G || got.B != ground.B {
				count++
			}
		}
	}
	return count
}

func painted(shot *image.RGBA, b Box, want color.NRGBA) int {
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

func TestTheBarUnderTheBoardIsFaintUnderAPanel(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	skin := Dark()
	first := OfferAt(w.Layout(), 0)
	for _, c := range []struct {
		what   string
		screen Screen
		lit    bool
	}{
		{"the board", OnTheBoard, true},
		{"a search", OnSearch, true},
		{"an open task", OnForm, false},
		{"a form", OnForm, false},
	} {
		v := w
		v.chosen, v.screen = 1, c.screen
		if c.screen == OnForm {
			v = v.Edit()
		}
		shot := drawn(t, v)
		lit, faint := painted(shot, first, skin.Chosen), painted(shot, first, skin.Faint)
		if c.lit && (lit == 0 || faint != 0) {
			t.Errorf("over %s the bar has %d lit and %d faint", c.what, lit, faint)
		}
		if !c.lit && (lit != 0 || faint == 0) {
			t.Errorf("under %s the bar has %d lit and %d faint", c.what, lit, faint)
		}
	}
}

func TestAChosenCardIsDrawnWithAnOutline(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	box := w.Layout().CardAt(1, 0, 0)
	edge := Dark().Chosen

	plain := painted(drawn(t, w), box, edge)
	chosen := painted(drawn(t, w.clicked(middleOf(w, 1, 0))), box, edge)
	if chosen <= plain {
		t.Errorf("a chosen card drew %d pixels of its outline and an unchosen one %d", chosen, plain)
	}
}

func TestTheFormIsDrawnOverTheBoard(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	shot := drawn(t, w)
	if got := ink(shot, Over(w.width, w.height), Dark().Back); got == 0 {
		t.Fatal("the form is blank")
	}
	l := w.Layout()
	if got := ink(shot, l.Head[0], Dark().Back); got == 0 {
		t.Error("the form hid the board behind it rather than sitting over it")
	}
}

func TestWhatWentWrongIsDrawnAlongTheBottom(t *testing.T) {
	f := newFake()
	f.fail = errNope
	w := New(f, 1200, 800).Wonder()

	shot := drawn(t, w)
	if got := ink(shot, WrongBox(w.Layout()), Dark().Back); got == 0 {
		t.Fatal("nothing was said about a store out of reach")
	}
}

func TestWhatWentWrongLeavesTheBarDrawn(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	f.fail = errNope
	w = w.Wonder()
	if w.Wrong() == nil {
		t.Fatal("nothing went wrong to be drawn")
	}
	shot := drawn(t, w)
	if got := ink(shot, WrongBox(w.Layout()), Dark().Back); got == 0 {
		t.Error("what went wrong is not drawn")
	}
	for i := range w.Offers() {
		if got := ink(shot, OfferAt(w.Layout(), i), Dark().Back); got == 0 {
			t.Errorf("offer %d is not drawn while something went wrong", i)
		}
	}
}

func TestOnlyTheLaneOfACardsKindHoldsIt(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Client's answer", State: "blocked"})

	shot := drawn(t, w)
	back := Dark().Back
	l := w.Layout()
	if got := ink(shot, l.Lane[0], back); got == 0 {
		t.Error("the blocked lane is empty and holds a blocked task")
	}
	for _, lane := range []int{1, 2} {
		card := l.LaneCardAt(lane, 0, 0)
		if got := ink(shot, card, back); got != 0 {
			t.Errorf("the %q lane drew %d pixels of a card that is not its kind", engine.Stalled[lane], got)
		}
	}
}

func TestTheDetailShowsOnlyTheTaskItWasOpenedOn(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "Write the spec", State: "todo", Description: "the first"},
		engine.Card{Number: 2, Title: "Ship it", State: "todo", Description: "the second"})

	box, written := Over(w.width, w.height), Dark().Ink
	first := painted(drawn(t, w.doubleClicked(middleOf(w, 1, 0))), box, written)
	second := painted(drawn(t, w.doubleClicked(middleOf(w, 1, 1))), box, written)
	if first == 0 || second == 0 {
		t.Fatal("a detail drew no title at all")
	}
	if first == second {
		t.Error("the detail drew the same thing for two different tasks")
	}
}

func TestTheFieldTheKeyboardIsOnIsDrawnDifferently(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	box := Over(w.width, w.height)
	lit := Dark().Chosen
	first := painted(drawn(t, w), box, lit)

	w.form = w.form.Set("priority", "high")
	moved := painted(drawn(t, w), box, lit)
	if first == moved {
		t.Error("moving the keyboard to another field drew the same thing")
	}
}

func TestTheTabBeingLookedAtIsTheOpenOne(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	l := w.Layout()
	open := Dark().Back

	shot := drawn(t, w)
	if painted(shot, TabAt(l, 0), open) == 0 {
		t.Error("the tab being looked at is not open")
	}
	if got := painted(shot, TabAt(l, 1), open); got != 0 {
		t.Errorf("the other tab drew %d pixels of an open folder", got)
	}

	second := TabAt(l, 1)
	w = w.clicked(Point{X: second.X + second.W/2, Y: second.Y + second.H/2})
	shot = drawn(t, w)
	if painted(shot, TabAt(l, 1), open) == 0 {
		t.Error("after clicking it, the second tab is not open")
	}
	if got := painted(shot, TabAt(l, 0), open); got != 0 {
		t.Errorf("the first tab still drew %d pixels of an open folder", got)
	}
}

func TestACardChosenInALaneIsDrawnChosen(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Client's answer", State: "blocked"})
	box := w.Layout().LaneCardAt(0, 0, 0)
	edge := Dark().Chosen

	plain := painted(drawn(t, w), box, edge)
	chosen := painted(drawn(t, w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})), box, edge)
	if chosen <= plain {
		t.Errorf("a chosen lane card drew %d pixels of outline and an unchosen one %d", chosen, plain)
	}
}

func TestTwoCardsOfOneKindStackDownTheirLane(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "Client's answer", State: "blocked"},
		engine.Card{Number: 2, Title: "Supplier's answer", State: "blocked"})

	shot := drawn(t, w)
	back := Dark().Back
	if got := ink(shot, w.Layout().LaneCardAt(0, 0, 0), back); got == 0 {
		t.Error("the first card of the lane is not there")
	}
	if got := ink(shot, w.Layout().LaneCardAt(0, 1, 0), back); got == 0 {
		t.Error("the second card of the lane is drawn on top of the first")
	}
}

func TestADetailDrawsTheTaskItIsOpenedOn(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.doubleClicked(middleOf(w, 1, 0))

	if got := painted(drawn(t, w), Over(w.width, w.height), Dark().Ink); got == 0 {
		t.Fatal("the detail drew no name at all")
	}
}

func TestOnlyTheFieldTheKeyboardIsOnIsLit(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	box := w.formBox()
	lit := Dark().Chosen
	rowOf := func(i int) Box {
		return Framed(formValue(box, w.Form().Fields(), i), formRows(box, w.Form().Fields(), i))
	}

	shot := drawn(t, w)
	if painted(shot, rowOf(0), lit) == 0 {
		t.Error("the field the keyboard is on is not lit")
	}
	for i := 1; i < len(w.Form().Fields()); i++ {
		if got := painted(shot, rowOf(i), lit); got != 0 {
			t.Errorf("field %d is lit and the keyboard is not on it: %d pixels", i, got)
		}
	}

	w.form = w.form.Set("priority", "high")
	on := w.Form().On()
	shot = drawn(t, w)
	if painted(shot, rowOf(on), lit) == 0 {
		t.Errorf("field %d is where the keyboard is and is not lit", on)
	}
	if got := painted(shot, rowOf(0), lit); got != 0 {
		t.Errorf("the field the keyboard left is still lit: %d pixels", got)
	}
}

func TestOnATwoPixelScreenEverythingIsDrawnTwiceAsLarge(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})

	window := offScreen(t, w.width*2, w.height*2)
	t.Cleanup(window.Release)

	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Exact(image.Pt(w.width*2, w.height*2)),
		Metric:      unit.Metric{PxPerDp: 2, PxPerSp: 2},
	}
	Draw(gtx, Theme(), Dark(), w)
	if err := window.Frame(gtx.Ops); err != nil {
		t.Fatalf("drawing: %v", err)
	}
	shot := image.NewRGBA(image.Rect(0, 0, w.width*2, w.height*2))
	if err := window.Screenshot(shot); err != nil {
		t.Fatalf("reading the pixels back: %v", err)
	}

	box := w.Layout().CardAt(1, 0, 0)
	twice := Box{X: box.X * 2, Y: box.Y * 2, W: box.W * 2, H: box.H * 2}
	if got := ink(shot, twice, Dark().Back); got == 0 {
		t.Error("nothing was drawn where two pixels to the dp puts the first card")
	}
	empty := Box{X: box.X, Y: box.Y + box.H/2, W: box.W, H: box.H / 4}
	if got := ink(shot, empty, Dark().Back); got != 0 {
		t.Errorf("%d pixels were drawn as though one pixel were one dp", got)
	}
}

func TestAPointerIsReadInDpAndNotInPixels(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "First", State: "todo"},
		engine.Card{Number: 2, Title: "Second", State: "todo"})

	second := w.Layout().CardAt(1, 1, 0)
	inDp := Point{X: second.X + second.W/2, Y: second.Y + second.H/2}
	inPixels := f32.Pt(float32(inDp.X*2), float32(inDp.Y*2))

	if got := Dp(inPixels, 2); got != inDp {
		t.Fatalf("a pointer at %v on a two-pixel screen is at %v in dp, want %v", inPixels, got, inDp)
	}
	if got := w.clicked(Dp(inPixels, 2)).Chosen(); got != 2 {
		t.Errorf("the pointer chose task %d", got)
	}
}

func TestTheObjectivesTabIsDrawnAsAGrid(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Goals: []engine.Goal{
		{Number: 3, Title: "Ship cairn", State: "active"},
		{Number: 4, Title: "Keep it honest", State: "active"},
	}}
	w.tab = Objectives

	shot := drawn(t, w)
	back := Dark().Back
	l := w.Layout()
	for at := range 2 {
		if got := ink(shot, GoalAt(l, at, 0), back); got == 0 {
			t.Errorf("objective %d is not drawn", at)
		}
	}
	if got := ink(shot, l.Head[1], back); got != 0 {
		t.Errorf("the objectives tab drew %d pixels of a column header", got)
	}
}

func TestTheBarDrawsWhatItOffers(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0))

	shot := drawn(t, w)
	for i := range w.Offers() {
		if got := ink(shot, OfferAt(w.Layout(), i), Dark().Back); got == 0 {
			t.Errorf("offer %d is not drawn", i)
		}
	}
}

func TestTheCardInMidAirAndItsLandingAreDrawn(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "First", State: "todo"},
		engine.Card{Number: 2, Title: "Second", State: "backlog"})

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(middleOf(w, 1, 0))

	shot := drawn(t, w)
	ghost, _ := w.Ghost()
	if got := painted(shot, ghost, Dark().Ghost); got == 0 {
		t.Error("the card being dragged is not drawn")
	}
	landing, _ := w.Landing()
	if got := painted(shot, landing, Dark().Chosen); got == 0 {
		t.Error("where the card would land is not drawn")
	}
}

func TestTextIsScaledOnceAndNotTwice(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})

	density := func(perDp float32) float64 {
		wide, high := int(float32(w.width)*perDp), int(float32(w.height)*perDp)
		window := offScreen(t, wide, high)
		defer window.Release()

		var ops op.Ops
		gtx := layout.Context{
			Ops:         &ops,
			Constraints: layout.Exact(image.Pt(wide, high)),
			Metric:      unit.Metric{PxPerDp: perDp, PxPerSp: perDp},
		}
		Draw(gtx, Theme(), Dark(), w)
		if err := window.Frame(gtx.Ops); err != nil {
			t.Fatalf("drawing: %v", err)
		}
		shot := image.NewRGBA(image.Rect(0, 0, wide, high))
		if err := window.Screenshot(shot); err != nil {
			t.Fatalf("reading the pixels back: %v", err)
		}
		title := PartsOf(w.Layout().CardAt(1, 0, 0)).Title
		box := Box{
			X: int(float32(title.X) * perDp), Y: int(float32(title.Y) * perDp),
			W: int(float32(title.W) * perDp), H: int(float32(title.H) * perDp),
		}
		return float64(ink(shot, box, Dark().Panel)) / float64(box.W*box.H)
	}

	one, two := density(1), density(2)
	if one == 0 || two == 0 {
		t.Fatal("a title row is empty")
	}
	if two > one {
		t.Errorf("a title fills %.1f%% of its row at one pixel to the dp and %.1f%% at two: "+
			"it is being scaled twice", 100*one, 100*two)
	}
}

func TestTheCardInMidAirCarriesItsOwnTitle(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "A very long name indeed for a task", State: "todo"},
		engine.Card{Number: 2, Title: "Short", State: "backlog"})

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(middleOf(w, 1, 0))
	ghost, _ := w.Ghost()

	written := painted(drawn(t, w), PartsOf(ghost).Title, Dark().Ink)
	if written == 0 {
		t.Fatal("the card in mid-air carries no title")
	}

	long := holding(f, engine.Card{Number: 2, Title: "A very long name indeed for a task", State: "backlog"})
	long = long.PressAt(middleOf(long, 0, 0)).MoveTo(middleOf(long, 1, 0))
	longGhost, _ := long.Ghost()
	if written >= painted(drawn(t, long), PartsOf(longGhost).Title, Dark().Ink) {
		t.Error("the card in mid-air is carrying somebody else's title")
	}
}

func TestNothingInHandDrawsNothingInMidAir(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	l := w.Layout()
	between := Box{X: l.Column[0].X + l.Column[0].W, Y: l.Column[0].Y, W: gap, H: l.Column[0].H}
	if got := painted(drawn(t, w), between, Dark().Ghost); got != 0 {
		t.Errorf("%d pixels of a card in mid-air were drawn with nothing in hand", got)
	}
}

func TestAnObjectiveChosenIsDrawnChosen(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Goals: []engine.Goal{
		{Number: 3, Title: "Ship cairn", State: "active"},
		{Number: 4, Title: "Keep it honest", State: "active"},
	}}
	w.tab = Objectives

	first := GoalAt(w.Layout(), 0, 0)
	edge := Dark().Chosen
	plain := painted(drawn(t, w), first, edge)

	chosen := w.clicked(Point{X: first.X + first.W/2, Y: first.Y + first.H/2})
	if chosen.Chosen() != 3 {
		t.Fatalf("clicking the first objective chose %d", chosen.Chosen())
	}
	shot := drawn(t, chosen)
	if painted(shot, first, edge) <= plain {
		t.Error("a chosen objective is drawn no differently")
	}
	second := GoalAt(w.Layout(), 1, 0)
	if got := painted(shot, second, edge); got != 0 {
		t.Errorf("an objective nobody chose drew %d pixels of the chosen colour", got)
	}
}

func TestASearchNarrowsWhatIsDrawn(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "Write the spec", State: "backlog"},
		engine.Card{Number: 2, Title: "Ship the board", State: "backlog"})

	both := ink(drawn(t, w), w.Layout().CardAt(0, 1, 0), Dark().Back)
	if both == 0 {
		t.Fatal("the second card is not drawn to begin with")
	}
	narrowed := w.Key("/").Types("spec")
	if got := ink(drawn(t, narrowed), narrowed.Layout().CardAt(0, 1, 0), Dark().Back); got != 0 {
		t.Errorf("a search left %d pixels of a card it does not reach", got)
	}
}

func TestWhatIsBeingSearchedForIsShown(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	w = w.Key("/").Types("spec")

	shot := drawn(t, w)
	if got := painted(shot, Searching(w.Layout()), Dark().Ink); got == 0 {
		t.Error("what is being searched for is not shown along the bottom")
	}
	if got := painted(shot, Message(w.width, w.height), Dark().Ink); got != 0 {
		t.Errorf("%d pixels of the search were written across the middle of the board", got)
	}
}

func TestTheHelpAndTheQuestionAreDrawn(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0))

	box := Over(w.width, w.height)
	if got := painted(drawn(t, w.Choose("Help")), box, Dark().Ink); got == 0 {
		t.Error("the help is blank")
	}
	if got := painted(drawn(t, w.Choose("Delete")), box, Dark().Ink); got == 0 {
		t.Error("the question asked before deleting is blank")
	}
}

func TestTheCaretIsDrawnWhereTypingGoes(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	box := Over(w.width, w.height)
	fields := w.Form().Fields()
	valueOf := func(i int) Box {
		value := formValue(box, fields, i)
		return value
	}

	on := w.On("title")
	with := painted(drawn(t, on), valueOf(0), Dark().Chosen)
	without := painted(drawn(t, on.On("priority")), valueOf(0), Dark().Chosen)
	if with <= without {
		t.Errorf("the title carried %d pixels of caret with the keyboard on it and %d without", with, without)
	}
	if without != 0 {
		t.Errorf("a field the keyboard is not on carries %d pixels of caret", without)
	}
}

func TestADescriptionIsDrawnOverSeveralRows(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{
		Number: 1, State: "todo", Title: "Write the spec",
		Description: strings.Repeat("two columns, the client says, and nothing else. ", 6),
	})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("description")

	box := Over(w.width, w.height)
	fields := w.Form().Fields()
	value := formValue(box, fields, 1)
	value.H = lineHigh * formRows(box, fields, 1)

	shot := drawn(t, w)
	rows := 0
	for row := range formRows(box, fields, 1) {
		line := Box{X: value.X, Y: value.Y + row*lineHigh, W: value.W, H: lineHigh}
		if painted(shot, line, Dark().Ink) > 0 {
			rows++
		}
	}
	if rows < 2 {
		t.Errorf("a long description was drawn on %d of its %d rows", rows, formRows(box, fields, 1))
	}
}

func TestTheOpenTabIsAFolderAndNotAColouredWord(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	l := w.Layout()
	shot := drawn(t, w)

	if got := painted(shot, TabAt(l, 0), Dark().Back); got == 0 {
		t.Error("the open tab does not carry the page's own ground")
	}
	if got := painted(shot, TabAt(l, 1), Dark().Back); got != 0 {
		t.Errorf("a shut tab drew %d pixels of the page's ground", got)
	}
	if got := painted(shot, TabAt(l, 1), Dark().Panel); got == 0 {
		t.Error("a shut tab does not sit in the band")
	}
	for i, side := range TabSides(TabAt(l, 0)) {
		if got := painted(shot, side, Dark().Edge); got == 0 {
			t.Errorf("side %d of the open tab was not drawn", i)
		}
	}
	for i, piece := range TabsLineAround(l, 0) {
		if piece.W == 0 {
			continue
		}
		if got := painted(shot, piece, Dark().Edge); got == 0 {
			t.Errorf("piece %d of the line closing the band was not drawn", i)
		}
	}
	under := Box{X: TabAt(l, 0).X + 2, Y: TabsLine(l).Y, W: TabAt(l, 0).W - 4, H: 1}
	if got := painted(shot, under, Dark().Edge); got != 0 {
		t.Errorf("%d pixels of line run under the open tab, which shuts the folder", got)
	}
}

func TestTheColumnsStandClearOfTheTabs(t *testing.T) {
	l := LayoutFor(1280, 840)
	if got := l.Head[0].Y - (l.Tabs.Y + l.Tabs.H); got < gap {
		t.Errorf("the columns start %d below the tabs, and a gap is %d", got, gap)
	}
}

func TestASearchSaysWhatItIs(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	l := w.Layout()
	box := Searching(l)

	off := painted(drawn(t, w), box, Dark().Panel)
	on := drawn(t, w.Key("/"))
	if got := painted(on, box, Dark().Panel); got <= off {
		t.Errorf("the search drew %d pixels of box and the bar behind it %d", painted(on, box, Dark().Panel), off)
	}
	if got := painted(on, SearchWord(box), Dark().Faint); got == 0 {
		t.Error("the search box does not say what it is")
	}
	typed := painted(drawn(t, w.Key("/").Types("spec")), SearchInto(box), Dark().Ink)
	if typed == 0 {
		t.Error("what was typed was not drawn where the search box types")
	}
}

func TestAPanelSaysWhatItIsAndDrawsItsKeysOnChips(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	box := Over(w.width, w.height)
	help := w.Key("?")

	shot := drawn(t, help)
	if got := painted(shot, Heading(box), Dark().Ink); got == 0 {
		t.Error("the help does not say what it is")
	}
	if got := painted(shot, Ruled(box), Dark().Edge); got < Ruled(box).W {
		t.Errorf("the heading is underlined across %d of %d", got, Ruled(box).W)
	}
	name, _ := FieldRow(box, FirstRow(box))
	if got := painted(shot, Chip(name), Dark().Back); got == 0 {
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
	if got := New(newFake(), 1200, 800).OnSystem("darwin").Said("New"); !strings.Contains(got, "⌘N") || !strings.Contains(got, "New") {
		t.Errorf("the bar writes New as %q", got)
	}
	if got := New(newFake(), 1200, 800).Said("nothing at all"); got != "nothing at all" {
		t.Errorf("an offer with no key is written %q, with a gap in front of it", got)
	}
}

func TestEveryFieldIsDrawnInABoxAndTheOneTypedIntoIsLit(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title")

	box := Over(w.width, w.height)
	fields := w.Form().Fields()
	frameOf := func(name string) Box {
		for i, field := range fields {
			if field.Name == name {
				value := formValue(box, fields, i)
				return Framed(value, formRows(box, fields, i))
			}
		}
		t.Fatalf("the form has no field %q", name)
		return Box{}
	}

	shot := drawn(t, w)
	on, off := frameOf("title"), frameOf("priority")
	for what, frame := range map[string]Box{"the field typed into": on, "a field not typed into": off} {
		if got := painted(shot, frame, Dark().Back); got == 0 {
			t.Errorf("%s is not drawn in a box", what)
		}
	}
	if painted(shot, Edges(on)[0], Dark().Chosen) == 0 {
		t.Error("the box of the field typed into is not lit")
	}
	if got := painted(shot, Edges(off)[0], Dark().Chosen); got != 0 {
		t.Errorf("a field nobody is typing into has %d pixels of lit edge", got)
	}
}

func TestALongFieldWrapsInsideItsBox(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("description")

	box := Over(w.width, w.height)
	fields := w.Form().Fields()
	var frame Box
	for i, field := range fields {
		if field.Name == "description" {
			value := formValue(box, fields, i)
			frame = Framed(value, formRows(box, fields, i))
		}
	}
	long := strings.Repeat("This description is far longer than one line of this form can hold. ", 3)
	shot := drawn(t, w.Types(long))

	inside := Inside(frame)
	first := Box{X: inside.X, Y: inside.Y, W: inside.W, H: lineHigh}
	second := Box{X: inside.X, Y: inside.Y + lineHigh, W: inside.W, H: lineHigh}
	if painted(shot, first, Dark().Ink) == 0 {
		t.Error("nothing was written on the first line")
	}
	if painted(shot, second, Dark().Ink) == 0 {
		t.Error("nothing went onto the second line: it was cut rather than wrapped")
	}
	past := Box{X: frame.X + frame.W, Y: frame.Y, W: 40, H: frame.H}
	if got := painted(shot, past, Dark().Ink); got != 0 {
		t.Errorf("%d pixels of writing ran out of the right of the box", got)
	}
	under := Box{X: frame.X, Y: frame.Y + frame.H, W: frame.W, H: between}
	if got := painted(shot, under, Dark().Ink); got != 0 {
		t.Errorf("%d pixels of writing ran out of the bottom of the box", got)
	}
}

func TestTheHistoryIsDrawnInItsBoxAndStaysThere(t *testing.T) {
	f := newFake()
	for range 40 {
		f.acts = append(f.acts, anAct("moved", "column", "todo", "current"))
	}
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Key("enter")
	if w.Screen() != OnForm {
		t.Fatalf("the window is on %v, not the detail", w.Screen())
	}

	box := w.formPanel()
	frame, track := BesideBox(box), BesideBar(box)
	shot := drawn(t, w)

	if painted(shot, Inside(frame), Dark().Faint) == 0 {
		t.Error("nothing of the history was drawn in its box")
	}
	under := Box{X: frame.X, Y: frame.Y + frame.H, W: frame.W, H: lineHigh}
	if got := painted(shot, under, Dark().Faint); got != 0 {
		t.Errorf("%d pixels of history ran out of the bottom of its box", got)
	}

	high := Box{X: track.X, Y: track.Y, W: track.W, H: track.H / 4}
	low := Box{X: track.X, Y: track.Y + track.H*3/4, W: track.W, H: track.H / 4}
	if painted(shot, high, Dark().Edge) == 0 {
		t.Error("the grip is not at the top of a history nobody has scrolled")
	}
	if got := painted(shot, low, Dark().Edge); got != 0 {
		t.Errorf("the grip has %d pixels at the bottom of a history nobody has scrolled", got)
	}
	ended := drawn(t, w.Scroll(Point{X: frame.X + frame.W/2, Y: frame.Y + frame.H/2}, 999*lineHigh))
	if painted(ended, low, Dark().Edge) == 0 {
		t.Error("the grip is not at the bottom of a history scrolled to its end")
	}
	if painted(ended, Inside(frame), Dark().Faint) == 0 {
		t.Error("a history scrolled to its end shows nothing")
	}
}

func TestTheObjectivesAreCutToTheirRun(t *testing.T) {
	f := newFake()
	var goals []engine.Goal
	for i := 1; i <= 12; i++ {
		goals = append(goals, engine.Goal{Number: i, Title: "Ship it", State: "live"})
	}
	w := holding(f)
	w.snap.Goals = goals
	l := w.Layout()
	second := TabAt(l, 1)
	w = w.clicked(Point{X: second.X + second.W/2, Y: second.Y + second.H/2})
	w = w.Scroll(Point{X: l.Column[0].X + 10, Y: l.Column[0].Y + 10}, 3*CardHeight)

	above := Box{X: l.Column[0].X, Y: l.Tabs.Y + l.Tabs.H, W: l.Column[0].W, H: gap}
	if got := painted(drawn(t, w), above, Dark().Panel); got != 0 {
		t.Errorf("%d pixels of objective were drawn above the run they belong in", got)
	}
}

func TestTheWorkUnderAnObjectiveIsDrawnAsTilesInItsBox(t *testing.T) {
	f := newFake()
	var cards []engine.Card
	for i := range 30 {
		cards = append(cards, engine.Card{
			Number: i + 1, Title: fmt.Sprintf("Task %d", i+1), State: "done",
			Objectives: []int{1},
		})
	}
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{
		Cards: cards,
		Goals: []engine.Goal{{Number: 1, Title: "Ship the desktop board", State: "active", Done: 7, Total: 30}},
	}
	w = w.Key("tab")
	w = w.clicked(goalMiddle(w, 1)).Key("enter")
	if w.Screen() != OnForm || !w.ChosenGoal() {
		t.Fatalf("the window is on %v with a goal chosen: %v", w.Screen(), w.ChosenGoal())
	}

	frame := BesideBox(w.formPanel())
	track := BesideBar(w.formPanel())
	shot := drawn(t, w)

	check := SubCheck(w.ServingCardAt(0))
	strip := func(b Box) Box { return Box{X: check.X, Y: b.Y, W: check.W, H: b.H} }
	if painted(shot, strip(frame), Dark().Faint) == 0 {
		t.Error("no tile was drawn in the box beside the form")
	}
	under := Box{X: frame.X, Y: frame.Y + frame.H, W: frame.W, H: SubHeight}
	if got := painted(shot, strip(under), Dark().Faint); got != 0 {
		t.Errorf("%d pixels of tile ran out of the bottom of the box", got)
	}

	high := Box{X: track.X, Y: track.Y, W: track.W, H: track.H / 4}
	if painted(shot, high, Dark().Edge) == 0 {
		t.Error("the grip is not at the top of a list nobody has scrolled")
	}
	walked := w
	for range w.ServingRows() + 2 {
		walked = walked.Scroll(Point{X: frame.X + 4, Y: frame.Y + 4}, SubHeight)
	}
	scrolled := drawn(t, walked)
	if painted(scrolled, high, Dark().Edge) == painted(shot, high, Dark().Edge) {
		t.Error("the grip did not move when the list was scrolled")
	}
}

func TestAnObjectiveNothingServesSaysSoInTheBox(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{
		Goals: []engine.Goal{{Number: 1, Title: "Ship the desktop board", State: "active"}},
	}
	w = w.Key("tab")
	w = w.clicked(goalMiddle(w, 1)).Key("enter")

	frame := BesideBox(Over(w.width, w.height))
	shot := drawn(t, w)
	if painted(shot, Inside(frame), Dark().Faint) == 0 {
		t.Error("the box says nothing where an objective nothing serves should say so")
	}
	if got := painted(shot, Box{X: frame.X, Y: frame.Y, W: edgeWide, H: frame.H}, Dark().Chosen); got != 0 {
		t.Errorf("%d pixels of tile were drawn for an objective nothing serves", got)
	}
}

func TestAnObjectiveCardSaysHowFarAlongItIs(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{
		Goals: []engine.Goal{{Number: 1, Title: "Ship the desktop board", State: "active", Done: 7, Total: 12}},
	}
	w = w.Key("tab")

	said := Counted(w.Reading().Goals[0])
	if said != "7/12 · 58%" {
		t.Fatalf("Counted = %q", said)
	}
	box := CountedAt(PartsOf(GoalAt(w.Layout(), 0, 0)).Number, said)
	if painted(drawn(t, w), box, Dark().Faint) == 0 {
		t.Error("the count was not drawn on the objective's card")
	}
}

func goalMiddle(w Window, number int) Point {
	for at, g := range engine.Sorted(w.Showing().Goals) {
		if g.Number == number {
			box := GoalAt(w.Layout(), at, w.goals)
			return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
		}
	}
	return Point{}
}

func TestTheListOfObjectivesIsDrawn(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{
		Cards: []engine.Card{{Number: 1, Title: "Write the spec", State: "todo"}},
		Goals: []engine.Goal{
			{Number: 1, Title: "Ship the desktop board", State: "active"},
			{Number: 2, Title: "Keep it honest", State: "active"},
		},
	}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "objective" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if w.Screen() != OnChoosing {
		t.Fatalf("the window is on %v, not the list", w.Screen())
	}

	box := w.ChoosingBox()
	first := ChoosingRow(box, ChoosingFrom(box))
	shot := drawn(t, w)

	if painted(shot, first, Dark().Chosen) == 0 {
		t.Error("the row the keyboard is on is not lit")
	}
	second := ChoosingRow(box, ChoosingFrom(box)+apart)
	if got := painted(shot, second, Dark().Chosen); got != 0 {
		t.Errorf("%d pixels lit a row the keyboard is not on", got)
	}

	moved := inked(drawn(t, w.Key("down")), first, Dark().Panel)
	ticked := inked(drawn(t, w.Key("enter").Key("down")), first, Dark().Panel)
	if ticked <= moved {
		t.Errorf("a ticked row paints %d and an unticked one %d: no mark was drawn", ticked, moved)
	}
}

func TestAListWithNothingToOfferSaysSo(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Cards: []engine.Card{{Number: 1, Title: "Write the spec", State: "todo"}}}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "objective" {
		w = w.Key("tab")
	}
	w = w.Key("enter")

	box := w.ChoosingBox()
	value := Row(box, ChoosingFrom(box), lineHigh)
	if painted(drawn(t, w), value, Dark().Faint) == 0 {
		t.Error("the list says nothing where it should say there is nothing to serve")
	}
}

func TestFits_cutsToTheRoomThereIs(t *testing.T) {
	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Exact(image.Pt(1200, 800)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}
	th := Theme()
	long := "T090f · Ajouter une étape pour réouvrir les droits avant intervention"

	if got := fits(gtx, th, "short", 400, small); got != "short" {
		t.Errorf("fits(short, 400) = %q, want it left alone", got)
	}
	cut := fits(gtx, th, long, 200, small)
	if !strings.HasSuffix(cut, "…") {
		t.Errorf("fits(long, 200) = %q, want it cut with an ellipsis", cut)
	}
	if wide(loosened(gtx), th, cut, small).Size.X > 200 {
		t.Errorf("fits(long, 200) = %q, which is wider than 200", cut)
	}
	if len([]rune(cut)) >= len([]rune(long)) {
		t.Errorf("fits(long, 200) = %q, which is no shorter than what it was given", cut)
	}
	exact := wide(loosened(gtx), th, "short", small).Size.X
	if got := fits(gtx, th, "short", exact, small); got != "short" {
		t.Errorf("fits(short, its own width) = %q, want it left alone", got)
	}
	if got := fits(gtx, th, "short", exact-1, small); got == "short" {
		t.Errorf("fits(short, one less than its width) = %q, want it cut", got)
	}

	oneCut := fits(gtx, th, long, 200, small)
	atCut := wide(loosened(gtx), th, oneCut, small).Size.X
	if got := fits(gtx, th, long, atCut, small); got != oneCut {
		t.Errorf("fits(long, the width of its own cut) = %q, want %q", got, oneCut)
	}

	one := string([]rune(long)[:1]) + "…"
	atOne := wide(loosened(gtx), th, one, small).Size.X
	if got := fits(gtx, th, long, atOne, small); got != one {
		t.Errorf("fits(long, room for one rune) = %q, want %q", got, one)
	}

	if got := fits(gtx, th, long, 1, small); got != "…" {
		t.Errorf("fits(long, 1) = %q, want the ellipsis on its own", got)
	}
}

func loosened(gtx layout.Context) layout.Context {
	loose := gtx
	loose.Constraints = layout.Constraints{Max: image.Pt(1<<20, lineHigh)}
	return loose
}

func TestGhost_isDrawnAsTheKindOfThingItIs(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	from := middleOf(w, 1, 0)
	to := Point{X: from.X, Y: from.Y - CardHeight}
	w = w.PressAt(from).MoveTo(to)
	ghost, dragging := w.Ghost()
	if !dragging {
		t.Fatal("nothing is in hand on the board")
	}
	shot := drawn(t, w)
	foot := Box{X: ghost.X + 2, Y: ghost.Y + ghost.H - between, W: ghost.W - 4, H: between - 2}
	if painted(shot, foot, Dark().Ghost) == 0 {
		t.Errorf("the foot of what is in hand is not filled: %+v", foot)
	}
}

func TestWhatACardCarriesIsDrawnBesideItsNumber(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "Bare", State: "todo"},
		engine.Card{Number: 2, Title: "Carries", State: "todo", Description: "something written", SubDone: 1, SubTotal: 3},
	)
	shot := drawn(t, w)
	beside := func(at int) Box {
		number := PartsOf(w.Layout().CardAt(1, at, 0)).Number
		return Box{X: number.X + number.W/2, Y: number.Y, W: number.W / 2, H: number.H}
	}
	if got := ink(shot, beside(0), Dark().Panel); got != 0 {
		t.Errorf("a card carrying nothing drew %d pixels beside its number", got)
	}
	if got := ink(shot, beside(1), Dark().Panel); got == 0 {
		t.Error("a card carrying a description and its parts drew nothing beside its number")
	}
}

func TestAHeldRunIsPaintedBehindTheWords(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Cards: []engine.Card{{Number: 1, Title: "Write the spec", State: "todo"}}}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title")

	box := Over(w.width, w.height)
	fields := w.Form().Fields()
	var value Box
	for i, field := range fields {
		if field.Name == "title" {
			value = formValue(box, fields, i)
		}
	}
	inside := Inside(Framed(value, 1))

	loose := ink(drawn(t, w), inside, Dark().Back)
	held := ink(drawn(t, w.Key("cmd+a")), inside, Dark().Back)
	if held <= loose {
		t.Errorf("holding the whole field painted %d pixels and letting go %d: "+
			"the run is not drawn", held, loose)
	}
	some := ink(drawn(t, w.Key("shift+left").Key("shift+left")), inside, Dark().Back)
	if some <= loose || some >= held {
		t.Errorf("a run of two characters painted %d, between %d and %d expected",
			some, loose, held)
	}
}

func TestEveryCardHasAnEdge(t *testing.T) {
	f := newFake()
	w := holding(f,
		card(1, "Write the spec", "todo"),
		card(2, "Ship it", "todo"))
	shot := drawn(t, w.clicked(middleOf(w, 1, 0)))

	for at, what := range map[int]string{0: "the chosen card", 1: "a card nobody chose"} {
		parts := PartsOf(w.Layout().CardAt(1, at, 0))
		top := Box{X: parts.Panel.X, Y: parts.Panel.Y, W: parts.Panel.W, H: 1}
		if ink(shot, top, Dark().Back) == 0 {
			t.Errorf("%s is drawn without an edge", what)
		}
	}
}

func TestEveryRunOfCardsSaysHowManyItHolds(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "Write the spec", State: "blocked"},
		engine.Card{Number: 2, Title: "Ship it", State: "blocked"},
		engine.Card{Number: 3, Title: "Later", State: "watching"})
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

func TestTheFormWritesWhatEachFieldSays(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo", Labels: []string{"backend"}})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	shot := drawn(t, w)
	box := w.formBox()
	for i, field := range w.Form().Fields() {
		if field.Name != "labels" && field.Name != "column" {
			continue
		}
		value, rows := formValue(box, w.Form().Fields(), i), formRows(box, w.Form().Fields(), i)
		if painted(shot, Inside(Framed(value, rows)), Dark().Ink) == 0 {
			t.Errorf("%q writes nothing", field.Name)
		}
	}
}

func TestAQuestionPutsWhatItNamesOnAChip(t *testing.T) {
	skin := Dark()
	cards := []engine.Card{card(1, "First", "todo"), {Number: 2, Title: "Gone", State: "todo", Archived: true}}
	w := holding(newFake(), cards...)
	w.archived = true
	w = w.clicked(middleOf(w, 1, 0)).KeysDown(key.ModShortcut).clicked(middleOf(w, 1, 1)).KeysDown(0)
	for _, c := range []struct {
		what  string
		asked Window
		chip  bool
	}{
		{"archiving what is already off", w.Key("x"), true},
		{"deleting one task", w.clicked(middleOf(w, 1, 0)).Choose("Delete"), false},
	} {
		if c.asked.Screen() != OnAsking {
			t.Fatalf("%s left the window on %v", c.what, c.asked.Screen())
		}
		box := c.asked.QuestionBox()
		name, _ := FieldRow(box, FirstRow(box))
		if got := painted(drawn(t, c.asked), Chip(name), skin.Back) > 0; got != c.chip {
			t.Errorf("%s: its first line on a chip is %v", c.what, got)
		}
	}
}

func TestTheTasksTabCountsWhatTheBoardLaysOut(t *testing.T) {
	w := holding(newFake(),
		engine.Card{Number: 1, Title: "Ship it", State: "todo", Steps: []engine.Part{{Number: 1, Title: "A step"}}},
		engine.Card{Number: 3, Title: "Gone", State: "done", Archived: true})
	if got := w.Holds(Tasks); got != 1 {
		t.Errorf("the tasks tab says %d, and the board lays out 1", got)
	}
}

func TestADragOpensRoomWhereTheCardWouldLand(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "First", State: "todo"},
		engine.Card{Number: 2, Title: "Second", State: "todo"},
		engine.Card{Number: 3, Title: "Third", State: "backlog"})
	l := w.Layout()
	second := l.CardAt(1, 1, 0)
	w = w.PressAt(middleOf(w, 0, 0)).MoveTo(Point{X: second.X + second.W/2, Y: second.Y + second.H/4})

	landing, ok := w.Landing()
	if !ok || landing != (Box{X: second.X, Y: second.Y, W: second.W, H: CardHeight}) {
		t.Errorf("the card would land in %+v, want the room %+v", landing, second)
	}
	slots := map[int]Box{}
	for _, p := range w.Placed() {
		slots[p.Card.Number] = p.Slot
	}
	if _, drawn := slots[3]; drawn {
		t.Error("the card in hand is still drawn in its column")
	}
	if slots[1] != l.CardAt(1, 0, 0) || slots[2] != l.CardAt(1, 2, 0) {
		t.Errorf("First is at %+v and Second at %+v: no room was made between them", slots[1], slots[2])
	}
	w = w.MoveTo(Point{X: second.X + second.W/2, Y: second.Y + second.H*3/4})
	if landing, _ := w.Landing(); landing.Y != l.CardAt(1, 2, 0).Y {
		t.Errorf("over the lower half of Second the card would land at %d, want after it", landing.Y)
	}
}

func paintedNear(shot *image.RGBA, b Box, want color.NRGBA, by int) int {
	count := 0
	for y := b.Y; y < b.Y+b.H; y++ {
		for x := b.X; x < b.X+b.W; x++ {
			got := shot.RGBAAt(x, y)
			if within(got.R, want.R, by) && within(got.G, want.G, by) && within(got.B, want.B, by) {
				count++
			}
		}
	}
	return count
}

func within(a, b uint8, by int) bool { return max(int(a)-int(b), int(b)-int(a)) <= by }

func dimmed(c color.NRGBA) color.NRGBA {
	veil := Dark().Veil
	mix := func(under, over uint8) uint8 {
		return uint8((int(under)*(255-int(veil.A)) + int(over)*int(veil.A)) / 255)
	}
	return color.NRGBA{R: mix(c.R, veil.R), G: mix(c.G, veil.G), B: mix(c.B, veil.B), A: 255}
}
