package engine

import (
	"errors"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/task"
)

func movedBehind(n, after string, more ...string) Input {
	in := given(append([]string{"number", n, "after", after}, more...)...)
	return in
}

func TestTaskMove_takesATaskIntoTheColumnOfTheOneItGoesBehind(t *testing.T) {
	e := openEngine(t)
	add(t, e, "In the backlog")
	if _, err := e.add("Already to do", Changes{State: of("todo")}); err != nil {
		t.Fatal(err)
	}
	a, err := e.Ask(TaskMove, movedBehind("1", "2"))
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	moved, _ := e.show(1)
	if factsOf(moved).State != domain.Todo {
		t.Errorf("task 1 is in %s, want todo", factsOf(moved).State)
	}
	if a.Crossed == nil || *a.Crossed != (Crossed{Task: 1, From: "backlog", To: "todo"}) {
		t.Errorf("Crossed = %+v, want task 1 from backlog to todo", a.Crossed)
	}
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
	if _, err := e.Ask(TaskMove, movedBehind("1", "2", "state", "current")); !errors.Is(err, ErrOtherColumn) {
		t.Errorf("err = %v, want ErrOtherColumn", err)
	}
	if kept, _ := e.show(1); factsOf(kept).State != domain.Backlog {
		t.Errorf("task 1 is in %s, want it left in the backlog", factsOf(kept).State)
	}
}

func TestTaskMove_refusesBehindATaskThatIsNotThere(t *testing.T) {
	e := openEngine(t)
	add(t, e, "In the backlog")
	if _, err := e.Ask(TaskMove, movedBehind("1", "9")); err == nil {
		t.Error("a move behind a task that is not there went through")
	}
}

func TestRenderText_saysATaskCrossedIntoAnotherColumn(t *testing.T) {
	tasks := []task.Task{builtTask(t, 1, "First")}
	got := RenderText(TaskMove, Answer{Tasks: tasks, Crossed: &Crossed{Task: 1, From: "backlog", To: "todo"}})
	if !strings.HasPrefix(got, "task 1 moved from backlog to todo\n") {
		t.Errorf("the answer does not open on the column it crossed:\n%s", got)
	}
}

func TestOf_carriesTheColumnATaskCrossed(t *testing.T) {
	crossed := &Crossed{Task: 1, From: "backlog", To: "todo"}
	if said, _ := Of(TaskMove, Answer{Crossed: crossed}).(Tasks); said.Crossed != crossed {
		t.Errorf("Of = %+v, want the crossing carried", said.Crossed)
	}
}
