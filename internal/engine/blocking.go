package engine

import (
	"errors"
	"fmt"
	"slices"

	"cairn/internal/task"
)

// What a blocking change can be turned down for.
var (
	errBlockerRequired    = errors.New("a blocker is a task number")
	errBlockedByItself    = errors.New("a task cannot block itself")
	errBlockingCycle      = errors.New("nothing would ever be able to start")
	errBlockersBothWays   = errors.New("the blockers are either replaced or changed, not both")
	errBlockerSignAtBirth = errors.New("a blocker is added or taken away only when changing a task")
	errStillBlocked       = task.ErrStillBlocked
)

func changedBlockers(given []string, current []int, atCreation bool) ([]int, error) {
	return changedNumbers(given, current, atCreation,
		errBlockerRequired, errBlockersBothWays, errBlockerSignAtBirth)
}

func mayBlock(number int, blockers []int, blocking func(int) ([]int, error)) error {
	if slices.Contains(blockers, number) {
		return fmt.Errorf("task %d: %w", number, errBlockedByItself)
	}
	seen := map[int]bool{number: true}
	via := map[int]int{}
	for _, b := range blockers {
		via[b] = b
	}
	waiting := slices.Clone(blockers)
	for len(waiting) > 0 {
		one := waiting[0]
		waiting = waiting[1:]
		if seen[one] {
			if one != number {
				continue
			}
			return fmt.Errorf("task %d and task %d: %w", number, via[one], errBlockingCycle)
		}
		seen[one] = true
		further, err := blocking(one)
		if err != nil {
			return err
		}
		for _, f := range further {
			if _, known := via[f]; !known || f == number {
				via[f] = via[one]
			}
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
