package gui

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/engine"
	"cairn/internal/trace"
)

// The edges: the exact point a click becomes a drag, the exact width a title
// stops fitting, the window that is short rather than narrow, and every line a
// history can be made of. None of these is reached by using the board normally,
// which is precisely why they are written down.

// Grip is a threshold, so both sides of it are worth saying out loud.
func TestTheExactPointAClickBecomesADrag(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	at := middleOf(w, 0, 0)

	held := w.PressAt(at).MoveTo(Point{X: at.X + Grip, Y: at.Y})
	if held.Dragging() {
		t.Errorf("moving exactly %d is still a click", Grip)
	}
	held = w.PressAt(at).MoveTo(Point{X: at.X + Grip + 1, Y: at.Y})
	if !held.Dragging() {
		t.Errorf("moving %d is a drag", Grip+1)
	}
	held = w.PressAt(at).MoveTo(Point{X: at.X, Y: at.Y - Grip - 1})
	if !held.Dragging() {
		t.Error("moving up is a drag as much as moving down")
	}
}

// Taking hold of a card and not moving it is not dragging, however long it is
// held.
func TestHoldingStillIsNotDragging(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	at := middleOf(w, 0, 0)
	if w.PressAt(at).Dragging() {
		t.Error("a card just taken hold of is being dragged")
	}
	if w.PressAt(at).MoveTo(at).Dragging() {
		t.Error("a card held still is being dragged")
	}
}

// Moving with nothing in hand is nothing at all.
func TestMovingWithNothingInHandDoesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	if got := w.MoveTo(Point{X: 500, Y: 500}); got.Dragging() {
		t.Error("moving the pointer with nothing held began a drag")
	}
}

// A detail that cannot read the history says so rather than showing none.
func TestADetailThatCannotReadTheHistorySaysSo(t *testing.T) {
	f := newFake()
	f.fail = errNope
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})

	w = w.doubleClicked(middleOf(w, 1, 0))
	if w.Wrong() == nil {
		t.Fatal("a history out of reach was drawn as an empty one")
	}
}

// Double-clicking where there is no card opens nothing.
func TestADoubleClickOnNothingOpensNothing(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	if got := w.doubleClicked(Point{X: 0, Y: 0}); got.Screen() != OnTheBoard {
		t.Fatalf("double-clicking the margin left the window on %v", got.Screen())
	}
}

// Resizing a window that is big enough already leaves it where it was: a form
// open across a resize is a form somebody is still typing in.
func TestResizingAWindowThatIsBigEnoughLeavesTheScreenAlone(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	if got := w.Resized(1400, 900); got.Screen() != OnForm {
		t.Fatalf("resizing closed the form: the window is on %v", got.Screen())
	}
	if got := w.Resized(320, 240); got.Screen() != TooSmallToDraw {
		t.Fatalf("shrinking past the smallest size left the window on %v", got.Screen())
	}
}

// A window short of height rather than width still has to give the band its
// share, which is the half that is not the columns'.
func TestAShortWindowSharesWhatIsLeftWithTheBand(t *testing.T) {
	l := LayoutFor(1200, Shortest)
	if l.Band.H <= 0 {
		t.Fatalf("the band got %d of a short window", l.Band.H)
	}
	if l.Column[1].H <= 0 {
		t.Fatalf("the columns got %d of a short window", l.Column[1].H)
	}
	// The band asks for the same two rows whatever the window, which is why
	// its share is a name and not a sum: the smallest window this board draws
	// in already has room for them.
	if l.Band.H != bandHigh {
		t.Errorf("the smallest window gave the band %d, want %d", l.Band.H, bandHigh)
	}
	if tall := LayoutFor(1200, 1400); tall.Band.H != bandHigh {
		t.Errorf("a tall window gave the band %d, want %d", tall.Band.H, bandHigh)
	}
}

// A card of exactly two characters' room keeps its title; one less and there is
// nowhere to put an ellipsis that would say anything.
func TestTheNarrowestCardThatCanCutATitle(t *testing.T) {
	two := padding + padding + 2*perGlyph
	if got := TitleOf("abcdef", two); got != "a…" {
		t.Errorf("a card with room for two answered %q", got)
	}
	if got := TitleOf("abcdef", two-perGlyph); got != "abcdef" {
		t.Errorf("a card with room for one answered %q", got)
	}
}

// A history is written from what an act tells, and an act tells different
// things depending on what was done to the task.
func TestEveryShapeAHistoryLineCanTake(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})

	f.acts = []trace.Act{
		anAct("added", "", "", ""),
		anAct("edited", "state", "todo", ""),
		anAct("edited", "title", "Write the whole spec", "board"),
	}
	w = w.doubleClicked(middleOf(w, 1, 0))

	lines := w.Acts()
	if len(lines) != 3 {
		t.Fatalf("the history is %d lines", len(lines))
	}
	if strings.Contains(lines[0], "→") {
		t.Errorf("an act that changed nothing wrote an arrow: %q", lines[0])
	}
	if !strings.Contains(lines[1], "state") || !strings.Contains(lines[1], "→ todo") {
		t.Errorf("an act that changed a field wrote %q", lines[1])
	}
	if !strings.Contains(lines[2], "(board)") {
		t.Errorf("an act somebody else did wrote %q", lines[2])
	}
	if strings.Contains(lines[1], "(") {
		t.Errorf("an act done at the terminal named a surface: %q", lines[1])
	}
}

// anAct builds one the only way there is, through its own membrane.
func anAct(did, field, became, through string) trace.Act {
	var told trace.AsTold
	told.On(1)
	told.Did(did)
	told.At(time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC))
	if field != "" {
		told.Changing(field)
		told.Became(became)
	} else {
		told.NothingNamed()
		told.BecameNothing()
	}
	if through != "" {
		told.Through(through)
	} else {
		told.ByHand()
	}
	return told.Act()
}

// Two engine operations of the same name are one operation. This is the only
// place the window names one by hand rather than reading it off the engine, so
// it is the only place a typo could hide.
func TestTheWindowNamesOperationsTheEngineDescribes(t *testing.T) {
	for _, op := range []engine.Operation{
		engine.TaskEdit, engine.TaskMove, engine.TaskNew,
		engine.ObjectiveNew, engine.ObjectiveEdit, engine.TraceShow,
	} {
		if op.Thing == "" || op.Verb == "" {
			t.Errorf("%+v names nothing", op)
		}
	}
}

// A scrolled column answers a different card at the same point, and a card
// drawn in one is drawn higher by exactly what it was scrolled by. Both are
// arithmetic nobody reaches by using the board gently.
func TestAScrolledColumnAnswersAndDrawsFromWhereItStands(t *testing.T) {
	f := newFake()
	var many []engine.Card
	for i := 1; i <= 12; i++ {
		many = append(many, engine.Card{Number: i, Title: "Task", State: "todo"})
	}
	w := holding(f, many...)

	l := w.Layout()
	if got, want := l.CardAt(1, 0, CardHeight).Y, l.Column[1].Y-CardHeight; got != want {
		t.Errorf("a column scrolled by one card drew its first at %d, want %d", got, want)
	}
	if got, want := l.CardAt(1, 2, CardHeight).Y, l.Column[1].Y+CardHeight; got != want {
		t.Errorf("the third card of a scrolled column is at %d, want %d", got, want)
	}

	at := Point{X: l.Column[1].X + 10, Y: l.Column[1].Y + 10}
	unscrolled := w.clicked(at).Chosen()
	w = w.Scroll(Point{X: l.Column[1].X + l.Column[1].W/2, Y: l.Column[1].Y + l.Column[1].H/2}, CardHeight*2)
	scrolled := w.clicked(at).Chosen()
	if scrolled != unscrolled+2 {
		t.Errorf("scrolled by two cards, the same point chose %d after %d", scrolled, unscrolled)
	}
}

// Dropping into a scrolled column lands where the card is seen to be, not
// where it would have been unscrolled.
func TestDroppingIntoAScrolledColumnLandsWhereItLooks(t *testing.T) {
	f := newFake()
	var many []engine.Card
	for i := 1; i <= 12; i++ {
		many = append(many, engine.Card{Number: i, Title: "Task", State: "todo"})
	}
	many = append(many, engine.Card{Number: 99, Title: "Newcomer", State: "backlog"})
	w := holding(f, many...)

	l := w.Layout()
	w = w.Scroll(Point{X: l.Column[1].X + l.Column[1].W/2, Y: l.Column[1].Y + l.Column[1].H/2}, CardHeight*3)
	onto := Point{X: l.Column[1].X + 10, Y: l.Column[1].Y + CardHeight + 10}

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(onto)
	w.LetGo(onto)

	in, asked := f.gave("task", "move")
	if !asked {
		t.Fatal("nothing was asked to move")
	}
	// Scrolled by three and dropped on the second slot showing: behind the
	// fourth card, not the first.
	if in.Given["after"] != "4" {
		t.Errorf("asked to move %+v, want it behind task 4", in.Given)
	}
}

// A card in a lane is found at the same point whatever the columns above it
// were scrolled to: a lane keeps its own place.
func TestALaneIsNotScrolledByTheColumnsAboveIt(t *testing.T) {
	f := newFake()
	var many []engine.Card
	for i := 1; i <= 12; i++ {
		many = append(many, engine.Card{Number: i, Title: "Task", State: "todo"})
	}
	many = append(many, engine.Card{Number: 99, Title: "Client's answer", State: "blocked"})
	w := holding(f, many...)

	box := w.Layout().LaneCardAt(0, 0, 0)
	at := Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
	if got := w.clicked(at).Chosen(); got != 99 {
		t.Fatalf("clicking the lane chose %d", got)
	}
	l := w.Layout()
	w = w.Scroll(Point{X: l.Column[1].X + l.Column[1].W/2, Y: l.Column[1].Y + l.Column[1].H/2}, CardHeight*3)
	if got := w.clicked(at).Chosen(); got != 99 {
		t.Errorf("after scrolling a column, the lane chose %d", got)
	}
}

// abs answers the distance and never a negative one, at zero as much as either
// side of it.
func TestTheDistanceIsNeverNegative(t *testing.T) {
	for _, c := range []struct{ n, want int }{{-3, 3}, {-1, 1}, {0, 0}, {1, 1}, {3, 3}} {
		if got := abs(c.n); got != c.want {
			t.Errorf("abs(%d) = %d, want %d", c.n, got, c.want)
		}
	}
}

// Moving exactly Grip upward is a click, the same as moving it sideways.
func TestTheGripIsTheSameInBothDirections(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	at := middleOf(w, 0, 0)
	if w.PressAt(at).MoveTo(Point{X: at.X, Y: at.Y + Grip}).Dragging() {
		t.Errorf("moving exactly %d down is still a click", Grip)
	}
	if !w.PressAt(at).MoveTo(Point{X: at.X, Y: at.Y + Grip + 1}).Dragging() {
		t.Errorf("moving %d down is a drag", Grip+1)
	}
}
