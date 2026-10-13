package task

import (
	"time"

	"cairn/internal/domain"
)

// AsTold builds a task one fact at a time.
type AsTold struct{ c Carried }

// From starts a builder on what a task already carries.
func From(t Task) AsTold { return AsTold{c: t.Carried()} }

// Archived is told the task is off the board.
func (a *AsTold) Archived() { a.c.Archived = true }

// Numbered is told the task's number.
func (a *AsTold) Numbered(n int) { a.c.Number = n }

// Titled is told the task's title.
func (a *AsTold) Titled(s string) { a.c.Title = s }

// InColumn is told the column the task sits in.
func (a *AsTold) InColumn(s domain.State) { a.c.State = s }

// Carries is told the priority the task carries.
func (a *AsTold) Carries(p domain.Priority) { a.c.Priority = p }

// Described is told the task's description.
func (a *AsTold) Described(s string) { a.c.Description = s }

// Labelled is told one more label the task carries.
func (a *AsTold) Labelled(s string) { a.c.Labels = append(a.c.Labels, s) }

// BlockedBy adds a task to what is in this one's way.
func (a *AsTold) BlockedBy(n int) { a.c.Blockers = append(a.c.Blockers, n) }

// Because names in words what is in its way.
func (a *AsTold) Because(s string) { a.c.Because = s }

// Due is told the task's deadline.
func (a *AsTold) Due(t time.Time) { a.c.Deadline = t }

// Estimated is told the task's estimate.
func (a *AsTold) Estimated(d time.Duration) { a.c.Estimate = d }

// Spent is told the time already spent on the task.
func (a *AsTold) Spent(d time.Duration) { a.c.Spent = d }

// Under is told the task the task hangs under.
func (a *AsTold) Under(n int) { a.c.Parent = n }

// Serves adds an objective the task serves.
func (a *AsTold) Serves(n int) { a.c.Objectives = append(a.c.Objectives, n) }

// Task hands back the task AsTold was told to build.
func (a *AsTold) Task() Task { return Made(a.c) }
