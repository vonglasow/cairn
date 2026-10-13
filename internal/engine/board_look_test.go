package engine

import (
	"slices"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

func TestLookAsksTheTwoListingsAndTakesTheFocusFromThem(t *testing.T) {
	f := newFake()
	f.focus = []int{7}

	found, err := Look(f)
	if err != nil {
		t.Fatalf("Look: %v", err)
	}
	if len(f.asked) != 2 {
		t.Fatalf("Look asked for %d operations, want 2", len(f.asked))
	}
	if f.asked[0].op.Thing != "task" || f.asked[0].op.Verb != "list" {
		t.Errorf("first asked %q %q, want task list", f.asked[0].op.Thing, f.asked[0].op.Verb)
	}
	if f.asked[1].op.Thing != "objective" || f.asked[1].op.Verb != "list" {
		t.Errorf("second asked %q %q, want objective list", f.asked[1].op.Thing, f.asked[1].op.Verb)
	}
	if !slices.Equal(found.Focus, []int{7}) {
		t.Errorf("Focus = %v, want 7 read off the listing", found.Focus)
	}
}

func TestLookTurnsTheTwoListingsIntoCardsAndGoals(t *testing.T) {
	f := newFake()
	f.tasks = []task.Task{aTask("Write the spec"), aTask("Ship the board")}
	f.goals = []objective.Objective{aGoal("Ship cairn")}
	f.focus = []int{2}
	f.version = 41

	found, err := Look(f)
	if err != nil {
		t.Fatalf("Look: %v", err)
	}
	if len(found.Cards) != 2 {
		t.Fatalf("read %d cards, want 2", len(found.Cards))
	}
	if found.Cards[0].Title != "Write the spec" {
		t.Errorf("the first card is %q", found.Cards[0].Title)
	}
	if len(found.Goals) != 1 || found.Goals[0].Title != "Ship cairn" {
		t.Errorf("read %v, want one objective", found.Goals)
	}
	if found.Version != 41 {
		t.Errorf("Version = %d, want 41", found.Version)
	}
}

func TestLookPassesOnAListingItCannotRead(t *testing.T) {
	for _, refused := range []string{"task list", "objective list"} {
		t.Run(refused, func(t *testing.T) {
			f := newFake()
			f.tasks = []task.Task{aTask("Write the spec")}
			f.goals = []objective.Objective{aGoal("Ship cairn")}
			f.refuse = map[string]error{refused: errNope}
			if _, err := Look(f); err == nil {
				t.Fatalf("%q was refused and the reading came back whole", refused)
			}
		})
	}
}

func TestACardIsToldWhetherAnythingIsWrittenAboutIt(t *testing.T) {
	f := newFake()
	var told task.AsTold
	aTask("Write the spec").Tell(&told)
	told.Described("why it matters")
	f.tasks = []task.Task{aTask("Nothing written"), told.Task()}

	found, err := Look(f)
	if err != nil {
		t.Fatalf("Look: %v", err)
	}
	if found.Cards[0].Described() {
		t.Error("a task with nothing written about it was said to have a description")
	}
	if !found.Cards[1].Described() {
		t.Error("a task with something written about it was said to have none")
	}
}

func TestACardKnowsWhetherItIsInHand(t *testing.T) {
	f := newFake()
	f.tasks = []task.Task{numberedTask(3, "First"), numberedTask(4, "Second")}
	f.focus = []int{4}

	read := found(t, f)
	if read.Cards[0].InFocus {
		t.Error("a task that is not in hand was marked as though it were")
	}
	if !read.Cards[1].InFocus {
		t.Error("the task in hand was not marked as such")
	}
}

func TestEveryCardInHandIsMarked(t *testing.T) {
	f := newFake()
	f.tasks = []task.Task{numberedTask(3, "First"), numberedTask(4, "Second"), numberedTask(5, "Third")}
	f.focus = []int{3, 5}

	read := found(t, f)
	if !read.Cards[0].InFocus || !read.Cards[2].InFocus {
		t.Error("a task in hand was not marked as such")
	}
	if read.Cards[1].InFocus {
		t.Error("a task that is not in hand was marked as though it were")
	}
}

func numberedTask(n int, title string) task.Task {
	var told task.AsTold
	aTask(title).Tell(&told)
	told.Numbered(n)
	return told.Task()
}

func found(t *testing.T, w Work) Snapshot {
	t.Helper()
	read, err := Look(w)
	if err != nil {
		t.Fatalf("Look: %v", err)
	}
	return read
}

func TestACardSaysWhenNothingBlocksItAnyMore(t *testing.T) {
	f := newFake()
	f.tasks = []task.Task{aTask("Write the spec"), aTask("Ship the board")}
	f.freed = []int{2}

	found, err := Look(f)
	if err != nil {
		t.Fatalf("Look: %v", err)
	}
	for _, card := range found.Cards {
		if card.Freed != (card.Number == 2) {
			t.Errorf("card %d says Freed = %v", card.Number, card.Freed)
		}
	}
}

func TestLookCarriesWhereEachTasksContextLives(t *testing.T) {
	f := newFake()
	f.tasks = []task.Task{aTask("Write the spec")}
	f.context = "/vault/1"
	found, err := Look(f)
	if err != nil {
		t.Fatalf("Look: %v", err)
	}
	if len(found.Cards) != 1 || found.Cards[0].Context != "/vault/1" {
		t.Errorf("the card carries %+v", found.Cards)
	}
}

func TestLookSaysWhenATaskHasGonePastItsEstimate(t *testing.T) {
	var a task.AsTold
	aTask("Write the spec").Tell(&a)
	a.Estimated(time.Hour)
	a.Spent(2 * time.Hour)
	f := newFake()
	f.tasks = []task.Task{a.Task(), aTask("Ship it")}
	found, err := Look(f)
	if err != nil {
		t.Fatal(err)
	}
	if !found.Cards[0].Overrun || found.Cards[1].Overrun {
		t.Errorf("overrun = %v and %v, want only the first", found.Cards[0].Overrun, found.Cards[1].Overrun)
	}
}

func TestLookAsksTheTwoListingsWhateverTheyHold(t *testing.T) {
	var parent task.AsTold
	aTask("Ship it").Tell(&parent)
	parent.Numbered(1)
	f := newFake()
	f.tasks = []task.Task{parent.Task(), aTask("Write the spec")}
	f.goals = []objective.Objective{aGoal("Ship cairn")}
	f.along = map[int]domain.Progress{1: {SubTasks: domain.Tally{Total: 2, Done: 1}}}
	found, err := Look(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.asked) != 2 {
		t.Errorf("Look asked %d operations for two tasks and an objective, want the two listings", len(f.asked))
	}
	if found.Cards[0].SubTotal != 2 || found.Cards[0].SubDone != 1 {
		t.Errorf("task 1 is along %d of %d, want 1 of 2", found.Cards[0].SubDone, found.Cards[0].SubTotal)
	}
}

func TestLookCarriesWhatASearchLooksIn(t *testing.T) {
	f := newFake()
	f.tasks = []task.Task{aTask("Write the spec")}
	f.goals = []objective.Objective{aGoal("Ship cairn")}
	found, err := Look(f)
	if err != nil {
		t.Fatalf("Look: %v", err)
	}
	if !slices.Equal(found.Cards[0].Said, task.Said(f.tasks[0])) {
		t.Errorf("the card says %v, the engine %v", found.Cards[0].Said, task.Said(f.tasks[0]))
	}
	if !slices.Equal(found.Goals[0].Said, objective.Said(f.goals[0])) {
		t.Errorf("the objective says %v, the engine %v", found.Goals[0].Said, objective.Said(f.goals[0]))
	}
}
