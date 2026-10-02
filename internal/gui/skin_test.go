package gui

import (
	"image/color"
	"math"
	"testing"
)

// The palette, held to the ratios it has to meet. Everything else in this
// repository is held by a gate; the palette was held by nothing, and drifted
// until a card on the board stood at 1.10 against it.
//
// The thresholds are the ones a person reading a screen needs: 4.5 for text,
// 3 for a shape that carries meaning by being seen at all — a border, a track,
// the edge of a card.
func TestEverySkinCanBeRead(t *testing.T) {
	for _, skin := range []struct {
		what string
		Skin
	}{{"dark", Dark()}} {
		ink := []struct {
			what string
			on   color.NRGBA
			said color.NRGBA
		}{
			{"ink on a panel", skin.Panel, skin.Ink},
			{"ink on the board", skin.Back, skin.Ink},
			{"faint on a panel", skin.Panel, skin.Faint},
			{"faint on the board", skin.Back, skin.Faint},
			{"an offer on the board", skin.Back, skin.Chosen},
			{"an error on the board", skin.Back, skin.Wrong},
			{"a day gone by, on a panel", skin.Panel, skin.Late},
			{"a day gone by, on the board", skin.Back, skin.Late},
		}
		for _, c := range ink {
			if got := Contrast(c.said, c.on); got < 4.5 {
				t.Errorf("%s: %s reads at %.2f, and text needs 4.5", skin.what, c.what, got)
			}
		}

		// A card is told from the board by its edge, not by its fill: a dark
		// theme cannot lift a panel to three times the ground without ceasing
		// to be dark. So the edge carries the threshold, and the fill only has
		// to be visible at all.
		shapes := []struct {
			what string
			on   color.NRGBA
			said color.NRGBA
			at   float64
		}{
			{"an edge against the board", skin.Back, skin.Edge, 3},
			{"an edge against a panel", skin.Panel, skin.Edge, 2},
			{"a panel against the board", skin.Back, skin.Panel, 1.3},
		}
		for _, c := range shapes {
			if got := Contrast(c.said, c.on); got < c.at {
				t.Errorf("%s: %s stands at %.2f, and it needs %.1f", skin.what, c.what, got, c.at)
			}
		}

		// A priority is told from the others and from what says "you are here".
		// They were the same blue at the bit, so a card of ordinary priority
		// wore the colour of being selected.
		for name, said := range skin.Priority {
			if near(said, skin.Chosen) {
				t.Errorf("%s: priority %q is the colour of what is chosen", skin.what, name)
			}
			if close(said, skin.Edge) {
				t.Errorf("%s: priority %q is the colour of no priority at all", skin.what, name)
			}
			for other, also := range skin.Priority {
				if other != name && close(said, also) {
					t.Errorf("%s: priority %q and %q are the same colour", skin.what, name, other)
				}
			}
		}
		// And an error is told from the priority that shouts loudest, which it
		// sat beside at 1.11.
		if near(skin.Wrong, skin.Priority["critical"]) {
			t.Errorf("%s: an error is the colour of a critical task", skin.what)
		}
	}
}

// near says whether two colours are close enough to be read as one. Measured
// as a distance between them and not as a contrast: contrast weighs light
// alone, so a blue and an orange of the same lightness come back identical —
// which is how this test first accused a palette of a fault it did not have.
// Sixty between a priority and what is chosen, where the fault was; forty-five
// between two things that merely have to be told apart.
func near(a, b color.NRGBA) bool  { return howFar(a, b) < 60 }
func close(a, b color.NRGBA) bool { return howFar(a, b) < 45 }

func howFar(a, b color.NRGBA) float64 {
	dr, dg, db := float64(a.R)-float64(b.R), float64(a.G)-float64(b.G), float64(a.B)-float64(b.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func TestContrast_answersTheKnownRatios(t *testing.T) {
	black := color.NRGBA{A: 255}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	if got := Contrast(white, black); math.Abs(got-21) > 0.01 {
		t.Errorf("white on black is %.2f, want 21", got)
	}
	if got := Contrast(black, black); math.Abs(got-1) > 0.01 {
		t.Errorf("black on black is %.2f, want 1", got)
	}
	// Either way round, since a ratio has no direction.
	if Contrast(white, black) != Contrast(black, white) {
		t.Error("the ratio changed when the two were swapped")
	}
}

// The curve a screen applies, straightened out. Pinned at the two values
// either side of where it changes shape, because that is the only place the
// two halves of it disagree — and at the two ends, which every ratio rests on.
func TestTheCurveIsStraightenedWhereItChangesShape(t *testing.T) {
	// To nine decimals, and that is not fussiness: the curve is continuous, so
	// its two halves agree at the join to within a millionth. A tolerance any
	// looser would accept either of them on the value that decides which is
	// used, and the comparison that chooses would be held by nothing.
	for _, c := range []struct {
		v    uint8
		want float64
	}{
		{0, 0},
		{9, 0.0027317429},
		{10, 0.0030352698},
		{11, 0.0033465358},
		{255, 1},
	} {
		if got := straight(c.v); math.Abs(got-c.want) > 0.000000001 {
			t.Errorf("straight(%d) = %.10f, want %.10f", c.v, got, c.want)
		}
	}
}

// Black against white is the widest contrast there is, and a colour against
// itself is none at all. Every judgement this file makes rests on those two.
func TestContrastIsTwentyOneAtItsWidestAndOneAtItsNarrowest(t *testing.T) {
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.NRGBA{A: 255}
	if got := Contrast(black, white); math.Abs(got-21) > 0.01 {
		t.Errorf("black on white is %.2f, want 21", got)
	}
	if got := Contrast(white, white); math.Abs(got-1) > 0.01 {
		t.Errorf("white on white is %.2f, want 1", got)
	}
}
