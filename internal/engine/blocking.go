package engine

// What blocks a task: how a change to it is read, what makes such a change
// impossible, and the one column change cairn makes on its own.

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

// changedBlockers reads what a caller said about what blocks a task. The
// grammar and the reading are changedNumbers'; only the words it refuses with
// belong here, because they have to name what the caller was naming.
func changedBlockers(given []string, current []int, atCreation bool) ([]int, error) {
	return changedNumbers(given, current, atCreation,
		ErrBlockerRequired, ErrBlockersBothWays, ErrBlockerSignAtBirth)
}

// mayBlock refuses a set of blockers that could never come unstuck.
//
// Two refusals, and the second is the first seen from far enough away. A task
// cannot block itself. And a task cannot be blocked, however long the way
// round, by something waiting on it — a ring of tasks each waiting on the next
// is a ring where nothing can ever start, so it is always somebody's mistake
// rather than a situation to record.
//
// The walk needs the store, so what blocks a task is asked for rather than
// carried: this is the one thing here that cannot be worked out from the task
// in hand.
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

// blockersOf reads what a task says blocks it, for the walk mayBlock makes.
func blockersOf(t task.Task) []int {
	return factsOf(t).blockers
}

// canBeBlocked refuses a set of blockers a task cannot carry: one of them being
// itself, one of them naming a task the store does not hold, or a ring in which
// nothing could ever start.
//
// A blocker naming nothing is refused rather than answered with an empty
// listing, the way `--parent` is: writing a link to a task that is not there
// leaves a fact in the store that was never true.
func (e *Engine) canBeBlocked(t task.Task) error {
	f := factsOf(t)
	if len(f.blockers) == 0 {
		return nil
	}
	// Every blocker is reached by the walk itself, which asks the store for each
	// one on its way round — so a blocker naming a task the store does not hold
	// comes back from there. Checked once, not twice: a loop here asking the
	// same question first was the same refusal written down again, and nothing
	// could tell whether it or the walk had produced it.
	return mayBlock(f.number, f.blockers, func(one int) ([]int, error) {
		found, err := e.tasks.Find(one)
		if err != nil {
			return nil, err
		}
		return blockersOf(found), nil
	})
}

// nothingBlocks hands back the tasks waiting on nothing any more, for a listing
// to mark.
func (e *Engine) nothingBlocks() ([]int, error) {
	return e.tasks.NothingBlocks()
}
