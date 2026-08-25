package domain

// Tally is how many of one kind of thing there are, and how many of those are
// done. The tags are the wire as much as the fields are: a machine reads one
// nested under a ProgressShape.
type Tally struct {
	Total int `json:"total"`
	Done  int `json:"done"`
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

// ProgressShape is how far along something is, for a machine. It is a
// pointer on the way out and null when nothing is under it: zeroes would say
// none of the work is done, where the truth is that there is no work.
//
// The share is published rather than left to be divided, so a client and a
// person reading the same objective never disagree about the rounding.
type ProgressShape struct {
	Tasks    *Tally `json:"tasks,omitempty"`
	SubTasks *Tally `json:"sub_tasks,omitempty"`
	Percent  int    `json:"percent"`
}

// OfProgress turns a progress into what crosses, or nil when there is nothing
// under it. A tally of nothing is left out entirely rather than written as
// zeroes: a task has no tasks under it, and saying "0 of 0 tasks" invents a
// kind of thing it cannot have.
func OfProgress(p Progress) *ProgressShape {
	if !p.Any() {
		return nil
	}
	out := ProgressShape{Percent: p.Percent()}
	if p.Tasks.Total > 0 {
		out.Tasks = &Tally{Total: p.Tasks.Total, Done: p.Tasks.Done}
	}
	if p.SubTasks.Total > 0 {
		out.SubTasks = &Tally{Total: p.SubTasks.Total, Done: p.SubTasks.Done}
	}
	return &out
}
