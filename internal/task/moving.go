package task

import (
	"errors"
	"fmt"
)

// ErrMovedBehindItself is returned when a task is asked to be moved behind itself.
var ErrMovedBehindItself = errors.New("a task cannot be moved behind itself")

// ErrOtherColumn is returned when a task is asked to be moved behind one in another column.
var ErrOtherColumn = errors.New("a task is moved inside its own column")

// MayMoveBehind says whether this task may take the place behind target.
func (t Task) MayMoveBehind(target Task) error {
	if t.number == target.number {
		return fmt.Errorf("task %d: %w", t.number, ErrMovedBehindItself)
	}
	if t.state != target.state {
		return fmt.Errorf("task %d is in %s and task %d is in %s: %w",
			t.number, t.state, target.number, target.state, ErrOtherColumn)
	}
	return nil
}
