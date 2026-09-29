package engine

import (
	"cairn/internal/domain"
	"errors"
	"path/filepath"
	"slices"
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

func TestAll_hasThirtyDistinctOperations(t *testing.T) {
	if got := TaskAdd.Name(); got != "task_add" {
		t.Errorf("TaskAdd.Name() = %q, want %q", got, "task_add")
	}
	if got := TaskAdd.Line(); got != "task add" {
		t.Errorf("TaskAdd.Line() = %q, want %q", got, "task add")
	}

	ops := All()
	if len(ops) != 30 {
		t.Fatalf("len(All()) = %d, want 30", len(ops))
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

// Archiving takes a task off the board and keeps everything it carries. It is
// the engine that says so, once, for every surface.
func TestArchive_takesATaskOffTheBoardAndKeepsIt(t *testing.T) {
	e := openEngine(t)
	number := factsOf(add(t, e, "Migration Jenkins")).number
	finished := string(domain.Done)
	if _, err := e.Edit(number, Changes{State: &finished}); err != nil {
		t.Fatalf("finishing it: %v", err)
	}

	if err := e.Archive(number); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	found := factsOf(shown(t, e, number))
	if !found.archived {
		t.Error("the task is not archived")
	}
	if found.state != domain.Done {
		t.Errorf("state = %q, want it still done — archiving is not a column", found.state)
	}

	// And it comes back.
	if err := e.Unarchive(number); err != nil {
		t.Fatalf("Unarchive: %v", err)
	}
	if factsOf(shown(t, e, number)).archived {
		t.Error("the task is still archived after being brought back")
	}
}

// Archiving what is already off the board is refused rather than written down
// again.
func TestArchive_refusesWhatIsAlreadyArchived(t *testing.T) {
	e := openEngine(t)
	number := factsOf(add(t, e, "Migration Jenkins")).number
	if err := e.Archive(number); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if err := e.Archive(number); !errors.Is(err, task.ErrAlreadyArchived) {
		t.Errorf("err = %v, want it to say the task is already archived", err)
	}
}

// A listing leaves what is off the board out, and hands it over when it is
// asked for by name.
func TestListing_leavesArchivedOutUntilAskedFor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	away := factsOf(add(t, e, "Migration Jenkins")).number
	kept := factsOf(add(t, e, "Spike Argo")).number
	if err := e.Archive(away); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	numbers := func(in Input) []int {
		answer, err := Do(TaskList, path, in)
		if err != nil {
			t.Fatalf("Do(TaskList): %v", err)
		}
		// Both fields: a listing nobody narrowed answers in Tasks and a
		// narrowed one in Matched, and asking for what is off the board is a
		// narrowing. Reading only the first said the archived ones were gone.
		var out []int
		for _, one := range append(answer.Tasks, answer.Matched...) {
			out = append(out, factsOf(one).number)
		}
		return out
	}
	if got := numbers(Input{}); !slices.Equal(got, []int{kept}) {
		t.Errorf("a listing gave %v, want only what is on the board (%d)", got, kept)
	}
	if got := numbers(Input{Given: map[string]string{"archived": "true"}}); !slices.Equal(got, []int{away}) {
		t.Errorf("asking for archived gave %v, want only what was put away (%d)", got, away)
	}
}

// Archiving takes a number or a narrowing, and says how many it reached.
func TestArchiving_takesANumberOrANarrowing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	done := string(domain.Done)
	for _, title := range []string{"Migration Jenkins", "Dépôt APP"} {
		n := factsOf(add(t, e, title)).number
		if _, err := e.Edit(n, Changes{State: &done}); err != nil {
			t.Fatalf("finishing %q: %v", title, err)
		}
	}
	add(t, e, "Spike Argo")
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	answer, err := Do(TaskArchive, path, Input{Given: map[string]string{"state": done}})
	if err != nil {
		t.Fatalf("Do(TaskArchive): %v", err)
	}
	if answer.Archived == nil || *answer.Archived != 2 {
		t.Fatalf("Archived = %v, want 2", answer.Archived)
	}

	// What it did not reach is still on the board.
	listed, err := Do(TaskList, path, Input{})
	if err != nil {
		t.Fatalf("Do(TaskList): %v", err)
	}
	if len(listed.Tasks) != 1 {
		t.Errorf("the board holds %d tasks, want the one that was not done", len(listed.Tasks))
	}
}

// A narrowing reaching nothing archives nothing and says so.
func TestArchiving_saysWhenItReachedNothing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	add(t, e, "Spike Argo")
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	answer, err := Do(TaskArchive, path, Input{Given: map[string]string{"state": string(domain.Done)}})
	if err != nil {
		t.Fatalf("Do(TaskArchive): %v", err)
	}
	if answer.Archived == nil || *answer.Archived != 0 {
		t.Errorf("Archived = %v, want 0", answer.Archived)
	}
}

// Neither a number nor a narrowing is refused: archiving everything by
// accident is not something to find out afterwards.
func TestArchiving_refusesWithNothingToGoOn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	add(t, e, "Spike Argo")
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := Do(TaskArchive, path, Input{}); err == nil {
		t.Error("archiving with nothing to go on was allowed")
	}
}

// A label is put on everything a narrowing reaches, keeping what each already
// carries.
func TestLabelling_putsOneOnWhatANarrowingReaches(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	done := string(domain.Done)
	first := factsOf(add(t, e, "Migration Jenkins")).number
	if _, err := e.Edit(first, Changes{State: &done, Labels: &[]string{"backend"}}); err != nil {
		t.Fatalf("setting it up: %v", err)
	}
	add(t, e, "Spike Argo")
	// A part of the work the narrowing reaches is left out with it, the way a
	// sweep leaves it out: both read the same world.
	part := factsOf(add(t, e, "Write the tests")).number
	if _, err := e.Edit(part, Changes{Parent: &first, State: &done}); err != nil {
		t.Fatalf("hanging it under: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	answer, err := Do(TaskLabel, path, Input{Given: map[string]string{"label": "sprint-41", "state": done}})
	if err != nil {
		t.Fatalf("Do(TaskLabel): %v", err)
	}
	if answer.Labelled == nil || *answer.Labelled != 1 {
		t.Fatalf("Labelled = %v, want 1", answer.Labelled)
	}

	back, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = back.Close() }()
	var carried labelReader
	found, err := back.Show(first)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	found.Tell(&carried)
	for _, want := range []string{"backend", "sprint-41"} {
		if !slices.Contains(carried.labels, want) {
			t.Errorf("task carries %v, want %q among them", carried.labels, want)
		}
	}
}

// labelReader reads back the labels a task carries.
type labelReader struct {
	task.AsTold
	labels []string
}

func (r *labelReader) Labelled(s string) { r.labels = append(r.labels, s) }

// Archiving one by number, and bringing it back. Both answer with the task as
// it ends up, and both refuse a number no task carries.
func TestArchiving_oneByNumber(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	n := factsOf(add(t, e, "Migration Jenkins")).number
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	number := map[string]string{"number": strconv.Itoa(n)}

	answer, err := Do(TaskArchive, path, Input{Given: number})
	if err != nil {
		t.Fatalf("Do(TaskArchive): %v", err)
	}
	if answer.Task == nil || factsOf(*answer.Task).number != n {
		t.Fatalf("archiving answered %v, want task %d", answer.Task, n)
	}
	if _, err := Do(TaskArchive, path, Input{Given: number}); err == nil {
		t.Error("archiving the same task twice was allowed")
	}

	back, err := Do(TaskUnarchive, path, Input{Given: number})
	if err != nil {
		t.Fatalf("Do(TaskUnarchive): %v", err)
	}
	if back.Task == nil || factsOf(*back.Task).number != n {
		t.Fatalf("bringing it back answered %v, want task %d", back.Task, n)
	}
	if _, err := Do(TaskUnarchive, path, Input{Given: number}); err == nil {
		t.Error("bringing back a task already on the board was allowed")
	}
	if _, err := Do(TaskUnarchive, path, Input{Given: map[string]string{"number": "404"}}); err == nil {
		t.Error("bringing back a task that is not there was allowed")
	}
}
