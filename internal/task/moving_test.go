package task

import (
	"errors"
	"strings"
	"testing"

	"cairn/internal/domain"
)

func sitting(number int, state domain.State) Task {
	var a AsTold
	a.Numbered(number)
	a.Titled("something")
	a.InColumn(state)
	return a.Task()
}

func TestMayMoveBehind_refusesWithTheReasonThatIsTrue(t *testing.T) {
	for _, tc := range []struct {
		what   string
		moving Task
		target Task
		want   error
	}{
		{"a task and itself share everything, so that is answered first",
			sitting(1, domain.Backlog), sitting(1, domain.Backlog), ErrMovedBehindItself},
		{"two tasks in different columns",
			sitting(1, domain.Backlog), sitting(2, domain.Todo), ErrOtherColumn},
	} {
		err := tc.moving.MayMoveBehind(tc.target)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: MayMoveBehind = %v, want %v", tc.what, err, tc.want)
		}
	}
}

func TestMayMoveBehind_allowsTwoTasksInOneColumn(t *testing.T) {
	if err := sitting(1, domain.Todo).MayMoveBehind(sitting(2, domain.Todo)); err != nil {
		t.Errorf("MayMoveBehind = %v, want it allowed", err)
	}
}

func TestMayMoveBehind_namesTheColumnEachIsIn(t *testing.T) {
	err := sitting(1, domain.Backlog).MayMoveBehind(sitting(2, domain.Todo))
	if err == nil {
		t.Fatal("MayMoveBehind = nil, want a refusal")
	}
	for _, want := range []string{"task 1 is in backlog", "task 2 is in todo"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the reason %q does not say %q", err, want)
		}
	}
}
