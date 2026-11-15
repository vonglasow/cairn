package tui

import (
	"cairn/internal/board"
	"strconv"
	"strings"
	"testing"
)

func TestWalkingAcrossStopsAtEitherEnd(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "ctrl+l", "ctrl+h", "ctrl+h")
	if b.Where().Column != 0 {
		t.Fatalf("the cursor walked past the first column to %d", b.Where().Column)
	}
}

func TestWalkingTheObjectives(t *testing.T) {
	b := hit(standing(manyGoals()), "tab")
	if b.Where().Goal != 1 {
		t.Fatalf("the cursor landed on %d, want the first", b.Where().Goal)
	}
	b = hit(b, "j")
	if b.Where().Goal != 2 {
		t.Fatalf("j moved to %d, want 2", b.Where().Goal)
	}
	b = hit(b, "k")
	if b.Where().Goal != 1 {
		t.Fatalf("k moved to %d, want 1", b.Where().Goal)
	}
	b = hit(b, "k")
	if b.Where().Goal != 1 {
		t.Fatal("k walked past the first objective")
	}
	for range 12 {
		b = hit(b, "j")
	}
	if b.Where().Goal != 9 {
		t.Fatalf("j walked to %d, want the last", b.Where().Goal)
	}
}

func TestWalkingAnEmptyObjectivesList(t *testing.T) {
	b := hit(standing(newFake()), "tab", "j", "k")
	if b.Where().Goal != 0 {
		t.Fatalf("the cursor landed on %d where there are none", b.Where().Goal)
	}
}

func TestTheCursorFollowsInsideAnObjective(t *testing.T) {
	f := newFake()
	f.goals = []board.Goal{{Number: 1, Title: "Ship it", State: "active"}}
	for i := 1; i <= 8; i++ {
		f.cards = append(f.cards, board.Card{Number: i, Title: "Task " + strconv.Itoa(i), State: "todo", Objectives: []int{1}})
	}
	b := hit(standing(f), "tab", "enter", "l", "i")
	for range 7 {
		b = hit(b, "j")
	}
	if b.Where().Task != 8 {
		t.Fatalf("the cursor reached %d inside the objective, want the last", b.Where().Task)
	}
	if b.offsets[1] == 0 {
		t.Fatal("the column inside an objective never scrolled")
	}
}

func aColumnOf(n int) *fake {
	f := newFake()
	for i := 1; i <= n; i++ {
		f.cards = append(f.cards, board.Card{Number: i, Title: "Task " + strconv.Itoa(i), State: "todo"})
	}
	return f
}

func TestGoesToTheEndsOfTheBandToo(t *testing.T) {
	f := newFake()
	f.cards = append(f.cards, board.Card{Number: 1, Title: "One", State: "todo"})
	for i := 20; i < 26; i++ {
		f.cards = append(f.cards, board.Card{Number: i, Title: "Stalled", State: "blocked"})
	}
	b := hit(standing(f), "ctrl+l", "j", "shift+g")
	if b.Where().Task != 25 {
		t.Fatalf("G in the band landed on %d, want the last", b.Where().Task)
	}
	if start := hit(b, "g", "g"); start.Where().Task != 20 {
		t.Fatalf("gg in the band landed on %d, want the first", start.Where().Task)
	}
}

func TestGoesToTheEndsOfTheObjectives(t *testing.T) {
	b := hit(standing(manyGoals()), "tab", "shift+g")
	if b.Where().Goal != 9 {
		t.Fatalf("G landed on objective %d, want the last", b.Where().Goal)
	}
	if start := hit(b, "g", "g"); start.Where().Goal != 1 {
		t.Fatalf("gg landed on objective %d, want the first", start.Where().Goal)
	}
}

func TestGIsALetterWhereALetterIsWhatItMeans(t *testing.T) {
	typed := hit(standing(aColumnOf(3)), "/", "g")
	if !strings.Contains(bare(typed.Draw()), "/g") {
		t.Fatalf("a g at the prompt did not reach it:\n%s", bare(typed.Draw()))
	}
	onColumn := hit(standing(aColumnOf(3)), "shift+g")
	if onColumn.Where().Task != 0 {
		t.Fatalf("G on a column took hold of task %d", onColumn.Where().Task)
	}
	inForm := hit(standing(aColumnOf(3)), "ctrl+l", "e", "g")
	if inForm.screen != board.OnForm {
		t.Fatal("a g in the form closed it")
	}
	if !strings.Contains(inForm.form.valueOf("title"), "g") {
		t.Fatalf("a g in the form did not reach the field: %q", inForm.form.valueOf("title"))
	}
	choosing := hit(standing(aColumnOf(1)), "ctrl+l", "shift+j")
	if after := hit(choosing, "shift+g"); !after.choosing {
		t.Fatal("G closed the popup")
	}
}

func TestGoesNowhereWhenThereIsNothingToGoTo(t *testing.T) {
	f := newFake()
	b := standing(f)
	if after := hit(b, "shift+g"); after.Where().Task != 0 {
		t.Fatalf("G on an empty column took hold of %d", after.Where().Task)
	}
	goals := hit(standing(newFake()), "tab")
	if after := hit(goals, "shift+g"); after.Where().Goal != 0 {
		t.Fatalf("G on an empty list took hold of %d", after.Where().Goal)
	}
}
