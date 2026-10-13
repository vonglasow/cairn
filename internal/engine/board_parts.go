package engine

import (
	"cairn/internal/domain"
	"cairn/internal/task"
)

// Added is the number of the task an add answered with: the parent the parts
// written in its form are added under. Nothing for an answer with no task.
func Added(answer Answer) int {
	if answer.Task == nil {
		return 0
	}
	return task.OfTask(*answer.Task).Number
}

// PartState is the state a part is put in: done when ticked, todo when not —
// the only two a part may stand in.
func PartState(done bool) string {
	if done {
		return string(domain.Done)
	}
	return string(domain.Todo)
}

// partsArgument is how a task's parts are handed over with it: one a line,
// "#4 [x] its title" for a part there already and "[ ] its title" for a new
// one — the lines a board's form writes them in.
var partsArgument = Argument{Name: "parts", Kind: Text,
	Means: `its parts, one a line: "#4 [x] its title" for one there already, "[ ] its title" for a new one; a part not named is left where it is`}

// settleParts settles what hangs under a task with the parts it was handed, in
// the act that handed them over: a line carrying a number is that part, named
// and ticked as it reads; a line carrying none is a new part. A part the lines
// no longer name is left where it is — deleting a task is an act of its own,
// asked for as one. One part refused and the act is, so none of them is
// written.
func (e *Engine) settleParts(parent int, said string) error {
	parts := PartsIn(said)
	added := map[int]bool{}
	for i, part := range parts {
		state := PartState(part.Done)
		if part.Number == 0 {
			made, err := e.add(part.Title, Changes{Parent: &parent, State: &state})
			if err != nil {
				return err
			}
			parts[i].Number = factsOf(made).number
			added[parts[i].Number] = true
			continue
		}
		title := part.Title
		if _, err := e.edit(part.Number, Changes{Title: &title, State: &state}); err != nil {
			return editReason(err)
		}
	}
	return e.orderingParts(parts, added)
}

// orderingParts puts the parts already there in the order the lines say,
// column by column: a task moves among the ones it sits with, so the parts
// done are ordered among themselves and those left among theirs. A part added
// on the way goes in behind them, where an add puts it.
func (e *Engine) orderingParts(parts []Part, added map[int]bool) error {
	for _, done := range []bool{false, true} {
		var order []int
		for _, part := range parts {
			if part.Done == done && !added[part.Number] {
				order = append(order, part.Number)
			}
		}
		for i, n := range order {
			at, behind := First, 0
			if i > 0 {
				at, behind = Behind, order[i-1]
			}
			if len(order) < 2 {
				break
			}
			if err := e.move(n, at, behind); err != nil {
				return err
			}
		}
	}
	return nil
}

// PartsGone is what hung under a task when its form was opened and its parts
// lines no longer name: the parts a board deletes, each as an act of its own,
// once the task and the rest of its parts are saved.
func PartsGone(said string, had []Card, parent int) []int {
	kept := map[int]bool{}
	for _, part := range PartsIn(said) {
		kept[part.Number] = true
	}
	var gone []int
	for _, was := range Hanging(had, parent) {
		if !kept[was.Number] {
			gone = append(gone, was.Number)
		}
	}
	return gone
}
