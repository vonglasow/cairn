package task

import (
	"cairn/internal/engine/change"
	"errors"
	"fmt"
)

// What a change to the links of a task can be turned down for.
var (
	ErrLinkRequired    = errors.New("a link is a task number")
	ErrLinksBothWays   = errors.New("the links are either replaced or changed, not both")
	ErrLinkSignAtBirth = errors.New("a link is added or taken away only when changing a task")
	ErrLinkedToItself  = errors.New("a task cannot be linked to itself")
)

// ChangedLinks reads a change to the links of a task.
func ChangedLinks(given []string, current []int, atCreation bool) ([]int, error) {
	return change.ChangedNumbers(given, current, atCreation,
		ErrLinkRequired, ErrLinksBothWays, ErrLinkSignAtBirth)
}

// CanBeLinked refuses a link to the task itself, or to no task.
func CanBeLinked(tasks Finder, t Task) error {
	f := t.Carried()
	for _, one := range f.Linked {
		if one == f.Number {
			return fmt.Errorf("task %d: %w", one, ErrLinkedToItself)
		}
		if _, err := tasks.Find(one); err != nil {
			return err
		}
	}
	return nil
}
