package task

import (
	"testing"

	"cairn/internal/domain"
)

func TestCounter_countsOnlyWhatIsDone(t *testing.T) {
	c := NewCounter(domain.Tally{})
	Task{number: 1, state: domain.Done, title: "Write the spec"}.Tell(c)
	Task{number: 2, state: domain.Done, title: "Write the tests"}.Tell(c)
	Task{number: 3, state: domain.Todo, title: "Ship it"}.Tell(c)

	if got := c.Tally(); got != (domain.Tally{Total: 3, Done: 2}) {
		t.Errorf("Tally() = %+v, want 2 of 3 done", got)
	}
}
