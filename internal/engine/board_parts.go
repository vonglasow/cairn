package engine

import (
	"cairn/internal/domain"
	"cairn/internal/task"
)

// Added is the number of the task an add answered with: the parent the parts written in its form are added under.
func Added(answer Answer) int {
	if answer.Task == nil {
		return 0
	}
	return task.OfTask(*answer.Task).Number
}

// PartState is the state a part is put in: done when ticked, todo when not — the only two a part may stand in.
func PartState(done bool) string {
	if done {
		return string(domain.Done)
	}
	return string(domain.Todo)
}

var partsArgument = Argument{Name: "parts", Kind: Text,
	Means: `its parts, one a line: "#4 [x] its title" for one there already, "[ ] its title" for a new one; a part not named is left where it is`}

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
			parts[i].Number = factsOf(made).Number
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

// PartsGone is what hung under a task when its form was opened and its parts lines no longer name: the parts a board deletes, each as an act of its own, once the task and the rest of its parts are saved.
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
