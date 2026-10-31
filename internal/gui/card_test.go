package gui

import (
	"cairn/internal/engine"
	"strings"
	"testing"
	"time"
)

func eightEach(said string) int { return 8 * len([]rune(said)) }

func TestATitleThatFitsTakesOneLine(t *testing.T) {
	if got := TitleLines("Write the spec", 200, eightEach); len(got) != 1 || got[0] != "Write the spec" {
		t.Fatalf("a title that fits was laid out %q", got)
	}
}

func TestATitleTooLongForOneLineTakesTwo(t *testing.T) {
	got := TitleLines("Write the spec of the board", 120, eightEach)
	if len(got) != 2 || got[0] != "Write the spec" || got[1] != "of the board" {
		t.Fatalf("a title of two lines was laid out %q", got)
	}
}

func TestATitleTooLongForTwoLinesIsCutAndSaysSo(t *testing.T) {
	got := TitleLines("Fix the band lane scroll when the window is resized", 120, eightEach)
	if len(got) != 2 || !strings.HasSuffix(got[1], "…") || eightEach(got[1]) > 120 {
		t.Fatalf("a title of three lines was laid out %q", got)
	}
	if strings.HasSuffix(got[1], " …") {
		t.Errorf("the cut left a space before the ellipsis: %q", got[1])
	}
}

func TestAWordLongerThanTheLineIsBroken(t *testing.T) {
	got := TitleLines(strings.Repeat("a", 40), 80, eightEach)
	if len(got) != 2 || eightEach(got[0]) > 80 || !strings.HasSuffix(got[1], "…") {
		t.Fatalf("a word longer than two lines was laid out %q", got)
	}
}

func TestACardWithNoRoomAnswersTheTitleWhole(t *testing.T) {
	if got := TitleLines("Write the spec", 0, eightEach); len(got) != 1 || got[0] != "Write the spec" {
		t.Fatalf("a card with no room answered %q", got)
	}
}

func TestAtMostTwoLabelsArePills(t *testing.T) {
	for _, c := range []struct {
		what   string
		labels []string
		room   int
		shown  []string
		more   int
	}{
		{"nothing to say", nil, 200, nil, 0},
		{"one that fits", []string{"devops"}, 200, []string{"devops"}, 0},
		{"two that fit", []string{"devops", "release"}, 200, []string{"devops", "release"}, 0},
		{"two at most, then how many more", []string{"devops", "release", "gui", "bug"}, 400, []string{"devops", "release"}, 2},
		{"room for one and the count", []string{"devops", "release", "gui"}, 110, []string{"devops"}, 2},
		{"room for the count alone", []string{"an-altogether-overlong-label"}, 40, nil, 1},
		{"room for nothing", []string{"devops"}, 10, nil, 0},
	} {
		shown, more := Pills(c.labels, c.room, eightEach)
		if strings.Join(shown, ",") != strings.Join(c.shown, ",") || more != c.more {
			t.Errorf("%s: pills %q and %d more, want %q and %d", c.what, shown, more, c.shown, c.more)
		}
	}
}

func TestCarried_saysOnlyWhatThereIs(t *testing.T) {
	g := Marks()
	for _, c := range []struct {
		what string
		card engine.Card
		want string
	}{
		{"nothing to say", engine.Card{}, ""},
		{"a description and no parts", engine.Card{Description: "something written"}, g.Described},
		{"steps and no description", engine.Card{SubDone: 1, SubTotal: 3}, g.Steps + " 1/3"},
		{"both", engine.Card{Description: "something written", SubDone: 0, SubTotal: 2},
			g.Described + " " + g.Steps + " 0/2"},
		{"nothing under it", engine.Card{SubDone: 0, SubTotal: 0}, ""},
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
			said, late := Due(engine.Card{Deadline: c.due}, now)
			if said != c.says || late != c.late {
				t.Fatalf("Due = %q %v, want %q %v", said, late, c.says, c.late)
			}
		})
	}
}

func TestACardSummarisesWhatItCostAndWhatItServes(t *testing.T) {
	c := engine.Card{Estimate: "2h", Spent: "9h", Overrun: true, Objectives: []int{4}}
	said, over := c.Cost()
	if said != "9h/2h" {
		t.Fatalf("Cost = %q, want 9h/2h", said)
	}
	if !over {
		t.Fatal("nine hours against two is not said to be over")
	}
	if Serves(c) != "◇4" {
		t.Fatalf("Serves = %q", Serves(c))
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
		if got := Standing(engine.Goal{State: standing}); got != want {
			t.Errorf("%q is marked %q, want %q", standing, got, want)
		}
	}
}

func TestACardsFactsRunTogetherAndWarnWhenTheyShould(t *testing.T) {
	now := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	if said, warn := Facts(engine.Card{State: "blocked", Freed: true}, now); said != FreedMark || !warn {
		t.Fatalf("a card waiting on nothing any more says %q %v", said, warn)
	}
	said, warn := Facts(engine.Card{}, now)
	if said != "" || warn {
		t.Fatalf("a card carrying nothing says %q %v", said, warn)
	}
	said, warn = Facts(engine.Card{
		Deadline: now.AddDate(0, 0, 2), Estimate: "2h", Spent: "9h", Overrun: true, Objectives: []int{4},
	}, now)
	for _, word := range []string{"2d", "9h/2h", "◇4"} {
		if !strings.Contains(said, word) {
			t.Errorf("%q does not say %q", said, word)
		}
	}
	if !warn {
		t.Error("nine hours against two is not warned about")
	}
}

func TestAnObjectiveWithNoStandingWritesNothing(t *testing.T) {
	if got := Standings(engine.Goal{}); got != nil {
		t.Fatalf("an objective standing nowhere says %v", got)
	}
	got := Standings(engine.Goal{State: "active"})
	if len(got) != 1 || !strings.Contains(got[0], "active") {
		t.Fatalf("an active objective says %v", got)
	}
}

func TestAStepSaysItsTitleAloneAndAServingTaskItsNumber(t *testing.T) {
	c := engine.Card{Number: 2, Title: "Draft the columns"}
	if got := StepLine(c); got != "Draft the columns" {
		t.Errorf("a step reads %q: its place is not a task's number, and must not read like one", got)
	}
	if got := SubLine(c); got != "#2  Draft the columns" {
		t.Errorf("a task serving an objective reads %q", got)
	}
}

func TestAnObjectivesCountSaysTheTimeSpentWhenThereIsAny(t *testing.T) {
	if got := Counted(engine.Goal{Done: 1, Total: 2, Spent: 90 * time.Minute}); got != "1/2 · 50% · 1h30m" {
		t.Errorf("Counted = %q", got)
	}
	if got := Counted(engine.Goal{Done: 1, Total: 2}); got != "1/2 · 50%" {
		t.Errorf("Counted with no time = %q", got)
	}
}

func TestTheBubbleHoldsACutTitleWhole(t *testing.T) {
	long := strings.Repeat("far too long ", 12)
	w := holding(newFake(), card(1, long, "backlog"), card(2, "Write the spec", "backlog"))
	cut := w.Hovering(middleOf(w, 0, 0))
	if cut.Bubble() != long {
		t.Fatalf("the bubble over a cut title holds %q", cut.Bubble())
	}
	if got := w.Hovering(middleOf(w, 0, 1)).Bubble(); got != "" {
		t.Fatalf("a title that fits has the bubble %q", got)
	}
	if got := cut.Away().Bubble(); got != "" {
		t.Fatalf("the bubble stays after the pointer left: %q", got)
	}
}

func TestTheBubbleIsDrawnLastAndInsideTheWindow(t *testing.T) {
	long := strings.Repeat("far too long ", 12)
	w := holding(newFake(), card(1, "first", "backlog"), card(2, "second", "backlog"), card(3, long, "backlog"))
	w = w.Hovering(middleOf(w, 0, 2))
	bubble, lines, ok := w.BubbleBox()
	if !ok || len(lines) < 2 {
		t.Fatalf("the cut title has no bubble of several lines: %v %v", ok, lines)
	}
	if bubble.X < 0 || bubble.Y < 0 || bubble.X+bubble.W > w.width || bubble.Y+bubble.H > w.height {
		t.Fatalf("the bubble %v is not inside the window", bubble)
	}
	shot := drawn(t, w)
	got := shot.RGBAAt(bubble.X+2, bubble.Y+bubble.H/2)
	ground := Dark().Panel
	if got.R != ground.R || got.G != ground.G || got.B != ground.B {
		t.Fatalf("the bubble's ground is %v where %v was drawn over the cards", got, ground)
	}
}
