package tui

import (
	"cairn/internal/board"
	"cairn/internal/engine/task"
	"strings"
	"testing"
)

func TestTheKindPopupWalksAndCancels(t *testing.T) {
	b := hit(standing(twoTasks()), "ctrl+l", "j", "shift+j")
	if !strings.Contains(bare(b.Draw()), "which kind of stopped") {
		t.Fatal("pushing the last task down asked nothing")
	}
	b = hit(b, "j", "j", "j")
	if b.choice != len(board.Stalled)-1 {
		t.Fatalf("the popup walked to %d, want its last answer", b.choice)
	}
	b = hit(b, "k", "k", "k")
	if b.choice != 0 {
		t.Fatalf("the popup walked to %d, want its first answer", b.choice)
	}
	b = hit(b, "esc")
	if strings.Contains(b.Draw(), "which kind of stopped") {
		t.Fatal("escape did not close the popup")
	}
}

func TestPushingDownFromTheBandDoesNothing(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "ctrl+l", "ctrl+j")
	b = hit(b, "shift+j")
	if strings.Contains(b.Draw(), "which kind of stopped") {
		t.Fatal("pushing down out of the band asked which kind")
	}
}

func TestPushingUpFromAColumnDoesNothing(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "ctrl+l", "shift+k")
	if len(f.edited) != 0 {
		t.Fatalf("pushing up from a column wrote %v", f.edited)
	}
	_ = b
}

func TestPushingRightOutOfTheBandDoesNothing(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "ctrl+l", "ctrl+j", "shift+l")
	if len(f.edited) != 0 {
		t.Fatalf("pushing right out of the band wrote %v", f.edited)
	}
	_ = b
}

func TestPushingWithNothingHeldDoesNothing(t *testing.T) {
	f := newFake()
	b := standing(f)
	if _, cmd := b.push(goRight); cmd != nil {
		t.Fatal("a push with no task under the cursor still asked for a change")
	}
}

func TestPushingUpMovesTheTaskOnePlaceHigher(t *testing.T) {
	f := newFake()
	f.cards = []board.Card{
		{Number: 1, Title: "One", State: "todo"},
		{Number: 2, Title: "Two", State: "todo"},
		{Number: 3, Title: "Three", State: "todo"},
	}
	b := hit(standing(f), "ctrl+l", "j", "j")
	if _, cmd := b.pressed("shift+k"); cmd != nil {
		cmd()
	}
	if got := f.moved[0]; got.number != 3 || got.at != task.Behind || got.target != 1 {
		t.Fatalf("pushing the third up asked for %+v, want task 3 behind task 1", got)
	}
	f.moved = nil
	second := hit(standing(f), "ctrl+l", "j")
	if _, cmd := second.pressed("shift+k"); cmd != nil {
		cmd()
	}
	if got := f.moved[0]; got.number != 2 || got.at != task.First {
		t.Fatalf("pushing the second up asked for %+v, want task 2 first", got)
	}
}
