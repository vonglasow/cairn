package task

// The rules a change to a task is held to that the task can answer on its own,
// given what it was and what it is about to become. They were worked out in
// the engine from facts pulled out of the task one by one, which put a task's
// rules beside every other caller that might one day need them.

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/domain"
)

// ErrStillBlocked is returned when a blocked task is moved into a column that
// says the work is yours to do while something is still in its way.
var ErrStillBlocked = errors.New("it is still blocked")

// workingColumns are the two columns a task cannot be moved into while
// something is still recorded as blocking it. Both say the work is yours to do
// now, and work waiting on somebody else is not.
var workingColumns = []domain.State{domain.Todo, domain.Current}

// inTheWay says whether anything is recorded as blocking the task: a task
// named as blocking it, or words naming something cairn does not hold.
func (t Task) inTheWay() bool {
	return len(t.blockers) > 0 || t.because != ""
}

// MayLeaveFor refuses moving a blocked task into a column that claims the work
// is yours to do while something is still in its way, and says what to take
// away.
//
// A refusal rather than a confirmation: a task waiting on somebody else is not
// a task you are about to do, and a column changed anyway would quietly make
// the record wrong.
func (t Task) MayLeaveFor(becomes Task) error {
	if t.state != domain.Blocked || !becomes.inTheWay() || !slices.Contains(workingColumns, becomes.state) {
		return nil
	}
	return fmt.Errorf("task %d: %w, %s. take them away first: %s",
		t.number, ErrStillBlocked, becomes.blockedBy(), becomes.whatToTakeAway())
}

// blockedBy names what is blocking the task, the way a card names it.
func (t Task) blockedBy() string {
	named := make([]string, 0, len(t.blockers)+1)
	for _, one := range t.blockers {
		named = append(named, domain.Ref(one))
	}
	if t.because != "" {
		named = append(named, strconv.Quote(t.because))
	}
	return "by " + strings.Join(named, " and by ")
}

// whatToTakeAway names what has to go, so a refusal is an instruction rather
// than a complaint. It names the things and not the flags: four surfaces stand
// behind the engine, and only one of them has a flag to type.
func (t Task) whatToTakeAway() string {
	how := make([]string, 0, len(t.blockers)+1)
	for _, one := range t.blockers {
		how = append(how, fmt.Sprintf("task %d", one))
	}
	if t.because != "" {
		how = append(how, "the reason it is waiting")
	}
	return strings.Join(how, ", ")
}

// FreedBy is the column the task lands in when the change to becomes took the
// last thing out of its way: todo, because that is the moment somebody said out
// loud that nothing is in its way any more.
//
// Only from blocked, and only when something was there to take away. A task
// nobody ever said was blocked is not put anywhere, and neither is one that
// still has something in its way.
func (t Task) FreedBy(becomes Task) (domain.State, bool) {
	if t.state != domain.Blocked || becomes.state != domain.Blocked {
		return "", false
	}
	if !t.inTheWay() || becomes.inTheWay() {
		return "", false
	}
	return domain.Todo, true
}

// MayFinishAs refuses moving the task to done while something hanging under it
// is not. What hangs under it is asked for only when it is being finished:
// every other change has no use for it.
func (t Task) MayFinishAs(now Task, hanging func(int) ([]Task, error)) error {
	if now.state != domain.Done || t.state == domain.Done {
		return nil
	}
	under, err := hanging(now.number)
	if err != nil {
		return err
	}
	return MayFinish(now.number, under)
}

// MayHang says whether the task may hang where it was told to: a task under
// nothing hangs anywhere; a part takes the column a part may, serves no
// objective of its own, never hangs under itself, and never under a part.
func (t Task) MayHang(parentOf func(int) (Task, error)) error {
	if t.parent == NoParent {
		return nil
	}
	for _, may := range []func() error{t.MayCarryItsColumn, t.MayServeAnObjective, t.MayHangUnderItsParent} {
		if err := may(); err != nil {
			return err
		}
	}
	parent, err := parentOf(t.parent)
	if err != nil {
		return err
	}
	return parent.MayCarrySubTasks()
}

// Work is what among these is work in its own right, its parts left out: a
// sweep acts on the work, and a part goes with the work it belongs to.
func Work(tasks []Task) []Task {
	work := make([]Task, 0, len(tasks))
	for _, one := range tasks {
		if one.parent == NoParent {
			work = append(work, one)
		}
	}
	return work
}
