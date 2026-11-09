package tui

import (
	"cairn/internal/engine"
	"strings"
	"testing"
	"time"
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
	g := engine.Goal{Number: 3, Title: "Ship the desktop board", Done: 7, Total: 12, Description: "something written", State: "active", Labels: []string{"product"}}
	rows := standing(newFake()).with(plainSkin).goalRows(g, 60)
	if len(rows) != 2 {
		t.Fatalf("an objective's row is %d lines, want 2", len(rows))
	}
	if !strings.Contains(rows[0], "3  Ship the desktop board") || !strings.Contains(rows[0], "⚑ product") {
		t.Fatalf("the first line does not say its number, title and type: %q", rows[0])
	}
	if !strings.Contains(rows[1], "7/12") {
		t.Fatalf("the count is missing from the second line: %q", rows[1])
	}
	if !strings.Contains(rows[1], "█") {
		t.Fatalf("the bar is missing from the second line: %q", rows[1])
	}
}

func TestGoalWithNothingUnderItSaysNoCount(t *testing.T) {
	rows := standing(newFake()).with(plainSkin).goalRows(engine.Goal{Number: 1, Title: "Nothing yet"}, 60)
	if strings.Contains(rows[1], "0/0") {
		t.Fatalf("an objective nothing serves claimed a count: %q", rows[1])
	}
}

func TestGoalIsAlwaysTwoRows(t *testing.T) {
	drawn := standing(newFake()).goalCard(engine.Branch{Goal: engine.Goal{Number: 1, Title: "Ship the desktop board", Done: 1, Total: 2}}, 60)
	if got := len(strings.Split(drawn, "\n")); got != goalHeight {
		t.Fatalf("an objective drew %d rows, want %d:\n%s", got, goalHeight, drawn)
	}
}

func TestTheGoalUnderTheCursorIsDrawnApart(t *testing.T) {
	g := engine.Branch{Goal: engine.Goal{Number: 1, Title: "One"}}
	b := standing(newFake())
	b.goal = 1
	elsewhere := standing(newFake())
	elsewhere.goal = 2
	if b.goalCard(g, 60) == elsewhere.goalCard(g, 60) {
		t.Fatal("the objective under the cursor is drawn exactly like the others")
	}
}

func TestABarOfOneColumn(t *testing.T) {
	if got := Bar(1, 2, 1); got != "░" {
		t.Fatalf("a bar one column wide is %q", got)
	}
	if got := Bar(2, 2, 1); got != "█" {
		t.Fatalf("a full bar one column wide is %q", got)
	}
}

func TestAGoalWithNothingBesideItsBar(t *testing.T) {
	if got := strip(standing(newFake()).goalFacts(engine.Goal{Number: 1}, 40)); strings.Contains(got, "  ") {
		t.Fatalf("a bar with nothing beside it trailed something: %q", got)
	}
}

func TestAnObjectivesFactsSayTheTimeSpentWhenThereIsAny(t *testing.T) {
	b := standing(newFake())
	if got := strip(b.goalFacts(engine.Goal{Done: 1, Total: 2, Spent: 90 * time.Minute}, 80)); !strings.Contains(got, "1/2  1h30m") {
		t.Errorf("the facts read %q", got)
	}
	if got := strip(b.goalFacts(engine.Goal{Done: 1, Total: 2}, 80)); !strings.HasSuffix(strings.TrimSpace(got), "1/2") {
		t.Errorf("the facts with no time spent read %q", got)
	}
}
