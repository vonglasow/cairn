package domain

// Tally is how many of one kind of thing there are, and how many of those are
// done.
type Tally struct {
	Total int
	Done  int
}

// Progress is how far along something is. Tasks and sub-tasks are counted
// apart because they are not the same thing — three tasks and four sub-tasks
// is never seven tasks — while the share is over both, since a sub-task is
// work that has to happen for the objective to be reached.
//
// A task fills SubTasks alone: nothing but sub-tasks hangs under a task.
type Progress struct {
	Tasks    Tally
	SubTasks Tally
}

// Any says whether anything is under it at all. Nothing under it is not zero
// per cent — zero per cent says none of the work is done, and there is no
// work. A surface asks this before showing a share.
func (p Progress) Any() bool { return p.Total() > 0 }

// Total is everything counted, tasks and sub-tasks together.
func (p Progress) Total() int { return p.Tasks.Total + p.SubTasks.Total }

// Done is everything counted that is done.
func (p Progress) Done() int { return p.Tasks.Done + p.SubTasks.Done }

// Percent is the share done, rounded to the nearest whole — except at the two
// ends, which are reserved.
//
// 100 is kept for the case where nothing is left: 299 of 300 rounds to 100,
// and 100 reads as "finished" to everybody, while the one thing still undone
// is exactly what somebody needed to see. 0 is kept for the case where nothing
// has been done, for the same reason in reverse — 1 of 300 rounding to 0 hides
// that somebody has started.
func (p Progress) Percent() int {
	total, done := p.Total(), p.Done()
	if total == 0 {
		return 0
	}
	if done == total {
		return 100
	}
	if done == 0 {
		return 0
	}
	// Rounded away from zero, so 12.5 is 13 rather than 12.
	share := (done*200 + total) / (total * 2)
	return min(max(share, 1), 99)
}
