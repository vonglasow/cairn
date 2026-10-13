package domain

// Tally is how many of one kind of thing there are, and how many of those are done.
type Tally struct {
	Total int `json:"total"`
	Done  int `json:"done"`
}

// Progress is how far along something is.
type Progress struct {
	Tasks    Tally
	SubTasks Tally
}

// Any says whether anything is under it at all.
func (p Progress) Any() bool { return p.Total() > 0 }

// Total is everything counted, tasks and sub-tasks together.
func (p Progress) Total() int { return p.Tasks.Total + p.SubTasks.Total }

// Done is everything counted that is done.
func (p Progress) Done() int { return p.Tasks.Done + p.SubTasks.Done }

// Percent is the share done, rounded to the nearest whole — except at the two ends, which are reserved.
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
	share := (done*200 + total) / (total * 2)
	return min(max(share, 1), 99)
}

// ProgressShape is how far along something is, for a machine.
type ProgressShape struct {
	Tasks    *Tally `json:"tasks,omitempty"`
	SubTasks *Tally `json:"sub_tasks,omitempty"`
	Percent  int    `json:"percent"`
}

// OfProgress turns a progress into what crosses, or nil when there is nothing under it.
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
