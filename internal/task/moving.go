package task

import (
	"errors"
	"fmt"
)

// ErrMovedBehindItself is returned when a task is asked to be moved behind itself.
var ErrMovedBehindItself = errors.New("a task cannot be moved behind itself")

// ErrOtherColumn is returned when a task is asked to be moved behind one in another column.
var ErrOtherColumn = errors.New("a task is moved inside its own column")

// ErrNotAmongItsSisters is returned when a sub-task is asked to be moved behind a task it does not sit with.
var ErrNotAmongItsSisters = errors.New("a sub-task is moved among its sisters")

// MayMoveBehind says whether this task may take the place behind target.
func (t Task) MayMoveBehind(target Task) error {
	if t.number == target.number {
		return fmt.Errorf("task %d: %w", t.number, ErrMovedBehindItself)
	}
	if !SitTogether(t.parent, target.parent) {
		return t.notAmongItsSisters(target)
	}
	if t.state != target.state {
		return fmt.Errorf("task %d is in %s and task %d is in %s: %w",
			t.number, t.state, target.number, target.state, ErrOtherColumn)
	}
	return nil
}

// SitTogether says whether two tasks take their places in the same run: two sub-tasks under one parent, or two tasks hanging under nothing.
func SitTogether(parent, otherParent int) bool { return parent == otherParent }

func (t Task) notAmongItsSisters(target Task) error {
	whose, hanging := t.number, t.parent
	if hanging == NoParent {
		whose, hanging = target.number, target.parent
	}
	return fmt.Errorf("task %d hangs under task %d: %w", whose, hanging, ErrNotAmongItsSisters)
}
