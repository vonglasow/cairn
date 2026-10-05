package tui

import (
	"strings"
	"testing"

	"cairn/internal/engine"
	"cairn/internal/reading"
)

// Pushing a task: sideways into another column, down into the band and the
// kind of stopped it becomes, up and down among its sisters.

func TestTheKindPopupWalksAndCancels(t *testing.T) {
	// From the last task in the column: from any other, pushing down moves it
	// one place lower rather than out of the column.
	b := hit(standing(twoTasks()), "l", "i", "j", "shift+j")
	if !strings.Contains(bare(b.Draw()), "which kind of stopped") {
		t.Fatal("pushing the last task down asked nothing")
	}
	b = hit(b, "j", "j", "j")
	if b.choice != len(kinds)-1 {
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

func TestPushingIntoTheBandWritesTheKindChosen(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i", "j", "shift+j", "j")
	_, cmd := b.pressed("enter")
	cmd()
	if got := f.edited[2].State; got == nil || *got != "watching" {
		t.Fatalf("the board wrote %v, want watching", got)
	}
}

func TestPushingDownFromTheBandDoesNothing(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "j", "i")
	b = hit(b, "shift+j")
	if strings.Contains(b.Draw(), "which kind of stopped") {
		t.Fatal("pushing down out of the band asked which kind")
	}
}

func TestPushingUpFromAColumnDoesNothing(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i", "shift+k")
	if len(f.edited) != 0 {
		t.Fatalf("pushing up from a column wrote %v", f.edited)
	}
	_ = b
}

func TestPushingRightOutOfTheBandDoesNothing(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "j", "i", "shift+l")
	if len(f.edited) != 0 {
		t.Fatalf("pushing right out of the band wrote %v", f.edited)
	}
	_ = b
}

func TestPushingPastTheLastColumnDoesNothing(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "Done", State: "done"}}
	b := hit(standing(f), "l", "l", "l", "i", "shift+l")
	if len(f.edited) != 0 {
		t.Fatalf("pushing past the last column wrote %v", f.edited)
	}
	_ = b
}

func TestPushingWithNothingHeldDoesNothing(t *testing.T) {
	f := newFake()
	b := standing(f)
	b.mode = Insert
	if _, cmd := b.push(goRight); cmd != nil {
		t.Fatal("a push with no task under the cursor still asked for a change")
	}
}

func TestPushingLeftAndRightLandsInTheRightColumn(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "current"}}
	// current is the third column: left is todo, right is done. A push that
	// went the other way, or one column too far, would land somewhere else.
	left := hit(standing(f), "l", "l", "i")
	if _, cmd := left.pressed("shift+h"); cmd != nil {
		cmd()
	}
	if got := f.edited[1].State; got == nil || *got != "todo" {
		t.Fatalf("pushing left wrote %v, want todo", got)
	}
	right := hit(standing(f), "l", "l", "i")
	if _, cmd := right.pressed("shift+l"); cmd != nil {
		cmd()
	}
	if got := f.edited[1].State; got == nil || *got != "done" {
		t.Fatalf("pushing right wrote %v, want done", got)
	}
}

func TestPushingLeftReachesTheFirstColumn(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "todo"}}
	b := hit(standing(f), "l", "i")
	if _, cmd := b.pressed("shift+h"); cmd != nil {
		cmd()
	}
	if got := f.edited[1].State; got == nil || *got != "backlog" {
		t.Fatalf("pushing left off todo wrote %v, want backlog", got)
	}
}

func TestPushingRightReachesTheLastColumn(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "current"}}
	b := hit(standing(f), "l", "l", "i")
	if _, cmd := b.pressed("shift+l"); cmd != nil {
		cmd()
	}
	if got := f.edited[1].State; got == nil || *got != "done" {
		t.Fatalf("pushing right off current wrote %v, want done", got)
	}
}

func TestThePopupMarksTheOneItIsOn(t *testing.T) {
	b := standing(twoTasks())
	b.skin = SkinOf(Default)
	b.choosing = true
	first := b.kindPopup()
	b.choice = 1
	second := b.kindPopup()
	if first == second {
		t.Fatal("the popup looks the same whichever answer it is on")
	}
	if !strings.Contains(first, b.skin.Picked.Render(" "+kinds[0]+" ")) {
		t.Fatalf("the first answer is not the one picked out:\n%q", first)
	}
	if !strings.Contains(second, b.skin.Picked.Render(" "+kinds[1]+" ")) {
		t.Fatalf("the second answer is not the one picked out:\n%q", second)
	}
}

func TestPushingDownMovesTheTaskOnePlaceLower(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "One", State: "todo"},
		{Number: 2, Title: "Two", State: "todo"},
		{Number: 3, Title: "Three", State: "todo"},
	}
	b := hit(standing(f), "l", "i")
	if _, cmd := b.pressed("shift+j"); cmd != nil {
		cmd()
	}
	if len(f.moved) != 1 {
		t.Fatalf("pushing down asked for %v", f.moved)
	}
	if got := f.moved[0]; got.number != 1 || got.at != engine.Behind || got.target != 2 {
		t.Fatalf("pushing the first down asked for %+v, want task 1 behind task 2", got)
	}
	if len(f.edited) != 0 {
		t.Fatalf("pushing down changed the column: %v", f.edited)
	}
}

func TestPushingUpMovesTheTaskOnePlaceHigher(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "One", State: "todo"},
		{Number: 2, Title: "Two", State: "todo"},
		{Number: 3, Title: "Three", State: "todo"},
	}
	// The third: it goes behind the first, which is one place higher.
	b := hit(standing(f), "l", "i", "j", "j")
	if _, cmd := b.pressed("shift+k"); cmd != nil {
		cmd()
	}
	if got := f.moved[0]; got.number != 3 || got.at != engine.Behind || got.target != 1 {
		t.Fatalf("pushing the third up asked for %+v, want task 3 behind task 1", got)
	}
	// And the second goes to the head, because there is nothing to sit behind.
	f.moved = nil
	second := hit(standing(f), "l", "i", "j")
	if _, cmd := second.pressed("shift+k"); cmd != nil {
		cmd()
	}
	if got := f.moved[0]; got.number != 2 || got.at != engine.First {
		t.Fatalf("pushing the second up asked for %+v, want task 2 first", got)
	}
}

func TestPushingTheFirstTaskUpDoesNothing(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i")
	if _, cmd := b.pressed("shift+k"); cmd != nil {
		t.Fatal("pushing the first task up asked for something")
	}
	_ = b
}

func TestAMoveThatRefusesIsSaid(t *testing.T) {
	f := twoTasks()
	f.failOp = errRefuse
	b := hit(standing(f), "l", "i")
	_, cmd := b.pressed("shift+j")
	if looked, ok := cmd().(Looked); !ok || looked.Err == nil {
		t.Fatal("a move that was refused was passed over")
	}
}
