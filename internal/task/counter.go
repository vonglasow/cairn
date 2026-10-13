package task

import "cairn/internal/domain"

// Count adds a task to a tally: one more, and one more done when it is.
func Count(tally domain.Tally, t Task) domain.Tally {
	tally.Total++
	if t.state == domain.Done {
		tally.Done++
	}
	return tally
}
