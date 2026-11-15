package task

import (
	"cairn/internal/engine/item"
	"errors"
	"testing"
)

func sitting(number int, state item.State) Task {
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
			sitting(1, item.Backlog), sitting(1, item.Backlog), ErrMovedBehindItself},
		{"two tasks in different columns",
			sitting(1, item.Backlog), sitting(2, item.Todo), ErrOtherColumn},
	} {
		err := tc.moving.MayMoveBehind(tc.target)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: MayMoveBehind = %v, want %v", tc.what, err, tc.want)
		}
	}
}
