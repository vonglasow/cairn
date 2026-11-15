package engine

import (
	"cairn/internal/engine/item"
	"errors"
	"testing"

	"cairn/internal/engine/task"
)

func movedBehind(n, after string, more ...string) Input {
	in := given(append([]string{"number", n, "after", after}, more...)...)
	return in
}

func TestTaskMove_saysNothingCrossedWhenTheColumnStays(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	add(t, e, "Second")
	a, err := e.Ask(TaskMove, movedBehind("1", "2"))
	if err != nil || a.Crossed != nil {
		t.Errorf("move = %+v, %v; want nothing crossed", a.Crossed, err)
	}
}

func TestTaskMove_refusesAColumnNamedAgainstTheOneItGoesBehind(t *testing.T) {
	e := openEngine(t)
	add(t, e, "In the backlog")
	if _, err := e.add("Already to do", Changes{State: of("todo")}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Ask(TaskMove, movedBehind("1", "2", "state", "current")); !errors.Is(err, task.ErrOtherColumn) {
		t.Errorf("err = %v, want ErrOtherColumn", err)
	}
	if kept, _ := e.show(1); kept.Carried().State != item.Backlog {
		t.Errorf("task 1 is in %s, want it left in the backlog", kept.Carried().State)
	}
}

func TestTaskMove_refusesBehindATaskThatIsNotThere(t *testing.T) {
	e := openEngine(t)
	add(t, e, "In the backlog")
	if _, err := e.Ask(TaskMove, movedBehind("1", "9")); err == nil {
		t.Error("a move behind a task that is not there went through")
	}
}
