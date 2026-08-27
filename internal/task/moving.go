package task

// The rule that says where a task may be moved. It lives here rather than in
// the engine because it is a rule about a task, and it lives beside the other
// three for the same reason: whoever changes what a task is will read them
// together.

import (
	"errors"
	"fmt"
)

// ErrMovedBehindItself is returned when a task is asked to be moved behind
// itself. Taking it out and putting it back relative to a place that has just
// shifted is exactly how an order stops running 1..N without anybody seeing
// it, so this is refused rather than quietly done.
var ErrMovedBehindItself = errors.New("a task cannot be moved behind itself")

// ErrOtherColumn is returned when a task is asked to be moved behind one in
// another column. Moving orders; changing column is an edit, and doing both at
// once would give one act two meanings.
var ErrOtherColumn = errors.New("a task is moved inside its own column")

// ErrNotAmongItsSisters is returned when a sub-task is asked to be moved
// behind a task it does not sit with. A sub-task takes its place among its
// sisters, never among the root tasks: where its parent sits carries the whole
// block.
//
// It is its own reason and not ErrOtherColumn because the two say different
// things. Between two columns it is the column that differs; between a
// sub-task and a root task it is the level, and those two may well share a
// column — saying "not in that column" of them would be false.
var ErrNotAmongItsSisters = errors.New("a sub-task is moved among its sisters")

// MayMoveBehind says whether this task may take the place behind target.
//
// The three refusals are asked in the order that makes each message the true
// one: a task and itself share everything, so that case would otherwise be
// answered by whichever check came first; and a sub-task and a root task may
// share a column, so the level is asked about before the column is.
func (t Task) MayMoveBehind(target Task) error {
	if t.number == target.number {
		return fmt.Errorf("task %d: %w", t.number, ErrMovedBehindItself)
	}
	if t.parent != target.parent {
		return t.notAmongItsSisters(target)
	}
	if t.state != target.state {
		return fmt.Errorf("task %d is in %s and task %d is in %s: %w",
			t.number, t.state, target.number, target.state, ErrOtherColumn)
	}
	return nil
}

// notAmongItsSisters says which of the two hangs under what, since knowing
// that is knowing where the task can go instead.
func (t Task) notAmongItsSisters(target Task) error {
	whose, hanging := t.number, t.parent
	if hanging == NoParent {
		whose, hanging = target.number, target.parent
	}
	return fmt.Errorf("task %d hangs under task %d: %w", whose, hanging, ErrNotAmongItsSisters)
}
