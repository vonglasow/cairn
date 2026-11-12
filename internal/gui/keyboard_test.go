package gui

import (
	"cairn/internal/engine"
	"math"
	"strings"
	"testing"

	"gioui.org/font"
	"gioui.org/text"
	"golang.org/x/image/math/fixed"

	"gioui.org/io/key"
	"gioui.org/io/pointer"
)

func TestAKeyIsWrittenTheWayItsSystemWritesIt(t *testing.T) {
	for _, c := range []struct{ system, said, want string }{
		{"darwin", "⌘N", "⌘N"},
		{"darwin", "⇧A", "⇧A"},
		{"windows", "⌘N", "Ctrl+N"},
		{"linux", "⌘⏎ or ⌘S in a form", "Ctrl+⏎ or Ctrl+S in a form"},
		{"windows", "⇧←→ in a field", "Shift+←→ in a field"},
		{"windows", "esc", "esc"},
		{"windows", "⌥←→ a word", "Alt+←→ a word"},
		{"linux", "⌦ ⌘⌫ ⌥⌫", "Delete Ctrl+Backspace Alt+Backspace"},
		{"darwin", "⌦ ⌘⌫ ⌥⌫", "⌦ ⌘⌫ ⌥⌫"},
	} {
		if got := KeyboardOf(c.system).Spelled(c.said); got != c.want {
			t.Errorf("on %s, %q is written %q, want %q", c.system, c.said, got, c.want)
		}
	}
}

func TestTheWindowWritesItsKeysForItsSystem(t *testing.T) {
	w := New(newFake(), 1200, 800).OnSystem("windows")
	for _, line := range w.Helping() {
		if strings.ContainsAny(line.Key+line.Means, "⌘⇧⌥⌫⌦") {
			t.Errorf("the help says %q — %q on Windows", line.Key, line.Means)
		}
	}
	if got := w.Said("New"); got != "New  Ctrl+N" {
		t.Errorf("the bar writes %q, want the key a Windows keyboard has", got)
	}
	if got := New(newFake(), 1200, 800).OnSystem("darwin").Said("New"); got != "New  ⌘N" {
		t.Errorf("on a Mac the bar writes %q", got)
	}
}

func TestALetterWrittenWithAltIsNoShortcut(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	chosen := w.clicked(middleOf(w, 1, 0))
	if got := Typed(chosen, "X", key.ModCtrl|key.ModAlt); f.what("task", "archive") != 0 || got.Screen() != OnTheBoard {
		t.Errorf("AltGr+X archived, or left the window on %v", got.Screen())
	}
	if got := Typed(chosen, "X", key.ModAlt); f.what("task", "archive") != 0 || got.Screen() != OnTheBoard {
		t.Errorf("⌥X archived, or left the window on %v", got.Screen())
	}
	if got := Typed(chosen, string(key.NameDeleteBackward), key.ModAlt); got.Screen() != OnTheBoard {
		t.Errorf("⌥⌫ on the board left the window on %v", got.Screen())
	}
}

func TestARightPressEndsTheCairn(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"}).Born(moment)
	w, _, _ = frames(t, w, pointer.Event{Kind: pointer.Press, Position: at(w, 1, 0), Buttons: pointer.ButtonSecondary})
	if _, on := w.Intro(); on {
		t.Error("a right press left the cairn building")
	}
}

func TestTheBarUnderAQuestionKeepsTheBoardsKeys(t *testing.T) {
	w := holding(newFake(), card(1, "Write the spec", "todo")).OnSystem("darwin")
	w = w.clicked(middleOf(w, 1, 0)).Key("cmd+backspace")
	if w.Screen() != OnAsking {
		t.Fatalf("delete left the window on %v", w.Screen())
	}
	if got := w.BarKey("Delete"); got != "⌘⌫" {
		t.Errorf("under the question the bar says Delete is %q", got)
	}
	if got := w.OfferKey("Delete"); got != "⏎" {
		t.Errorf("the question says its Delete is %q", got)
	}
}

// rasterSlack is the room a key chip keeps over what this system writes its key in: the face standing in for a mark draws it a point or two wider on another system, Linux in particular.
const rasterSlack = 3

func TestAKeyChipHasRoomForTheMarksGoMonoLacks(t *testing.T) {
	th := drawingTheme()
	monoWritten := func(said string) int {
		th.Shaper.LayoutString(text.Parameters{
			Font: font.Font{Typeface: MonoFace, Weight: font.SemiBold}, PxPerEm: fixed.I(small),
			MaxLines: 1, MaxWidth: math.MaxInt32,
		}, said)
		end := 0
		for g, ok := th.Shaper.NextGlyph(); ok; g, ok = th.Shaper.NextGlyph() {
			end = max(end, (g.X + g.Advance).Ceil())
		}
		return end
	}
	for _, key := range []string{"Ctrl+⏎", "⌘⌫", "⇧F", "Ctrl+K"} {
		if got, want := KeyChipWide(key), monoWritten(key)+2*chipInset+rasterSlack; got < want {
			t.Errorf("the chip of %s is %d wide, its key is written %d wide", key, got, want-2*chipInset)
		}
	}
}
