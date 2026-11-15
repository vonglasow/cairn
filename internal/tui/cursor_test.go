package tui

import (
	"cairn/internal/board"
	"strconv"
	"strings"
	"testing"
)

func TestWalkingAcrossKeepsHoldWhereThereIsSomething(t *testing.T) {
	f := twoTasks()
	f.cards = append(f.cards, board.Card{Number: 4, Title: "Over there", State: "current"})
	b := hit(standing(f), "ctrl+l", "l")
	if b.Where().Task != 4 {
		t.Fatalf("the cursor holds %d, want 4", b.Where().Task)
	}
}

func TestWalkingAcrossStopsAtEitherEnd(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "ctrl+l", "ctrl+h", "ctrl+h")
	if b.Where().Column != 0 {
		t.Fatalf("the cursor walked past the first column to %d", b.Where().Column)
	}
}

func TestSteppingUpAndDownAColumn(t *testing.T) {
	b := hit(standing(twoTasks()), "ctrl+l")
	if b.Where().Task != 1 {
		t.Fatalf("the cursor took hold of %d, want 1", b.Where().Task)
	}
	b = hit(b, "j")
	if b.Where().Task != 2 {
		t.Fatalf("j moved to %d, want 2", b.Where().Task)
	}
	b = hit(b, "j")
	if !b.Where().InBand || b.Where().Task != 3 {
		t.Fatalf("j off the last task landed on %d, band %v", b.Where().Task, b.Where().InBand)
	}
	b = hit(b, "k")
	if b.Where().InBand || b.Where().Task != 2 {
		t.Fatalf("k out of the band landed on %d, band %v", b.Where().Task, b.Where().InBand)
	}
	b = hit(b, "k", "k")
	if b.Where().Task != 1 {
		t.Fatal("k walked past the first task")
	}
}

func TestTheCursorFallsToTheNextOneAlong(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "ctrl+l")
	fewer := board.Snapshot{Cards: []board.Card{{Number: 2, Title: "Second", State: "todo"}}}
	after := b.took(Looked{Snapshot: fewer})
	if after.Where().Task != 2 {
		t.Fatalf("the cursor fell to %d, want 2", after.Where().Task)
	}
}

func TestTheCursorFallsBackToTheOneBefore(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "ctrl+l", "j")
	fewer := board.Snapshot{Cards: []board.Card{{Number: 1, Title: "First", State: "todo"}}}
	after := b.took(Looked{Snapshot: fewer})
	if after.Where().Task != 1 {
		t.Fatalf("the cursor fell to %d, want 1", after.Where().Task)
	}
}

func TestTheCursorLetsGoWhenTheColumnEmpties(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "ctrl+l")
	after := b.took(Looked{Snapshot: board.Snapshot{}})
	if after.Where().Task != 0 {
		t.Fatalf("the cursor still holds %d", after.Where().Task)
	}
}

func TestWalkingRightToTheLastColumnAndNoFurther(t *testing.T) {
	f := newFake()
	f.cards = []board.Card{
		{Number: 1, Title: "One", State: "backlog"},
		{Number: 2, Title: "Four", State: "done"},
	}
	b := hit(standing(f), "ctrl+l", "ctrl+l", "ctrl+l")
	if b.Where().Column != lastColumn() {
		t.Fatalf("walking right stopped at %d, want %d", b.Where().Column, lastColumn())
	}
	b = hit(b, "l")
	if b.Where().Column != lastColumn() {
		t.Fatal("walking right went past the last column")
	}
	full := newFake()
	full.cards = []board.Card{
		{Number: 1, Title: "One", State: "backlog"},
		{Number: 2, Title: "Two", State: "todo"},
		{Number: 3, Title: "Three", State: "current"},
		{Number: 4, Title: "Four", State: "done"},
	}
	held := hit(standing(full), "l", "l", "l")
	if held.Where().Task != 4 {
		t.Fatalf("walking right in insert landed on %d, want the task in done", held.Where().Task)
	}
	held = hit(held, "l")
	if held.Where().Column != lastColumn() || held.Where().Task != 4 {
		t.Fatalf("walking right in insert went past the last column, to %d", held.Where().Column)
	}
}

func TestTheCursorFallsToWhateverStandsInItsPlace(t *testing.T) {
	f := newFake()
	f.cards = []board.Card{
		{Number: 5, Title: "First", State: "todo"},
		{Number: 3, Title: "Second", State: "todo"},
		{Number: 9, Title: "Third", State: "todo"},
	}
	b := hit(standing(f), "ctrl+l", "j")
	fewer := board.Snapshot{Cards: []board.Card{
		{Number: 5, Title: "First", State: "todo"},
		{Number: 9, Title: "Third", State: "todo"},
	}}
	after := b.took(Looked{Snapshot: fewer})
	if after.Where().Task != 9 {
		t.Fatalf("the cursor fell to %d, want 9 — the one now standing where it was", after.Where().Task)
	}
}

func TestTheCursorReachesTheFirstColumnAndTheLast(t *testing.T) {
	b := standing(twoTasks())
	b = hit(b, "ctrl+l", "ctrl+l", "ctrl+l", "ctrl+l")
	if b.Where().Column != lastColumn() {
		t.Fatalf("walking right stopped at %d", b.Where().Column)
	}
	b = hit(b, "ctrl+h", "ctrl+h", "ctrl+h", "ctrl+h")
	if b.Where().Column != 0 {
		t.Fatalf("walking left stopped at %d", b.Where().Column)
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

func TestSidewaysDoesNothingOnTheObjectives(t *testing.T) {
	b := hit(standing(manyGoals()), "tab", "j")
	if after := hit(b, "l", "h"); after.Where().Goal != b.Where().Goal {
		t.Fatalf("sideways moved the cursor to %d", after.Where().Goal)
	}
}

func TestWalkingAnEmptyObjectivesList(t *testing.T) {
	b := hit(standing(newFake()), "tab", "j", "k")
	if b.Where().Goal != 0 {
		t.Fatalf("the cursor landed on %d where there are none", b.Where().Goal)
	}
}

func TestTheObjectivesListScrollsToTheCursor(t *testing.T) {
	f := newFake()
	for i := 1; i <= 30; i++ {
		f.goals = append(f.goals, board.Goal{Number: i, Title: "Objective " + strconv.Itoa(i), State: "active"})
	}
	b := hit(standing(f), "tab")
	if b.goalAt != 0 {
		t.Fatalf("the list opened scrolled to %d", b.goalAt)
	}
	for range 28 {
		b = hit(b, "j")
	}
	if b.goalAt == 0 {
		t.Fatal("the list never scrolled, so the last objective is off screen")
	}
	for range 28 {
		b = hit(b, "k")
	}
	if b.goalAt != 0 {
		t.Fatalf("coming back up left the list scrolled to %d", b.goalAt)
	}
}

func TestAColumnScrollsToTheCursor(t *testing.T) {
	f := newFake()
	for i := 1; i <= 12; i++ {
		f.cards = append(f.cards, board.Card{Number: i, Title: "Task " + strconv.Itoa(i), State: "todo"})
	}
	b := hit(standing(f), "ctrl+l")
	if b.offsets[1] != 0 {
		t.Fatalf("the column opened scrolled to %d", b.offsets[1])
	}
	for range 11 {
		b = hit(b, "j")
	}
	if b.offsets[1] == 0 {
		t.Fatal("the column never scrolled, so the last task is off screen")
	}
	if b.Where().Task != 12 {
		t.Fatalf("the cursor reached %d, want the last of twelve", b.Where().Task)
	}
	for range 11 {
		b = hit(b, "k")
	}
	if b.offsets[1] != 0 {
		t.Fatalf("coming back up left the column scrolled to %d", b.offsets[1])
	}
}

func TestTheBandScrollsToTheCursor(t *testing.T) {
	f := newFake()
	f.cards = append(f.cards, board.Card{Number: 1, Title: "One", State: "todo"})
	for i := 20; i < 32; i++ {
		f.cards = append(f.cards, board.Card{Number: i, Title: "Stalled " + strconv.Itoa(i), State: "blocked"})
	}
	b := hit(standing(f), "ctrl+l", "j")
	if !b.Where().InBand {
		t.Fatal("j off the last task did not reach the band")
	}
	for range 11 {
		b = hit(b, "j")
	}
	if b.laneAt[0] == 0 {
		t.Fatal("the band never scrolled, so what is at its bottom is unreachable")
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

func TestWalkingSidewaysKeepsTheLevel(t *testing.T) {
	f := aColumnOf(4)
	f.cards = append(f.cards,
		board.Card{Number: 11, Title: "Also one", State: "current"},
		board.Card{Number: 12, Title: "Also two", State: "current"},
		board.Card{Number: 13, Title: "Also three", State: "current"},
	)
	b := hit(standing(f), "ctrl+l", "j", "j")
	if b.Where().Task != 3 {
		t.Fatalf("the cursor is on %d, want the third of todo", b.Where().Task)
	}
	b = hit(b, "l")
	if b.Where().Task != 13 {
		t.Fatalf("walking across landed on %d, want the third of current", b.Where().Task)
	}
	if back := hit(b, "h"); back.Where().Task != 3 {
		t.Fatalf("walking back landed on %d, want the third of todo", back.Where().Task)
	}
}

func TestWalkingIntoAShorterColumnStopsAtItsEnd(t *testing.T) {
	f := aColumnOf(5)
	f.cards = append(f.cards, board.Card{Number: 11, Title: "Only one", State: "current"})
	b := hit(standing(f), "ctrl+l", "j", "j", "j", "j")
	if b.Where().Task != 5 {
		t.Fatalf("the cursor is on %d, want the fifth", b.Where().Task)
	}
	b = hit(b, "l")
	if b.Where().Task != 11 {
		t.Fatalf("walking into a column of one landed on %d", b.Where().Task)
	}
}

func TestGoesToTheEndOfAColumnAndBackToItsStart(t *testing.T) {
	b := hit(standing(aColumnOf(6)), "ctrl+l")
	b = hit(b, "shift+g")
	if b.Where().Task != 6 {
		t.Fatalf("G landed on %d, want the last", b.Where().Task)
	}
	b = hit(b, "g", "g")
	if b.Where().Task != 1 {
		t.Fatalf("gg landed on %d, want the first", b.Where().Task)
	}
}

func TestOneGOnItsOwnDoesNothingAndIsForgotten(t *testing.T) {
	b := hit(standing(aColumnOf(6)), "ctrl+l")
	after := hit(b, "g")
	if after.Where().Task != 1 || after.pending != "g" {
		t.Fatalf("one g moved to %d, pending %q", after.Where().Task, after.pending)
	}
	after = hit(after, "j")
	if after.pending != "" {
		t.Fatalf("a key after a g left it pending as %q", after.pending)
	}
	if after.Where().Task != 2 {
		t.Fatalf("the key after a g was swallowed: the cursor is on %d", after.Where().Task)
	}
	if again := hit(after, "g"); again.Where().Task != 2 {
		t.Fatalf("a g after that jumped to %d", again.Where().Task)
	}
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
