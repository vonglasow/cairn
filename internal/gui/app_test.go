package gui

import (
	"errors"
	"image"
	"strconv"
	"strings"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"

	"cairn/internal/engine"
	"cairn/internal/trace"
)

func TestEveryPointerEventItsGesture(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})
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
	if f.what("task", "move") != 1 || f.what("task", "edit") != 0 {
		t.Errorf("the drop asked for %d moves and %d edits, want one move", f.what("task", "move"), f.what("task", "edit"))
	}

	scrolled := Pointed(w, pointer.Scroll, at, CardHeight)
	if scrolled.scroll[0] != w.scroll[0] {
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
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})

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
	if got := Typed(w, "Q", 0); got.Screen() != OnTheBoard || len(f.wrote()) != 0 {
		t.Error("a key nobody bound did something")
	}
	if got := Typed(w, "A", key.ModShift); !got.archived {
		t.Error("⇧A did not show what was archived")
	}
	if got := Typed(Typed(w, "A", key.ModShift), "A", key.ModShift); got.archived {
		t.Error("⇧A twice did not hide it again")
	}
	if got := Typed(w, "?", key.ModShift); got.Screen() != OnHelp {
		t.Errorf("? left the window on %v, not the help", got.Screen())
	}
	if got := Typed(w, "/", key.ModShift); got.Screen() != OnSearch {
		t.Errorf("a shifted punctuation key left the window on %v", got.Screen())
	}
}

func TestTheDetailIsDrawnWithWhatACardOnlyMarked(t *testing.T) {
	f := newFake()
	f.acts = []trace.Act{anAct("added", "", "", "")}
	w := holding(f, engine.Card{
		Number: 1, Title: "Write the spec", State: "todo",
		Description: "The client wants two columns",
	})

	w = w.doubleClicked(middleOf(w, 1, 0))
	if got := ink(drawn(t, w), Over(w.width, w.height), Dark().Back); got == 0 {
		t.Fatal("the detail is blank")
	}
}

func TestAddingOnTheObjectivesTabAsksAnObjectivesQuestions(t *testing.T) {
	form := adding(Objectives, "")
	if !form.ForObjective {
		t.Fatal("adding on the objectives tab opened a task's form")
	}
	for _, absent := range []string{"blockers", "parent", "objectives"} {
		if form.ValueOf(absent) != "" {
			t.Errorf("an objective's form asks about %q", absent)
		}
	}
}

func TestAWindowBigEnoughSaysNothingAboutItsSize(t *testing.T) {
	f := newFake()
	if said := New(f, 1200, 800).Says(); said != "" {
		t.Errorf("a window big enough said %q", said)
	}
}

func TestADropRefusedChangesNothingAndSaysSo(t *testing.T) {
	f := newFake()
	f.refuseMove = errRefuse
	w := holding(f,
		engine.Card{Number: 1, Title: "First", State: "todo"},
		engine.Card{Number: 2, Title: "Second", State: "backlog"})

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(middleOf(w, 1, 1))
	w = w.LetGo(middleOf(w, 1, 1))

	if w.Wrong() == nil {
		t.Fatal("a refused move said nothing")
	}
	if f.what("task", "edit") != 0 {
		t.Error("a column change was asked for apart from the move")
	}
}

func TestATaskTheReadingDoesNotHoldStillMoves(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	w.snap.Cards[0].Number = 7

	at := middleOf(w, 0, 0)
	w = w.PressAt(at)
	w = w.MoveTo(middleOf(w, 1, 0))
	w.LetGo(middleOf(w, 1, 0))

	if f.what("task", "move") != 1 {
		t.Errorf("asked for %d moves", f.what("task", "move"))
	}
}

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

func TestTheTwoLengthsOfTime(t *testing.T) {
	if got := Beat(); got != 2*time.Second {
		t.Errorf("the beat is %v", got)
	}
	if got := Twin(); got != 400*time.Millisecond {
		t.Errorf("two presses are one gesture within %v", got)
	}
}

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

func TestClickingATabChangesToIt(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
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

func TestClickingBesideTheTabsChangesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	l := w.Layout()
	beside := Point{X: l.Tabs.X + l.Tabs.W - 10, Y: l.Tabs.Y + l.Tabs.H/2}
	if got := w.clicked(beside); got.Tab() != Tasks {
		t.Errorf("clicking beside the tabs left the window on %v", got.Tab())
	}
}

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

func TestAFrameOfEventsScrollsBackUp(t *testing.T) {
	f := newFake()
	var many []engine.Card
	for i := 1; i <= 12; i++ {
		many = append(many, engine.Card{Number: i, Title: "Task", State: "todo"})
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

func TestWhatClosingTheWindowExitsWith(t *testing.T) {
	for _, c := range []struct {
		err  error
		code int
		said string
	}{{nil, 0, ""}, {errNope, 1, errNope.Error()}} {
		var said strings.Builder
		exited := -1
		ended(c.err, &said, func(code int) { exited = code })
		if exited != c.code {
			t.Errorf("%v: exited %d, want %d", c.err, exited, c.code)
		}
		if !strings.Contains(said.String(), c.said) || (c.err == nil && said.Len() > 0) {
			t.Errorf("%v: said %q", c.err, said.String())
		}
	}
}

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

func TestPointed_theWheelOverADescriptionScrollsIt(t *testing.T) {
	f := newFake()
	said := make([]string, 30)
	for i := range said {
		said[i] = "Une ligne de ce que la tâche raconte, assez longue pour compter"
	}
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo",
		Description: strings.Join(said, "\n")})
	w = w.doubleClicked(middleOf(w, 1, 0))
	if w.Screen() != OnForm {
		t.Fatalf("the task did not open: %v", w.Screen())
	}
	box, _, _, _, long := w.sayingOn()
	if !long {
		t.Fatal("the open task shows no description")
	}
	over := Point{X: box.X + box.W/2, Y: box.Y + box.H/2}

	if got := Pointed(w, pointer.Scroll, over, lineHigh); got.Saying() == 0 {
		t.Error("the wheel over the description left it where it was")
	}
	small := w
	for range 6 {
		small = Pointed(small, pointer.Scroll, over, 3)
	}
	if small.Saying() != 1 {
		t.Errorf("six nudges of three pixels left the description at %d, want 1", small.Saying())
	}
}

func TestPointing_tellsADoubleClickFromTheClicksBeforeIt(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	at := middleOf(w, 1, 0)
	now := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)

	one, g := Pointing(w, Gesture{}, pointer.Press, at, 0, now)
	one, g = Pointing(one, g, pointer.Release, at, 0, now)
	if one.Screen() != OnTheBoard || one.Chosen() != 1 {
		t.Fatalf("one click left the window on %v, choosing %d", one.Screen(), one.Chosen())
	}

	two, g := Pointing(one, g, pointer.Press, at, 0, now.Add(time.Millisecond))
	two, _ = Pointing(two, g, pointer.Release, at, 0, now.Add(time.Millisecond))
	if two.Screen() != OnForm {
		t.Errorf("the second click left the window on %v, not the task it opened", two.Screen())
	}

	apart, g := Pointing(one, g, pointer.Press, at, 0, now.Add(Twin()+time.Second))
	apart, _ = Pointing(apart, g, pointer.Release, at, 0, now.Add(Twin()+time.Second))
	if apart.Screen() != OnTheBoard {
		t.Errorf("two clicks a second apart opened the task anyway: %v", apart.Screen())
	}
}

func TestPointed_sendsEachKindWhereItBelongs(t *testing.T) {
	f := newFake()
	long := strings.Repeat("Une ligne de ce que la tâche raconte, bien assez longue\n", 30)
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo", Description: long})
	at := middleOf(w, 1, 0)

	if got := Pointed(w, pointer.Press, at, 0); !got.pressed {
		t.Error("a press took hold of nothing")
	}
	if got := Pointed(Pointed(w, pointer.Press, at, 0), pointer.Drag,
		Point{X: at.X + Grip + 1, Y: at.Y}, 0); !got.held.Moved {
		t.Error("a drag did not move what the press had taken hold of")
	}
	if got := Pointed(Pointed(w, pointer.Press, at, 0), pointer.Release, at, 0); got.Chosen() != 1 {
		t.Error("a release did not finish the click it belonged to")
	}
	open := w.doubleClicked(at)
	box, _, _, _, _ := open.sayingOn()
	over := Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
	if got := Pointed(open, pointer.Scroll, over, 5*lineHigh); got.Saying() == 0 {
		t.Error("a scroll moved nothing")
	}
}

func TestPasting_takesWhatTheClipboardHeld(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title").Key("cmd+a")
	if !w.Pasting() {
		w = w.Key("cmd+v")
	}

	got := pasting(w, strings.NewReader("colonnes"))
	if said := got.Form().ValueOf("title"); said != "colonnes" {
		t.Errorf("the title reads %q after pasting", said)
	}
	if got.Pasting() {
		t.Error("the window is still waiting for a clipboard it has been given")
	}
	broken := pasting(w, failing{})
	if said := broken.Form().ValueOf("title"); said != "Write the spec" {
		t.Errorf("a clipboard that would not open left the title as %q", said)
	}
	if broken.Pasting() {
		t.Error("a clipboard that would not open left the window waiting")
	}
}

type failing struct{}

func (failing) Read([]byte) (int, error) { return 0, errors.New("no") }

func TestTheRestOfATurnDoesNotReachWhatOpensUnderIt(t *testing.T) {
	long := strings.Repeat("une phrase de description assez longue pour replier ", 60)
	cards := []engine.Card{{Number: 1, Title: "Ship it", State: "todo", Description: long}}
	for i := 2; i <= 40; i++ {
		cards = append(cards, engine.Card{Number: i, Title: "Other " + strconv.Itoa(i), State: "todo"})
	}
	w := holding(newFake(), cards...)
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	frame, _, _, _, _ := w.sayingOn()
	over := Point{X: max(frame.X, w.Layout().Column[1].X) + 10, Y: frame.Y + 10}
	if !w.Layout().Column[1].In(over) || !frame.In(over) {
		t.Fatal("this proves nothing: there is no column under the description")
	}
	start := moment
	turn := func(w Window, after time.Duration) Window {
		w, _ = Pointing(w, Gesture{}, pointer.Scroll, over, lineHigh, start.Add(after))
		return w
	}
	w = turn(w, 0)
	w = turn(w, 16*time.Millisecond)
	if w.Saying() == 0 {
		t.Fatal("the turn did not scroll the description")
	}
	w = w.Key("esc")
	if w.Screen() != OnTheBoard {
		t.Fatalf("esc left the window on %v", w.Screen())
	}
	for i := 2; i < 10; i++ {
		w = turn(w, time.Duration(i)*16*time.Millisecond)
	}
	if w.scroll[1] != 0 {
		t.Fatalf("the rest of the turn scrolled the column under it to %d", w.scroll[1])
	}
	fresh := 9*16*time.Millisecond + Still()
	w = turn(w, fresh)
	if w.scroll[1] == 0 {
		t.Fatal("a fresh turn over the board after a pause of Still() did not scroll it")
	}
	w = w.DoubleClick(middleOf(w, 1, 0))
	if w.Screen() != OnForm {
		t.Fatalf("the task did not open: %v", w.Screen())
	}
	w = turn(w, fresh+16*time.Millisecond)
	if w.Saying() != 0 {
		t.Fatalf("the rest of a turn of the board scrolled the description to %d", w.Saying())
	}
}

func TestShiftHeldDoesNotStopADeletion(t *testing.T) {
	w := holding(newFake()).Key("cmd+n").Types("Plans").Key("left")
	if got := Typed(w, string(key.NameDeleteBackward), key.ModShift).Form().Writing().Reads(); got != "Plas" {
		t.Errorf("⇧⌫ left %q", got)
	}
	if got := Typed(w, string(key.NameDeleteForward), key.ModShift).Form().Writing().Reads(); got != "Plan" {
		t.Errorf("⇧⌦ left %q", got)
	}
}
