package tui

import (
	"cairn/internal/reading"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
)

var today = time.Date(2026, 9, 1, 15, 30, 0, 0, time.UTC)

func on(day string) time.Time {
	parsed, err := time.Parse("2006-01-02", day)
	if err != nil {
		panic(err)
	}
	return parsed
}

func TestStoppedGlyph(t *testing.T) {
	g := GlyphsFor(Unicode)
	for state, want := range map[string]string{
		"blocked":  g.Blocked,
		"watching": g.Watching,
		"paused":   g.Paused,
		"todo":     "",
	} {
		if got := StoppedGlyph(state, g); got != want {
			t.Fatalf("%q is marked %q, want %q", state, got, want)
		}
	}
}

func TestMarks(t *testing.T) {
	g := GlyphsFor(Unicode)
	if got := marks(reading.Card{}, g); got != "" {
		t.Fatalf("a card carrying nothing is marked %q", got)
	}
	if got := marks(reading.Card{Described: true}, g); got != g.Described {
		t.Fatalf("a described card is marked %q", got)
	}
	if got := marks(reading.Card{State: "blocked"}, g); got != g.Blocked {
		t.Fatalf("a blocked card is marked %q", got)
	}
	// The state anchors the right-hand end, so the eye finds it in the same
	// place whether or not there is a description beside it.
	both := marks(reading.Card{Described: true, State: "paused"}, g)
	if !strings.HasSuffix(both, g.Paused) {
		t.Fatalf("the state is not last: %q", both)
	}
	if !strings.HasPrefix(both, g.Described) {
		t.Fatalf("the description mark is not first: %q", both)
	}
}

func TestFactsInAColumn(t *testing.T) {
	g := GlyphsFor(Unicode)
	c := reading.Card{State: "todo", Deadline: on("2026-09-03"), SubDone: 3, SubTotal: 5, Labels: []string{"spec"}}
	got := facts(c, g, plainSkin, today, 40)
	for _, want := range []string{"2d", g.SubTasks + "3/5", g.Label + " spec"} {
		if !strings.Contains(got, want) {
			t.Fatalf("the facts row %q is missing %q", got, want)
		}
	}
}

func TestFactsDropWhatWillNotFit(t *testing.T) {
	g := GlyphsFor(Unicode)
	c := reading.Card{State: "todo", Deadline: on("2026-09-03"), SubDone: 3, SubTotal: 5, Labels: []string{"spec"}}
	got := facts(c, g, plainSkin, today, 10)
	if strings.Contains(got, "spec") {
		t.Fatalf("the label survived a row too narrow for it: %q", got)
	}
	if !strings.Contains(got, "2d") {
		t.Fatalf("the deadline was dropped before the label: %q", got)
	}
}

func TestFactsInTheBand(t *testing.T) {
	g := GlyphsFor(Unicode)
	blocked := facts(reading.Card{State: "blocked", Blockers: []int{12, 13}}, g, plainSkin, today, 40)
	if blocked != g.InTheWay+" 12 13" {
		t.Fatalf("a blocked card says %q", blocked)
	}
	because := facts(reading.Card{State: "watching", Because: "she confirms"}, g, plainSkin, today, 40)
	if because != g.InTheWay+" she confirms" {
		t.Fatalf("a watched card says %q", because)
	}
	nothing := facts(reading.Card{State: "paused"}, g, plainSkin, today, 40)
	if nothing != "" {
		t.Fatalf("a paused card with nothing in its way says %q", nothing)
	}
	// A stalled card says what is in the way instead of when it is due: in the
	// band that is the only question being asked.
	both := facts(reading.Card{State: "blocked", Blockers: []int{1}, Deadline: on("2026-09-02")}, g, plainSkin, today, 40)
	if strings.Contains(both, "1d") {
		t.Fatalf("a stalled card mentioned its deadline: %q", both)
	}
}

func TestCardRowsAreFour(t *testing.T) {
	rows := Rows(reading.Card{Number: 14, Title: "Write the spec"}, GlyphsFor(Unicode), plainSkin, today, 20)
	if len(rows) != 4 {
		t.Fatalf("a card has %d rows, want 4", len(rows))
	}
	if !strings.HasPrefix(rows[0], "14") {
		t.Fatalf("the number is not first: %q", rows[0])
	}
	if rows[1] != "Write the spec" {
		t.Fatalf("the title is not on the second row: %q", rows[1])
	}
}

func TestCardIsAlwaysSixRows(t *testing.T) {
	l := LayoutFor(100, 40)
	for _, c := range []reading.Card{
		{Number: 1, Title: "Short"},
		{Number: 2, Title: "A title long enough to run over two whole lines and then some", Described: true, SubDone: 1, SubTotal: 2, Labels: []string{"a", "b"}, Deadline: on("2026-09-02")},
		{Number: 3, Title: "In focus", InFocus: true},
	} {
		drawn := standing(newFake()).with(plainSkin).card(c, l)
		if got := len(strings.Split(drawn, "\n")); got != CardHeight {
			t.Fatalf("task %d drew %d rows, want %d:\n%s", c.Number, got, CardHeight, drawn)
		}
	}
}

func TestTheOneInFocusIsDrawnUnlikeTheOthers(t *testing.T) {
	l := LayoutFor(100, 40)
	b := standing(newFake()).with(plainSkin)
	plainCard := b.card(reading.Card{Number: 1, Title: "One"}, l)
	held := b.card(reading.Card{Number: 1, Title: "One", InFocus: true}, l)
	if !strings.Contains(held, "╔") {
		t.Fatalf("the task in focus is not drawn with a double border:\n%s", held)
	}
	if strings.Contains(plainCard, "╔") {
		t.Fatalf("a task that is not in focus was drawn with a double border:\n%s", plainCard)
	}
}

func TestTheCardTheCursorHoldsIsDrawnApart(t *testing.T) {
	l := LayoutFor(100, 40)
	c := reading.Card{Number: 1, Title: "One", State: "todo"}
	f := newFake()
	f.cards = []reading.Card{c}
	held := hit(standing(f), "l", "i").with(plainSkin)
	if held.card(c, l) == standing(f).with(plainSkin).card(c, l) {
		t.Fatal("the card under the cursor is drawn exactly like the others")
	}
}

// An archived task showed exactly like a live one. Turn the archive on and the
// board gained rows that behaved differently from their neighbours with nothing
// on them to say so.
func TestAnArchivedTaskCarriesItsOwnMark(t *testing.T) {
	g := GlyphsFor(Unicode)
	live := reading.Card{Number: 1, Title: "One", State: "todo"}
	away := reading.Card{Number: 2, Title: "Two", State: "todo", Archived: true}
	if marks(live, g) == marks(away, g) {
		t.Fatalf("an archived card is marked like a live one: %q", marks(away, g))
	}
	if !strings.Contains(marks(away, g), g.Archived) {
		t.Fatalf("the archived mark is not on the card: %q", marks(away, g))
	}
}

// Labels past the end of the row were dropped without a word, so a card said it
// carried one label when it carried four.
func TestLabelsThatDoNotFitAreCounted(t *testing.T) {
	c := reading.Card{Number: 1, Title: "One", State: "todo",
		Labels: []string{"sprint-172", "backend", "urgent", "review"}}
	row := facts(c, GlyphsFor(Unicode), PlainSkin(), today, 18)
	if !strings.Contains(row, "+3") {
		t.Fatalf("a row that dropped three labels does not say so: %q", row)
	}
}

// The one thing about a blocked task that has to be acted on: every task it
// waits on is done. The command line said so and the board did not.
func TestABlockedCardSaysWhenNothingBlocksItAnyMore(t *testing.T) {
	c := reading.Card{Number: 1, Title: "Ship it", State: "blocked",
		Blockers: []int{2}, Freed: true}
	row := facts(c, GlyphsFor(Unicode), PlainSkin(), today, 40)
	if !strings.Contains(row, "nothing blocks it") {
		t.Fatalf("a freed card says %q", row)
	}
}

// The two ends of the row that counts what it dropped: a width that holds the
// count exactly, and one that holds nothing at all.
func TestARowThatCountsWhatItDropped(t *testing.T) {
	c := reading.Card{Number: 1, Title: "One", State: "todo",
		Labels: []string{"aaa", "bbb", "ccc"}}
	g, s := GlyphsFor(Unicode), PlainSkin()

	// "⚑ aaa  +2" is nine cells. At nine it fits exactly; at eight the last
	// fact goes too and nothing is left to count beside.
	if got := facts(c, g, s, today, 9); got != "⚑ aaa  +2" {
		t.Errorf("at nine the row reads %q", got)
	}
	if got := facts(c, g, s, today, 8); got != "" {
		t.Errorf("at eight the row reads %q, want nothing at all", got)
	}
	if got := facts(c, g, s, today, 2); got != "" {
		t.Errorf("at two the row reads %q, want nothing at all", got)
	}
}

// A long field says how much it holds, beside its name: the half of what a bar
// says that a terminal box can carry. The other half — where in it you are — is
// the line numbers down its side.
func TestALongFieldSaysHowMuchItHolds(t *testing.T) {
	for _, c := range []struct {
		value, want string
	}{
		{"", "description"},
		{"une ligne", "description · 1 line"},
		{"un\ndeux\ntrois", "description · 3 lines"},
	} {
		if got := LongTitle("description", c.value); got != c.want {
			t.Errorf("a description of %q is titled %q, want %q", c.value, got, c.want)
		}
	}
}

// The parts on an open task, the window's way: what is left first, a box
// before each — empty or ticked, the same marks the form's field writes — and
// the heading counting both.
func TestThePartsOfAnOpenTaskSayWhatIsLeft(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "Ship it", State: "todo"},
		{Number: 2, Title: "Draft the columns", State: "done", Parent: 1},
		{Number: 3, Title: "Get the client's nod", State: "todo", Parent: 1},
	}
	b := standing(f)
	// What the rows say, and not how they are dressed: the dressing is checked
	// in the pinned frames, where it is decided.
	var rows []string
	for _, row := range b.subTasks(f.cards[0], 60) {
		rows = append(rows, strip(row))
	}
	if len(rows) != 2 {
		t.Fatalf("%d rows for two parts", len(rows))
	}
	if !strings.Contains(rows[0], "[ ]") || !strings.Contains(rows[0], "Get the client's nod") {
		t.Errorf("the first row is %q, want the part left, with an empty box", rows[0])
	}
	if !strings.Contains(rows[1], "[x]") || !strings.Contains(rows[1], "Draft the columns") {
		t.Errorf("the second row is %q, want the part done, ticked", rows[1])
	}
	if head := b.partsHead(f.cards[0]); !strings.Contains(head, "1 to do · 1 done") {
		t.Errorf("the heading reads %q", head)
	}
}

// A part whose name is longer than its panel is cut to the panel, and the panel
// keeps one row for it: a row cut to the wrong width wraps, and every row that
// wraps pushes the history and everything under it down a line.
func TestALongPartIsCutToItsPanel(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "Ship it", State: "todo"},
		{Number: 2, Title: strings.Repeat("a long name for a part ", 10), State: "todo", Parent: 1},
	}
	b := standing(f)
	under := b.detailLeftUnder(f.cards[0], 40)
	if len(under) != 1 {
		t.Fatalf("%d panels under the description, want the parts' one", len(under))
	}
	// Its border, its name, one row, its border.
	if got := lipgloss.Height(under[0]); got != 4 {
		t.Fatalf("a panel holding one long part is %d rows tall, want 4:\n%s", got, under[0])
	}
}
