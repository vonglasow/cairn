package engine

// What blocks a task: how a change to it is read, what makes such a change
// impossible, and the one column change cairn makes on its own.

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/task"
)

// What a blocking change can be turned down for.
var (
	ErrBlockerRequired    = errors.New("a blocker is a task number")
	ErrBlockedByItself    = errors.New("a task cannot block itself")
	ErrBlockingCycle      = errors.New("nothing would ever be able to start")
	ErrBlockersBothWays   = errors.New("the blockers are either replaced or changed, not both")
	ErrBlockerSignAtBirth = errors.New("a blocker is added or taken away only when changing a task")
	ErrStillBlocked       = errors.New("it is still blocked")
)

// blockedColumns are the two columns a task cannot be moved into while
// something is still recorded as blocking it. Both say the work is yours to do
// now, and work waiting on somebody else is not.
var blockedColumns = []domain.State{domain.Todo, domain.Current}

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

// stillBlocked says whether anything is recorded as blocking a task: a task
// named as blocking it, or words naming something cairn does not hold.
func stillBlocked(f taskFacts) bool {
	return len(f.blockers) > 0 || f.because != ""
}

// mayLeave refuses moving a blocked task into a column that claims the work is
// yours to do while something is still in its way, and says what to take away.
//
// A refusal rather than a confirmation: a task waiting on somebody else is not
// a task you are about to do, and a column changed anyway would quietly make
// the record wrong.
func mayLeave(was, becomes taskFacts) error {
	if was.state != domain.Blocked || !stillBlocked(becomes) {
		return nil
	}
	if !slices.Contains(blockedColumns, becomes.state) {
		return nil
	}
	return fmt.Errorf("task %d: %w, %s. take them away first: %s",
		was.number, ErrStillBlocked, inTheWay(becomes), howToTakeAway(becomes))
}

// inTheWay names what is blocking a task, the way a card names it.
func inTheWay(f taskFacts) string {
	named := make([]string, 0, len(f.blockers)+1)
	for _, one := range f.blockers {
		named = append(named, fmt.Sprintf("#%d", one))
	}
	if f.because != "" {
		named = append(named, strconv.Quote(f.because))
	}
	return "by " + strings.Join(named, " and by ")
}

// howToTakeAway names what has to go, so a refusal is an instruction rather
// than a complaint. It names the things and not the flags: the engine is behind
// four surfaces, and only one of them has a flag to type. Somebody on the board
// or in the window was being told to type --blocker -4, which is not something
// either of them can be told.
func howToTakeAway(f taskFacts) string {
	how := make([]string, 0, len(f.blockers)+1)
	for _, one := range f.blockers {
		how = append(how, fmt.Sprintf("task %d", one))
	}
	if f.because != "" {
		how = append(how, "the reason it is waiting")
	}
	return strings.Join(how, ", ")
}

// freedFrom is the column a task lands in when the last thing blocking it is
// taken away: todo, because that is the moment somebody said out loud that
// nothing is in its way any more.
//
// Only from blocked, and only when something was there to take away. A task
// nobody ever said was blocked is not put anywhere, and neither is one that
// still has something in its way.
func freedFrom(was, becomes taskFacts) (domain.State, bool) {
	if was.state != domain.Blocked || becomes.state != domain.Blocked {
		return "", false
	}
	if !stillBlocked(was) || stillBlocked(becomes) {
		return "", false
	}
	return domain.Todo, true
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

// NothingBlocks hands back the tasks waiting on nothing any more, for a listing
// to mark.
func (e *Engine) NothingBlocks() ([]int, error) {
	return e.tasks.NothingBlocks()
}
