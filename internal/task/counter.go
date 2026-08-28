package task

import (
	"time"

	"cairn/internal/domain"
)

// Counter is a task's membrane for a tally: it is told every fact a task
// carries and keeps only the one a tally needs — which column it sits in.
// A membrane hears everything, and keeping only what it needs is exactly the
// point: counting what is done is not a question to ask a task, it is a task
// telling its column and a counter keeping score.
type Counter struct {
	tally domain.Tally
}

// NewCounter starts a counter from a tally already under way, so a running
// total can be told one task at a time rather than rebuilt from nothing.
func NewCounter(start domain.Tally) *Counter { return &Counter{tally: start} }

// Numbered is heard and ignored: a tally does not care which task it counts.
func (c *Counter) Numbered(int) {}

// Titled is heard and ignored: a tally does not care what a task is called.
func (c *Counter) Titled(string) {}

// InColumn is the one fact a tally needs: it counts the task, and counts it
// again as done when the column it sits in is Done.
func (c *Counter) InColumn(s domain.State) {
	c.tally.Total++
	if s == domain.Done {
		c.tally.Done++
	}
}

// Carries is heard and ignored: a tally does not care what a task carries.
func (c *Counter) Carries(domain.Priority) {}

// Described is heard and ignored: a tally does not care what a task says.
func (c *Counter) Described(string) {}

// NotDescribed is heard and ignored, the same way Described is.
func (c *Counter) NotDescribed() {}

// Labelled is heard and ignored: a tally does not care what a task is labelled.
func (c *Counter) Labelled(string) {}

// NotLabelled is heard and ignored, the same way Labelled is.
func (c *Counter) NotLabelled() {}

// LabelsTold is heard and ignored: a tally never waited on the labels.
func (c *Counter) LabelsTold() {}

// BlockedBy is heard and ignored: a task waiting on something is still a task
// to be counted, and only the column says whether it is done.
func (c *Counter) BlockedBy(int) {}

// NotBlocked is heard and ignored, the same way BlockedBy is.
func (c *Counter) NotBlocked() {}

// BlockersTold is heard and ignored: a tally never waited on the blockers.
func (c *Counter) BlockersTold() {}

// Because is heard and ignored: why a task is stopped changes no tally.
func (c *Counter) Because(string) {}

// NotBecause is heard and ignored, the same way Because is.
func (c *Counter) NotBecause() {}

// Due is heard and ignored: a tally does not care when a task is due.
func (c *Counter) Due(time.Time) {}

// NotDue is heard and ignored, the same way Due is.
func (c *Counter) NotDue() {}

// Estimated is heard and ignored: a tally does not care how long a task takes.
func (c *Counter) Estimated(time.Duration) {}

// NotEstimated is heard and ignored, the same way Estimated is.
func (c *Counter) NotEstimated() {}

// Spent is heard and ignored: a tally does not care how long a task took.
func (c *Counter) Spent(time.Duration) {}

// NothingSpent is heard and ignored, the same way Spent is.
func (c *Counter) NothingSpent() {}

// Under is heard and ignored: a tally does not care what a task hangs under.
func (c *Counter) Under(int) {}

// UnderNothing is heard and ignored, the same way Under is.
func (c *Counter) UnderNothing() {}

// Serves is heard and ignored: a tally does not care what a task serves.
func (c *Counter) Serves(int) {}

// ServesNothing is heard and ignored, the same way Serves is.
func (c *Counter) ServesNothing() {}

// ObjectivesTold is heard and ignored: a tally never waited on the objectives.
func (c *Counter) ObjectivesTold() {}

// Tally hands back the tally the counter has kept so far.
func (c *Counter) Tally() domain.Tally { return c.tally }
