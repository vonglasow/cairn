package gui

import (
	"image"
	"strings"
	"testing"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"cairn/internal/engine"
	"cairn/internal/trace"
)

func openTask(t *testing.T) (Window, *fake) {
	t.Helper()
	f := newFake()
	f.acts = []trace.Act{anAct("changed", "title", "Write the spec", "")}
	w := holding(f,
		engine.Card{Number: 1, Title: "Write the spec", State: "todo", Labels: []string{"backend"},
			Description: "Deux colonnes et une bande", Context: "/vault/1"},
		engine.Card{Number: 2, Title: "Relire l'ADR", State: "todo", Parent: 1})
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	if w.Screen() != OnForm {
		t.Fatalf("the task did not open: %v", w.Screen())
	}
	return w, f
}

func readable(t *testing.T, w Window, key string) Readable {
	t.Helper()
	for _, r := range w.Readables() {
		if r.Key == key {
			return r
		}
	}
	t.Fatalf("the column beside the form writes nothing called %q", key)
	return Readable{}
}

func on(r Readable, at int) Point {
	return Point{X: r.Box.X + at*glyphOf(r.Size), Y: r.Box.Y + lineHigh/2}
}

func TestEveryTextAnOpenTaskWritesCanBeCopied(t *testing.T) {
	w, _ := openTask(t)
	for _, c := range []struct {
		key, word string
	}{
		{"history", "TRACE"},
		{"name:context", "context"},
	} {
		t.Run(c.key, func(t *testing.T) {
			r := readable(t, w, c.key)
			got, _ := w.DoubleClick(on(r, 0)).Key("cmd+c").Copying()
			if got != c.word {
				t.Fatalf("a double click on the first word of %q copied %q, want %q", c.key, got, c.word)
			}
		})
	}
	history := readable(t, w, "history:0")
	at := strings.Index(history.Said, "changed")
	if at < 0 {
		t.Fatalf("the history reads %q", history.Said)
	}
	if got, _ := w.DoubleClick(on(history, at)).Key("cmd+c").Copying(); got != "changed" {
		t.Fatalf("a double click on a word of the history copied %q", got)
	}
}

func TestARunOfAnOpenTasksTextIsHeldByDraggingAcrossIt(t *testing.T) {
	w, _ := openTask(t)
	r := readable(t, w, "field:context")
	w = w.PressAt(on(r, 0)).MoveTo(on(r, 6)).LetGo(on(r, 6))
	if got, _ := w.Key("cmd+c").Copying(); got != "/vault" {
		t.Fatalf("dragging across the first six letters copied %q", got)
	}
	w = w.PressAt(on(r, 2)).LetGo(on(r, 2))
	if got, wants := w.Key("cmd+c").Copying(); wants {
		t.Fatalf("a click held %q", got)
	}
}

func TestAllOfATextIsHeldAtOnce(t *testing.T) {
	w, _ := openTask(t)
	r := readable(t, w, "history:0")
	w = w.PressAt(on(r, 0)).LetGo(on(r, 0))
	if got, _ := w.Key("cmd+a").Key("cmd+c").Copying(); got != r.Said {
		t.Fatalf("⌘A then ⌘C copied %q", got)
	}
}

func TestWhatIsHeldIsWhatIsPainted(t *testing.T) {
	w, _ := openTask(t)
	r := readable(t, w, "name:context")
	w = w.DoubleClick(on(r, 0))
	if held := w.Holding("name:context", r.Said); held.Holds() != "context" {
		t.Fatalf("the name of the context is drawn holding %q", held.Holds())
	}
	if held := w.Holding("history", "TRACE"); held.Holding() {
		t.Fatalf("the history's head is drawn holding %q, though nothing was picked in it", held.Holds())
	}
}

func TestAWordInTheFormIsHeldByADoubleClick(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	box := w.formBox()
	fields := w.Form().Fields()
	value := formValue(box, fields, 0)
	inside := Inside(Framed(value, formRows(box, fields, 0)))
	at := Point{X: inside.X + 11*perGlyph + perGlyph/2, Y: inside.Y + lineHigh/2}
	w, _ = Pointing(w, Gesture{}, pointer.Press, at, 0, moment)
	if got := w.DoubleClick(at).Form().Writing().Holds(); got != "spec" {
		t.Fatalf("a double click on the third word of the title held %q", got)
	}
}

func TestWhatIsHeldOnAnOpenTaskIsPainted(t *testing.T) {
	for _, key := range []string{"field:context", "history:0"} {
		t.Run(key, func(t *testing.T) {
			w, _ := openTask(t)
			r := readable(t, w, key)
			before := held(drawn(t, w), r)
			w = w.DoubleClick(on(r, 0))
			if after := held(drawn(t, w), r); after <= before {
				t.Fatalf("holding a word of %q painted %d points of the held ground, against %d before", key, after, before)
			}
		})
	}
}

func held(shot *image.RGBA, r Readable) int {
	ghost, count := Dark().Ghost, 0
	for y := r.Box.Y; y < r.Box.Y+lineHigh; y++ {
		for x := r.Box.X; x < r.Box.X+r.Box.W; x++ {
			if got := shot.RGBAAt(x, y); got.R == ghost.R && got.G == ghost.G && got.B == ghost.B {
				count++
			}
		}
	}
	return count
}

func TestACharacterIsReckonedAtEverySize(t *testing.T) {
	for size, want := range map[int]int{
		small - 1: perSmallGlyph, small: perSmallGlyph,
		reads: perGlyph, plain: perGlyph, heading - 1: perGlyph,
		heading: perGlyph + 1, heading + 1: perGlyph + 1,
	} {
		if got := glyphOf(size); got != want {
			t.Errorf("at %d a character is reckoned %d wide, want %d", size, got, want)
		}
	}
}

func TestTheCharacterUnderAPoint(t *testing.T) {
	r := Readable{Box: Box{X: 100, Y: 200, W: 300, H: 3 * lineHigh}, Size: reads,
		Lines: Wrapped("un deux\ntrois\nquatre", 40)}
	at := func(column, row int) Point {
		return Point{X: r.Box.X + column*perGlyph + 1, Y: r.Box.Y + row*lineHigh + 1}
	}
	for _, c := range []struct {
		what string
		p    Point
		from int
		want int
	}{
		{"the first character", at(0, 0), 0, 0},
		{"two in on the second line", at(2, 1), 0, 10},
		{"the first line in sight, scrolled one down", at(1, 0), 1, 9},
		{"above the text", Point{X: r.Box.X + 1, Y: r.Box.Y - 50}, 0, 0},
		{"below the text", at(1, 9), 0, 15},
		{"left of the text", Point{X: r.Box.X - 40, Y: r.Box.Y + 1}, 0, 0},
		{"past the end of a line", at(30, 1), 0, 13},
	} {
		r.From = c.from
		if got := r.At(c.p, Reckoned); got != c.want {
			t.Errorf("%s: character %d, want %d", c.what, got, c.want)
		}
	}
}

func TestTheRoomAPieceOfTextTakes(t *testing.T) {
	r := Readable{Box: Box{X: 100, Y: 200, W: 300, H: lineHigh}, Size: reads,
		Lines: Wrapped("1\n2\n3\n4\n5", 40), From: 2}
	if got := r.inSight(); got != r.Box {
		t.Errorf("five lines in a box one line high take %v, want the box %v", got, r.Box)
	}
	r.Box.H, r.From = 10*lineHigh, 0
	if got := r.inSight(); got != r.Box {
		t.Errorf("a box taller than its lines takes %v, want the box %v", got, r.Box)
	}
}

func TestAPartIsInSightWhenAllOfItIs(t *testing.T) {
	frame := Box{X: 0, Y: 100, W: 300, H: 100}
	for _, c := range []struct {
		y    int
		want bool
	}{
		{100, true}, {99, false}, {200 - lineHigh, true}, {201 - lineHigh, false},
	} {
		if got := Within(Box{Y: c.y, H: lineHigh}, frame); got != c.want {
			t.Errorf("a line at %d: in sight %v, want %v", c.y, got, c.want)
		}
	}
}

func TestWhereAHeldRunIsPainted(t *testing.T) {
	box := Box{X: 100, Y: 200, W: 300, H: lineHigh}
	if got, want := HeldRun(box, 40, 25), (Box{X: 140, Y: 200, W: 25, H: lineHigh}); got != want {
		t.Fatalf("the run is painted at %+v, want %+v", got, want)
	}
}

func TestTheFaceMeasuresWhatItDraws(t *testing.T) {
	th := Theme()
	gtx := layout.Context{Ops: new(op.Ops), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(1000, lineHigh)}}
	said := []rune("Write the spec")
	edges := Measured(th)(string(said), reads)
	if len(edges) != len(said)+1 || edges[0] != 0 {
		t.Fatalf("%d characters measured %d edges from %v", len(said), len(edges), edges)
	}
	for _, n := range []int{1, 5, len(said)} {
		want := wide(gtx, th, string(said[:n]), reads).Size.X
		if got := edges[n]; got < want-1 || got > want+1 {
			t.Errorf("the edge after %d characters is at %d, drawn at %d", n, got, want)
		}
	}
	thin, broad := Measured(th)("iiii", reads), Measured(th)("WWWW", reads)
	if thin[4] >= broad[4] {
		t.Errorf("four i measure %d and four W %d", thin[4], broad[4])
	}
}

func TestAPointLandsOnTheNearestEdge(t *testing.T) {
	m := Measure(func(string, int) []int { return []int{0, 3, 13, 15} })
	for x, want := range map[int]int{-5: 0, 1: 0, 2: 1, 7: 1, 9: 2, 14: 3, 100: 3} {
		if got := m.Nearest("iWi", reads, x); got != want {
			t.Errorf("x %d lands on %d, want %d", x, got, want)
		}
	}
}

func TestTextIsReachedWithTheFaceItIsDrawnIn(t *testing.T) {
	m := Measure(func(string, int) []int { return []int{0, 3, 13, 15} })
	r := Readable{Box: Box{X: 10, W: 100, H: lineHigh}, Said: "iWi", Size: reads,
		Lines: Lines{Said: []string{"iWi"}, From: []int{0}}}
	if got := r.At(Point{X: 19, Y: 2}, m); got != 2 {
		t.Errorf("an open task reaches %d, want 2", got)
	}

	f := newFake()
	w := holding(f, card(1, "iWi", "todo")).Measuring(m)
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	box := w.formBox()
	fields := w.Form().Fields()
	value := formValue(box, fields, 0)
	inside := Inside(Framed(value, 1))
	w = w.PressAt(Point{X: inside.X + 9, Y: inside.Y + 2})
	if got := w.Form().Writing().At; got != 2 {
		t.Errorf("a field reaches %d, want 2", got)
	}
}

func TestARunGrowsFromItsAnchorByWholeUnits(t *testing.T) {
	said := "Write the spec"
	word := Picked{From: 6, To: 9}
	for _, c := range []struct {
		what          string
		by            Unit
		anchor        Picked
		at            int
		fixed, moving int
	}{
		{"a character on", Characters, Picked{From: 2, To: 2}, 7, 2, 7},
		{"a character back", Characters, Picked{From: 7, To: 7}, 2, 7, 2},
		{"inside its own word", Words, word, 7, 6, 9},
		{"into the word after", Words, word, 11, 6, 14},
		{"into the word before", Words, word, 2, 9, 0},
		{"a paragraph", Paragraphs, Picked{From: 0, To: 14}, 3, 0, 14},
	} {
		fixed, moving := Stretched(said, c.by, c.anchor, c.at)
		if fixed != c.fixed || moving != c.moving {
			t.Errorf("%s: %d to %d, want %d to %d", c.what, fixed, moving, c.fixed, c.moving)
		}
	}
}

func TestAPressLandsOnTheLineAFieldShows(t *testing.T) {
	f := newFake()
	title := strings.Repeat("word ", 40) + "end"
	w := holding(f, card(1, title, "todo"))
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	w.form = w.form.onto(0)
	box := w.formBox()
	fields := w.Form().Fields()
	value := formValue(box, fields, 0)
	inside := Inside(Framed(value, 1))
	lines := w.wrapping(title, inside, FormSize(fields[0]))
	if len(lines.Said) < 2 {
		t.Fatalf("the title fits on %d line", len(lines.Said))
	}
	w = w.PressAt(Point{X: inside.X, Y: inside.Y + 2})
	if got, want := w.Form().Writing().At, lines.From[len(lines.From)-1]; got != want {
		t.Errorf("a press at the start of the line in sight lands on %d, want %d", got, want)
	}
}

func TestAnOpenTaskToldNowhereSaysNoPlace(t *testing.T) {
	w := holding(newFake(), card(1, "Write the spec", "todo"))
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	if w.Screen() != OnForm {
		t.Fatalf("the task did not open: %v", w.Screen())
	}
	if said := w.ContextSays(); said != nil {
		t.Errorf("an open task told nowhere writes %+v", said)
	}
	for _, r := range w.Readables() {
		if strings.HasSuffix(r.Key, ":context") {
			t.Errorf("an open task told nowhere holds %q", r.Key)
		}
	}
}
