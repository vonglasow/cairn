package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine"
	"cairn/internal/gui/geom"
	"strings"
	"testing"
	"time"
)

func eightEach(said string) int { return 8 * len([]rune(said)) }

func TestAWordLongerThanTheLineIsCutAndSaysSo(t *testing.T) {
	got := TitleLines(strings.Repeat("a", 40)+" more", 80, 80, eightEach)
	if len(got) != 1 || eightEach(got[0]) > 80 || !strings.HasSuffix(got[0], "…") {
		t.Fatalf("a word longer than the line was laid out %q", got)
	}
}

func TestCarried_saysOnlyWhatThereIs(t *testing.T) {
	g := Marks()
	for _, c := range []struct {
		what string
		card board.Card
		want string
	}{
		{"nothing to say", board.Card{}, ""},
		{"a description and no parts", board.Card{Description: "something written"}, g.Described},
		{"steps and no description", board.Card{SubDone: 1, SubTotal: 3}, g.Steps + " 1/3"},
		{"both", board.Card{Description: "something written", SubDone: 0, SubTotal: 2},
			g.Described + " " + g.Steps + " 0/2"},
		{"nothing under it", board.Card{SubDone: 0, SubTotal: 0}, ""},
	} {
		if got := Carried(c.card, g); got != c.want {
			t.Errorf("%s: Carried = %q, want %q", c.what, got, c.want)
		}
	}
}

func TestACardSaysWhenItIsDue(t *testing.T) {
	now := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	for _, c := range []struct {
		what string
		due  time.Time
		says string
		late bool
	}{
		{"nothing at all", time.Time{}, "", false},
		{"in two days", now.AddDate(0, 0, 2), "2d", false},
		{"three days ago", now.AddDate(0, 0, -3), "−2d", true},
		{"months away", now.AddDate(0, 3, 0), "", false},
	} {
		t.Run(c.what, func(t *testing.T) {
			said, late := Due(board.Card{Deadline: c.due}, now)
			if said != c.says || late != c.late {
				t.Fatalf("Due = %q %v, want %q %v", said, late, c.says, c.late)
			}
		})
	}
}

func TestAnObjectiveCardSaysWhereItStands(t *testing.T) {
	for standing, want := range map[string]string{
		"active":  "▶",
		"on-hold": "‖",
		"reached": "✓",
		"dropped": "✕",
		"":        "",
	} {
		if got := Standing(board.Goal{State: standing}); got != want {
			t.Errorf("%q is marked %q, want %q", standing, got, want)
		}
	}
}

func TestACardsFactsRunTogetherAndWarnWhenTheyShould(t *testing.T) {
	now := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	if said, warn := CardFacts(board.Card{State: "blocked", Freed: true}, now, engine.Pages{}); said != FreedMark || !warn {
		t.Fatalf("a card waiting on nothing any more says %q %v", said, warn)
	}
	said, warn := CardFacts(board.Card{}, now, engine.Pages{})
	if said != "" || warn {
		t.Fatalf("a card carrying nothing says %q %v", said, warn)
	}
	said, warn = CardFacts(board.Card{
		Deadline: now.AddDate(0, 0, 2), Estimate: "2h", Spent: "9h", Overrun: true, Objectives: []int{4}, Serves: []string{"Firmware"},
	}, now, engine.Pages{})
	for _, word := range []string{"2d", "9h/2h", "Firmware"} {
		if !strings.Contains(said, word) {
			t.Errorf("%q does not say %q", said, word)
		}
	}
	if !warn {
		t.Error("nine hours against two is not warned about")
	}
}

func TestAnObjectiveWithNoStandingWritesNothing(t *testing.T) {
	if got := Standings(board.Goal{}); got != nil {
		t.Fatalf("an objective standing nowhere says %v", got)
	}
	got := Standings(board.Goal{State: "active"})
	if len(got) != 1 || !strings.Contains(got[0], "active") {
		t.Fatalf("an active objective says %v", got)
	}
}

func TestAStepSaysItsTitleAloneAndAServingTaskItsNumber(t *testing.T) {
	c := board.Card{Number: 2, Title: "Draft the columns"}
	if got := StepLine(c); got != "Draft the columns" {
		t.Errorf("a step reads %q: its place is not a task's number, and must not read like one", got)
	}
	if got := SubLine(c); got != "#2  Draft the columns" {
		t.Errorf("a task serving an objective reads %q", got)
	}
}

func titleOf(w Window, column, at int) geom.Point {
	title := PartsOf(w.Layout().CardAt(column, at, w.scroll[column])).Title
	return geom.Point{X: title.X + title.W/2, Y: title.Y + lineHigh/2}
}

func TestTheBubbleHoldsACutTitleWhole(t *testing.T) {
	long := strings.Repeat("far too long ", 12)
	w := holding(newFake(), card(1, long, "backlog"), card(2, "Write the spec", "backlog"))
	cut := w.Hovering(titleOf(w, 0, 0))
	if cut.Bubble() != long {
		t.Fatalf("the bubble over a cut title holds %q", cut.Bubble())
	}
	if got := w.Hovering(titleOf(w, 0, 1)).Bubble(); got != "" {
		t.Fatalf("a title that fits has the bubble %q", got)
	}
	if got := cut.Away().Bubble(); got != "" {
		t.Fatalf("the bubble stays after the pointer left: %q", got)
	}
}

func TestTheObjectiveMarkOpensABubbleOfWhatItServes(t *testing.T) {
	c := board.Card{Number: 1, Title: "Write the spec", State: "todo", Objectives: []int{1, 2}, Serves: []string{"Ship the desktop board", "Write the docs"}}
	w := holding(newFake(), c)
	w.Snap.Goals = []board.Goal{{Number: 1, Title: "Ship the desktop board"}, {Number: 2, Title: "Write the docs"}}
	row := CardRowOf(c, PartsOf(w.Layout().CardAt(1, 0, 0)).Number, w.measured(), w.Now, engine.Pages{})
	if row.ServesAt.W == 0 || row.ServesAt.X+row.ServesAt.W != row.FactsAt.X+row.FactsAt.W {
		t.Fatalf("the objective mark is at %+v in the facts %+v", row.ServesAt, row.FactsAt)
	}
	got := w.Hovering(geom.Point{X: row.ServesAt.X + row.ServesAt.W/2, Y: row.ServesAt.Y + row.ServesAt.H/2}).Bubble()
	if got != "#1  Ship the desktop board\n#2  Write the docs" {
		t.Fatalf("the bubble over the objective mark holds %q", got)
	}
}
