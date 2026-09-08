package reading

import (
	"cairn/internal/objective"
	"cairn/internal/task"
	"testing"
)

// Reading is asked for the same way as writing, and the focus arrives with the
// listing that marks it rather than through a question of its own.
func TestLookAsksTheTwoListingsAndTakesTheFocusFromThem(t *testing.T) {
	f := newFake()
	f.focus = 7

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
	if found.Focus != 7 {
		t.Errorf("Focus = %d, want 7 read off the listing", found.Focus)
	}
}

// A reading is the two listings turned into what a board draws. These cover
// the turning itself and the three places it can give up: either listing, and
// the store's own version.
func TestLookTurnsTheTwoListingsIntoCardsAndGoals(t *testing.T) {
	f := newFake()
	f.tasks = []task.Task{aTask("Write the spec"), aTask("Ship the board")}
	f.goals = []objective.Objective{aGoal("Ship cairn")}
	f.focus = 2
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
	for _, refused := range []string{"task list", "objective list", "task show", "objective show"} {
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

// Two facts a card carries that are not on the task itself: whether anything
// is written about it, and whether it is the one being worked on. Both are
// worked out here, so both are checked here.
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
	if found.Cards[0].Described {
		t.Error("a task with nothing written about it was said to have a description")
	}
	if !found.Cards[1].Described {
		t.Error("a task with something written about it was said to have none")
	}
}

func TestACardKnowsWhetherItIsTheOneInHand(t *testing.T) {
	f := newFake()
	f.tasks = []task.Task{numbered(3, "First"), numbered(4, "Second")}
	f.focus = 4

	read := found(t, f)
	if read.Cards[0].InFocus {
		t.Error("a task that is not in hand was marked as though it were")
	}
	if !read.Cards[1].InFocus {
		t.Error("the task in hand was not marked as such")
	}
}

// numbered builds a task carrying a number, which only the store gives one in
// earnest — and a card that marks the focus can only be checked against two
// tasks that are telling different numbers.
func numbered(n int, title string) task.Task {
	var told task.AsTold
	aTask(title).Tell(&told)
	told.Numbered(n)
	return told.Task()
}

// found reads the board and fails the test rather than the caller.
func found(t *testing.T, w Work) Snapshot {
	t.Helper()
	read, err := Look(w)
	if err != nil {
		t.Fatalf("Look: %v", err)
	}
	return read
}
