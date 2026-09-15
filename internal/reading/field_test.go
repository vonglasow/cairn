package reading

import (
	"slices"
	"testing"
)

func TestLookPassesOnWhatItCannotRead(t *testing.T) {
	f := newFake()
	f.fail = errNope
	if _, err := Look(f); err == nil {
		t.Fatal("a store that refused was read as an empty one")
	}
}

func TestMaybe(t *testing.T) {
	if got := Maybe(0); got != "" {
		t.Fatalf("no parent is written %q", got)
	}
	if got := Maybe(7); got != "7" {
		t.Fatalf("a parent is written %q", got)
	}
}

func TestWordsNeverAnswerNothingAtAll(t *testing.T) {
	if got := Words("  "); got == nil || len(got) != 0 {
		t.Fatalf("an empty field became %v, and a nil would mean nobody said anything", got)
	}
	if got := Words("a  b"); len(got) != 2 {
		t.Fatalf("two words became %v", got)
	}
}

func TestGoalFieldsAreFewerAndSayNothingOfBlockers(t *testing.T) {
	fields := GoalFields(Goal{Title: "One", State: "active"})
	for _, f := range fields {
		if f.Name == "blockers" || f.Name == "parent" {
			t.Fatalf("an objective was given a field it has not got: %q", f.Name)
		}
	}
	if len(fields) >= len(TaskFields(Card{})) {
		t.Fatal("an objective's form is not shorter than a task's")
	}
}

func TestNumbersWritesThemSpacedApart(t *testing.T) {
	if got := Numbers([]int{3, 14}); got != "3 14" {
		t.Fatalf("wrote %q, want \"3 14\"", got)
	}
	if got := Numbers(nil); got != "" {
		t.Fatalf("nothing was written %q", got)
	}
}

// A store that answers its listings and then cannot say which version it is at
// is still a store a reading has to give up on: the board follows the version,
// so a reading without one would never notice the next change.
func TestLookPassesOnAStoreThatCannotSayItsVersion(t *testing.T) {
	f := newFake()
	f.refuse = map[string]error{}
	f.versionFails = errNope
	if _, err := Look(f); err == nil {
		t.Fatal("a store that could not say its version was read as though it could")
	}
}

// NumbersIn reads back what Numbers wrote. A form hands its field to the engine
// as the engine takes it, so a surface that showed names had to be able to get
// the numbers again — and getting them by splitting on spaces in the surface
// would be the spelling of an argument living in two places.
func TestNumbersIn_readsBackWhatNumbersWrote(t *testing.T) {
	for _, tc := range []struct {
		what string
		said string
		want []int
	}{
		{"what Numbers writes", Numbers([]int{1, 3}), []int{1, 3}},
		{"one", "4", []int{4}},
		{"nothing at all", "", nil},
		{"spaces and nothing else", "   ", nil},
		{"ragged spacing", " 2   7 ", []int{2, 7}},
		{"a word that is not a number", "1 lamp 3", []int{1, 3}},
	} {
		if got := NumbersIn(tc.said); !slices.Equal(got, tc.want) {
			t.Errorf("%s: NumbersIn(%q) = %v, want %v", tc.what, tc.said, got, tc.want)
		}
	}
}
