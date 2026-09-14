package gui

import (
	"strings"
	"testing"
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
