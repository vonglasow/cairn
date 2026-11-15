package task

import (
	"cairn/internal/engine/change"
	"errors"
	"fmt"
	"slices"
)

// What a blocking change can be turned down for.
var (
	ErrBlockerRequired    = errors.New("a blocker is a task number")
	ErrBlockedByItself    = errors.New("a task cannot block itself")
	ErrBlockingCycle      = errors.New("nothing would ever be able to start")
	ErrBlockersBothWays   = errors.New("the blockers are either replaced or changed, not both")
	ErrBlockerSignAtBirth = errors.New("a blocker is added or taken away only when changing a task")
)

// ChangedBlockers reads a change to the blockers of a task.
func ChangedBlockers(given []string, current []int, atCreation bool) ([]int, error) {
	return change.ChangedNumbers(given, current, atCreation,
		ErrBlockerRequired, ErrBlockersBothWays, ErrBlockerSignAtBirth)
}

// MayBlock refuses blockers that would block a task by itself, or in a cycle.
func MayBlock(number int, blockers []int, blocking func(int) ([]int, error)) error {
	if slices.Contains(blockers, number) {
		return fmt.Errorf("task %d: %w", number, ErrBlockedByItself)
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
			return fmt.Errorf("task %d and task %d: %w", number, via[one], ErrBlockingCycle)
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

func blockersOf(t Task) []int {
	return t.Carried().Blockers
}

// CanBeBlocked refuses a task whose blockers would never let anything start.
func CanBeBlocked(tasks Finder, t Task) error {
	f := t.Carried()
	if len(f.Blockers) == 0 {
		return nil
	}
	return MayBlock(f.Number, f.Blockers, func(one int) ([]int, error) {
		found, err := tasks.Find(one)
		if err != nil {
			return nil, err
		}
		return blockersOf(found), nil
	})
}
