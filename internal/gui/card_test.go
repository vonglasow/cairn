package gui

import (
	"cairn/internal/domain"
	"cairn/internal/engine"
	"strings"
	"testing"
	"time"
)

func eightEach(said string) int { return 8 * len([]rune(said)) }

func TestATitleThatFitsTakesOneLine(t *testing.T) {
	if got := TitleLines("Write the spec", 200, 200, eightEach); len(got) != 1 || got[0] != "Write the spec" {
		t.Fatalf("a title that fits was laid out %q", got)
	}
}

func TestATitleTooLongForOneLineTakesTwo(t *testing.T) {
	got := TitleLines("Write the spec of the board", 120, 120, eightEach)
	if len(got) != 2 || got[0] != "Write the spec" || got[1] != "of the board" {
		t.Fatalf("a title of two lines was laid out %q", got)
	}
}

func TestATitleTooLongForTwoLinesIsCutAndSaysSo(t *testing.T) {
	got := TitleLines("Fix the band lane scroll when the window is resized", 120, 120, eightEach)
	if len(got) != 2 || !strings.HasSuffix(got[1], "…") || eightEach(got[1]) > 120 {
		t.Fatalf("a title of three lines was laid out %q", got)
	}
	if strings.HasSuffix(got[1], " …") {
		t.Errorf("the cut left a space before the ellipsis: %q", got[1])
	}
}

func TestAWordLongerThanTheLineIsCutAndSaysSo(t *testing.T) {
	got := TitleLines(strings.Repeat("a", 40)+" more", 80, 80, eightEach)
	if len(got) != 1 || eightEach(got[0]) > 80 || !strings.HasSuffix(got[0], "…") {
		t.Fatalf("a word longer than the line was laid out %q", got)
	}
}

func TestAWordThatDoesNotFitTheFirstLineMovesToTheNext(t *testing.T) {
	got := TitleLines("Write the spec", 32, 80, eightEach)
	if len(got) != 2 || got[0] != "" || got[1] != "Write the…" {
		t.Fatalf("a first word too wide for the first line was laid out %q", got)
	}
}

func TestACardWithNoRoomAnswersTheTitleWhole(t *testing.T) {
	if got := TitleLines("Write the spec", 0, 0, eightEach); len(got) != 1 || got[0] != "Write the spec" {
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
	c := engine.Card{Estimate: "2h", Spent: "9h", Overrun: true, Objectives: []int{4}, Serves: []string{"Firmware"}}
	said, over := c.Cost()
	if said != "9h/2h" {
		t.Fatalf("Cost = %q, want 9h/2h", said)
	}
	if !over {
		t.Fatal("nine hours against two is not said to be over")
	}
	if Serves(c) != "Firmware" {
		t.Fatalf("Serves = %q", Serves(c))
	}
	c.Objectives, c.Serves = []int{4, 5}, []string{"Enclosure", "Firmware"}
	if Serves(c) != "Enclosure +1" {
		t.Fatalf("Serves of two = %q", Serves(c))
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
	if said, warn := CardFacts(engine.Card{State: "blocked", Freed: true}, now, engine.Pages{}); said != FreedMark || !warn {
		t.Fatalf("a card waiting on nothing any more says %q %v", said, warn)
	}
	said, warn := CardFacts(engine.Card{}, now, engine.Pages{})
	if said != "" || warn {
		t.Fatalf("a card carrying nothing says %q %v", said, warn)
	}
	said, warn = CardFacts(engine.Card{
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

func titleOf(w Window, column, at int) Point {
	title := PartsOf(w.Layout().CardAt(column, at, w.scroll[column])).Title
	return Point{X: title.X + title.W/2, Y: title.Y + lineHigh/2}
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

func TestOnlyTheTitleOpensItsBubble(t *testing.T) {
	long := strings.Repeat("far too long ", 12)
	w := holding(newFake(), card(1, long, "backlog"))
	steps := PartsOf(w.Layout().CardAt(0, 0, 0)).Steps
	if got := w.Hovering(Point{X: steps.X + steps.W/2, Y: steps.Y}).Bubble(); got != "" {
		t.Fatalf("the card's foot opens the bubble %q", got)
	}
}

func TestTheObjectiveMarkOpensABubbleOfWhatItServes(t *testing.T) {
	c := engine.Card{Number: 1, Title: "Write the spec", State: "todo", Objectives: []int{1, 2}, Serves: []string{"Ship the desktop board", "Write the docs"}}
	w := holding(newFake(), c)
	w.snap.Goals = []engine.Goal{{Number: 1, Title: "Ship the desktop board"}, {Number: 2, Title: "Write the docs"}}
	row := CardRowOf(c, PartsOf(w.Layout().CardAt(1, 0, 0)).Number, w.measured(), w.now, engine.Pages{})
	if row.ServesAt.W == 0 || row.ServesAt.X+row.ServesAt.W != row.FactsAt.X+row.FactsAt.W {
		t.Fatalf("the objective mark is at %+v in the facts %+v", row.ServesAt, row.FactsAt)
	}
	got := w.Hovering(Point{X: row.ServesAt.X + row.ServesAt.W/2, Y: row.ServesAt.Y + row.ServesAt.H/2}).Bubble()
	if got != "#1  Ship the desktop board\n#2  Write the docs" {
		t.Fatalf("the bubble over the objective mark holds %q", got)
	}
}

func TestTheFactsMakeRoomForTheMarkDrawnBeforeWhatTheyServe(t *testing.T) {
	c := engine.Card{Number: 1, Objectives: []int{1}, Serves: []string{"Firmware"}}
	m := holding(newFake()).measured()
	row := CardRowOf(c, Box{W: 300, H: lineHigh}, m, time.Time{}, engine.Pages{})
	if want := mixedWide(m, "Firmware", small) + servesSide + servesGap; row.FactsAt.W < want {
		t.Errorf("the facts %q are given %d wide, want %d at least", row.Facts, row.FactsAt.W, want)
	}
	if row.MarkAt.X != row.ServesAt.X || row.MarkAt.W != servesSide {
		t.Errorf("the mark %+v does not lead what it serves %+v", row.MarkAt, row.ServesAt)
	}
}

func TestTheBubbleIsDrawnLastAndInsideTheWindow(t *testing.T) {
	long := strings.Repeat("far too long ", 12)
	w := holding(newFake(), card(1, "first", "backlog"), card(2, "second", "backlog"), card(3, long, "backlog"))
	w = w.Hovering(titleOf(w, 0, 2))
	bubble, lines, ok := w.BubbleBox()
	if !ok || len(lines) < 2 {
		t.Fatalf("the cut title has no bubble of several lines: %v %v", ok, lines)
	}
	if bubble.X < 0 || bubble.Y < 0 || bubble.X+bubble.W > w.width || bubble.Y+bubble.H > w.height {
		t.Fatalf("the bubble %v is not inside the window", bubble)
	}
	shot := drawn(t, w)
	got := shot.RGBAAt(bubble.X+2, bubble.Y+bubble.H/2)
	ground := Dark().Colour.Surface.Raised
	if got.R != ground.R || got.G != ground.G || got.B != ground.B {
		t.Fatalf("the bubble's ground is %v where %v was drawn over the cards", got, ground)
	}
}

func TestGoMonoIsGivenItsWholeAdvance(t *testing.T) {
	if got := monoWide("1/3", small); got*5 < 3*small*3 {
		t.Errorf("three figures at %d are given %d, under the %v they take", small, got, float64(3*small*3)/5)
	}
}

func TestALabelPillWearsItsColour(t *testing.T) {
	c := engine.Card{Number: 1, Title: "Write the spec", State: "todo", Labels: []string{"release"}}
	w := holding(newFake(), c)
	w.snap.Colours = map[string]string{"release": "peach"}
	skin := Dark()
	if len(skin.Colour.Labels) != 14 {
		t.Fatalf("the skin knows %d label colours", len(skin.Colour.Labels))
	}
	ground, ink := skin.LabelPill("peach")
	if ink != skin.Colour.Labels["peach"] || ground == skin.Colour.Surface.Control {
		t.Fatalf("a peach pill is %v on %v", ink, ground)
	}
	pill := CardRowOf(c, PartsOf(w.Layout().CardAt(1, 0, 0)).Number, w.measured(), w.now, engine.Pages{}).Pills[0]
	if painted(drawn(t, w), Box{X: pill.X + pill.W/2, Y: pill.Y + 1, W: 1, H: 1}, ground) != 1 {
		t.Error("the card's label is not drawn on its colour")
	}
	if got := w.LabelColour("doc"); got != domain.ColourOf("doc") {
		t.Errorf("a label the board has no colour for wears %q", got)
	}
}

func TestALabelsNameStandsInTheMiddleOfItsPill(t *testing.T) {
	c := engine.Card{Number: 1, Title: "A", State: "todo", Labels: []string{"release"}}
	w := holding(newFake(), c)
	shot := drawn(t, w)
	pill := CardRowOf(c, PartsOf(w.Layout().CardAt(1, 0, 0)).Number, w.measured(), w.now, engine.Pages{}).Pills[0]
	if above, under := inkGaps(shot, pill); above-under > 1 || under-above > 1 {
		t.Errorf("release stands %d under its pill's top and %d over its foot", above, under)
	}
}

func TestTheFactsGiveTheirLastPiecesUpBeforeTheyCoverTheNumber(t *testing.T) {
	now := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	c := engine.Card{Number: 10, Deadline: now.AddDate(0, 0, 2), Estimate: "2h", Spent: "9h", Overrun: true, Objectives: []int{4}, Serves: []string{"Ops"}}
	m := holding(newFake()).measured()
	whole, _ := CardFacts(c, now, engine.Pages{})
	number := monoWide("#10", small) + 2*pillGap
	for _, c2 := range []struct {
		room  int
		facts string
		late  bool
	}{
		{number + mixedWide(m, whole, small) + servesSide + servesGap, whole, true},
		{number + mixedWide(m, "2d  9h/2h", small), "2d  9h/2h", true},
		{number + mixedWide(m, "2d", small), "2d", false},
		{number + mixedWide(m, "2d", small) - 1, "", false},
	} {
		row := CardRowOf(c, Box{W: c2.room, H: lineHigh}, m, now, engine.Pages{})
		if row.Facts != c2.facts || row.Late != c2.late {
			t.Errorf("in %d the facts say %q late %v, want %q %v", c2.room, row.Facts, row.Late, c2.facts, c2.late)
		}
		if row.Facts != "" && row.FactsAt.X < row.NumberAt.X+row.NumberAt.W {
			t.Errorf("in %d the facts at %+v cover the number at %+v", c2.room, row.FactsAt, row.NumberAt)
		}
		if strings.HasSuffix(row.Facts, "Ops") != (row.ServesAt != Box{}) {
			t.Errorf("in %d the facts %q mark what they serve at %+v", c2.room, row.Facts, row.ServesAt)
		}
	}
}

func TestCardFactsSayFirstHowManyTasksItIsInTheWayOf(t *testing.T) {
	now := time.Date(2026, 11, 13, 9, 0, 0, 0, time.Local)
	if said, warn := CardFacts(engine.Card{Blocks: 2, Serves: []string{"Launch"}}, now, engine.Pages{}); said != "⊳2  Launch" || warn {
		t.Errorf("CardFacts = %q, %v; want ⊳2 before what it serves, and no warning", said, warn)
	}
}
