package engine

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

func TestLookPassesOnAStoreThatCannotSayItsVersion(t *testing.T) {
	f := newFake()
	f.refuse = map[string]error{}
	f.versionFails = errNope
	if _, err := Look(f); err == nil {
		t.Fatal("a store that could not say its version was read as though it could")
	}
}

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

func TestThePartsOfATaskAreWrittenAndReadBackAsLines(t *testing.T) {
	said := PartsSaid([]Card{
		{Number: 2, Title: "Relire l'ADR", State: "todo"},
		{Number: 3, Title: "Trancher", State: "done"},
	})
	want := "#2 [ ] Relire l'ADR\n#3 [x] Trancher"
	if said != want {
		t.Fatalf("the parts read\n%q\nwant\n%q", said, want)
	}

	back := PartsIn(said + "\n[ ] Une nouvelle\n   \n[x] Déjà faite")
	if len(back) != 4 {
		t.Fatalf("read back %d parts, want 4 — a line of spaces is not one", len(back))
	}
	for i, c := range []struct {
		number int
		title  string
		done   bool
	}{
		{2, "Relire l'ADR", false},
		{3, "Trancher", true},
		{0, "Une nouvelle", false},
		{0, "Déjà faite", true},
	} {
		if back[i].Number != c.number || back[i].Title != c.title || back[i].Done != c.done {
			t.Errorf("part %d read %+v, want %d %q %v", i, back[i], c.number, c.title, c.done)
		}
	}
}

func TestATaskWithNoPartsSaysNothing(t *testing.T) {
	if said := PartsSaid(nil); said != "" {
		t.Fatalf("a task with no parts reads %q", said)
	}
	if back := PartsIn(""); len(back) != 0 {
		t.Fatalf("nothing read back as %v", back)
	}
}
