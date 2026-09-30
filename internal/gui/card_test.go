package gui

import (
	"strings"
	"testing"

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
	if len([]rune(got)) != 33 {
		t.Fatalf("the cut title is %d characters, and the room is 33", len([]rune(got)))
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
		{"a description and no parts", reading.Card{Described: true}, g.Described},
		{"parts and no description", reading.Card{SubDone: 1, SubTotal: 3}, g.SubTasks + " 1/3"},
		{"both", reading.Card{Described: true, SubDone: 0, SubTotal: 2},
			g.Described + " " + g.SubTasks + " 0/2"},
		{"nothing under it", reading.Card{SubDone: 0, SubTotal: 0}, ""},
	} {
		if got := Carried(c.card, g); got != c.want {
			t.Errorf("%s: Carried = %q, want %q", c.what, got, c.want)
		}
	}
}
