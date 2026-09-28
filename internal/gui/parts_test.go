package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/reading"
)

// Where every piece of every panel goes. This exists because of what mutation
// found: arithmetic inside the drawing is arithmetic no assertion can judge —
// a title moved six pixels looks like a title, and the only test that would
// notice is a pinned pixel frame, which goes red on another machine for a
// reason that is not a defect.
//
// So the sums live here, in numbers a test can pin, and drawing is left with
// nothing to get wrong but a colour.

func TestTheParcelsOfACardSitInsideIt(t *testing.T) {
	box := Box{X: 100, Y: 200, W: 285, H: CardHeight}
	parts := PartsOf(box)

	if parts.Panel.H != CardHeight-gap {
		t.Errorf("the panel is %d tall in a card of %d", parts.Panel.H, CardHeight)
	}
	for what, part := range map[string]Box{
		"the edge":   parts.Edge,
		"the number": parts.Number,
		"the title":  parts.Title,
	} {
		if part.X < parts.Panel.X || part.X+part.W > parts.Panel.X+parts.Panel.W {
			t.Errorf("%s runs from %d to %d, outside the panel", what, part.X, part.X+part.W)
		}
		if part.Y < parts.Panel.Y || part.Y+part.H > parts.Panel.Y+parts.Panel.H {
			t.Errorf("%s runs from %d to %d, outside the panel", what, part.Y, part.Y+part.H)
		}
	}
	if parts.Title.Y <= parts.Number.Y {
		t.Error("the title is drawn above the number")
	}
	if parts.Edge.W >= parts.Number.X-parts.Panel.X {
		t.Error("the edge runs under the number rather than beside it")
	}
}

func TestThePanelOverTheBoardLeavesTheBoardShowing(t *testing.T) {
	over := Over(1200, 800)
	if over.X <= 0 || over.Y <= 0 {
		t.Errorf("the panel starts at %d,%d", over.X, over.Y)
	}
	if over.X+over.W >= 1200 || over.Y+over.H >= 800 {
		t.Errorf("the panel ends at %d,%d in a window of 1200 by 800", over.X+over.W, over.Y+over.H)
	}
}

func TestARowSitsInsideItsPanel(t *testing.T) {
	panel := Box{X: 100, Y: 100, W: 400, H: 300}
	row := Row(panel, 150, 20)
	if row.X <= panel.X || row.X+row.W >= panel.X+panel.W {
		t.Errorf("a row runs from %d to %d in a panel from %d to %d",
			row.X, row.X+row.W, panel.X, panel.X+panel.W)
	}
	if row.Y != 150 || row.H != 20 {
		t.Errorf("a row asked for at 150 by 20 is at %d by %d", row.Y, row.H)
	}
}

// A field's name and its value share a row, and the value starts after the
// widest name a form asks about.
func TestAFieldsNameAndItsValueShareARowWithoutOverlapping(t *testing.T) {
	panel := Box{X: 100, Y: 100, W: 600, H: 400}
	name, value := FieldRow(panel, 120)
	if name.Y != value.Y || name.H != value.H {
		t.Error("a field's name and its value are not on the same row")
	}
	if name.X+name.W > value.X {
		t.Errorf("the name runs to %d and the value starts at %d", name.X+name.W, value.X)
	}
	if value.X+value.W >= panel.X+panel.W {
		t.Errorf("the value runs to %d, past the panel at %d", value.X+value.W, panel.X+panel.W)
	}
}

func TestTheTabsSitOneAfterAnother(t *testing.T) {
	l := LayoutFor(1200, 800)
	first, second := TabAt(l, 0), TabAt(l, 1)
	if first.X >= second.X {
		t.Error("the second tab is drawn before the first")
	}
	if first.X+first.W > second.X {
		t.Error("the two tabs overlap")
	}
	if second.X+second.W > l.Tabs.X+l.Tabs.W {
		t.Error("the tabs run past the strip they sit in")
	}
}

// The message a window too small shows is in the middle of it, and inside it.
func TestTheMessageOfAWindowTooSmallSitsInIt(t *testing.T) {
	box := Message(400, 300)
	if box.X < 0 || box.X+box.W > 400 {
		t.Errorf("the message runs from %d to %d in a window 400 wide", box.X, box.X+box.W)
	}
	if box.Y < 0 || box.Y+box.H > 300 {
		t.Errorf("the message runs from %d to %d in a window 300 tall", box.Y, box.Y+box.H)
	}
}

// Every parcel of every panel, written down at one size. A sum changed by
// accident shows up here as a number that moved, which is the only way a
// drawing's arithmetic can be noticed at all.
func TestTheParcelsAtOneSize(t *testing.T) {
	var out strings.Builder
	say := func(name string, b Box) {
		fmt.Fprintf(&out, "%-14s x=%-5d y=%-5d w=%-5d h=%d\n", name, b.X, b.Y, b.W, b.H)
	}
	l := LayoutFor(1200, 800)
	say("tab 0", TabAt(l, 0))
	say("tab 1", TabAt(l, 1))
	say("tab name 0", TabName(TabAt(l, 0)))
	say("goals run", GoalsRun(l))
	say("tab side top", TabSides(TabAt(l, 0))[0])
	say("tab side left", TabSides(TabAt(l, 0))[1])
	say("tab side right", TabSides(TabAt(l, 0))[2])
	say("edge right", Edges(Over(1200, 800))[3])
	say("goal 0", GoalAt(l, 0, 0))
	say("goal 1", GoalAt(l, 1, 0))
	say("goal wrapped", GoalAt(l, GoalsAcross(), 0))
	say("tabs line", TabsLine(l))
	say("tabs line left", TabsLineAround(l, 1)[0])
	say("tabs line right", TabsLineAround(l, 1)[1])
	say("search", Searching(l))
	say("search word", SearchWord(Searching(l)))
	say("search into", SearchInto(Searching(l)))
	say("lane name 0", LaneName(l, 0))

	parts := PartsOf(l.CardAt(1, 0, 0))
	say("card panel", parts.Panel)
	say("card edge", parts.Edge)
	say("card number", parts.Number)
	say("card title", parts.Title)

	over := Over(1200, 800)
	say("over", over)
	say("over row", Row(over, over.Y+padding, 24))
	name, value := FieldRow(over, over.Y+padding)
	say("field name", name)
	say("field value", value)
	say("heading", Heading(over))
	say("ruled", Ruled(over))
	say("chip", Chip(name))
	say("framed one", Framed(value, 1))
	say("framed four", Framed(value, 4))
	say("inside four", Inside(Framed(value, 4)))
	say("field top 1", Box{X: over.X, Y: FieldTop(over, reading.TaskFields(reading.Card{}), 1), W: 1, H: 1})
	goalFields := reading.GoalFields(reading.Goal{})
	say("sub panel", SubPanel(Box{X: 100, Y: 200, W: 400, H: SubHeight}))
	say("sub writing", SubWriting(Box{X: 100, Y: 200, W: 400, H: SubHeight}))
	say("choosing over", ChoosingOver(1200, 800))
	say("serving head", ServingHead(over, goalFields))
	say("serving box", ServingBox(over, goalFields))
	say("serving bar", ServingBar(over, goalFields))
	say("counted at", CountedAt(parts.Number, "7/12 · 58%"))
	say("history head", HistoryHead(over))
	say("history box", HistoryBox(over))
	say("history bar", HistoryBar(over))
	say("thumb at the top", Thumb(HistoryBar(over), 100, 10, 0))
	say("thumb halfway", Thumb(HistoryBar(over), 100, 10, 45))
	say("thumb at the end", Thumb(HistoryBar(over), 100, 10, 90))
	say("line 0", LineAt(Inside(HistoryBox(over)), 0, 0))
	say("line 3 of 2", LineAt(Inside(HistoryBox(over)), 3, 2))
	say("caret", CaretAt(Inside(HistoryBox(over)), 1, 0, 40))
	say("panel offer 0", PanelOfferAt(over, 0))
	say("panel offer 1", PanelOfferAt(over, 1))
	say("message", Message(400, 300))

	got := out.String()
	path := filepath.Join("testdata", "parcels.txt")
	if os.Getenv("UPDATE_FRAMES") != "" {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v — run the suite with UPDATE_FRAMES=1 to write it", err)
	}
	if strings.TrimSpace(string(want)) != strings.TrimSpace(got) {
		t.Errorf("a parcel moved:\n%s\nwant:\n%s", got, want)
	}
}

// The four edges of an outline are the four sides of a box and nothing else.
func TestTheFourEdgesOfAnOutline(t *testing.T) {
	box := Box{X: 10, Y: 20, W: 100, H: 50}
	edges := Edges(box)
	if len(edges) != 4 {
		t.Fatalf("an outline has %d edges", len(edges))
	}
	for i, edge := range edges {
		if edge.X < box.X || edge.X+edge.W > box.X+box.W {
			t.Errorf("edge %d runs from %d to %d, outside the box", i, edge.X, edge.X+edge.W)
		}
		if edge.Y < box.Y || edge.Y+edge.H > box.Y+box.H {
			t.Errorf("edge %d runs from %d to %d, outside the box", i, edge.Y, edge.Y+edge.H)
		}
	}
	// The bottom edge is on the last row of the box, not past it.
	if got := edges[1].Y; got != box.Y+box.H-1 {
		t.Errorf("the bottom edge is at %d, want %d", got, box.Y+box.H-1)
	}
	if got := edges[3].X; got != box.X+box.W-1 {
		t.Errorf("the right edge is at %d, want %d", got, box.X+box.W-1)
	}
}

// The first row of a panel is inside it, below its top.
func TestTheFirstRowOfAPanel(t *testing.T) {
	box := Box{X: 10, Y: 20, W: 100, H: 50}
	if got := FirstRow(box); got <= box.Y || got >= box.Y+box.H {
		t.Errorf("the first row is at %d in a panel from %d to %d", got, box.Y, box.Y+box.H)
	}
}

// The bar along the bottom offers what can be done, and each offer is a place
// a pointer can land.
func TestTheOffersInTheBarSitOneAfterAnother(t *testing.T) {
	l := LayoutFor(1200, 800)
	first, second := OfferAt(l, 0), OfferAt(l, 1)
	if first.X >= second.X {
		t.Error("the second offer is drawn before the first")
	}
	if first.X+first.W > second.X {
		t.Error("two offers overlap, so a click would hit both")
	}
	if first.Y != l.Bar.Y || first.H != l.Bar.H {
		t.Error("an offer does not fill the bar's height")
	}
}

// The objectives are a grid: across first, then down, each the width of a
// card. One to a row and the width of the whole board, an objective was a
// card's height stretched over four columns of nothing, and a dozen of them
// ran off the bottom of the window.
func TestTheObjectivesAreLaidOutAcrossAndThenDown(t *testing.T) {
	l := LayoutFor(1200, 800)
	across := GoalsAcross()
	if across < 2 {
		t.Fatalf("%d objectives stand side by side, which is a list", across)
	}
	first, second := GoalAt(l, 0, 0), GoalAt(l, 1, 0)
	if second.Y != first.Y {
		t.Error("the second objective is a row below the first: they are one to a row")
	}
	if second.X <= first.X {
		t.Errorf("the second objective is at %d and the first at %d", second.X, first.X)
	}
	if first.W != l.Column[0].W {
		t.Errorf("an objective is %d wide and a column %d", first.W, l.Column[0].W)
	}
	// The one after the last of a row goes back to the left, a row down.
	wrapped := GoalAt(l, across, 0)
	if wrapped.X != first.X {
		t.Errorf("the first of the second row is at %d and the first of the first at %d", wrapped.X, first.X)
	}
	if wrapped.Y != first.Y+CardHeight {
		t.Errorf("the second row is at %d, want %d", wrapped.Y, first.Y+CardHeight)
	}
	// And every one of them stands inside the board.
	run := GoalsRun(l)
	for at := range 2 * across {
		box := GoalAt(l, at, 0)
		if box.X < run.X || box.X+box.W > run.X+run.W {
			t.Errorf("objective %d runs from %d to %d, and the board from %d to %d",
				at, box.X, box.X+box.W, run.X, run.X+run.W)
		}
	}
}

// How many rows a number of objectives comes to, including the part-full one:
// a row nobody counted is a row nobody can scroll to.
func TestHowManyRowsTheObjectivesComeTo(t *testing.T) {
	across := GoalsAcross()
	if got := GoalRows(0); got != 0 {
		t.Errorf("no objectives came to %d rows", got)
	}
	if got := GoalRows(1); got != 1 {
		t.Errorf("one objective came to %d rows", got)
	}
	if got := GoalRows(across); got != 1 {
		t.Errorf("exactly one row of objectives came to %d rows", got)
	}
	if got := GoalRows(across + 1); got != 2 {
		t.Errorf("one past a full row came to %d rows", got)
	}
}

// An objective moves with the grid, as a card moves with its column.
func TestAnObjectiveMovesWithTheGrid(t *testing.T) {
	l := LayoutFor(1200, 800)
	at := GoalsAcross() + 2
	still, scrolled := GoalAt(l, at, 0), GoalAt(l, at, CardHeight)
	if scrolled.Y != still.Y-CardHeight {
		t.Errorf("scrolled by one row, it is at %d, want %d", scrolled.Y, still.Y-CardHeight)
	}
	if scrolled.X != still.X {
		t.Error("scrolling moved an objective sideways")
	}
	// The board the grid is laid out in spans every column, first to last.
	run := GoalsRun(l)
	if got, want := run.W, l.Column[Columns-1].X+l.Column[Columns-1].W-l.Column[0].X; got != want {
		t.Errorf("the objectives' board is %d wide, want %d", got, want)
	}
	if run.H != l.Column[0].H {
		t.Errorf("the objectives' board is %d tall and the first column %d", run.H, l.Column[0].H)
	}
}

// A search is a box along the bottom — and not a message across the middle of
// the board, which is where a window says it is too small and nowhere a person
// types. It takes the whole bar, and says what it is before what is in it.
func TestTheSearchIsABoxAlongTheBottom(t *testing.T) {
	l := LayoutFor(1200, 800)
	box := Searching(l)

	if box != l.Bar {
		t.Errorf("the search box is %+v and the bar %+v", box, l.Bar)
	}
	word, into := SearchWord(box), SearchInto(box)
	if word.X < box.X || word.X+word.W > into.X {
		t.Errorf("the word runs from %d to %d and what is typed starts at %d", word.X, word.X+word.W, into.X)
	}
	if got, want := into.X+into.W, box.X+box.W-padding; got != want {
		t.Errorf("what is typed ends at %d and the box at %d", got, want)
	}
	if into.W <= box.W/2 {
		t.Errorf("what is typed has %d of a box %d wide: the word has taken it", into.W, box.W)
	}
	// Both sit on one line, in the middle of a bar taller than a line.
	if word.Y != into.Y || word.H != lineHigh || into.H != lineHigh {
		t.Errorf("the word is %+v and what is typed %+v: they are not on one line", word, into)
	}
	if word.Y <= box.Y || word.Y+word.H >= box.Y+box.H {
		t.Errorf("the line sits at %d in a box from %d to %d", word.Y, box.Y, box.Y+box.H)
	}
}

// A panel draws what can be done along its own bottom, and each of them is a
// place a pointer can land. Drawing a word that looks like a button and is not
// one is worse than drawing nothing.
func TestAPanelsOwnOffersSitAlongItsBottom(t *testing.T) {
	box := Over(1200, 800)
	first, second := PanelOfferAt(box, 0), PanelOfferAt(box, 1)

	if first.Y != second.Y {
		t.Error("a panel's offers are not on one row")
	}
	if first.Y < box.Y+box.H/2 {
		t.Error("a panel's offers are not along its bottom")
	}
	if first.Y+first.H > box.Y+box.H {
		t.Error("a panel's offers run past its bottom")
	}
	if first.X+first.W > second.X {
		t.Error("two offers overlap, so a click would hit both")
	}
	if first.X < box.X || second.X+second.W > box.X+box.W {
		t.Error("a panel's offers run outside it")
	}
}

// A field that wants more than one line gets more than one, and the fields
// under it move down by what it took. Without that a description is a slot the
// size of a word.
func TestAFieldThatWantsMoreThanOneLineGetsIt(t *testing.T) {
	fields := []reading.Field{
		{Name: "title"},
		{Name: "description", Long: true},
		{Name: "column"},
	}
	if got := RowsOf(fields[0]); got != 1 {
		t.Errorf("a plain field takes %d rows", got)
	}
	if got := RowsOf(fields[1]); got <= 1 {
		t.Errorf("a long field takes %d rows", got)
	}

	box := Over(1200, 800)
	first := FieldTop(box, fields, 0)
	second := FieldTop(box, fields, 1)
	third := FieldTop(box, fields, 2)
	// A field takes its rows, the room its frame keeps around them, and what it
	// leaves before the next one.
	if second-first != fieldHigh(fields[0]) {
		t.Errorf("the second field is %d below the first, want %d", second-first, fieldHigh(fields[0]))
	}
	if third-second != fieldHigh(fields[1]) {
		t.Errorf("the third field is %d below the second, which takes %d rows",
			third-second, RowsOf(fields[1]))
	}
	if third-second <= second-first {
		t.Error("a field of four lines is no further from the next than a field of one")
	}
}

// A value longer than the room a one-line field has shows its end, because the
// end is where typing is happening — and it shows it whole, wrapped onto a
// line of its own rather than cut down to fit.
func TestAValueTooLongForOneLineShowsItsEnd(t *testing.T) {
	room := 20
	long := "a value far longer than this"
	lines := Wrapped(long, room)
	line, _ := lines.At(len([]rune(long)))
	from := lines.Window(1, line)

	if from == 0 {
		t.Fatal("a value too long for one line still started at its beginning")
	}
	if shown := lines.Said[from]; !strings.HasSuffix(long, shown) {
		t.Errorf("it shows %q, which is not the end of %q", shown, long)
	}
	// Nothing is lost: every line put back together is what was typed.
	if back := strings.Join(lines.Said, " "); back != long {
		t.Errorf("the lines came back as %q", back)
	}
}

// Where a one-line field starts showing its end instead of its beginning.
func TestWhereAValueStartsShowingItsEndInstead(t *testing.T) {
	room := 20
	exactly := strings.Repeat("a", room)
	if got := Wrapped(exactly, room); len(got.Said) != 1 {
		t.Errorf("a value of exactly the room came to %d lines", len(got.Said))
	}
	one := strings.Repeat("a", room+1)
	got := Wrapped(one, room)
	if len(got.Said) != 2 {
		t.Errorf("a value one past the room came to %d lines", len(got.Said))
	}
	if line, _ := got.At(room + 1); got.Window(1, line) != 1 {
		t.Error("with the caret past the first line, the first line is still what is shown")
	}
	// And asking past the end of a list of fields is the whole list.
	fields := []reading.Field{{Name: "one"}, {Name: "two"}}
	box := Over(1200, 800)
	if past, last := FieldTop(box, fields, 9), FieldTop(box, fields, len(fields)); past != last {
		t.Errorf("past the last field is %d and the end is %d", past, last)
	}
}

// How many characters a box can show, which every cut in this window is
// reckoned from.
func TestHowManyCharactersABoxCanShow(t *testing.T) {
	if got := room(Box{W: padding + padding + 10*perGlyph}); got != 10 {
		t.Errorf("a box with room for ten answered %d", got)
	}
	if got := room(Box{W: padding + padding}); got != 0 {
		t.Errorf("a box with room for none answered %d", got)
	}
	wider := room(Box{W: padding + padding + 20*perGlyph})
	if narrower := room(Box{W: padding + padding + 10*perGlyph}); wider <= narrower {
		t.Errorf("a wider box shows %d and a narrower one %d", wider, narrower)
	}
}

// A text too long for its box goes onto the next line rather than out of it,
// and breaks at a space when there is one to break at.
func TestATextTooLongForItsBoxGoesOntoTheNextLine(t *testing.T) {
	lines := Wrapped("the quick brown fox", 9)
	for i, line := range lines.Said {
		if len([]rune(line)) > 9 {
			t.Errorf("line %d is %q, %d characters in room for 9", i, line, len([]rune(line)))
		}
	}
	if len(lines.Said) < 2 {
		t.Fatalf("nineteen characters in room for nine came to %d line(s)", len(lines.Said))
	}
	if strings.Join(strings.Fields(strings.Join(lines.Said, " ")), " ") != "the quick brown fox" {
		t.Errorf("the words came back as %q", lines.Said)
	}
	// Broken at a space, so no line starts or ends inside a word that fits.
	if strings.HasPrefix(lines.Said[1], "k ") {
		t.Errorf("%q was broken inside a word that would have fitted", lines.Said)
	}
	// A word longer than the room has to break inside itself: the alternative
	// is a line that runs out of the box, which is the thing being prevented.
	long := Wrapped("antidisestablishmentarianism", 9)
	if len(long.Said) < 3 {
		t.Errorf("a word of 28 characters in room for 9 came to %d lines", len(long.Said))
	}
	for _, line := range long.Said {
		if len([]rune(line)) > 9 {
			t.Errorf("%q is longer than the room there is", line)
		}
	}
	// No room at all is one line: a box nothing fits in is not a reason to
	// loop for ever.
	if none := Wrapped("anything", 0); len(none.Said) != 1 || none.Said[0] != "anything" {
		t.Errorf("with no room it came to %q", none.Said)
	}
	if empty := Wrapped("", 9); len(empty.Said) != 1 || empty.Said[0] != "" {
		t.Errorf("nothing came to %q", empty.Said)
	}
}

// The caret is counted in the whole text, and has to be found on the line it
// fell on — otherwise it is drawn on the first line whatever was typed.
func TestWhichLineTheCaretFellOn(t *testing.T) {
	lines := Wrapped("the quick brown fox", 9)
	if line, column := lines.At(0); line != 0 || column != 0 {
		t.Errorf("the beginning is line %d column %d", line, column)
	}
	line, column := lines.At(len([]rune("the quick brown fox")))
	if line != len(lines.Said)-1 {
		t.Errorf("the end is on line %d of %d", line, len(lines.Said))
	}
	if column != len([]rune(lines.Said[line])) {
		t.Errorf("the end is %d into a line of %d", column, len([]rune(lines.Said[line])))
	}
	// Past the end, and before the beginning, are the ends.
	if l, c := lines.At(999); l != len(lines.Said)-1 || c != len([]rune(lines.Said[l])) {
		t.Errorf("past the end is line %d column %d", l, c)
	}
	if l, c := lines.At(-4); l != 0 || c != 0 {
		t.Errorf("before the beginning is line %d column %d", l, c)
	}
}

// More lines than there is room for shows the ones around the caret: what is
// being typed is what has to stay in sight.
func TestWhichLinesAreShownWhenThereAreMoreThanFit(t *testing.T) {
	lines := Wrapped("one two three four five six seven eight", 9)
	if len(lines.Said) < 5 {
		t.Fatalf("that came to %d lines, which is not enough to go out of sight", len(lines.Said))
	}
	if from := lines.Window(4, 0); from != 0 {
		t.Errorf("with the caret on the first line, the window starts at %d", from)
	}
	last := len(lines.Said) - 1
	if from := lines.Window(4, last); from != last-3 {
		t.Errorf("with the caret on the last line, the window starts at %d and the last line is %d", from, last)
	}
	// Room for more lines than there are shows them all, from the first.
	if from := lines.Window(99, last); from != 0 {
		t.Errorf("with room for every line, the window starts at %d", from)
	}
}

// A field is a box, the way the search is a box: the same shape wherever
// somebody types, and room inside it that the frame does not take.
func TestAFieldIsABoxWithRoomInsideIt(t *testing.T) {
	over := Over(1200, 800)
	_, value := FieldRow(over, FirstRow(over))

	one, four := Framed(value, 1), Framed(value, 4)
	if four.H <= one.H {
		t.Errorf("a field of four lines is %d tall and one of a single line %d", four.H, one.H)
	}
	if one.W != value.W || one.X != value.X {
		t.Errorf("the frame is %+v and what it holds %+v", one, value)
	}
	inside := Inside(four)
	if inside.X <= four.X || inside.X+inside.W >= four.X+four.W {
		t.Errorf("the room inside runs from %d to %d in a frame from %d to %d", inside.X, inside.X+inside.W, four.X, four.X+four.W)
	}
	if inside.H < 4*lineHigh {
		t.Errorf("a frame of four lines has room for %d, and a line is %d", inside.H, lineHigh)
	}
}

// Every field of a form fits in the panel, above what the panel offers. Fields
// grew a frame around them, and a frame is height: the last field of the
// longest form is the one that goes off the bottom first, and going off the
// bottom is a field somebody cannot fill in.
func TestEveryFieldFitsInThePanelAboveWhatItOffers(t *testing.T) {
	box := Over(Narrowest, Shortest)
	for _, fields := range [][]reading.Field{
		reading.TaskFields(reading.Card{}),
		reading.GoalFields(reading.Goal{}),
	} {
		last := len(fields) - 1
		_, value := FieldRow(box, FieldTop(box, fields, last))
		bottom := Framed(value, RowsOf(fields[last]))
		if got, want := bottom.Y+bottom.H, PanelOfferAt(box, 0).Y; got >= want {
			t.Errorf("the last of %d fields ends at %d and the offers start at %d",
				len(fields), got, want)
		}
	}
}

// The history is beside what a task carries, not under it — and its box and
// its bar both stop short of what the panel offers along its bottom.
func TestTheHistoryHasItsOwnColumnOfThePanel(t *testing.T) {
	for _, size := range [][2]int{{Narrowest, Shortest}, {1280, 840}, {1920, 1200}} {
		box := Over(size[0], size[1])
		fields, history := DetailFields(box), DetailHistory(box)
		if fields.X+fields.W != history.X {
			t.Errorf("at %v the fields end at %d and the history starts at %d", size, fields.X+fields.W, history.X)
		}
		if fields.W+history.W != box.W {
			t.Errorf("at %v the two columns come to %d of a panel %d wide", size, fields.W+history.W, box.W)
		}
		if fields.W <= history.W {
			t.Errorf("at %v the fields have %d and the history %d: the fields are the longer column", size, fields.W, history.W)
		}
		frame, track := HistoryBox(box), HistoryBar(box)
		if frame.X+frame.W > track.X {
			t.Errorf("at %v the history runs to %d and its bar starts at %d", size, frame.X+frame.W, track.X)
		}
		if track.X+track.W > box.X+box.W {
			t.Errorf("at %v the bar ends at %d and the panel at %d", size, track.X+track.W, box.X+box.W)
		}
		if bottom, offers := frame.Y+frame.H, PanelOfferAt(box, 0).Y; bottom > offers {
			t.Errorf("at %v the history ends at %d and what the panel offers starts at %d", size, bottom, offers)
		}
		if frame.Y <= HistoryHead(box).Y {
			t.Errorf("at %v the history starts at %d and its heading at %d", size, frame.Y, HistoryHead(box).Y)
		}
		if HistoryRows(box) < 1 {
			t.Errorf("at %v the history has room for %d lines", size, HistoryRows(box))
		}
	}
}

// The grip says two things at once: how tall it is, how much of the history is
// showing; where it sits, which part of it.
func TestTheGripSaysHowMuchIsShowingAndWhichPart(t *testing.T) {
	track := Box{X: 100, Y: 200, W: barWide, H: 300}

	// Nothing to scroll: the grip fills the track, and the bar is still there.
	if got := Thumb(track, 5, 10, 0); got != track {
		t.Errorf("with everything showing the grip is %+v and the track %+v", got, track)
	}

	little, lots := Thumb(track, 100, 10, 0), Thumb(track, 20, 10, 0)
	if little.H >= lots.H {
		t.Errorf("ten lines of a hundred gave a grip of %d and ten of twenty %d", little.H, lots.H)
	}
	if little.H < leastThumb {
		t.Errorf("a grip of %d is too small to take hold of", little.H)
	}

	// It starts at the top and reaches the bottom, and never passes it.
	top, bottom := Thumb(track, 100, 10, 0), Thumb(track, 100, 10, 90)
	if top.Y != track.Y {
		t.Errorf("at the beginning the grip is at %d and the track at %d", top.Y, track.Y)
	}
	if got, want := bottom.Y+bottom.H, track.Y+track.H; got != want {
		t.Errorf("at the end the grip ends at %d and the track at %d", got, want)
	}
	if past := Thumb(track, 100, 10, 999); past.Y+past.H > track.Y+track.H {
		t.Errorf("past the end the grip ends at %d, below the track", past.Y+past.H)
	}
	if middle := Thumb(track, 100, 10, 45); middle.Y <= top.Y || middle.Y >= bottom.Y {
		t.Errorf("halfway the grip is at %d, between %d and %d", middle.Y, top.Y, bottom.Y)
	}
}

// Where a point on the track falls in the history, which is what makes the bar
// something to pull rather than something to look at.
func TestWhereAPointOnTheBarFallsInTheHistory(t *testing.T) {
	track := Box{X: 100, Y: 200, W: barWide, H: 300}
	if got := Reached(track, 100, 10, Point{Y: track.Y}); got > 0 {
		t.Errorf("the top of the track is line %d", got)
	}
	if got := Reached(track, 100, 10, Point{Y: track.Y + track.H}); got < 90 {
		t.Errorf("the bottom of the track is line %d, and the end is 90", got)
	}
	middle := Reached(track, 100, 10, Point{Y: track.Y + track.H/2})
	if middle <= 0 || middle >= 90 {
		t.Errorf("halfway down the track is line %d", middle)
	}
	// A history that fits has nowhere to be pulled to.
	if got := Reached(track, 5, 10, Point{Y: track.Y + track.H}); got != 0 {
		t.Errorf("a history that fits was pulled to line %d", got)
	}
	if got := Reached(Box{H: 2}, 100, 10, Point{Y: 1}); got != 0 {
		t.Errorf("a track with no room answered line %d", got)
	}
}

// Acts are told one after another with a blank line between two of them: an
// act that wrapped onto three lines, with the next beginning straight after
// it, read as one act nobody could tell the end of.
func TestTheHistoryIsToldOneActAfterAnother(t *testing.T) {
	acts := []string{"30 Aug  moved column → current", "29 Aug  added"}
	lines := Told(acts, 12)
	if len(lines) < 4 {
		t.Fatalf("two acts too long for the room came to %d lines: %q", len(lines), lines)
	}
	blank := 0
	for _, line := range lines {
		if line == "" {
			blank++
		}
	}
	if blank != 1 {
		t.Errorf("two acts are told with %d blank lines between them: %q", blank, lines)
	}
	if lines[0] == "" || lines[len(lines)-1] == "" {
		t.Errorf("the history begins or ends on nothing: %q", lines)
	}
	if got := Told(nil, 12); len(got) != 0 {
		t.Errorf("no acts at all came to %q", got)
	}
	if got := Told([]string{"one"}, 12); len(got) != 1 {
		t.Errorf("one act came to %q", got)
	}
}

// Moving a run so that what was chosen is inside it. Written once and used by
// the columns and by the objectives alike: it was the same arithmetic twice,
// and the second copy is where a boundary goes wrong unnoticed.
func TestMovingARunSoThatWhatIsChosenIsInsideIt(t *testing.T) {
	const high, run = 10, 50 // five rows fit
	// Already in sight: nothing moves, at either end of what is showing.
	for _, at := range []int{0, 4} {
		if got := inSight(at, high, run, 0); got != 0 {
			t.Errorf("row %d is already in sight and the run moved to %d", at, got)
		}
	}
	// One past the bottom: the run moves by exactly one row and no more.
	if got := inSight(5, high, run, 0); got != high {
		t.Errorf("one past the bottom moved the run to %d, want %d", got, high)
	}
	// Far past the bottom: the row ends up against the bottom of the run.
	if got := inSight(9, high, run, 0); got != 10*high-run {
		t.Errorf("row 9 moved the run to %d, want %d", got, 10*high-run)
	}
	// Above what is showing: the row ends up against the top.
	if got := inSight(2, high, run, 40); got != 2*high {
		t.Errorf("row 2 from a run at 40 moved to %d, want %d", got, 2*high)
	}
	if got := inSight(4, high, run, 40); got != 40 {
		t.Errorf("row 4 is the first one showing and the run moved to %d", got)
	}
	// A run with no room at all still answers, rather than looping or dividing
	// by nothing: the row's top goes to the top of the run.
	if got := inSight(3, high, 0, 0); got != 3*high {
		t.Errorf("a run with no room answered %d", got)
	}
	// A run shorter than one row shows that row's top. It never happens on
	// this board — a column is always taller than a card — but it is what the
	// two clamps say, and what they say is worth being able to read back.
	if got := inSight(2, high, high/2, 2*high); got != 2*high {
		t.Errorf("a row taller than its run answered %d, want %d", got, 2*high)
	}
}

// Where one line of a wrapped text is drawn, and which is the last one drawn.
// Both drawings that show wrapped lines worked this out for themselves, which
// is arithmetic inside a drawing — the one thing no assertion here can judge.
func TestWhereEachLineOfAWrappedTextGoes(t *testing.T) {
	inside := Box{X: 100, Y: 200, W: 300, H: 4 * lineHigh}

	first := LineAt(inside, 0, 0)
	if first.X != inside.X || first.Y != inside.Y {
		t.Errorf("the first line is at %+v and the box at %+v", first, inside)
	}
	if first.W != inside.W || first.H != lineHigh {
		t.Errorf("a line is %dx%d in a box %d wide", first.W, first.H, inside.W)
	}
	if second := LineAt(inside, 1, 0); second.Y != first.Y+lineHigh {
		t.Errorf("the second line is at %d, want %d", second.Y, first.Y+lineHigh)
	}
	// Scrolled, the first line shown is the one at the top of the box.
	if scrolled := LineAt(inside, 3, 3); scrolled.Y != inside.Y {
		t.Errorf("with the fourth line first, it is at %d and the box at %d", scrolled.Y, inside.Y)
	}
	if scrolled := LineAt(inside, 5, 3); scrolled.Y != inside.Y+2*lineHigh {
		t.Errorf("two below the first shown is at %d", scrolled.Y)
	}

	// The last line drawn is the last there is, or the last there is room for.
	if got := Lasting(2, 0, 4); got != 2 {
		t.Errorf("two lines in room for four came to %d", got)
	}
	if got := Lasting(9, 0, 4); got != 4 {
		t.Errorf("nine lines in room for four came to %d", got)
	}
	if got := Lasting(9, 6, 4); got != 9 {
		t.Errorf("from the seventh of nine, in room for four, came to %d", got)
	}
}

// Where the caret stands: on the line it fell on, after what comes before it.
func TestWhereTheCaretStandsAmongWrappedLines(t *testing.T) {
	inside := Box{X: 100, Y: 200, W: 300, H: 4 * lineHigh}

	first := CaretAt(inside, 0, 0, 40)
	if first.X != inside.X+40 {
		t.Errorf("the caret is at %d and the box starts at %d", first.X, inside.X)
	}
	if first.Y != inside.Y || first.H != lineHigh {
		t.Errorf("the caret is %+v on the first line of %+v", first, inside)
	}
	if first.W != caretWide {
		t.Errorf("the caret is %d wide, want %d", first.W, caretWide)
	}
	// On a later line, and on a later line of a text that has been scrolled.
	if third := CaretAt(inside, 2, 0, 0); third.Y != inside.Y+2*lineHigh {
		t.Errorf("on the third line the caret is at %d", third.Y)
	}
	if scrolled := CaretAt(inside, 5, 4, 0); scrolled.Y != inside.Y+lineHigh {
		t.Errorf("the sixth line with the fifth first puts the caret at %d", scrolled.Y)
	}
	// Nothing before it puts it hard against the left of the box.
	if none := CaretAt(inside, 0, 0, 0); none.X != inside.X {
		t.Errorf("with nothing before it the caret is at %d", none.X)
	}
}

// A line is broken at a space, but never at the first character of the line
// itself: that gives an empty line, and the text never gets any shorter.
func TestALineIsNeverBrokenAtItsOwnBeginning(t *testing.T) {
	lines := Wrapped(" abcdefgh", 4)
	if lines.Said[0] == "" {
		t.Errorf("the first line is empty: %q", lines.Said)
	}
	if lines.Said[0] != " abc" {
		t.Errorf("the first line is %q, and four characters of %q are \" abc\"", lines.Said[0], " abcdefgh")
	}
}

// The last line is left whole: it fits, so there is nothing to break. Broken
// anyway, a text ending in two short words came apart for no reason.
func TestTheLastLineIsLeftWhole(t *testing.T) {
	lines := Wrapped("abcdefghij k l", 11)
	if len(lines.Said) != 2 {
		t.Fatalf("that came to %d lines: %q", len(lines.Said), lines.Said)
	}
	if lines.Said[1] != "k l" {
		t.Errorf("the last line is %q, and it fits whole", lines.Said[1])
	}
}

// The grip's height, at the sizes the two guards around it turn on.
func TestHowTallTheGripIs(t *testing.T) {
	// A track shorter than the smallest grip, with everything showing: the
	// grip is the track, not the floor. Told apart, because a grip taller than
	// its own track is a grip drawn outside the bar.
	if got := thumbHigh(Box{H: 10}, 5, 5); got != 10 {
		t.Errorf("a track of 10 with everything showing has a grip of %d", got)
	}
	if got := thumbHigh(Box{H: 300}, 100, 10); got != 30 {
		t.Errorf("a tenth of a track of 300 is a grip of %d", got)
	}
	if got := thumbHigh(Box{H: 300}, 1000, 10); got != leastThumb {
		t.Errorf("a hundredth of a track of 300 is a grip of %d, and the floor is %d", got, leastThumb)
	}
}

// A track with exactly one point of room to move in still moves: it is the
// boundary between a bar that can be pulled and one that cannot.
func TestABarWithOnePointOfRoomStillMoves(t *testing.T) {
	track := Box{X: 5, Y: 100, W: barWide, H: leastThumb + 1}
	if got := thumbHigh(track, 100, 1); got != leastThumb {
		t.Fatalf("the grip is %d, and this test needs one point of room", got)
	}
	top, bottom := Thumb(track, 100, 1, 0), Thumb(track, 100, 1, 99)
	if top.H != leastThumb || bottom.H != leastThumb {
		t.Errorf("the grip is %d at the top and %d at the end, want %d", top.H, bottom.H, leastThumb)
	}
	if bottom.Y != top.Y+1 {
		t.Errorf("with one point of room the grip went from %d to %d", top.Y, bottom.Y)
	}
	// And a point at the bottom of that track reaches the end of the history.
	if got := Reached(track, 100, 1, Point{Y: track.Y + track.H}); got <= 0 {
		t.Errorf("the bottom of a track with one point of room is line %d", got)
	}
}

// Where a point on the track falls, exactly. The grip is taken hold of by its
// middle, so half of it comes off the reckoning — a half nobody pinned is a
// grip that jumps by half its height the moment it is touched.
func TestExactlyWhereAPointOnTheBarFalls(t *testing.T) {
	track := Box{Y: 100, W: barWide, H: 300}
	// A grip of 30 in a track of 300 leaves 270 to move in, over 90 lines.
	if got := Reached(track, 100, 10, Point{Y: track.Y + 150}); got != 45 {
		t.Errorf("halfway down the track is line %d, want 45", got)
	}
	if got := Reached(track, 100, 10, Point{Y: track.Y + 15}); got != 0 {
		t.Errorf("a grip's half-height down the track is line %d", got)
	}
}

// Room for one character at a time is room: it wraps rather than giving up and
// writing the whole text on one line.
func TestRoomForOneCharacterIsStillRoom(t *testing.T) {
	if got := Wrapped("abc", 1); len(got.Said) != 3 {
		t.Errorf("three characters in room for one came to %d lines: %q", len(got.Said), got.Said)
	}
	if got := Wrapped("abc", 0); len(got.Said) != 1 || got.Said[0] != "abc" {
		t.Errorf("with no room at all it came to %q", got.Said)
	}
}

// The caret exactly at the start of a line belongs to that line and not to the
// one before it, where it would be drawn at the far right of the wrong row.
func TestTheCaretAtTheStartOfALineBelongsToThatLine(t *testing.T) {
	lines := Wrapped("the quick brown fox", 9)
	if len(lines.Said) < 2 || lines.From[1] == 0 {
		t.Fatalf("that came to %q beginning at %v", lines.Said, lines.From)
	}
	line, column := lines.At(lines.From[1])
	if line != 1 {
		t.Errorf("the caret at the start of the second line fell on line %d", line)
	}
	if column != 0 {
		t.Errorf("it fell %d into that line, and it is at its start", column)
	}
	// And two characters on, it is two into that same line.
	if line, column := lines.At(lines.From[1] + 2); line != 1 || column != 2 {
		t.Errorf("two into the second line is line %d column %d", line, column)
	}
}

// Which line the window of shown lines starts at, exactly.
func TestExactlyWhereTheWindowOfShownLinesStarts(t *testing.T) {
	ten := Lines{Said: make([]string, 10)}
	if got := ten.Window(4, 9); got != 6 {
		t.Errorf("the last of ten lines in room for four starts the window at %d, want 6", got)
	}
	if got := ten.Window(4, 4); got != 1 {
		t.Errorf("the fifth of ten lines in room for four starts the window at %d, want 1", got)
	}
	if got := ten.Window(4, 3); got != 0 {
		t.Errorf("the fourth of ten lines in room for four starts the window at %d, want 0", got)
	}
}

// Where a tile of the work under an objective sits: down by its rank, up by
// what has been scrolled. Nothing else pinned this, and every sum in it
// survived being changed.
func TestAServingCardSitsByItsRankAndItsScroll(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{
		Cards: []reading.Card{{Number: 1, Title: "Write the spec", Objectives: []int{1}}},
		Goals: []reading.Goal{{Number: 1, Title: "Ship it", State: "active"}},
	}
	w.chosen, w.chosenGoal, w.screen = 1, true, OnDetail

	box := ServingBox(Over(w.width, w.height), w.Shown())
	for _, tc := range []struct {
		what       string
		at, scroll int
		want       int
	}{
		{"the first, unscrolled", 0, 0, box.Y},
		{"the second, unscrolled", 1, 0, box.Y + SubHeight},
		{"the third, unscrolled", 2, 0, box.Y + SubHeight + SubHeight},
		{"the first, scrolled one tile", 0, SubHeight, box.Y - SubHeight},
		{"the second, scrolled one tile", 1, SubHeight, box.Y},
	} {
		w.serving = tc.scroll
		got := w.ServingCardAt(tc.at)
		if got.Y != tc.want {
			t.Errorf("%s: y = %d, want %d", tc.what, got.Y, tc.want)
		}
		if got.X != box.X || got.W != box.W || got.H != SubHeight {
			t.Errorf("%s: %v, want the box's width at a tile's height", tc.what, got)
		}
	}
}
