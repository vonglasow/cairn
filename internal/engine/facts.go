package engine

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
)

type taskFacts struct {
	number, parent      int
	title, description  string
	state               domain.State
	priority            domain.Priority
	labels              []string
	blockers            []int
	objectives          []int
	because             string
	deadline            time.Time
	estimate, timeSpent time.Duration
	archived            bool
}

func (f *taskFacts) Numbered(n int)            { f.number = n }
func (f *taskFacts) Titled(s string)           { f.title = s }
func (f *taskFacts) InColumn(s domain.State)   { f.state = s }
func (f *taskFacts) Archived()                 { f.archived = true }
func (f *taskFacts) NotArchived()              { f.archived = false }
func (f *taskFacts) Carries(p domain.Priority) { f.priority = p }
func (f *taskFacts) Described(s string)        { f.description = s }
func (f *taskFacts) NotDescribed()             { f.description = "" }
func (f *taskFacts) Labelled(l string)         { f.labels = append(f.labels, l) }
func (f *taskFacts) NotLabelled()              { f.labels = nil }
func (f *taskFacts) LabelsTold()               {}
func (f *taskFacts) BlockedBy(n int)           { f.blockers = append(f.blockers, n) }
func (f *taskFacts) NotBlocked()               { f.blockers = nil }
func (f *taskFacts) BlockersTold()             {}
func (f *taskFacts) Because(s string)          { f.because = s }
func (f *taskFacts) NotBecause()               { f.because = "" }
func (f *taskFacts) Due(t time.Time)           { f.deadline = t }
func (f *taskFacts) NotDue()                   { f.deadline = time.Time{} }
func (f *taskFacts) Estimated(d time.Duration) { f.estimate = d }
func (f *taskFacts) NotEstimated()             { f.estimate = 0 }
func (f *taskFacts) Spent(d time.Duration)     { f.timeSpent = d }
func (f *taskFacts) NothingSpent()             { f.timeSpent = 0 }
func (f *taskFacts) Under(n int)               { f.parent = n }
func (f *taskFacts) UnderNothing()             { f.parent = task.NoParent }
func (f *taskFacts) Serves(n int)              { f.objectives = append(f.objectives, n) }
func (f *taskFacts) ServesNothing()            { f.objectives = nil }
func (f *taskFacts) ObjectivesTold()           {}

func factsOf(t task.Task) taskFacts {
	var f taskFacts
	t.Tell(&f)
	return f
}
