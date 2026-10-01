package gui

import (
	"testing"
)

// A piece of text somebody is typing into, and where the caret is in it. One
// mechanism for every text in this window — the fields of a form and the
// search line — because a caret that behaved differently in two places would
// be two things to learn.

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

// A character is a rune: the caret walks letters, not bytes.
func TestTheCaretWalksCharactersAndNotBytes(t *testing.T) {
	written := Written{Value: "réunion", At: 7}
	if got := written.Left().Left(); got.At != 5 {
		t.Fatalf("two back from the end of a seven-letter word is %d", got.At)
	}
	if got := (Written{Value: "réunion", At: 2}).Types("X"); got.Value != "réXunion" {
		t.Errorf("typing after the é gave %q", got.Value)
	}
}

func TestWhatIsTakenBackIsWhatIsBeforeTheCaret(t *testing.T) {
	if got := (Written{Value: "abcd", At: 3}).Back(); got.Value != "abd" || got.At != 2 {
		t.Errorf("taking back at 3 gave %q with the caret at %d", got.Value, got.At)
	}
	if got := (Written{Value: "abcd", At: 0}).Back(); got.Value != "abcd" {
		t.Errorf("taking back at the beginning gave %q", got.Value)
	}
	if got := (Written{}).Back(); got.Value != "" {
		t.Errorf("taking back from nothing gave %q", got.Value)
	}
}

// The caret can be put anywhere, and anywhere outside the text is its nearest
// end: a click past the last letter means the end, not an error.
func TestTheCaretIsPutWhereItIsAskedForOrAtTheNearestEnd(t *testing.T) {
	written := Written{Value: "abcdefgh"}
	if got := written.CaretAt(3); got.At != 3 {
		t.Errorf("asked for 3, the caret is at %d", got.At)
	}
	if got := written.CaretAt(99); got.At != 8 {
		t.Errorf("asked for past the end, the caret is at %d", got.At)
	}
	if got := written.CaretAt(-4); got.At != 0 {
		t.Errorf("asked for before the beginning, the caret is at %d", got.At)
	}
}

// Ending is the caret put after the last character, which is where a field
// opens: somebody who has just arrived is adding, not correcting.
func TestAFieldOpensWithTheCaretAtTheEnd(t *testing.T) {
	if got := Ending("spec"); got.At != 4 || got.Value != "spec" {
		t.Errorf("a field of %q opens with the caret at %d", got.Value, got.At)
	}
	if got := Ending(""); got.At != 0 {
		t.Errorf("an empty field opens with the caret at %d", got.At)
	}
}

// A run of text can be held, which is what selecting is: an anchor where it
// began and the caret where it stands now. Everything else here already works
// on one place; this is the only thing that works on two.
func TestWritten_holdsARunOfText(t *testing.T) {
	w := Ending("Deux colonnes")
	if w.Holding() {
		t.Error("a text nobody selected is holding a run")
	}

	// Held back from the end, the run is what lies between the two places,
	// whichever way round they are.
	held := w.CaretAt(5).Selects(9)
	if got := held.Holds(); got != "colo" {
		t.Errorf("held %q, want %q", got, "colo")
	}
	if got := w.CaretAt(9).Selects(5).Holds(); got != "colo" {
		t.Errorf("held backwards %q, want %q", got, "colo")
	}

	// Typing over a held run replaces it, which is what every text field does.
	if got := held.Types("X").Reads(); got != "Deux Xnnes" {
		t.Errorf("typing over the run gave %q", got)
	}
	// And so does deleting, in one stroke rather than four.
	if got := held.Back().Reads(); got != "Deux nnes" {
		t.Errorf("deleting the run gave %q", got)
	}
	// Putting the caret anywhere lets the run go.
	if held.CaretAt(2).Holding() {
		t.Error("moving the caret kept the run held")
	}
	// All of it is a run from one end to the other.
	if got := w.All().Holds(); got != "Deux colonnes" {
		t.Errorf("holding all of it gave %q", got)
	}
}
