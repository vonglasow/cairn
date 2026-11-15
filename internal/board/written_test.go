package board

import (
	"testing"
)

func TestTypingGoesInAtTheCaret(t *testing.T) {
	written := Written{Value: "spec", At: 2}
	if got := written.Types("XY"); got.Value != "spXYec" {
		t.Fatalf("typing in the middle gave %q", got.Value)
	}
	if got := written.Types("XY"); got.At != 4 {
		t.Errorf("the caret is at %d, and two characters were typed at 2", got.At)
	}
	if got := (Written{}).Types("a"); got.Value != "a" || got.At != 1 {
		t.Errorf("typing into nothing gave %q with the caret at %d", got.Value, got.At)
	}
	if got := written.Types(""); got != written {
		t.Error("typing nothing changed something")
	}
}

func TestTheCaretWalksAndStopsAtBothEnds(t *testing.T) {
	written := Written{Value: "ab", At: 2}
	if got := written.Right(); got.At != 2 {
		t.Errorf("right of the end is %d", got.At)
	}
	if got := written.Left().Left().Left(); got.At != 0 {
		t.Errorf("left of the beginning is %d", got.At)
	}
	if got := written.Left().Right(); got.At != 2 {
		t.Errorf("left then right is %d", got.At)
	}
}

func TestWhatIsTakenBackIsWhatIsBeforeTheCaret(t *testing.T) {
	if got := (Written{Value: "abcd", At: 3}).Back(); got.Value != "abd" || got.At != 2 {
		t.Errorf("taking back at 3 gave %q with the caret at %d", got.Value, got.At)
	}
	if got := (Written{Value: "abcd", At: 0}).Back(); got.Value != "abcd" {
		t.Errorf("taking back at the beginning gave %q", got.Value)
	}
	if got := Ending("héllo").Back(); got.Value != "héll" {
		t.Fatalf("back after an accented word gave %q", got.Value)
	}
	if got := (Written{}).Back(); got.Value != "" {
		t.Errorf("taking back from nothing gave %q", got.Value)
	}
}

func TestWritten_holdsARunOfText(t *testing.T) {
	w := Ending("Two columns")
	if w.Holding() {
		t.Error("a text nobody selected is holding a run")
	}

	held := w.CaretAt(4).Selects(8)
	if got := held.Holds(); got != "colu" {
		t.Errorf("held %q, want %q", got, "colu")
	}
	if got := w.CaretAt(8).Selects(4).Holds(); got != "colu" {
		t.Errorf("held backwards %q, want %q", got, "colu")
	}

	if got := held.Types("X").Reads(); got != "Two Xmns" {
		t.Errorf("typing over the run gave %q", got)
	}
	if got := held.Back().Reads(); got != "Two mns" {
		t.Errorf("deleting the run gave %q", got)
	}
	if held.CaretAt(2).Holding() {
		t.Error("moving the caret kept the run held")
	}
	if got := w.All().Holds(); got != "Two columns" {
		t.Errorf("holding all of it gave %q", got)
	}
}

func TestWritten_walksToEitherEndOfItsLine(t *testing.T) {
	w := Written{Value: "first line\nsecond line\nthird", At: 14}
	if got := w.LineStart(); got.At != 11 || got.Held != 0 {
		t.Errorf("the start of the line is at %d holding %d, want 11", got.At, got.Held)
	}
	if got := w.LineEnd(); got.At != 22 || got.Held != 0 {
		t.Errorf("the end of the line is at %d holding %d, want 22", got.At, got.Held)
	}
	if got := (Written{Value: "one", At: 1}).LineEnd(); got.At != 3 {
		t.Errorf("the end of the only line is at %d, want 3", got.At)
	}
	if got := (Written{Value: "a\nb", At: 2}).LineStart(); got.At != 2 {
		t.Errorf("a caret already at the start of its line moved to %d", got.At)
	}
}

func TestWritten_walksAWordAtATime(t *testing.T) {
	w := Written{Value: "write  the spec", At: 9}
	if got := w.WordBack(); got.At != 7 {
		t.Errorf("a word back is at %d, want 7", got.At)
	}
	if got := w.WordBack().WordBack(); got.At != 0 {
		t.Errorf("two words back is at %d, want 0", got.At)
	}
	if got := w.WordOn(); got.At != 10 {
		t.Errorf("a word on is at %d, want 10", got.At)
	}
	if got := w.WordOn().WordOn(); got.At != 15 {
		t.Errorf("two words on is at %d, want 15", got.At)
	}
	if got := (Written{Value: "ab", At: 0}).WordBack(); got.At != 0 {
		t.Errorf("a word back from the start is at %d", got.At)
	}
	if got := (Written{Value: "ab", At: 2}).WordOn(); got.At != 2 {
		t.Errorf("a word on from the end is at %d", got.At)
	}
}

func TestWritten_takesOutWhatIsAfterTheCaret(t *testing.T) {
	w := Written{Value: "spec", At: 1}
	if got := w.Forward(); got.Value != "sec" || got.At != 1 {
		t.Errorf("forward = %q at %d, want \"sec\" at 1", got.Value, got.At)
	}
	if got := (Written{Value: "spec", At: 4}).Forward(); got.Value != "spec" {
		t.Errorf("forward at the end took %q", got.Value)
	}
	held := Written{Value: "one\ntwo\nthree", At: 9, Held: -7}
	if got := held.Forward(); got.Value != "onhree" || got.At != 2 {
		t.Errorf("forward over a run across lines = %q at %d", got.Value, got.At)
	}
}

func TestWritten_takesBackToAPlace(t *testing.T) {
	w := Written{Value: "one\ntwo three", At: 13}
	if got := w.BackTo(w.WordBack().At); got.Value != "one\ntwo " || got.At != 8 {
		t.Errorf("back a word = %q at %d", got.Value, got.At)
	}
	if got := w.BackTo(w.LineStart().At); got.Value != "one\n" || got.At != 4 {
		t.Errorf("back to the start of the line = %q at %d", got.Value, got.At)
	}
	held := Written{Value: "one two", At: 7, Held: -3}
	if got := held.BackTo(0); got.Value != "one " {
		t.Errorf("a held run is taken back alone, whatever the place: %q", got.Value)
	}
}
