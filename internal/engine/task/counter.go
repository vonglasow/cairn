package task

import "cairn/internal/engine/item"

// Count adds a task to a tally: one more, and one more done when it is.
func Count(tally item.Tally, t Task) item.Tally {
	tally.Total++
	if t.state == item.Done {
		tally.Done++
	}
	return tally
}
