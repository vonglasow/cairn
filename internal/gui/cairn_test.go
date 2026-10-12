package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gioui.org/io/pointer"

	"cairn/internal/reading"
)

// area is a polygon's area, by the shoelace.
func area(points []Point) int {
	twice := 0
	for i, p := range points {
		q := points[(i+1)%len(points)]
		twice += p.X*q.Y - q.X*p.Y
	}
	return max(twice, -twice) / 2
}

// A card is built of stones the way an old wall is: courses of irregular
// stones fitted together, not bricks — every stone inside the card, the
// card mostly covered, the same stones every time for the same card and
// other stones for another.
func TestACardIsBuiltOfStones(t *testing.T) {
	panel := Box{X: 100, Y: 200, W: 280, H: 96}
	stones := StonesOf(panel, 7)
	if len(stones) < 4 {
		t.Fatalf("a card is built of %d stones", len(stones))
	}
	covered, rows, slanted := 0, map[int]bool{}, false
	for i, stone := range stones {
		if len(stone) < 4 {
			t.Errorf("stone %d has %d corners", i, len(stone))
		}
		top := stone[0].Y
		for j, p := range stone {
			if !panel.In(p) {
				t.Errorf("stone %d reaches %+v, outside the card %+v", i, p, panel)
			}
			q := stone[(j+1)%len(stone)]
			if p.X != q.X && p.Y != q.Y {
				slanted = true
			}
			top = min(top, p.Y)
		}
		rows[(top-panel.Y)/(panel.H/3)] = true
		covered += area(stone)
	}
	if whole := panel.W * panel.H; covered*100 < whole*70 || covered > whole {
		t.Errorf("the stones cover %d of a card of %d", covered, whole)
	}
	if len(rows) < 2 {
		t.Errorf("the stones lie in %d course", len(rows))
	}
	if !slanted {
		t.Error("every edge is square: bricks, not stones")
	}
	again := StonesOf(panel, 7)
	for i := range stones {
		for j := range stones[i] {
			if again[i][j] != stones[i][j] {
				t.Fatalf("the same card was built of other stones the second time")
			}
		}
	}
	if other := StonesOf(panel, 8); len(other) == len(stones) && other[0][1] == stones[0][1] && other[1][1] == stones[1][1] {
		t.Error("two cards are built of the same stones")
	}
}

// The cards arrive one after another: a card's stones fall, then give way to
// the card; the first lands before the last, and all are cards at the end.
func TestTheCardsArriveOneAfterAnother(t *testing.T) {
	if fall, fade := Arriving(0, 0, 5); fall != 0 || fade != 0 {
		t.Errorf("at the start the first card has fallen %v, faded %v", fall, fade)
	}
	for k := range 5 {
		if fall, fade := Arriving(1, k, 5); fall != 1 || fade != 1 {
			t.Errorf("at the end card %d has fallen %v, faded %v", k, fall, fade)
		}
	}
	first, _ := Arriving(0.4, 0, 5)
	last, _ := Arriving(0.4, 4, 5)
	if first <= last {
		t.Errorf("at two fifths the first card has fallen %v and the last %v", first, last)
	}
	for _, at := range []float64{0.1, 0.3, 0.5, 0.7, 0.9} {
		fall, fade := Arriving(at, 2, 5)
		if fade > 0 && fall < 1 {
			t.Errorf("at %v a card fades into itself before its stones have landed", at)
		}
		later, laterFade := Arriving(at+0.05, 2, 5)
		if later < fall || laterFade < fade {
			t.Errorf("at %v a card goes back up", at)
		}
	}
	// Many cards still all arrive inside the breath.
	if fall, fade := Arriving(0.999, 59, 60); fall != 1 || fade < 0.9 {
		t.Errorf("the sixtieth card of sixty has fallen %v, faded %v, at the very end", fall, fade)
	}
}

// Nothing is built over a board with no task: the animation is the content
// arriving, and there is none.
func TestNoTaskNoCairn(t *testing.T) {
	start := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	if _, on := New(newFake(), 1200, 800).Born(start).Intro(); on {
		t.Error("an empty board builds a cairn")
	}
	if _, on := holding(newFake(), card(1, "A", "todo")).Born(start).Intro(); !on {
		t.Error("a board with a task builds none")
	}
}

// The cairn is built from the first frame, for as long as it lasts, and is
// ended by a press.
func TestTheCairnLastsABreath(t *testing.T) {
	start := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	if _, on := holding(newFake(), card(1, "A", "todo")).Intro(); on {
		t.Error("a window never drawn builds a cairn")
	}
	w := holding(newFake(), card(1, "A", "todo")).Born(start)
	if progress, on := w.Intro(); !on || progress != 0 {
		t.Errorf("on its first frame the cairn is %v, %v", progress, on)
	}
	if progress, on := w.At(start.Add(IntroLasts() / 2)).Intro(); !on || progress != 0.5 {
		t.Errorf("half way through it is %v, %v", progress, on)
	}
	if _, on := w.At(start.Add(IntroLasts())).Intro(); on {
		t.Error("the cairn is still being built once its time is up")
	}
	if again := w.At(start.Add(time.Hour)).Born(start.Add(time.Hour)); again.born != start {
		t.Error("a later frame was taken for a birth")
	}
	pressed, _ := Pointing(w, Gesture{}, pointer.Press, Point{X: 1, Y: 1}, 0, start)
	if _, on := pressed.Intro(); on {
		t.Error("a press left the cairn building")
	}
	moved, _ := Pointing(w, Gesture{}, pointer.Move, Point{X: 1, Y: 1}, 0, start)
	if _, on := moved.Intro(); !on {
		t.Error("a pointer passing over ended the cairn")
	}
}

// While it is built, a card not yet arrived is not drawn — the board shows
// its ground there — and its stones are, falling; once arrived it is the card.
func TestTheCardsAreBuiltOnScreen(t *testing.T) {
	start := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	cards := []reading.Card{card(1, "First", "todo"), card(2, "Second", "todo"), card(3, "Third", "todo")}
	w := holding(newFake(), cards...).Born(start)
	l := w.Layout()
	last := PartsOf(l.CardAt(1, 2, 0)).Panel
	early := w.At(start)
	if got := painted(drawn(t, early), last, Dark().Panel); got != 0 {
		t.Errorf("before it arrives, %d pixels of the last card are drawn", got)
	}
	// Half way through its fall, its stones are above where it lands.
	at := 0.0
	for fall, _ := Arriving(at, 2, 3); fall < 0.5; fall, _ = Arriving(at, 2, 3) {
		at += 0.01
	}
	falling := w.At(start.Add(time.Duration(at * float64(IntroLasts()))))
	above := Box{X: last.X, Y: 0, W: last.W, H: last.Y}
	if got := inked(drawn(t, falling), above, Dark().Back); got == 0 {
		t.Error("half way through its fall nothing of the last card is above it")
	}
	ended := w.At(start.Add(IntroLasts()))
	if got := painted(drawn(t, ended), Box{X: last.X + last.W/2, Y: last.Y + 2, W: 1, H: 1}, Dark().Panel); got != 1 {
		t.Error("once built the card is not drawn")
	}
}

// The stones of a few cards, and how the cards arrive, to the point and the
// millionth: a sum gone wrong builds another wall, or another arrival, and
// nothing else would say so. UPDATE_STONES=1 writes it anew, after a change
// meant to build other stones.
func TestTheStonesAreTheSameStones(t *testing.T) {
	var out strings.Builder
	for _, panel := range []Box{{X: 100, Y: 200, W: 280, H: 96}, {X: 0, Y: 0, W: 150, H: 60}} {
		for seed := 1; seed <= 4; seed++ {
			fmt.Fprintf(&out, "%+v seed %d:", panel, seed)
			for _, stone := range StonesOf(panel, seed) {
				fmt.Fprintf(&out, " %v", stone)
			}
			out.WriteString("\n")
		}
	}
	for _, at := range []float64{0, 0.1, 0.25, 0.4, 0.55, 0.7, 0.85, 1} {
		for _, k := range []int{0, 2, 4} {
			fall, fade := Arriving(at, k, 5)
			fmt.Fprintf(&out, "at %.2f card %d of 5: fell %.6f faded %.6f\n", at, k, fall, fade)
		}
	}
	path := filepath.Join("testdata", "stones.txt")
	if os.Getenv("UPDATE_STONES") != "" {
		if err := os.WriteFile(path, []byte(out.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v — run with UPDATE_STONES=1 to write it", err)
	}
	if string(want) != out.String() {
		t.Errorf("the stones moved:\n%s\nwant:\n%s", out.String(), want)
	}
}

// What the building is drawn by, at its edges.
func TestWhatTheBuildingIsDrawnBy(t *testing.T) {
	if got := cut(0, 200, 100); got != 100 {
		t.Errorf("leaving a stone's width, a cut is at %d", got)
	}
	if got := cut(0, 200, 200-stoneLeast); got != 200-stoneLeast {
		t.Errorf("leaving exactly the narrowest stone, a cut is at %d", got)
	}
	if got := cut(0, 200, 200-stoneLeast+1); got != 200 {
		t.Errorf("leaving less than a stone, a cut is at %d", got)
	}
	if got := cut(0, 200, 300); got != 200 {
		t.Errorf("past the course, a cut is at %d", got)
	}
	panel := Box{X: 10, Y: 100, W: 50, H: 40}
	if Dropped(panel, 0) != -140 || Dropped(panel, 1) != 0 || Dropped(panel, 0.5) != -70 {
		t.Errorf("drops are %d, %d, %d", Dropped(panel, 0), Dropped(panel, 1), Dropped(panel, 0.5))
	}
	if got := Fallen([]Point{{X: 1, Y: 2}, {X: 3, Y: 4}}, -10); got[0] != (Point{X: 1, Y: -8}) || got[1] != (Point{X: 3, Y: -6}) {
		t.Errorf("a stone fell to %v", got)
	}
	within := Box{X: 0, Y: 100, W: 50, H: 200}
	for slot, want := range map[Box]bool{
		{Y: 300, H: 10}: false, {Y: 299, H: 10}: true,
		{Y: 90, H: 10}: false, {Y: 91, H: 10}: true,
	} {
		if got := (Placed{Slot: slot, Within: within}).seen(); got != want {
			t.Errorf("a card at %d, %d high, seen: %v", slot.Y, slot.H, got)
		}
	}
}
