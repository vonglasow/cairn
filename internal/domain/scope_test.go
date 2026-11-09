package domain

import (
	"errors"
	"testing"
)

func TestALabelsScopeIsOneOfThree(t *testing.T) {
	if got, err := ParseScope(" Objectives "); err != nil || got != ForObjectives {
		t.Errorf("Objectives reads %q, %v", got, err)
	}
	if got, err := ParseScope(""); err != nil || got != ForBoth {
		t.Errorf("nothing reads %q, %v; want both", got, err)
	}
	if _, err := ParseScope("everything"); !errors.Is(err, ErrNoSuchScope) {
		t.Errorf("everything reads %v", err)
	}
}

func TestAScopeAdmitsWhatItIsFor(t *testing.T) {
	cases := []struct {
		scope             Scope
		tasks, objectives bool
	}{
		{ForBoth, true, true},
		{ForTasks, true, false},
		{ForObjectives, false, true},
	}
	for _, c := range cases {
		if c.scope.Admits(OnTasks) != c.tasks || c.scope.Admits(OnObjectives) != c.objectives {
			t.Errorf("%s admits tasks %v and objectives %v", c.scope, c.scope.Admits(OnTasks), c.scope.Admits(OnObjectives))
		}
	}
}

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
