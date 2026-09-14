package task

import (
	"errors"
	"strings"
	"testing"

	"cairn/internal/domain"
)

// sitting builds a task carrying only what the moving rule reads: which task it
// is, what it hangs under, and which column it is in.
func sitting(number, parent int, state domain.State) Task {
	var a AsTold
	a.Numbered(number)
	a.Titled("something")
	a.Under(parent)
	a.InColumn(state)
	return a.Task()
}

// Each case names the reason it expects rather than only that it was refused. A
// test asserting "it failed" lets every mutant that fails differently live: swap
// two of these three checks and the refusals come back in the wrong order, which
// only a test reading the reason can see.
func TestMayMoveBehind_refusesWithTheReasonThatIsTrue(t *testing.T) {
	for _, tc := range []struct {
		what   string
		moving Task
		target Task
		want   error
	}{
		{
			"a task and itself share everything, so that is answered first",
			sitting(1, NoParent, domain.Backlog), sitting(1, NoParent, domain.Backlog),
			ErrMovedBehindItself,
		},
		{
			"a sub-task and a root task may share a column, so the level is answered before it",
			sitting(2, 1, domain.Backlog), sitting(3, NoParent, domain.Backlog),
			ErrNotAmongItsSisters,
		},
		{
			"a root task and a sub-task, the other way round",
			sitting(3, NoParent, domain.Backlog), sitting(2, 1, domain.Backlog),
			ErrNotAmongItsSisters,
		},
		{
			"two sub-tasks of different parents do not sit together either",
			sitting(2, 1, domain.Backlog), sitting(4, 3, domain.Backlog),
			ErrNotAmongItsSisters,
		},
		{
			"two root tasks in different columns",
			sitting(1, NoParent, domain.Backlog), sitting(2, NoParent, domain.Todo),
			ErrOtherColumn,
		},
		{
			"two sisters in different columns",
			sitting(2, 1, domain.Backlog), sitting(3, 1, domain.Current),
			ErrOtherColumn,
		},
	} {
		err := tc.moving.MayMoveBehind(tc.target)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: MayMoveBehind = %v, want %v", tc.what, err, tc.want)
		}
	}
}

func TestMayMoveBehind_allowsTwoTasksThatSitTogether(t *testing.T) {
	for _, tc := range []struct {
		what           string
		moving, target Task
	}{
		{"two root tasks in one column", sitting(1, NoParent, domain.Todo), sitting(2, NoParent, domain.Todo)},
		{"two sisters in one column", sitting(2, 1, domain.Current), sitting(3, 1, domain.Current)},
	} {
		if err := tc.moving.MayMoveBehind(tc.target); err != nil {
			t.Errorf("%s: MayMoveBehind = %v, want it allowed", tc.what, err)
		}
	}
}

// The refusal has to say where the task can go instead, which means naming the
// parent — and naming the right one whichever of the two hangs under something.
func TestMayMoveBehind_namesWhoseSubTaskItIs(t *testing.T) {
	for _, tc := range []struct {
		what           string
		moving, target Task
		want           string
	}{
		{"the one moving hangs under something", sitting(2, 1, domain.Backlog), sitting(3, NoParent, domain.Backlog), "task 2 hangs under task 1"},
		{"the target hangs under something", sitting(3, NoParent, domain.Backlog), sitting(2, 1, domain.Backlog), "task 2 hangs under task 1"},
	} {
		err := tc.moving.MayMoveBehind(tc.target)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: MayMoveBehind = %v, want it to say %q", tc.what, err, tc.want)
		}
	}
}

// The column refusal names both columns, since a refusal naming one leaves
// somebody not knowing which of the two to move.
func TestMayMoveBehind_namesTheColumnEachIsIn(t *testing.T) {
	err := sitting(1, NoParent, domain.Backlog).MayMoveBehind(sitting(2, NoParent, domain.Todo))
	if err == nil {
		t.Fatal("MayMoveBehind = nil, want a refusal")
	}
	for _, want := range []string{"task 1 is in backlog", "task 2 is in todo"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the reason %q does not say %q", err, want)
		}
	}
}

// SitTogether is the level rule on its own, asked without a task in hand: a
// board has the same question before it draws, and a board working it out for
// itself is the rule living twice.
func TestSitTogether_isTheLevelAndNotTheColumn(t *testing.T) {
	for _, tc := range []struct {
		what string
		a, b int
		want bool
	}{
		{"two sub-tasks under one parent", 1, 1, true},
		{"two tasks hanging under nothing", NoParent, NoParent, true},
		{"a sub-task and a task", 1, NoParent, false},
		{"a task and a sub-task", NoParent, 1, false},
		{"sub-tasks under different parents", 1, 2, false},
	} {
		if got := SitTogether(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: SitTogether(%d, %d) = %v, want %v", tc.what, tc.a, tc.b, got, tc.want)
		}
	}
}
