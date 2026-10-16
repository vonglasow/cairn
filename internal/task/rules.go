package task

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/domain"
)

// ErrStillBlocked is returned when a blocked task is moved into a column that says the work is yours to do while something is still in its way.
var ErrStillBlocked = errors.New("it is still blocked")

var workingColumns = []domain.State{domain.Todo, domain.Current}

func (t Task) inTheWay() bool {
	return len(t.blockers) > 0 || t.because != ""
}

// MayLeaveFor refuses moving a blocked task into a column that claims the work is yours to do while something is still in its way, and says what to take away.
func (t Task) MayLeaveFor(becomes Task) error {
	if t.state != domain.Blocked || !becomes.inTheWay() || !slices.Contains(workingColumns, becomes.state) {
		return nil
	}
	return fmt.Errorf("task %d: %w, %s. take them away first: %s",
		t.number, ErrStillBlocked, becomes.blockedBy(), becomes.whatToTakeAway())
}

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

// FreedBy is the column the task lands in when the change to becomes took the last thing out of its way: todo, because that is the moment somebody said out loud that nothing is in its way any more.
func (t Task) FreedBy(becomes Task) (domain.State, bool) {
	if t.state != domain.Blocked || becomes.state != domain.Blocked {
		return "", false
	}
	if !t.inTheWay() || becomes.inTheWay() {
		return "", false
	}
	return domain.Todo, true
}

// ErrStepStillOpen is returned when a task is asked to be finished while one of its steps is not.
var ErrStepStillOpen = errors.New("a step of it is not done")

// MayFinishAs refuses moving the task to done while one of its steps is not.
func (t Task) MayFinishAs(now Task) error {
	if now.state != domain.Done || t.state == domain.Done {
		return nil
	}
	var open []string
	for i, s := range now.steps {
		if !s.Done {
			open = append(open, strconv.Itoa(i+1))
		}
	}
	if len(open) == 0 {
		return nil
	}
	return fmt.Errorf("task %d: step %s: %w", now.number, strings.Join(open, ", "), ErrStepStillOpen)
}
