package board

import (
	"cairn/internal/engine"
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

func TestLookPassesOnAStoreThatCannotSayItsVersion(t *testing.T) {
	f := newFake()
	f.refuse = map[string]error{}
	f.versionFails = errNope
	if _, err := Look(f); err == nil {
		t.Fatal("a store that could not say its version was read as though it could")
	}
}

func TestTheStepsOfATaskAreWrittenWhatIsLeftFirstAndReadBackAsLines(t *testing.T) {
	said := PartsWritten(Hanging(Card{Steps: []engine.Part{
		{Number: 2, Title: "Trancher", Done: true},
		{Number: 3, Title: "Relire l'ADR"},
	}}))
	want := "#3 [ ] Relire l'ADR\n#2 [x] Trancher"
	if said != want {
		t.Fatalf("the parts read\n%q\nwant\n%q", said, want)
	}

	back := engine.PartsIn(said + "\n[ ] A new one\n   \n[x] Already done")
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
		{0, "A new one", false},
		{0, "Already done", true},
	} {
		if back[i].Number != c.number || back[i].Title != c.title || back[i].Done != c.done {
			t.Errorf("part %d read %+v, want %d %q %v", i, back[i], c.number, c.title, c.done)
		}
	}
}

func TestPartsIn_readsAMarkOnlyWhereItIsWhole(t *testing.T) {
	for _, tc := range []struct {
		said string
		want []engine.Part
	}{
		{"#3", []engine.Part{{Title: "#3"}}},
		{"a]b", []engine.Part{{Title: "a]b"}}},
		{"[ ]\nNext", []engine.Part{{Title: "Next"}}},
	} {
		if got := engine.PartsIn(tc.said); !slices.Equal(got, tc.want) {
			t.Errorf("PartsIn(%q) = %+v, want %+v", tc.said, got, tc.want)
		}
	}
}
