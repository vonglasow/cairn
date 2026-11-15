package task

import (
	"cairn/internal/engine/item"
	"testing"
)

func TestCount_countsOnlyWhatIsDone(t *testing.T) {
	var got item.Tally
	for _, one := range []Task{{number: 1, state: item.Done}, {number: 2, state: item.Done}, {number: 3, state: item.Todo}} {
		got = Count(got, one)
	}
	if got != (item.Tally{Total: 3, Done: 2}) {
		t.Errorf("Count = %+v, want 2 of 3 done", got)
	}
}
