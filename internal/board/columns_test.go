package board

import (
	"slices"
	"testing"
)

func TestDealDrawsTheTasksInFocusFirstInCurrent(t *testing.T) {
	all, _ := Deal([]Card{
		{Number: 1, State: "current"},
		{Number: 2, State: "current", InFocus: true, Focused: 2},
		{Number: 3, State: "current"},
		{Number: 4, State: "current", InFocus: true, Focused: 1},
		{Number: 5, State: "todo", InFocus: true, Focused: 3},
		{Number: 6, State: "todo"},
	}, false)
	current, _ := At(all, "current")
	var got []int
	for _, c := range current.Cards {
		got = append(got, c.Number)
	}
	if !slices.Equal(got, []int{4, 2, 1, 3}) {
		t.Errorf("current reads %v, want the tasks in focus first, in the order they were focused: [4 2 1 3]", got)
	}
	todo, _ := At(all, "todo")
	if todo.Cards[0].Number != 5 || todo.Cards[1].Number != 6 {
		t.Errorf("todo reads %v: only current puts the focus first", todo.Cards)
	}
}

func TestAt(t *testing.T) {
	all, _ := Deal(nil, false)
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

func TestFound_reachesATaskThroughItsSteps(t *testing.T) {
	all := []Card{
		{Number: 1, Title: "Ship it", Said: []string{"1", "Ship it", "Write the pipeline", "Fix the pipeline"}},
		{Number: 4, Title: "Buy a lamp", Said: []string{"4", "Buy a lamp"}},
	}
	numbers := func(cards []Card) []int {
		var out []int
		for _, c := range cards {
			out = append(out, c.Number)
		}
		return out
	}
	for _, tc := range []struct {
		what  string
		query string
		want  []int
	}{
		{"a word only a step carries", "pipeline", []int{1}},
		{"two steps of one task, once", "the pipeline", []int{1}},
		{"a word the task itself carries", "Ship", []int{1}},
		{"a word carried by neither", "nothing at all", nil},
		{"a word another task carries", "lamp", []int{4}},
		{"nothing asked for, everything laid out", "", []int{1, 4}},
	} {
		if got := numbers(Found(all, tc.query, false)); !slices.Equal(got, tc.want) {
			t.Errorf("%s: Found(%q, false) = %v, want %v", tc.what, tc.query, got, tc.want)
		}
	}
}
