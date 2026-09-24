package reading

import (
	"slices"
	"testing"
)

func TestDeal(t *testing.T) {
	all, band := Deal([]Card{
		{Number: 1, State: "backlog"},
		{Number: 2, State: "todo"},
		{Number: 3, State: "current"},
		{Number: 4, State: "done"},
		{Number: 5, State: "blocked"},
		{Number: 6, State: "watching"},
		{Number: 7, State: "paused"},
	})
	if len(all) != Columns {
		t.Fatalf("dealt into %d columns, want %d", len(all), Columns)
	}
	for i, want := range []string{"backlog", "todo", "current", "done"} {
		if all[i].State != want {
			t.Fatalf("column %d is %q, want %q", i, all[i].State, want)
		}
		if len(all[i].Cards) != 1 {
			t.Fatalf("column %q holds %d cards, want 1", want, len(all[i].Cards))
		}
	}
	if len(band) != 3 {
		t.Fatalf("the band holds %d cards, want 3", len(band))
	}
}

func TestDealKeepsTheOrderItWasGiven(t *testing.T) {
	all, _ := Deal([]Card{
		{Number: 9, State: "todo"},
		{Number: 2, State: "todo"},
	})
	got, _ := At(all, "todo")
	if got.Cards[0].Number != 9 || got.Cards[1].Number != 2 {
		t.Fatalf("the order changed: %v", got.Cards)
	}
}

func TestDealPassesOverAStateItDoesNotKnow(t *testing.T) {
	all, band := Deal([]Card{{Number: 1, State: "elsewhere"}})
	if len(band) != 0 {
		t.Fatal("an unknown state went into the band")
	}
	for _, c := range all {
		if len(c.Cards) != 0 {
			t.Fatalf("an unknown state went into %q", c.State)
		}
	}
}

func TestAt(t *testing.T) {
	all, _ := Deal(nil)
	if _, ok := At(all, "todo"); !ok {
		t.Fatal("there is no todo column")
	}
	if _, ok := At(all, "blocked"); ok {
		t.Fatal("the band's states must not have a column of their own")
	}
}

func TestIndexOfAndStateAt(t *testing.T) {
	for i, state := range []string{"backlog", "todo", "current", "done"} {
		if got := IndexOf(state); got != i {
			t.Fatalf("%q stands at %d, want %d", state, got, i)
		}
		if got := StateAt(i); got != state {
			t.Fatalf("place %d is %q, want %q", i, got, state)
		}
	}
	if got := IndexOf("blocked"); got != -1 {
		t.Fatalf("a state with no column answered %d", got)
	}
	if got := StateAt(-1); got != "" {
		t.Fatalf("before the first column answered %q", got)
	}
	if got := StateAt(Columns); got != "" {
		t.Fatalf("past the last column answered %q", got)
	}
}

// OnTheBoard is what a board lays out, which is every task hanging under
// nothing. Anything counting what is shown, or walking it with the arrows, has
// to agree with what is drawn — so both ask this rather than each deciding.
func TestOnTheBoard_isEveryTaskHangingUnderNothing(t *testing.T) {
	cards := []Card{
		{Number: 1, Title: "Ship it"},
		{Number: 2, Title: "Write the spec", Parent: 1},
		{Number: 3, Title: "Something else"},
	}
	var got []string
	for _, c := range OnTheBoard(cards) {
		got = append(got, c.Title)
	}
	if want := []string{"Ship it", "Something else"}; !slices.Equal(got, want) {
		t.Errorf("OnTheBoard = %v, want %v", got, want)
	}
}

// A sub-task is not dealt onto a board at all. It is a part of the work its
// parent names, so the parent is what stands in a column and the part is read
// where the parent is read.
func TestDeal_leavesSubTasksOffTheBoard(t *testing.T) {
	dealt, band := Deal([]Card{
		{Number: 1, Title: "Ship it", State: "todo"},
		{Number: 2, Title: "Write the spec", State: "todo", Parent: 1},
		{Number: 3, Title: "Client's answer", State: "blocked"},
		{Number: 4, Title: "A part of it", State: "done", Parent: 3},
		{Number: 5, Title: "Something else", State: "done"},
	})
	var drawn []string
	for _, held := range dealt {
		for _, c := range held.Cards {
			drawn = append(drawn, c.Title)
		}
	}
	for _, c := range band {
		drawn = append(drawn, c.Title)
	}
	want := []string{"Ship it", "Something else", "Client's answer"}
	if !slices.Equal(drawn, want) {
		t.Errorf("Deal drew %v, want %v", drawn, want)
	}
}
