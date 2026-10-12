package gui

import (
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/reading"
)

// listOn opens the list of the field named, on a task's form.
func listOn(t *testing.T, arg string, cards ...reading.Card) Window {
	t.Helper()
	w := holding(newFake(), cards...)
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != arg {
		w = w.Key("tab")
	}
	w = w.choosingList()
	if w.Screen() != OnChoosing {
		t.Fatalf("the %s field opened no list", arg)
	}
	return w
}

// A list of four priorities is a panel of four rows, as wide as what it says
// and hung under the field it answers, the way a field's list opens anywhere
// else. It used to take most of the window whatever it held.
func TestAListIsTheSizeOfWhatItHolds(t *testing.T) {
	w := listOn(t, "priority", reading.Card{Number: 1, Title: "A", State: "todo", Priority: "normal"})
	box := w.ChoosingBox()
	if got := ChoosingRows(box); got != len(reading.Priorities()) {
		t.Errorf("four priorities are drawn in a run of %d rows", got)
	}
	if over := Over(w.width, w.height); box.W > over.W/2 {
		t.Errorf("the list is %d wide, in a panel of %d", box.W, over.W)
	}
	_, value, rows := FormPlace(w.formBox(), w.Form().Fields(), w.Form().On())
	field := Framed(value, rows)
	if box.X != field.X || box.Y != field.Y+field.H+between {
		t.Errorf("the list is at %d,%d; the field ends at %d,%d", box.X, box.Y, field.X, field.Y+field.H)
	}
	// Wide enough for every row, its heading and its offers.
	for _, one := range w.Choosing() {
		said := Reckoned(w.RowSays(one), reads)
		if line := ChoosingLine(box, 0); line.W < said[len(said)-1] {
			t.Errorf("%q needs %d and its line is %d", w.RowSays(one), said[len(said)-1], line.W)
		}
	}
	offers := w.PanelOffers()
	if last := offers[len(offers)-1]; last.X+last.W > box.X+box.W-padding {
		t.Errorf("the last offer ends at %d, past the list's edge at %d", last.X+last.W, box.X+box.W)
	}
	head := Reckoned(Heads(w), heading)
	if Heading(box).W < head[len(head)-1] {
		t.Errorf("the heading %q needs %d and has %d", Heads(w), head[len(head)-1], Heading(box).W)
	}
	// Narrowed, it keeps its size: a panel jumping under the typing moves
	// every row and offer from under the pointer.
	if got := w.Types("hi").ChoosingBox(); got != box {
		t.Errorf("narrowed, the list became %+v from %+v", got, box)
	}
}

// A list longer than the window has room for stops at the panel's foot and
// scrolls; one with nothing in it keeps a row to say so.
func TestALongListStopsWithinTheWindow(t *testing.T) {
	cards := []reading.Card{{Number: 1, Title: "A", State: "todo"}}
	for i := 2; i <= 60; i++ {
		cards = append(cards, reading.Card{Number: i, Title: "Another task", State: "backlog"})
	}
	w := listOn(t, "parent", cards...)
	box, over := w.ChoosingBox(), Over(w.width, w.height)
	if box.Y < over.Y || box.Y+box.H > over.Y+over.H || box.X < over.X || box.X+box.W > over.X+over.W {
		t.Errorf("the list %+v runs out of %+v", box, over)
	}
	if rows := ChoosingRows(box); rows >= len(w.Choosing()) || rows < 3 {
		t.Errorf("sixty tasks are drawn in a run of %d rows", rows)
	}
	empty := listOn(t, "objective", reading.Card{Number: 1, Title: "A", State: "todo"})
	if got := ChoosingRows(empty.ChoosingBox()); got != 1 {
		t.Errorf("an empty list has a run of %d rows", got)
	}
}

// The list of backups is as long as the backups, in the middle of the window.
func TestTheBackupsListIsTheSizeOfTheBackups(t *testing.T) {
	w := New(newFake(), 1200, 800)
	w.screen, w.backups = OnBackups, []domain.Backup{{Number: 2}, {Number: 1}}
	box := w.BackupsBox()
	if got := w.BackupsRoom(); got != 2 {
		t.Errorf("two backups are drawn in room for %d", got)
	}
	if mid := box.X + box.W/2; mid < w.width/2-1 || mid > w.width/2+1 {
		t.Errorf("the list %+v is not across the middle of %d", box, w.width)
	}
	if box.W > Over(w.width, w.height).W/2 {
		t.Errorf("two lines of a backup take %d across", box.W)
	}
	if w.panel() != box {
		t.Errorf("the panel is %+v and the list %+v", w.panel(), box)
	}
}

// A form open on a task ends a bar's height under what it holds, as a form
// adding one does: a task of few fields sat over a floor of empty ground, its
// offers a hand's width below its last field. The history beside it is as
// tall as the fields, and stays inside the panel.
func TestAnOpenFormIsTheHeightOfItsFields(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	panel, over := w.formPanel(), Over(w.width, w.height)
	_, _, end := formLayout(FormFields(over), w.Form().Fields())
	if want := end - over.Y + between + barHigh + padding; panel.H != want || panel.Y != over.Y || panel.W != over.W {
		t.Errorf("the form is %+v, want %d high from %+v", panel, want, over)
	}
	if beside := BesideBox(panel); beside.Y+beside.H > PanelOfferAt(panel, 0).Y {
		t.Errorf("the history %+v runs into the offers at %d", beside, PanelOfferAt(panel, 0).Y)
	}
	if w.panel() != panel || w.Rows() != BesideRows(panel) {
		t.Errorf("the panel answered is %+v and the history's rows %d", w.panel(), w.Rows())
	}
}

// A list with no room under its field opens above it, ending a gap over it;
// one too tall for that too starts at the panel's top. A list wider than the
// room right of its field ends at the panel's right edge.
func TestAListOpensWhereItHasRoom(t *testing.T) {
	cards := []reading.Card{{Number: 1, Title: "A", State: "todo"}}
	for i := 2; i <= 6; i++ {
		cards = append(cards, reading.Card{Number: i, Title: "Another", State: "backlog"})
	}
	w := listOn(t, "parent", cards...)
	box, frame, over := w.ChoosingBox(), fieldFrame(w), Over(w.width, w.height)
	if frame.Y+frame.H+between+box.H <= over.Y+over.H {
		t.Fatalf("the list %+v fits under its field %+v: nothing here says where it goes otherwise", box, frame)
	}
	if box.Y+box.H != frame.Y-between || box.X != frame.X {
		t.Errorf("the list is %+v, over a field at %+v", box, frame)
	}
	for i := 7; i <= 60; i++ {
		cards = append(cards, reading.Card{Number: i, Title: "Another", State: "backlog"})
	}
	if got := listOn(t, "parent", cards...).ChoosingBox(); got.Y != over.Y {
		t.Errorf("a list too tall for either side starts at %d, not the panel's top %d", got.Y, over.Y)
	}
	long := reading.Card{Number: 2, Title: strings.Repeat("a long title ", 6), State: "backlog"}
	right := listOn(t, "objective", reading.Card{Number: 1, Title: "A", State: "todo"}, long)
	right.snap.Goals = []reading.Goal{{Number: 1, Title: strings.Repeat("an objective worth a long name ", 3), State: "active"}}
	right = right.choosingList()
	if got := right.ChoosingBox(); got.X+got.W != over.X+over.W || got.X >= fieldFrame(right).X {
		t.Errorf("a wide list is %+v beside a field at %+v, in %+v", got, fieldFrame(right), over)
	}
}

// A row's words and the state written hard right of them have the line to
// themselves and no more: the widest row sets the width, to the pixel.
func TestAListIsAsWideAsItsWidestRow(t *testing.T) {
	long := reading.Card{Number: 2, Title: strings.Repeat("a long title ", 6), State: "backlog"}
	w := listOn(t, "parent", reading.Card{Number: 1, Title: "A", State: "todo"}, long)
	var widest Choice
	for _, one := range w.Choosing() {
		if one.Number == 2 {
			widest = one
		}
	}
	want := ends(Reckoned, w.RowSays(widest), reads) + padding + len([]rune(widest.State))*perGlyph
	if got := ChoosingLine(w.ChoosingBox(), 0).W; got != want {
		t.Errorf("the line is %d wide for a row needing %d", got, want)
	}
}

// The list of backups is as wide as its longest line, a padding either side of
// the line and either side of the row it is written in.
func TestTheBackupsListIsAsWideAsItsLongestLine(t *testing.T) {
	w := New(newFake(), 1200, 800)
	w.screen, w.backups = OnBackups, []domain.Backup{{Number: 1234567890123}, {Number: 1}}
	if got, want := w.BackupsBox().W, ends(Reckoned, backupSays(w.backups[0]), reads)+4*padding; got != want {
		t.Errorf("the list is %d wide, want %d", got, want)
	}
	if got, want := w.BackupsBox().Y, Over(1200, 800).Y+(Over(1200, 800).H-w.BackupsBox().H)/2; got != want {
		t.Errorf("the list starts at %d, want %d", got, want)
	}
}

// A heading is written whole: the width a panel needs for it leaves a padding
// either side of it.
func TestAHeadingSetsTheLeastWidthOfItsPanel(t *testing.T) {
	w := Window{screen: OnBackups}
	if got := Heading(Box{W: w.headWide()}).W; got != ends(Reckoned, Heads(w), heading) {
		t.Errorf("a panel as wide as its heading needs writes it in %d", got)
	}
}

// What hangs from a field stays under it when it ends exactly at the panel's
// foot, and goes above it a pixel past that.
func TestWhatHangsFromAFieldGoesAboveOnlyWhenItMust(t *testing.T) {
	over := Box{X: 0, Y: 0, W: 1000, H: 500}
	field := Box{X: 100, Y: 400, W: 100, H: 20}
	if got := Hung(field, over, 300, 74); got != (Point{X: 100, Y: 426}) {
		t.Errorf("a panel ending at the foot hangs at %+v", got)
	}
	if got := Hung(field, over, 300, 75); got != (Point{X: 100, Y: 319}) {
		t.Errorf("a panel a pixel too tall hangs at %+v", got)
	}
}

// The bar's offers are as wide as what they say, the way a panel's are:
// spelled for Windows, "Shift+A  Archived" ran into the "?  Help" after it.
func TestTheBarsOffersAreAsWideAsWhatTheySay(t *testing.T) {
	w := New(newFake(), 1200, 800)
	w.keyboard = KeyboardOf("windows")
	l, boxes, offers := w.Layout(), w.BarOffers(), w.Offers()
	if len(boxes) != len(offers) || boxes[0].X != l.Bar.X {
		t.Fatalf("%d boxes for %d offers, the first at %+v", len(boxes), len(offers), boxes[0])
	}
	for i, offer := range offers {
		said := Reckoned(w.Said(offer), reads)
		if boxes[i].W < said[len(said)-1] || boxes[i].W < offerWide {
			t.Errorf("%q needs %d and has %d", w.Said(offer), said[len(said)-1], boxes[i].W)
		}
		if i > 0 && boxes[i].X != boxes[i-1].X+boxes[i-1].W {
			t.Errorf("%q starts at %d, not where the one before it ends", offer, boxes[i].X)
		}
		if boxes[i].Y != OfferAt(l, i).Y || boxes[i].H != OfferAt(l, i).H {
			t.Errorf("%q is off the bar: %+v", offer, boxes[i])
		}
	}
	archived := boxes[2]
	if got, ok := w.offerAt(Point{X: archived.X + archived.W - 2, Y: archived.Y + archived.H/2}); !ok || got != "Archived" {
		t.Errorf("the end of Archived answers %q, %v", got, ok)
	}
	// Under a panel the bar says the board's offers, laid out the same way.
	if under := w.Edit(); len(under.barAlong(under.BoardOffers())) != len(under.BoardOffers()) {
		t.Error("under a panel the bar lays out another number of offers")
	}
}
