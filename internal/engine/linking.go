package engine

import (
	"errors"
	"fmt"

	"cairn/internal/task"
)

// What a change to the links of a task can be turned down for.
var (
	errLinkRequired    = errors.New("a link is a task number")
	errLinksBothWays   = errors.New("the links are either replaced or changed, not both")
	errLinkSignAtBirth = errors.New("a link is added or taken away only when changing a task")
	errLinkedToItself  = errors.New("a task cannot be linked to itself")
)

func changedLinks(given []string, current []int, atCreation bool) ([]int, error) {
	return changedNumbers(given, current, atCreation,
		errLinkRequired, errLinksBothWays, errLinkSignAtBirth)
}

func (e *Engine) canBeLinked(t task.Task) error {
	f := factsOf(t)
	for _, one := range f.Linked {
		if one == f.Number {
			return fmt.Errorf("task %d: %w", one, errLinkedToItself)
		}
		if _, err := e.tasks.Find(one); err != nil {
			return err
		}
	}
	return nil
}
