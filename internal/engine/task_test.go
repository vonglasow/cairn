package engine

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"cairn/internal/domain"

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

func TestAll_namesEachOperationOnce(t *testing.T) {
	if got := TaskAdd.Name(); got != "task_add" {
		t.Errorf("TaskAdd.Name() = %q, want %q", got, "task_add")
	}
	if got := TaskAdd.Line(); got != "task add" {
		t.Errorf("TaskAdd.Line() = %q, want %q", got, "task add")
	}

	ops := All()
	if len(ops) != 35 {
		t.Fatalf("len(All()) = %d, want 35", len(ops))
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
	if _, err := e.edit(under, Changes{Parent: &parent}); err != nil {
		t.Fatalf("hanging it under: %v", err)
	}

	finished := string(domain.Done)
	if _, err := e.edit(parent, Changes{State: &finished}); !errors.Is(err, task.ErrSubTaskStillOpen) {
		t.Fatalf("err = %v, want it to say something under it is open", err)
	}

	// Once the sub-task is done, the parent may be.
	if _, err := e.edit(under, Changes{State: &finished}); err != nil {
		t.Fatalf("finishing the sub-task: %v", err)
	}
	if _, err := e.edit(parent, Changes{State: &finished}); err != nil {
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
	if _, err := e.edit(parent, Changes{State: &finished}); err != nil {
		t.Fatalf("finishing it: %v", err)
	}
	under := factsOf(add(t, e, "Write the spec")).number
	if _, err := e.edit(under, Changes{Parent: &parent}); err != nil {
		t.Fatalf("hanging an open one under it: %v", err)
	}

	title := "Ship it properly"
	if _, err := e.edit(parent, Changes{Title: &title}); err != nil {
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
	if _, err := e.edit(under, Changes{State: &working}); err != nil {
		t.Fatalf("putting it in current while it hangs under nothing: %v", err)
	}
	// Hung under a parent, what it was doing on its own stops being true.
	if _, err := e.edit(under, Changes{Parent: &parent}); err != nil {
		t.Fatalf("hanging it under: %v", err)
	}
	if got := factsOf(shown(t, e, under)).state; got != domain.Todo {
		t.Errorf("state = %q, want it settled to todo", got)
	}
	// And it will not go back.
	if _, err := e.edit(under, Changes{State: &working}); !errors.Is(err, task.ErrSubTaskHasNoColumn) {
		t.Errorf("err = %v, want it to say a sub-task is done or it is not", err)
	}
	finished := string(domain.Done)
	if _, err := e.edit(under, Changes{State: &finished}); err != nil {
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
	if _, err := e.edit(under, Changes{State: &finished}); err != nil {
		t.Fatalf("finishing it: %v", err)
	}
	if _, err := e.edit(under, Changes{Parent: &parent}); err != nil {
		t.Fatalf("hanging it under: %v", err)
	}
	if got := factsOf(shown(t, e, under)).state; got != domain.Done {
		t.Errorf("state = %q, want it still done", got)
	}
}

// shown reads a task back the way anything else does.
func shown(t *testing.T, e *Engine, number int) task.Task {
	t.Helper()
	found, err := e.show(number)
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
	if _, err := e.edit(number, Changes{State: &finished}); err != nil {
		t.Fatalf("finishing it: %v", err)
	}

	if err := e.archive(number); err != nil {
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
	if err := e.unarchive(number); err != nil {
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
	if err := e.archive(number); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if err := e.archive(number); !errors.Is(err, task.ErrAlreadyArchived) {
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
	if err := e.archive(away); err != nil {
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
		if _, err := e.edit(n, Changes{State: &done}); err != nil {
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
	if _, err := e.edit(first, Changes{State: &done, Labels: &[]string{"backend"}}); err != nil {
		t.Fatalf("setting it up: %v", err)
	}
	add(t, e, "Spike Argo")
	// A part of the work the narrowing reaches is left out with it, the way a
	// sweep leaves it out: both read the same world.
	part := factsOf(add(t, e, "Write the tests")).number
	if _, err := e.edit(part, Changes{Parent: &first, State: &done}); err != nil {
		t.Fatalf("hanging it under: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	answer, err := Do(TaskLabel, path, Input{Given: map[string]string{"put": "sprint-41", "state": done}})
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
	found, err := back.show(first)
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

// Labelling keeps by everything a listing keeps by, lists included. The labels
// and the objectives a narrowing names travel apart from the rest, and they
// used to be dropped on the way: `--objective 3` labelled every task the
// other filters reached, without a word.
func TestTaskLabel_keepsByWhatItWasListed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	backend := factsOf(add(t, e, "Migration Jenkins")).number
	if _, err := e.edit(backend, Changes{Labels: &[]string{"backend"}}); err != nil {
		t.Fatalf("setting it up: %v", err)
	}
	add(t, e, "Spike Argo")
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	answer, err := Do(TaskLabel, path, Input{
		Given:  map[string]string{"put": "sprint-41"},
		Listed: map[string][]string{"label": {"backend"}},
	})
	if err != nil {
		t.Fatalf("Do(TaskLabel): %v", err)
	}
	if answer.Labelled == nil || *answer.Labelled != 1 {
		t.Errorf("Labelled = %v, want 1: only the task labelled backend", answer.Labelled)
	}
}

// An operation that took two arguments under one name could only ever be
// handed one of them by a caller that names its arguments: an agent's call is
// a JSON object, and an object holds a key once.
func TestAll_noOperationTakesTwoArgumentsUnderOneName(t *testing.T) {
	for _, op := range All() {
		seen := map[string]bool{}
		for _, arg := range op.Takes {
			if seen[arg.Name] {
				t.Errorf("%s takes %q twice", op.Name(), arg.Name)
			}
			seen[arg.Name] = true
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

// A search can be narrowed by what a listing narrows by. The one place the two
// vocabularies meet: what is looked for is free text and what narrows it is a
// closed set, and a word common enough to be worth searching is a word that
// answers the whole board without one.
func TestSearching_canBeNarrowed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	done := string(domain.Done)
	kept := factsOf(add(t, e, "Write the spec")).number
	gone := factsOf(add(t, e, "Rewrite the spec")).number
	if _, err := e.edit(gone, Changes{State: &done}); err != nil {
		t.Fatalf("finishing it: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	looking := map[string][]string{"named": {"spec"}}
	answer, err := Do(TaskSearch, path, Input{Listed: looking, Given: map[string]string{"state": done}})
	if err != nil {
		t.Fatalf("Do(TaskSearch): %v", err)
	}
	if len(answer.Matched) != 1 || factsOf(answer.Matched[0]).number != gone {
		t.Fatalf("narrowed to done, the search answered %d tasks, want only %d", len(answer.Matched), gone)
	}
	// And told nothing to narrow by, it still answers the whole board.
	all, err := Do(TaskSearch, path, Input{Listed: looking})
	if err != nil {
		t.Fatalf("Do(TaskSearch): %v", err)
	}
	if len(all.Matched) != 2 {
		t.Errorf("with nothing to narrow by, the search answered %d tasks, want 2 including %d",
			len(all.Matched), kept)
	}
}

// TestTaskShow_saysWhereItsContextLives: showing a task answers the directory
// its outside context lives in, composed from the notes the surface found and
// the task's number — and touches nothing on disk to say so. A caller that
// never said where the notes are is told nothing about them.
func TestTaskShow_saysWhereItsContextLives(t *testing.T) {
	path := storeAt(t)
	notes := filepath.Join(t.TempDir(), "vault")
	if _, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "one"}}); err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}
	at := Places{Store: path, Notes: notes}
	answer, err := DoAt(TaskShow, at, Input{Given: map[string]string{"number": "1"}}, ByHand)
	if err != nil {
		t.Fatalf("DoAt(TaskShow): %v", err)
	}
	if want := filepath.Join(notes, "1"); answer.Context != want {
		t.Errorf("Context = %q, want %q", answer.Context, want)
	}
	if said := Of(TaskShow, answer).(Task); said.Context != answer.Context {
		t.Errorf("the shape says %q, the answer %q", said.Context, answer.Context)
	}
	if !strings.Contains(RenderText(TaskShow, answer), "context  "+answer.Context) {
		t.Errorf("the card does not say where:\n%s", RenderText(TaskShow, answer))
	}
	if _, err := os.Stat(notes); err == nil {
		t.Errorf("showing a task made %s", notes)
	}

	held, err := OpenAt(at)
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer func() { _ = held.Close() }()
	asked, err := held.Ask(TaskShow, Input{Given: map[string]string{"number": "1"}})
	if err != nil || asked.Context != answer.Context {
		t.Errorf("an engine held open says %q (%v), want %q", asked.Context, err, answer.Context)
	}

	unsaid, err := Do(TaskShow, path, Input{Given: map[string]string{"number": "1"}})
	if err != nil {
		t.Fatalf("Do(TaskShow): %v", err)
	}
	if unsaid.Context != "" || strings.Contains(RenderText(TaskShow, unsaid), "context") {
		t.Errorf("a caller that named no notes is told %q", unsaid.Context)
	}
}

// A listing says how far along each task is and where its context lives, so a
// board drawing every card reads them in the one answer rather than showing
// each task in turn to learn two facts about it.
func TestTaskList_saysHowFarAlongEachTaskIsAndWhereItsContextLives(t *testing.T) {
	path := storeAt(t)
	for _, step := range []Input{
		given("title", "Ship it"),
		given("title", "Write it", "parent", "1", "state", "done"),
		given("title", "Read it", "parent", "1"),
		given("title", "Alone"),
	} {
		if _, err := Do(TaskAdd, path, step); err != nil {
			t.Fatal(err)
		}
	}
	answer, err := DoAt(TaskList, Places{Store: path, Notes: "/notes"}, given("with-archived", "true"), ByHand)
	if err != nil {
		t.Fatal(err)
	}
	if p := answer.Along[1]; p.SubTasks.Total != 2 || p.SubTasks.Done != 1 {
		t.Errorf("task 1 is along %+v, want one of two", p)
	}
	if p := answer.Along[4]; p.Any() {
		t.Errorf("task 4 is along %+v, want nothing under it", p)
	}
	if answer.Contexts[4] != filepath.Join("/notes", "4") {
		t.Errorf("task 4's context is %q", answer.Contexts[4])
	}
	narrowed, err := Do(TaskList, path, given("state", "backlog"))
	if err != nil {
		t.Fatal(err)
	}
	if p := narrowed.Along[1]; p.SubTasks.Total != 2 {
		t.Errorf("a narrowed listing says task 1 is along %+v, want its two parts counted", p)
	}
}

// The same for objectives: how far along each one is, in the listing.
func TestObjectiveList_saysHowFarAlongEachObjectiveIs(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(ObjectiveAdd, path, given("title", "Ship cairn")); err != nil {
		t.Fatal(err)
	}
	for _, step := range []Input{
		{Given: map[string]string{"title": "Write it", "state": "done"}, Listed: map[string][]string{"objective": {"1"}}},
		{Given: map[string]string{"title": "Read it"}, Listed: map[string][]string{"objective": {"1"}}},
		given("title", "A part", "parent", "2"),
	} {
		if _, err := Do(TaskAdd, path, step); err != nil {
			t.Fatal(err)
		}
	}
	answer, err := Do(ObjectiveList, path, Input{})
	if err != nil {
		t.Fatal(err)
	}
	p := answer.Along[1]
	if p.Tasks.Total != 2 || p.Tasks.Done != 1 || p.SubTasks.Total != 1 {
		t.Errorf("objective 1 is along %+v, want one of two tasks and one part", p)
	}
	shown, err := Do(ObjectiveShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if shown.Progress != p {
		t.Errorf("the listing says %+v and showing it says %+v", p, shown.Progress)
	}
}
