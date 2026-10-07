package reading

import (
	"errors"
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
			if _, err := work.Ask(engine.TaskAdd, engine.Input{Given: map[string]string{
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
	return nil
}
