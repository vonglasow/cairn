package reading

import (
	"errors"
	"slices"
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/task"
)

// Added is the number of the task an add answered with: the parent the parts
// written in its form are added under. Nothing for an answer with no task.
func Added(answer engine.Answer) int {
	if answer.Task == nil {
		return 0
	}
	return task.OfTask(*answer.Task).Number
}

// errNoParent is what settling parts under no task says: something upstream
// lost the number of the task they belong to, and nothing is touched.
var errNoParent = errors.New("the parts belong to no task: nothing was saved under it")

// PartState is the state a part is put in: done when ticked, todo when not —
// the only two a part may stand in.
func PartState(done bool) string {
	if done {
		return string(domain.Done)
	}
	return string(domain.Todo)
}

// SettleParts settles what hangs under a task with what its form's parts field
// says: a line carrying a number is that part, named and ticked as it reads; a
// line carrying none is a part nobody has written down yet; a number the field
// no longer carries is a part that has gone. had is what hung under it when the
// form was opened.
//
// One operation each, through the engine like everything else. The engine has
// no argument for a set of parts and should not grow one: what it is asked for
// here is adding a task, changing a task and deleting a task, which is what
// these are. Both boards save a task's form, so this is written once.
func SettleParts(work Work, parent int, said string, had []Card) error {
	// Number 0 is no task: under it, every task with no parent would read as
	// a part the field no longer names, and be deleted.
	if parent == 0 {
		return errNoParent
	}
	kept := map[int]bool{}
	for _, part := range PartsIn(said) {
		if part.Number == 0 {
			if _, err := work.Ask(engine.TaskNew, engine.Input{Given: map[string]string{
				"title": part.Title, "parent": strconv.Itoa(parent), "state": PartState(part.Done),
			}}); err != nil {
				return err
			}
			continue
		}
		kept[part.Number] = true
		if _, err := work.Ask(engine.TaskEdit, engine.Input{Given: map[string]string{
			"number": strconv.Itoa(part.Number), "title": part.Title, "state": PartState(part.Done),
		}}); err != nil {
			return err
		}
	}
	for _, was := range Hanging(had, parent) {
		if kept[was.Number] {
			continue
		}
		if _, err := work.Ask(engine.TaskDelete, Numbered(was.Number)); err != nil {
			return err
		}
	}
	return ordering(work, PartsIn(said), Hanging(had, parent))
}

// ordering puts the parts already there in the order the field says, column
// by column: a task moves among the ones it sits with, so the parts done are
// ordered among themselves and those left among theirs — one done put behind
// one to do is a move the engine refuses, and the board shows what is left
// first whatever the order. A column moves nothing when its order is the one
// it had and none of it changed column, or holds a single part; a part ticked
// or unticked on the way
// lands where the engine puts it, so its column is ordered again. A part
// added on the way past goes in behind them, where an add puts it.
func ordering(work Work, said []Part, had []Card) error {
	for _, done := range []bool{false, true} {
		var wanted, was []int
		moved := false
		for _, part := range said {
			if part.Number != 0 && part.Done == done {
				wanted = append(wanted, part.Number)
			}
		}
		for _, c := range had {
			if slices.Contains(wanted, c.Number) {
				was = append(was, c.Number)
				moved = moved || Finished(c) != done
			}
		}
		// One alone in its column has no order to keep.
		if len(wanted) < 2 || !moved && slices.Equal(wanted, was) {
			continue
		}
		if err := movingInto(work, wanted); err != nil {
			return err
		}
	}
	return nil
}

// movingInto puts parts of one column in the order given: the first at its
// head, each next directly behind the one before it.
func movingInto(work Work, order []int) error {
	for i, n := range order {
		in := Numbered(n)
		if i == 0 {
			in.Given["first"] = "true"
		} else {
			in.Given["after"] = strconv.Itoa(order[i-1])
		}
		if _, err := work.Ask(engine.TaskMove, in); err != nil {
			return err
		}
	}
	return nil
}
