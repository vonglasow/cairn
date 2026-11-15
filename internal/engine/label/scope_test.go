package label

import (
	"errors"
	"testing"
)

func TestAScopeRefusesALabelOutsideIt(t *testing.T) {
	err := ForObjectives.Refuses("area", OnTasks)
	if !errors.Is(err, ErrOutOfScope) || err.Error() != `"area" is a label for objectives only, not for a task` {
		t.Errorf("area on a task reads %v", err)
	}
	if err := ForTasks.Refuses("chore", OnObjectives); err == nil || err.Error() != `"chore" is a label for tasks only, not for an objective` {
		t.Errorf("chore on an objective reads %v", err)
	}
	if err := ForBoth.Refuses("doc", OnTasks); err != nil {
		t.Errorf("doc on a task reads %v", err)
	}
}
