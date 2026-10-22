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
		if f.Name == "blockers" || f.Name == "steps" {
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

func TestTheStepsOfATaskAreWrittenWhatIsLeftFirstAndReadBackAsLines(t *testing.T) {
	said := PartsWritten(Hanging(Card{Steps: []Part{
		{Number: 2, Title: "Trancher", Done: true},
		{Number: 3, Title: "Relire l'ADR"},
	}}))
	want := "#3 [ ] Relire l'ADR\n#2 [x] Trancher"
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
		{3, "Relire l'ADR", false},
		{2, "Trancher", true},
		{0, "Une nouvelle", false},
		{0, "Déjà faite", true},
	} {
		if back[i].Number != c.number || back[i].Title != c.title || back[i].Done != c.done {
			t.Errorf("part %d read %+v, want %d %q %v", i, back[i], c.number, c.title, c.done)
		}
	}
}

func TestATaskWithNoPartsSaysNothing(t *testing.T) {
	if said := PartsWritten(Hanging(Card{})); said != "" {
		t.Fatalf("a task with no parts reads %q", said)
	}
	if back := PartsIn(""); len(back) != 0 {
		t.Fatalf("nothing read back as %v", back)
	}
}

func TestFields_giveAnExampleOfWhatIsTypedInThem(t *testing.T) {
	for _, fields := range [][]Field{TaskFields(Card{}), GoalFields(Goal{}), AddingFields("todo")} {
		for _, f := range fields {
			typed := len(f.Choices) == 0 && f.Name != "description" && f.Name != "objectives"
			if typed && f.Example == "" {
				t.Errorf("the field %q gives no example of what to type in it", f.Name)
			}
			if !typed && f.Example != "" {
				t.Errorf("the field %q, which is not typed, gives the example %q", f.Name, f.Example)
			}
		}
	}
	for name, want := range map[string]string{"deadline": "2026-12-25", "estimate": "1h30m"} {
		for _, f := range TaskFields(Card{}) {
			if f.Name == name && f.Example != want {
				t.Errorf("the field %q gives the example %q, want %q", name, f.Example, want)
			}
		}
	}
}

func TestTheTaskFormAsksForItsLinks(t *testing.T) {
	names := func(fields []Field) []string {
		var all []string
		for _, f := range fields {
			all = append(all, f.Name)
		}
		return all
	}
	edit := EditFields(Card{Number: 1, Linked: []int{3, 5}})
	at := slices.Index(names(edit), "links")
	if at < 1 || edit[at-1].Name != "blockers" {
		t.Fatalf("links are not right after blockers: %v", names(edit))
	}
	if f := edit[at]; f.Arg != "link" || f.Value != "3 5" || f.Was != "3 5" || f.Example != "3 5" {
		t.Errorf("the links field is %+v", f)
	}
	if got, _ := valueOf(AddingFields("todo"), "links"); got != "" {
		t.Errorf("a new task asks for links already filled: %q", got)
	}
	if _, ok := valueOf(AddingFields("todo"), "links"); !ok {
		t.Error("a new task's form has no links field")
	}
}
