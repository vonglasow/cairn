package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"errors"
	"image"
	"strconv"
	"strings"
	"testing"
	"time"

	"gioui.org/io/pointer"
)

func TestEveryPointerEventItsGesture(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	at := middleOf(w, 0, 0)

	pressed := Pointed(w, pointer.Press, at, 0)
	if pressed.held.Number != 1 {
		t.Errorf("a press took hold of %d", pressed.held.Number)
	}
	dragged := Pointed(pressed, pointer.Drag, geom.Point{X: at.X + 50, Y: at.Y}, 0)
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

	scrolled := Pointed(w, pointer.Scroll, at, geom.CardHeight)
	if scrolled.scroll[0] != w.scroll[0] {
		t.Log("scrolling a column that fits changes nothing, as it should")
	}
	if same := Pointed(w, pointer.Enter, at, 0); same.Chosen != w.Chosen {
		t.Error("a kind of event this board has no use for did something")
	}
}

func TestAWindowBigEnoughSaysNothingAboutItsSize(t *testing.T) {
	f := newFake()
	if said := New(f, 1200, 800).Says(); said != "" {
		t.Errorf("a window big enough said %q", said)
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

func TestPasting_takesWhatTheClipboardHeld(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title").Key("cmd+a")
	if !w.Pasting() {
		w = w.Key("cmd+v")
	}

	got := pasting(w, strings.NewReader("colonnes"))
	if said := valueOf(got.Form, "title"); said != "colonnes" {
		t.Errorf("the title reads %q after pasting", said)
	}
	if got.Pasting() {
		t.Error("the window is still waiting for a clipboard it has been given")
	}
	broken := pasting(w, failing{})
	if said := valueOf(broken.Form, "title"); said != "Write the spec" {
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
	cards := []board.Card{{Number: 1, Title: "Ship it", State: "todo", Description: long}}
	for i := 2; i <= 40; i++ {
		cards = append(cards, board.Card{Number: i, Title: "Other " + strconv.Itoa(i), State: "todo"})
	}
	w := holding(newFake(), cards...)
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	frame, _, _, _, _ := w.sayingOn()
	over := geom.Point{X: max(frame.X, w.Layout().Column[1].X) + 10, Y: frame.Y + 10}
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
	if w.Screen != board.OnTheBoard {
		t.Fatalf("esc left the window on %v", w.Screen)
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
	if w.Screen != board.OnForm {
		t.Fatalf("the task did not open: %v", w.Screen)
	}
	w = turn(w, fresh+16*time.Millisecond)
	if w.Saying() != 0 {
		t.Fatalf("the rest of a turn of the board scrolled the description to %d", w.Saying())
	}
}

func TestTheNativeMenuIsMadeOnTheWindowsOwnThread(t *testing.T) {
	ran := false
	run := func(f func()) {
		ran = true
		f()
	}
	if !onTheWindowsThread(run, func() bool { return ran }) {
		t.Error("the menu was installed outside what runs on the window's thread")
	}
}

func TestASymbolTypedIsTheKeyWhenThePressNamedAnother(t *testing.T) {
	for _, c := range []struct {
		text, named string
		key         string
		taken       bool
	}{
		{"?", ",", "?", true},
		{"/", "/", "", false},
		{"?", "?", "", false},
		{"/", ":", "/", true},
		{"a", "A", "", false},
		{"7", "7", "", false},
		{"ab", "", "", false},
		{" ", "Space", "", false},
	} {
		if key, taken := symbolTyped(c.text, c.named); key != c.key || taken != c.taken {
			t.Errorf("%q typed after a press named %q is taken as %q %v", c.text, c.named, key, taken)
		}
	}
}
