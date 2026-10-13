package task

import (
	"testing"

	"cairn/internal/domain"
)

func TestCount_countsOnlyWhatIsDone(t *testing.T) {
	var got domain.Tally
	for _, one := range []Task{{number: 1, state: domain.Done}, {number: 2, state: domain.Done}, {number: 3, state: domain.Todo}} {
		got = Count(got, one)
	}
	if got != (domain.Tally{Total: 3, Done: 2}) {
		t.Errorf("Count = %+v, want 2 of 3 done", got)
	}
}
