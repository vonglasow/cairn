package gui

import (
	"cairn/internal/engine"
	"strings"
	"testing"
	"time"
)

func TestATitleThatFitsIsLeftAlone(t *testing.T) {
	if got := TitleOf("Write the spec", 285); got != "Write the spec" {
		t.Fatalf("a title that fits was written %q", got)
	}
}

func TestATitleTooLongIsCutAndSaysSo(t *testing.T) {
	long := strings.Repeat("a", 60)
	got := TitleOf(long, 285)
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("a title that was cut does not say so: %q", got)
	}
	if len([]rune(got)) != 32 {
		t.Fatalf("the cut title is %d characters, and the room is 32", len([]rune(got)))
	}
}

func TestTheCutFallsWhereTheRoomEnds(t *testing.T) {
	room := (285 - padding - padding) / perGlyph
	exactly := strings.Repeat("a", room)
	if got := TitleOf(exactly, 285); got != exactly {
		t.Errorf("a title of exactly the room was cut to %q", got)
	}
	one := strings.Repeat("a", room+1)
	if got := TitleOf(one, 285); got == one {
		t.Error("a title one past the room was not cut")
	}
}

func TestACardWithNoRoomAnswersTheTitleWhole(t *testing.T) {
	if got := TitleOf("Write the spec", 0); got != "Write the spec" {
		t.Fatalf("a card with no room answered %q", got)
	}
	if got := TitleOf("Write the spec", padding+padding+perGlyph); got != "Write the spec" {
		t.Fatalf("a card with room for one answered %q", got)
	}
}

func TestASpaceAtTheCutGoesWithIt(t *testing.T) {
	room := (285 - padding - padding) / perGlyph
	title := strings.Repeat("a", room-1) + " more"
	got := TitleOf(title, 285)
	if strings.HasSuffix(got, " …") {
		t.Fatalf("the cut left a space before the ellipsis: %q", got)
	}
}

func TestTagged_measuresAtTheSizeItIsDrawnIn(t *testing.T) {
	c := engine.Card{Labels: []string{"ado:36467", "chantier:tests-qualite", "sp26-20"}}
	got := Tagged(c, 270)
	if !strings.Contains(got, "chantier:tests-qualite") {
		t.Errorf("Tagged = %q, want room for the second label at the size it is drawn in", got)
	}
}

func TestTagged_atItsEdges(t *testing.T) {
	room := func(glyphs int) int { return glyphs * perSmallGlyph }
	for _, c := range []struct {
		what   string
		labels []string
		room   int
		want   string
	}{
		{"nothing to say", nil, room(20), ""},
		{"one that fits", []string{"devops"}, room(20), "devops"},
		{"two that fit", []string{"devops", "release"}, room(20), "devops  release"},
		{"exactly the room there is", []string{"devops", "release"}, room(15), "devops  release"},
		{"one glyph short of it", []string{"devops", "release"}, room(14), "devops  +1"},
		{"none of them fits", []string{"an-altogether-overlong-label"}, room(4), "+1"},
		{"the count is what is left, not what was dropped last",
			[]string{"devops", "release", "suivi", "backend"}, room(14), "devops  +3"},
	} {
		if got := Tagged(engine.Card{Labels: c.labels}, c.room); got != c.want {
			t.Errorf("%s: Tagged = %q, want %q", c.what, got, c.want)
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
	got := shot.RGBAAt(bubble.X, bubble.Y+bubble.H/2)
	edge := Dark().Edge
	if got.R != edge.R || got.G != edge.G || got.B != edge.B {
		t.Fatalf("the bubble's edge is %v where %v was drawn over the cards", got, edge)
	}
}
