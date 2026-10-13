package engine

import (
	"errors"
	"fmt"
	"slices"

	"cairn/internal/task"
)

// What a blocking change can be turned down for.
var (
	ErrBlockerRequired    = errors.New("a blocker is a task number")
	ErrBlockedByItself    = errors.New("a task cannot block itself")
	ErrBlockingCycle      = errors.New("nothing would ever be able to start")
	ErrBlockersBothWays   = errors.New("the blockers are either replaced or changed, not both")
	ErrBlockerSignAtBirth = errors.New("a blocker is added or taken away only when changing a task")
	ErrStillBlocked       = task.ErrStillBlocked
)

func changedBlockers(given []string, current []int, atCreation bool) ([]int, error) {
	return changedNumbers(given, current, atCreation,
		ErrBlockerRequired, ErrBlockersBothWays, ErrBlockerSignAtBirth)
}

func mayBlock(number int, blockers []int, blocking func(int) ([]int, error)) error {
	if slices.Contains(blockers, number) {
		return fmt.Errorf("task %d: %w", number, ErrBlockedByItself)
	}
	seen := map[int]bool{number: true}
	waiting := slices.Clone(blockers)
	for len(waiting) > 0 {
		one := waiting[0]
		waiting = waiting[1:]
		if seen[one] {
			return fmt.Errorf("task %d and task %d: %w", number, one, ErrBlockingCycle)
		}
		seen[one] = true
		further, err := blocking(one)
		if err != nil {
			return err
		}
		waiting = append(waiting, further...)
	}
	return nil
}

func blockersOf(t task.Task) []int {
	return factsOf(t).Blockers
}

func (e *Engine) canBeBlocked(t task.Task) error {
	f := factsOf(t)
	if len(f.Blockers) == 0 {
		return nil
	}
	return mayBlock(f.Number, f.Blockers, func(one int) ([]int, error) {
		found, err := e.tasks.Find(one)
		if err != nil {
			return nil, err
		}
		return blockersOf(found), nil
	})
}

func (e *Engine) nothingBlocks() ([]int, error) {
	return e.tasks.NothingBlocks()
}
