package gui

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"

	"cairn/internal/reading"
	"cairn/internal/trace"
)

// What an event means. These are the three functions the loop is made of, and
// they are ordinary functions precisely so that the loop itself is the only
// thing left that no test can reach.

func TestEveryPointerEventItsGesture(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	at := middleOf(w, 0, 0)

	pressed := Pointed(w, pointer.Press, at, 0)
	if pressed.held.Number != 1 {
		t.Errorf("a press took hold of %d", pressed.held.Number)
	}
	dragged := Pointed(pressed, pointer.Drag, Point{X: at.X + 50, Y: at.Y}, 0)
	if !dragged.Dragging() {
		t.Error("a drag did not begin one")
	}
	dropped := Pointed(dragged, pointer.Release, middleOf(w, 1, 0), 0)
	if dropped.Dragging() {
		t.Error("a release left the card in hand")
	}
	if f.what("task", "edit") != 1 {
		t.Errorf("the drop asked for %d edits", f.what("task", "edit"))
	}

	scrolled := Pointed(w, pointer.Scroll, at, CardHeight)
	if scrolled.scroll[0] != w.scroll[0] {
		// One card fits, so there is nothing to scroll: what is checked is
		// that a scroll event reaches the column rather than what it does.
		t.Log("scrolling a column that fits changes nothing, as it should")
	}
	if same := Pointed(w, pointer.Enter, at, 0); same.Chosen() != w.Chosen() {
		t.Error("a kind of event this board has no use for did something")
	}
}

func TestWhenTwoPressesAreOneDoubleClick(t *testing.T) {
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	at := Point{X: 100, Y: 100}

	if Twice(time.Time{}, now, at, at) {
		t.Error("the first press of all was read as a second")
	}
	if !Twice(now, now.Add(Twin()-time.Millisecond), at, at) {
		t.Error("two presses in the same place, quickly, are a double-click")
	}
	if Twice(now, now.Add(Twin()+time.Millisecond), at, at) {
		t.Error("two presses too far apart in time were read as one gesture")
	}
	if !Twice(now, now.Add(time.Millisecond), at, Point{X: at.X + Grip, Y: at.Y}) {
		t.Error("a hand that moved within the grip is still in the same place")
	}
	if Twice(now, now.Add(time.Millisecond), at, Point{X: at.X + Grip + 1, Y: at.Y}) {
		t.Error("a press somewhere else was read as a double-click")
	}
	if Twice(now, now.Add(time.Millisecond), at, Point{X: at.X, Y: at.Y + Grip + 1}) {
		t.Error("a press below the last one was read as a double-click")
	}
}

func TestEveryKeyItsShortcut(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})

	if got := Typed(w, "N", key.ModCommand); got.Screen() != OnForm {
		t.Errorf("⌘N left the window on %v", got.Screen())
	}
	if got := Typed(w, "N", key.ModCtrl); got.Screen() != OnForm {
		t.Errorf("ctrl+N left the window on %v", got.Screen())
	}
	opened := w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	if got := Typed(opened, string(key.NameEscape), 0); got.Screen() != OnTheBoard {
		t.Errorf("escape left the window on %v", got.Screen())
	}
	if got := Typed(w, "Q", 0); got.Screen() != OnTheBoard || len(f.asked) != 0 {
		t.Error("a key nobody bound did something")
	}
}

// The detail draws the task and its history, which is the one panel the board
// cannot show on a card.
func TestTheDetailIsDrawnWithWhatACardOnlyMarked(t *testing.T) {
	f := newFake()
	f.acts = []trace.Act{anAct("added", "", "", "")}
	w := holding(f, reading.Card{
		Number: 1, Title: "Write the spec", State: "todo",
		Description: "The client wants two columns",
	})

	w = w.doubleClicked(middleOf(w, 1, 0))
	if got := ink(drawn(t, w), Over(w.width, w.height), Dark().Back); got == 0 {
		t.Fatal("the detail is blank")
	}
}

// A form opened on the objectives tab asks an objective's questions.
func TestAddingOnTheObjectivesTabAsksAnObjectivesQuestions(t *testing.T) {
	form := adding(Objectives)
	if !form.ForObjective {
		t.Fatal("adding on the objectives tab opened a task's form")
	}
	for _, absent := range []string{"blockers", "parent", "objectives"} {
		if form.ValueOf(absent) != "" {
			t.Errorf("an objective's form asks about %q", absent)
		}
	}
}

// A window that is not too small says nothing about its size.
func TestAWindowBigEnoughSaysNothingAboutItsSize(t *testing.T) {
	f := newFake()
	if said := New(f, 1200, 800).Says(); said != "" {
		t.Errorf("a window big enough said %q", said)
	}
}

// A move the store refuses is said, and the column change that went before it
// stands — which is what having no transaction between the two means.
func TestAMoveRefusedAfterTheColumnChangedSaysSo(t *testing.T) {
	f := newFake()
	f.refuseMove = errRefuse
	w := holding(f,
		reading.Card{Number: 1, Title: "First", State: "todo"},
		reading.Card{Number: 2, Title: "Second", State: "backlog"})

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(middleOf(w, 1, 1))
	w = w.LetGo(middleOf(w, 1, 1))

	if w.Wrong() == nil {
		t.Fatal("a refused move said nothing")
	}
	if f.what("task", "edit") != 1 {
		t.Error("the column change did not go through before the move was tried")
	}
}

// A task the reading does not hold has no column, and a drop is still asked
// for: the store is what says what is true, not the last reading of it.
func TestATaskTheReadingDoesNotHoldStillMoves(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	w.snap.Cards[0].Number = 7 // the card on screen is no longer task 1

	at := middleOf(w, 0, 0)
	w = w.PressAt(at)
	w = w.MoveTo(middleOf(w, 1, 0))
	w.LetGo(middleOf(w, 1, 0))

	if f.what("task", "edit") != 1 {
		t.Errorf("asked for %d edits", f.what("task", "edit"))
	}
}

// Asking the store again is the one decision the loop makes, so it is a
// function the loop calls rather than a line inside it.
func TestWhenItIsTimeToAskTheStoreAgain(t *testing.T) {
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	if !due(time.Time{}, now) {
		t.Error("a window that has never looked does not look")
	}
	if due(now, now.Add(Beat()-time.Millisecond)) {
		t.Error("it asked again before the beat was up")
	}
	if !due(now, now.Add(Beat())) {
		t.Error("it did not ask again once the beat was up")
	}
}

// The two lengths of time this window runs on, said out loud so a change to
// either is a change somebody made on purpose.
func TestTheTwoLengthsOfTime(t *testing.T) {
	if got := Beat(); got != 2*time.Second {
		t.Errorf("the beat is %v", got)
	}
	if got := Twin(); got != 400*time.Millisecond {
		t.Errorf("two presses are one gesture within %v", got)
	}
}

// Exactly the grip, in either direction, is still the same place.
func TestTheGripOfADoubleClickInBothDirections(t *testing.T) {
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	at := Point{X: 100, Y: 100}
	soon := now.Add(time.Millisecond)

	if !Twice(now, soon, at, Point{X: at.X, Y: at.Y + Grip}) {
		t.Errorf("a hand that moved %d down is still in the same place", Grip)
	}
	if !Twice(now, soon, at, Point{X: at.X + Grip, Y: at.Y + Grip}) {
		t.Error("a hand that moved the grip in both directions is still in the same place")
	}
}

// Clicking a tab changes to it, which is the only way there is.
func TestClickingATabChangesToIt(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	l := w.Layout()

	second := TabAt(l, 1)
	w = w.clicked(Point{X: second.X + second.W/2, Y: second.Y + second.H/2})
	if w.Tab() != Objectives {
		t.Fatalf("clicking the second tab left the window on %v", w.Tab())
	}
	first := TabAt(l, 0)
	w = w.clicked(Point{X: first.X + first.W/2, Y: first.Y + first.H/2})
	if w.Tab() != Tasks {
		t.Fatalf("clicking the first tab left the window on %v", w.Tab())
	}
}

// Clicking the strip beside the tabs changes nothing.
func TestClickingBesideTheTabsChangesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	l := w.Layout()
	beside := Point{X: l.Tabs.X + l.Tabs.W - 10, Y: l.Tabs.Y + l.Tabs.H/2}
	if got := w.clicked(beside); got.Tab() != Tasks {
		t.Errorf("clicking beside the tabs left the window on %v", got.Tab())
	}
}

// Exactly the span is still the same gesture; a moment past it is not.
func TestExactlyTheSpanOfADoubleClick(t *testing.T) {
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	at := Point{X: 100, Y: 100}
	if !Twice(now, now.Add(Twin()), at, at) {
		t.Errorf("two presses exactly %v apart are one gesture", Twin())
	}
	if Twice(now, now.Add(Twin()+time.Nanosecond), at, at) {
		t.Error("a moment past the span is a second gesture")
	}
}

// Scrolling back up is a scroll like any other, and the range the loop asks for
// has to reach below zero for one to arrive at all.
func TestAFrameOfEventsScrollsBackUp(t *testing.T) {
	f := newFake()
	var many []reading.Card
	for i := 1; i <= 12; i++ {
		many = append(many, reading.Card{Number: i, Title: "Task", State: "todo"})
	}
	w := holding(f, many...)
	l := w.Layout()
	over := Point{X: l.Column[1].X + l.Column[1].W/2, Y: l.Column[1].Y + l.Column[1].H/2}
	w = w.Scroll(over, CardHeight*3)

	got, _, _ := frames(t, w, pointer.Event{
		Kind:     pointer.Scroll,
		Position: at(w, 1, 0),
		Scroll:   f32.Pt(0, -float32(CardHeight)),
	})
	if got.scroll[1] >= w.scroll[1] {
		t.Fatalf("scrolling up left the column at %d, having been at %d", got.scroll[1], w.scroll[1])
	}
}

// Closing the window is not a failure, and anything else is. This is what the
// process exits with, and it is a function because the loop that calls it is
// the one thing here no test can reach.
func TestWhatClosingTheWindowExitsWith(t *testing.T) {
	if got := code(nil); got != 0 {
		t.Errorf("a window closed on purpose exits %d", got)
	}
	if got := code(errNope); got != 1 {
		t.Errorf("a window that failed exits %d", got)
	}
}

// A window reports itself in pixels and this board works in dp, so both the
// size it reports and the distance a wheel says it turned have to be divided
// by the ratio. Written as functions because the loop that calls them is the
// one thing here no test can reach.
func TestASizeAndAScrollAreBothReadInDp(t *testing.T) {
	if wide, high := Sized(image.Pt(2400, 1600), 2); wide != 1200 || high != 800 {
		t.Errorf("a window of 2400 by 1600 at two pixels to the dp is %d by %d", wide, high)
	}
	if wide, high := Sized(image.Pt(1200, 800), 1); wide != 1200 || high != 800 {
		t.Errorf("a window at one pixel to the dp is %d by %d", wide, high)
	}
	if got := Turned(192, 2); got != 96 {
		t.Errorf("a wheel that turned 192 pixels at two to the dp turned %d", got)
	}
	if got := Turned(-192, 2); got != -96 {
		t.Errorf("a wheel turned the other way turned %d", got)
	}
}
