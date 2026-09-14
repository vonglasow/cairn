package tui

import (
	"cairn/internal/reading"
	"strings"
	"testing"
)

func TestStandingGlyph(t *testing.T) {
	g := GlyphsFor(Unicode)
	for state, want := range map[string]string{
		"active":  g.Active,
		"on-hold": g.OnHold,
		"reached": g.Reached,
		"dropped": g.Dropped,
		"":        "",
	} {
		if got := StandingGlyph(state, g); got != want {
			t.Fatalf("%q is marked %q, want %q", state, got, want)
		}
	}
}

func TestBar(t *testing.T) {
	for _, c := range []struct {
		done, total, width int
		want               string
	}{
		{0, 0, 4, "░░░░"},
		{0, 4, 4, "░░░░"},
		{2, 4, 4, "██░░"},
		{4, 4, 4, "████"},
		{1, 3, 6, "██░░░░"},
	} {
		if got := Bar(c.done, c.total, c.width); got != c.want {
			t.Fatalf("Bar(%d, %d, %d) = %q, want %q", c.done, c.total, c.width, got, c.want)
		}
	}
	if got := Bar(1, 2, 0); got != "" {
		t.Fatalf("a bar of no width is %q", got)
	}
}

func TestGoalRows(t *testing.T) {
	g := reading.Goal{Number: 3, Title: "Ship the desktop board", Done: 7, Total: 12, Described: true, State: "active", Labels: []string{"product"}}
	rows := standing(newFake()).with(plainSkin).goalRows(g, 60)
	if len(rows) != 4 {
		t.Fatalf("an objective's card has %d rows, want 4", len(rows))
	}
	if !strings.HasPrefix(rows[0], "3") {
		t.Fatalf("the number is not first: %q", rows[0])
	}
	if !strings.Contains(rows[3], "7/12") {
		t.Fatalf("the count is missing from the last row: %q", rows[3])
	}
	if !strings.Contains(rows[3], "█") {
		t.Fatalf("the bar is missing from the last row: %q", rows[3])
	}
}

func TestGoalWithNothingUnderItSaysNoCount(t *testing.T) {
	rows := standing(newFake()).with(plainSkin).goalRows(reading.Goal{Number: 1, Title: "Nothing yet"}, 60)
	if strings.Contains(rows[3], "0/0") {
		t.Fatalf("an objective nothing serves claimed a count: %q", rows[3])
	}
}

func TestGoalIsAlwaysSixRows(t *testing.T) {
	drawn := standing(newFake()).goalCard(reading.Goal{Number: 1, Title: "Ship the desktop board", Done: 1, Total: 2}, 60)
	if got := len(strings.Split(drawn, "\n")); got != CardHeight {
		t.Fatalf("an objective drew %d rows, want %d:\n%s", got, CardHeight, drawn)
	}
}

func TestTheGoalUnderTheCursorIsDrawnApart(t *testing.T) {
	g := reading.Goal{Number: 1, Title: "One"}
	b := standing(newFake())
	b.goal = 1
	elsewhere := standing(newFake())
	elsewhere.goal = 2
	if b.goalCard(g, 60) == elsewhere.goalCard(g, 60) {
		t.Fatal("the objective under the cursor is drawn exactly like the others")
	}
}

func TestABarOfOneColumn(t *testing.T) {
	// One column is still a bar: it says nothing is done yet, which is true.
	if got := Bar(1, 2, 1); got != "░" {
		t.Fatalf("a bar one column wide is %q", got)
	}
	if got := Bar(2, 2, 1); got != "█" {
		t.Fatalf("a full bar one column wide is %q", got)
	}
}

func TestTheFactsRowKeepsWhatFitsBesideTheBar(t *testing.T) {
	// The bar takes its width out of the row, and what is left decides how many
	// facts follow it. A budget two columns out drops a label or keeps one that
	// should have gone.
	g := reading.Goal{Number: 1, Title: "One", Done: 1, Total: 2, Labels: []string{"ab"}}
	gl := GlyphsFor(Unicode)
	// "1/2  ⚑ ab" is nine columns of facts, and what is left for them is the
	// width less the bar and the two spaces beside it.
	const room = barWidth + 2 + 9
	if got := factsRow(g, gl, room); !strings.Contains(got, "⚑ ab") {
		t.Fatalf("a row with room for the label reads %q", got)
	}
	if got := factsRow(g, gl, room-1); strings.Contains(got, "⚑ ab") {
		t.Fatalf("a row one column too narrow kept the label: %q", got)
	}
}

func TestAGoalWithNothingBesideItsBar(t *testing.T) {
	if got := factsRow(reading.Goal{Number: 1}, GlyphsFor(Unicode), 40); strings.Contains(got, "  ") {
		t.Fatalf("a bar with nothing beside it trailed something: %q", got)
	}
}
