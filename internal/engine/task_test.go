package engine

import (
	"cairn/internal/domain"
	"errors"
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

func TestTaskSearch_listsWhatCarriesEveryNamedThing(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "Write the spec"}}); err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}
	if _, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "Ship it"}}); err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}

	answer, err := Do(TaskSearch, path, Input{Listed: map[string][]string{"named": {"spec"}}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(answer.Matched) != 1 || factsOf(answer.Matched[0]).title != "Write the spec" {
		t.Fatalf("answer.Matched = %v, want one task titled %q", answer.Matched, "Write the spec")
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

func TestAll_hasTwentySixDistinctOperations(t *testing.T) {
	if got := TaskAdd.Name(); got != "task_add" {
		t.Errorf("TaskAdd.Name() = %q, want %q", got, "task_add")
	}
	if got := TaskAdd.Line(); got != "task add" {
		t.Errorf("TaskAdd.Line() = %q, want %q", got, "task add")
	}

	ops := All()
	if len(ops) != 27 {
		t.Fatalf("len(All()) = %d, want 27", len(ops))
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

// A parent is not finished while something under it is open, and the engine is
// where that is enforced once for all three surfaces.
func TestEdit_refusesToFinishAParentWithAnOpenSubTask(t *testing.T) {
	e := openEngine(t)
	parent := factsOf(add(t, e, "Ship it")).number
	under := factsOf(add(t, e, "Write the spec")).number
	if _, err := e.Edit(under, Changes{Parent: &parent}); err != nil {
		t.Fatalf("hanging it under: %v", err)
	}

	finished := string(domain.Done)
	if _, err := e.Edit(parent, Changes{State: &finished}); !errors.Is(err, task.ErrSubTaskStillOpen) {
		t.Fatalf("err = %v, want it to say something under it is open", err)
	}

	// Once the sub-task is done, the parent may be.
	if _, err := e.Edit(under, Changes{State: &finished}); err != nil {
		t.Fatalf("finishing the sub-task: %v", err)
	}
	if _, err := e.Edit(parent, Changes{State: &finished}); err != nil {
		t.Errorf("finishing the parent: %v", err)
	}
}

// A parent already done may still be edited for anything else. Asking the rule
// of every edit would make a finished parent fail for a reason nobody caused
// today — the sub-task was open when it was finished, or it was reopened
// since, and neither is this edit's doing.
func TestEdit_leavesAFinishedParentEditableForAnythingElse(t *testing.T) {
	e := openEngine(t)
	parent := factsOf(add(t, e, "Ship it")).number
	finished := string(domain.Done)
	if _, err := e.Edit(parent, Changes{State: &finished}); err != nil {
		t.Fatalf("finishing it: %v", err)
	}
	under := factsOf(add(t, e, "Write the spec")).number
	if _, err := e.Edit(under, Changes{Parent: &parent}); err != nil {
		t.Fatalf("hanging an open one under it: %v", err)
	}

	title := "Ship it properly"
	if _, err := e.Edit(parent, Changes{Title: &title}); err != nil {
		t.Errorf("editing a finished parent: %v", err)
	}
}

// A sub-task takes done or todo, and hanging a task under another settles the
// column it was carrying: done stays done, anything else becomes todo.
func TestEdit_holdsASubTaskToDoneOrNot(t *testing.T) {
	e := openEngine(t)
	parent := factsOf(add(t, e, "Ship it")).number
	under := factsOf(add(t, e, "Write the spec")).number

	working := string(domain.Current)
	if _, err := e.Edit(under, Changes{State: &working}); err != nil {
		t.Fatalf("putting it in current while it hangs under nothing: %v", err)
	}
	// Hung under a parent, what it was doing on its own stops being true.
	if _, err := e.Edit(under, Changes{Parent: &parent}); err != nil {
		t.Fatalf("hanging it under: %v", err)
	}
	if got := factsOf(shown(t, e, under)).state; got != domain.Todo {
		t.Errorf("state = %q, want it settled to todo", got)
	}
	// And it will not go back.
	if _, err := e.Edit(under, Changes{State: &working}); !errors.Is(err, task.ErrSubTaskHasNoColumn) {
		t.Errorf("err = %v, want it to say a sub-task is done or it is not", err)
	}
	finished := string(domain.Done)
	if _, err := e.Edit(under, Changes{State: &finished}); err != nil {
		t.Errorf("finishing a sub-task: %v", err)
	}
}

// A task already done keeps being done when it is hung under one: finished is
// the one thing a part was doing on its own that stays true.
func TestEdit_keepsAFinishedTaskDoneWhenItIsHungUnderOne(t *testing.T) {
	e := openEngine(t)
	parent := factsOf(add(t, e, "Ship it")).number
	under := factsOf(add(t, e, "Write the spec")).number
	finished := string(domain.Done)
	if _, err := e.Edit(under, Changes{State: &finished}); err != nil {
		t.Fatalf("finishing it: %v", err)
	}
	if _, err := e.Edit(under, Changes{Parent: &parent}); err != nil {
		t.Fatalf("hanging it under: %v", err)
	}
	if got := factsOf(shown(t, e, under)).state; got != domain.Done {
		t.Errorf("state = %q, want it still done", got)
	}
}

// shown reads a task back the way anything else does.
func shown(t *testing.T, e *Engine, number int) task.Task {
	t.Helper()
	found, err := e.Show(number)
	if err != nil {
		t.Fatalf("Show(%d): %v", number, err)
	}
	return found
}
