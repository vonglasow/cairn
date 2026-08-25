package engine

import (
	"strconv"
	"strings"
	"testing"

	"cairn/internal/task"
)

func TestTaskList_listsEveryTask(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "one"}}); err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}
	if _, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "two"}}); err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}

	answer, err := Do(TaskList, path, Input{})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(answer.Tasks) != 2 {
		t.Fatalf("len(answer.Tasks) = %d, want 2", len(answer.Tasks))
	}
}

func TestTaskShow_showsATaskItsChildrenAndProgress(t *testing.T) {
	path := storeAt(t)
	parent, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "parent"}})
	if err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}
	number := strconv.Itoa(factsOf(*parent.Task).number)
	if _, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "child", "parent": number}}); err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}

	answer, err := Do(TaskShow, path, Input{Given: map[string]string{"number": number}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Task == nil || factsOf(*answer.Task).title != "parent" {
		t.Fatalf("answer.Task = %v, want the parent task", answer.Task)
	}
	if len(answer.Under) != 1 || factsOf(answer.Under[0]).title != "child" {
		t.Fatalf("answer.Under = %v, want one task titled %q", answer.Under, "child")
	}
	if answer.Progress.SubTasks.Total != 1 {
		t.Fatalf("answer.Progress.SubTasks.Total = %d, want 1", answer.Progress.SubTasks.Total)
	}
}

func TestTaskEdit_editsATask(t *testing.T) {
	path := storeAt(t)
	parent, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "parent"}})
	if err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}
	child, err := Do(TaskAdd, path, Input{Given: map[string]string{
		"title": "child", "parent": strconv.Itoa(factsOf(*parent.Task).number),
	}})
	if err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}
	number := strconv.Itoa(factsOf(*child.Task).number)

	// An empty --parent sets a task free — the same translation task_add uses.
	answer, err := Do(TaskEdit, path, Input{Given: map[string]string{"number": number, "parent": ""}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Task == nil || factsOf(*answer.Task).parent != task.NoParent {
		t.Fatalf("answer.Task.Parent = %v, want %d: an empty --parent sets a task free", answer.Task, task.NoParent)
	}

	if _, err := Do(TaskEdit, path, Input{Given: map[string]string{"number": number}}); err == nil ||
		!strings.Contains(err.Error(), "--title") {
		t.Fatalf("err = %v, want a refusal naming the flags to change", err)
	}
}

func TestTaskDelete_deletesATask(t *testing.T) {
	path := storeAt(t)
	added, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "gone soon"}})
	if err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}

	answer, err := Do(TaskDelete, path, Input{Given: map[string]string{"number": strconv.Itoa(factsOf(*added.Task).number)}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Deleted == nil || *answer.Deleted != factsOf(*added.Task).number {
		t.Fatalf("answer.Deleted = %v, want %d", answer.Deleted, factsOf(*added.Task).number)
	}

	list, err := Do(TaskList, path, Input{})
	if err != nil {
		t.Fatalf("Do(TaskList): %v", err)
	}
	if len(list.Tasks) != 0 {
		t.Fatalf("len(list.Tasks) = %d, want 0: the task was not deleted", len(list.Tasks))
	}
}

func TestTaskAdd_addsATask(t *testing.T) {
	path := storeAt(t)
	in := Input{Given: map[string]string{"title": "write the report"}}

	answer, err := Do(TaskAdd, path, in)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Task == nil || factsOf(*answer.Task).title != "write the report" {
		t.Fatalf("answer.Task = %v, want a task titled %q", answer.Task, "write the report")
	}
}

func TestAll_hasNineteenDistinctOperations(t *testing.T) {
	if got := TaskAdd.Name(); got != "task_add" {
		t.Errorf("TaskAdd.Name() = %q, want %q", got, "task_add")
	}
	if got := TaskAdd.Line(); got != "task add" {
		t.Errorf("TaskAdd.Line() = %q, want %q", got, "task add")
	}

	ops := All()
	if len(ops) != 19 {
		t.Fatalf("len(All()) = %d, want 19", len(ops))
	}
	seen := map[string]bool{}
	for _, op := range ops {
		name := op.Name()
		if seen[name] {
			t.Errorf("%s appears more than once", name)
		}
		seen[name] = true
		if op.Level == 0 {
			t.Errorf("%s carries no Level", name)
		}
		if op.Needs == 0 {
			t.Errorf("%s carries no Needs", name)
		}
		if op.Short == "" {
			t.Errorf("%s carries no Short", name)
		}
		if op.Does == "" {
			t.Errorf("%s carries no Does", name)
		}
	}
}
