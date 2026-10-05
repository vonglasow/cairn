package reading

import (
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

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
