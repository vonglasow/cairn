package task

import (
	"time"

	"cairn/internal/domain"
)

// AsTold builds a task from what it is told.
type AsTold struct {
	task       Task
	labels     []string
	blockers   []int
	objectives []int
}

// Archived is told the task is off the board.
func (a *AsTold) Archived() { a.task.archived = true }

// NotArchived is told it is not.
func (a *AsTold) NotArchived() { a.task.archived = false }

// Numbered is told the task's number.
func (a *AsTold) Numbered(n int) { a.task.number = n }

// Titled is told the task's title.
func (a *AsTold) Titled(s string) { a.task.title = s }

// InColumn is told the column the task sits in.
func (a *AsTold) InColumn(s domain.State) { a.task.state = s }

// Carries is told the priority the task carries.
func (a *AsTold) Carries(p domain.Priority) { a.task.priority = p }

// Described is told the task's description.
func (a *AsTold) Described(s string) { a.task.description = s }

// NotDescribed is told the task carries no description.
func (a *AsTold) NotDescribed() { a.task.description = "" }

// Labelled is told one more label the task carries.
func (a *AsTold) Labelled(s string) { a.labels = append(a.labels, s) }

// NotLabelled is told the task carries no labels at all.
func (a *AsTold) NotLabelled() { a.labels = nil }

// LabelsTold is told there are no more labels coming, and closes the set.
func (a *AsTold) LabelsTold() { a.task.labels = a.labels }

// BlockedBy adds a task to what is in this one's way.
func (a *AsTold) BlockedBy(n int) { a.blockers = append(a.blockers, n) }

// NotBlocked says nothing is in its way.
func (a *AsTold) NotBlocked() { a.blockers = nil }

// BlockersTold closes the set.
func (a *AsTold) BlockersTold() { a.task.blockers = a.blockers }

// Because names in words what is in its way and is not a task cairn holds.
func (a *AsTold) Because(s string) { a.task.because = s }

// NotBecause says nothing outside cairn was named.
func (a *AsTold) NotBecause() { a.task.because = "" }

// Due is told the task's deadline.
func (a *AsTold) Due(t time.Time) { a.task.deadline = t }

// NotDue is told the task carries no deadline.
func (a *AsTold) NotDue() { a.task.deadline = time.Time{} }

// Estimated is told the task's estimate.
func (a *AsTold) Estimated(d time.Duration) { a.task.estimate = d }

// NotEstimated is told the task carries no estimate.
func (a *AsTold) NotEstimated() { a.task.estimate = 0 }

// Spent is told the time already spent on the task.
func (a *AsTold) Spent(d time.Duration) { a.task.timeSpent = d }

// NothingSpent is told no time has been spent on the task.
func (a *AsTold) NothingSpent() { a.task.timeSpent = 0 }

// Under is told the task the task hangs under.
func (a *AsTold) Under(n int) { a.task.parent = n }

// UnderNothing is told the task hangs under nothing.
func (a *AsTold) UnderNothing() { a.task.parent = NoParent }

// Serves adds an objective the task serves.
func (a *AsTold) Serves(n int) { a.objectives = append(a.objectives, n) }

// ServesNothing is told the task serves no objective.
func (a *AsTold) ServesNothing() { a.objectives = nil }

// ObjectivesTold closes the set, the way LabelsTold and BlockersTold do.
func (a *AsTold) ObjectivesTold() { a.task.objectives = a.objectives }

// Task hands back the task AsTold was told to build.
func (a *AsTold) Task() Task { return a.task }
