package task

import (
	"errors"
	"testing"
)

func TestMayUnarchive_refusesWhatIsOnTheBoard(t *testing.T) {
	var here AsTold
	here.Numbered(1)
	here.Titled("Migration Jenkins")
	if err := here.Task().MayUnarchive(); !errors.Is(err, ErrNotArchived) {
		t.Errorf("MayUnarchive = %v, want it to say the task is on the board", err)
	}
	var away AsTold
	away.Numbered(2)
	away.Titled("Spike Argo")
	away.Archived()
	if err := away.Task().MayUnarchive(); err != nil {
		t.Errorf("MayUnarchive on an archived task = %v, want it allowed", err)
	}
}
