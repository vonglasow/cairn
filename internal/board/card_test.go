package board

import (
	"strings"
	"testing"
	"time"
)

var today = time.Date(2026, 9, 1, 15, 30, 0, 0, time.UTC)

func on(day string) time.Time {
	parsed, err := time.Parse("2006-01-02", day)
	if err != nil {
		panic(err)
	}
	return parsed
}

func TestStopped(t *testing.T) {
	for _, state := range []string{"blocked", "watching", "paused"} {
		if !Stopped(state) {
			t.Fatalf("%q means not moving, and was not read that way", state)
		}
	}
	for _, state := range []string{"backlog", "todo", "current", "done", ""} {
		if Stopped(state) {
			t.Fatalf("%q was read as not moving", state)
		}
	}
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

func TestDeadline(t *testing.T) {
	for _, c := range []struct {
		name       string
		due        time.Time
		text       string
		show, late bool
	}{
		{"none at all", time.Time{}, "", false, false},
		{"today", on("2026-09-01"), "0d", true, false},
		{"in two days", on("2026-09-03"), "2d", true, false},
		{"at the edge of soon", on("2026-09-08"), "7d", true, false},
		{"one day past soon", on("2026-09-09"), "", false, false},
		{"gone by three days", on("2026-08-29"), "−3d", true, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			text, show, late := Deadline(c.due, today)
			if text != c.text || show != c.show || late != c.late {
				t.Fatalf("Deadline = %q %v %v, want %q %v %v", text, show, late, c.text, c.show, c.late)
			}
		})
	}
}

func TestDeadlineIgnoresTheTimeOfDay(t *testing.T) {
	// The answer to "how many days" must not depend on what o'clock it is when
	// the board is opened.
	morning := time.Date(2026, 9, 1, 0, 1, 0, 0, time.UTC)
	evening := time.Date(2026, 9, 1, 23, 59, 0, 0, time.UTC)
	first, _, _ := Deadline(on("2026-09-03"), morning)
	second, _, _ := Deadline(on("2026-09-03"), evening)
	if first != second {
		t.Fatalf("the same deadline read %q in the morning and %q at night", first, second)
	}
}

func TestMarks(t *testing.T) {
	g := GlyphsFor(Unicode)
	if got := marks(Card{}, g); got != "" {
		t.Fatalf("a card carrying nothing is marked %q", got)
	}
	if got := marks(Card{Described: true}, g); got != g.Described {
		t.Fatalf("a described card is marked %q", got)
	}
	if got := marks(Card{State: "blocked"}, g); got != g.Blocked {
		t.Fatalf("a blocked card is marked %q", got)
	}
	// The state anchors the right-hand end, so the eye finds it in the same
	// place whether or not there is a description beside it.
	both := marks(Card{Described: true, State: "paused"}, g)
	if !strings.HasSuffix(both, g.Paused) {
		t.Fatalf("the state is not last: %q", both)
	}
	if !strings.HasPrefix(both, g.Described) {
		t.Fatalf("the description mark is not first: %q", both)
	}
}

func TestFactsInAColumn(t *testing.T) {
	g := GlyphsFor(Unicode)
	c := Card{State: "todo", Deadline: on("2026-09-03"), SubDone: 3, SubTotal: 5, Labels: []string{"spec"}}
	got := facts(c, g, plainSkin, today, 40)
	for _, want := range []string{"2d", g.SubTasks + "3/5", g.Label + " spec"} {
		if !strings.Contains(got, want) {
			t.Fatalf("the facts row %q is missing %q", got, want)
		}
	}
}

func TestFactsDropWhatWillNotFit(t *testing.T) {
	g := GlyphsFor(Unicode)
	c := Card{State: "todo", Deadline: on("2026-09-03"), SubDone: 3, SubTotal: 5, Labels: []string{"spec"}}
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
	blocked := facts(Card{State: "blocked", Blockers: []int{12, 13}}, g, plainSkin, today, 40)
	if blocked != g.InTheWay+" 12 13" {
		t.Fatalf("a blocked card says %q", blocked)
	}
	because := facts(Card{State: "watching", Because: "she confirms"}, g, plainSkin, today, 40)
	if because != g.InTheWay+" she confirms" {
		t.Fatalf("a watched card says %q", because)
	}
	nothing := facts(Card{State: "paused"}, g, plainSkin, today, 40)
	if nothing != "" {
		t.Fatalf("a paused card with nothing in its way says %q", nothing)
	}
	// A stalled card says what is in the way instead of when it is due: in the
	// band that is the only question being asked.
	both := facts(Card{State: "blocked", Blockers: []int{1}, Deadline: on("2026-09-02")}, g, plainSkin, today, 40)
	if strings.Contains(both, "1d") {
		t.Fatalf("a stalled card mentioned its deadline: %q", both)
	}
}

func TestCardRowsAreFour(t *testing.T) {
	rows := Card{Number: 14, Title: "Write the spec"}.Rows(GlyphsFor(Unicode), plainSkin, today, 20)
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
	for _, c := range []Card{
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
	plainCard := b.card(Card{Number: 1, Title: "One"}, l)
	held := b.card(Card{Number: 1, Title: "One", InFocus: true}, l)
	if !strings.Contains(held, "╔") {
		t.Fatalf("the task in focus is not drawn with a double border:\n%s", held)
	}
	if strings.Contains(plainCard, "╔") {
		t.Fatalf("a task that is not in focus was drawn with a double border:\n%s", plainCard)
	}
}

func TestTheCardTheCursorHoldsIsDrawnApart(t *testing.T) {
	l := LayoutFor(100, 40)
	c := Card{Number: 1, Title: "One", State: "todo"}
	f := newFake()
	f.cards = []Card{c}
	held := hit(standing(f), "l", "i").with(plainSkin)
	if held.card(c, l) == standing(f).with(plainSkin).card(c, l) {
		t.Fatal("the card under the cursor is drawn exactly like the others")
	}
}
