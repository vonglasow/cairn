package gui

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/reading"
)

// A card is one line of title and a fixed height, so a long name is cut rather
// than allowed to change the shape of the board. These say where the cut falls.

func TestATitleThatFitsIsLeftAlone(t *testing.T) {
	// A column of 285 leaves (285-20)/8 = 33 characters.
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

// Exactly the room it has is not too long. One more is.
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

// A card too narrow to hold anything answers the title whole rather than an
// ellipsis on its own, which says nothing at all.
func TestACardWithNoRoomAnswersTheTitleWhole(t *testing.T) {
	if got := TitleOf("Write the spec", 0); got != "Write the spec" {
		t.Fatalf("a card with no room answered %q", got)
	}
	if got := TitleOf("Write the spec", padding+padding+perGlyph); got != "Write the spec" {
		t.Fatalf("a card with room for one answered %q", got)
	}
}

// A space left at the cut is dropped, so nothing reads "the …".
func TestASpaceAtTheCutGoesWithIt(t *testing.T) {
	room := (285 - padding - padding) / perGlyph
	title := strings.Repeat("a", room-1) + " more"
	got := TitleOf(title, 285)
	if strings.HasSuffix(got, " …") {
		t.Fatalf("the cut left a space before the ellipsis: %q", got)
	}
}

// Labels are measured at the size they are drawn in. They are written in the
// small face while perGlyph is the wide one the title uses, so measuring with
// the title's gabarit left a fifth of every card's label line empty — and on a
// card carrying three, the third was the one dropped.
func TestTagged_measuresAtTheSizeItIsDrawnIn(t *testing.T) {
	c := reading.Card{Labels: []string{"ado:36467", "chantier:tests-qualite", "sp26-20"}}
	got := Tagged(c, 270)
	if !strings.Contains(got, "chantier:tests-qualite") {
		t.Errorf("Tagged = %q, want room for the second label at the size it is drawn in", got)
	}
}

// What a card's label line says, at the edges. Room is counted in glyphs of
// the small face, so the widths here are that many times perSmallGlyph.
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
		if got := Tagged(reading.Card{Labels: c.labels}, c.room); got != c.want {
			t.Errorf("%s: Tagged = %q, want %q", c.what, got, c.want)
		}
	}
}

// What a card says beside its number, at its edges. A card with nothing under
// it says nothing about parts: "0/0" is a card claiming to be a piece of work
// that was split, which it was not.
func TestCarried_saysOnlyWhatThereIs(t *testing.T) {
	g := Marks()
	for _, c := range []struct {
		what string
		card reading.Card
		want string
	}{
		{"nothing to say", reading.Card{}, ""},
		{"a description and no parts", reading.Card{Description: "something written"}, g.Described},
		{"parts and no description", reading.Card{SubDone: 1, SubTotal: 3}, g.SubTasks + " 1/3"},
		{"both", reading.Card{Description: "something written", SubDone: 0, SubTotal: 2},
			g.Described + " " + g.SubTasks + " 0/2"},
		{"nothing under it", reading.Card{SubDone: 0, SubTotal: 0}, ""},
	} {
		if got := Carried(c.card, g); got != c.want {
			t.Errorf("%s: Carried = %q, want %q", c.what, got, c.want)
		}
	}
}

// The window knew nothing about time: no deadline, no notion of late. Deciding
// what to do without knowing what is due is not deciding.
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
			said, late := Due(reading.Card{Deadline: c.due}, now)
			if said != c.says || late != c.late {
				t.Fatalf("Due = %q %v, want %q %v", said, late, c.says, c.late)
			}
		})
	}
}

// Estimated at two hours and nine hours spent is a fact no overview carried.
func TestACardSummarisesWhatItCostAndWhatItServes(t *testing.T) {
	c := reading.Card{Estimate: "2h", Spent: "9h", Overrun: true, Objectives: []int{4}}
	said, over := Spent(c)
	if said != "9h/2h" {
		t.Fatalf("Spent = %q, want 9h/2h", said)
	}
	if !over {
		t.Fatal("nine hours against two is not said to be over")
	}
	if Serves(c) != "◇4" {
		t.Fatalf("Serves = %q", Serves(c))
	}
}

// An objective's standing — active, on hold, reached, dropped — was nowhere on
// its card, so a list mixed what is running with what was abandoned.
func TestAnObjectiveCardSaysWhereItStands(t *testing.T) {
	for standing, want := range map[string]string{
		"active":  "▶",
		"on-hold": "‖",
		"reached": "✓",
		"dropped": "✕",
		"":        "",
	} {
		if got := Standing(reading.Goal{State: standing}); got != want {
			t.Errorf("%q is marked %q, want %q", standing, got, want)
		}
	}
}

// The right-hand end of a card's label row: what is due, what it has cost, and
// what it serves — the three facts the detail carried and no overview did.
func TestACardsFactsRunTogetherAndWarnWhenTheyShould(t *testing.T) {
	now := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	if said, warn := Facts(reading.Card{State: "blocked", Freed: true}, now); said != FreedMark || !warn {
		t.Fatalf("a card waiting on nothing any more says %q %v", said, warn)
	}
	said, warn := Facts(reading.Card{}, now)
	if said != "" || warn {
		t.Fatalf("a card carrying nothing says %q %v", said, warn)
	}
	said, warn = Facts(reading.Card{
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

// Each half of what a task cost says something on its own, and neither is the
// other: an estimate with no time against it is still worth knowing. Whether
// it ran over is the card's to say — reading works it out from the lengths.
func TestWhatATaskCostIsSaidEvenWhenOnlyHalfOfItIsKnown(t *testing.T) {
	for _, c := range []struct {
		what                  string
		spent, estimate, says string
		over                  bool
	}{
		{"neither", "", "", "", false},
		{"only what it cost", "9h", "", "9h", false},
		{"only what was reckoned", "", "2h", "2h", false},
		{"exactly what was reckoned", "2h", "2h", "2h/2h", false},
		{"a minute over", "2h1m", "2h", "2h1m/2h", true},
	} {
		t.Run(c.what, func(t *testing.T) {
			said, over := Spent(reading.Card{Spent: c.spent, Estimate: c.estimate, Overrun: c.over})
			if said != c.says || over != c.over {
				t.Fatalf("Spent = %q %v, want %q %v", said, over, c.says, c.over)
			}
		})
	}
}

// An objective with no standing writes nothing where it would write one, rather
// than an empty mark taking a label's room.
func TestAnObjectiveWithNoStandingWritesNothing(t *testing.T) {
	if got := Standings(reading.Goal{}); got != nil {
		t.Fatalf("an objective standing nowhere says %v", got)
	}
	got := Standings(reading.Goal{State: "active"})
	if len(got) != 1 || !strings.Contains(got[0], "active") {
		t.Fatalf("an active objective says %v", got)
	}
}
