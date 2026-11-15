package engine

import (
	"slices"
	"strings"
	"testing"

	"cairn/internal/engine/task"
)

func titlesOf(a Answer) []string {
	titles := []string{}
	for _, one := range append(append([]task.Task{}, a.Tasks...), a.Matched...) {
		titles = append(titles, task.OfTask(one).Title)
	}
	return titles
}

func TestTaskList_refusesOnlyWhatAClosedSetExcludes(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")

	if _, err := TaskList.Run(workedOn{Engine: e}, Input{
		Given: map[string]string{"state": "nowhere"},
	}); err == nil {
		t.Error("a column no task can sit in was accepted")
	}

	answer, err := TaskList.Run(workedOn{Engine: e}, Input{
		Listed: map[string][]string{"objective": {"9"}},
	})
	if err != nil {
		t.Fatalf("an objective nobody added was refused: %v", err)
	}
	if got := titlesOf(answer); len(got) != 0 {
		t.Errorf("it listed %v, want nothing serving objective 9", got)
	}
}

func TestObjectiveList_isHeldToTheObjectiveStates(t *testing.T) {
	e := openEngine(t)
	if _, err := ObjectiveList.Run(workedOn{Engine: e}, Input{
		Given: map[string]string{"state": "todo"},
	}); err == nil {
		t.Error("an objective listing accepted a column a task sits in")
	}
}

func TestTaskList_refusesAnObjectiveNothingCouldBeNumbered(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	for _, named := range []string{"soon", "0", "-1"} {
		if _, err := TaskList.Run(workedOn{Engine: e}, Input{
			Listed: map[string][]string{"objective": {named}},
		}); err == nil {
			t.Errorf("%q was accepted as an objective to narrow by", named)
		}
	}
}

func TestTaskList_refusesTheTasksInFocusAndTheOnesThatAreNot(t *testing.T) {
	_, err := keeping(Input{Given: map[string]string{"focus": "true", "no-focus": "true"}}, taskNarrowing, nil)
	if err == nil || !strings.Contains(err.Error(), "--focus and --no-focus") {
		t.Errorf("both asked gave %v", err)
	}
}

func TestTaskList_keepsTheTasksNothingStops(t *testing.T) {
	path := storeAt(t)
	for _, title := range []string{"free", "waits on 1", "waits on 4, done", "done", "waits on the client"} {
		did(t, path, TaskNew, ask("title", title))
	}
	did(t, path, TaskEdit, ask("number", "2").with("blocker", "1"))
	did(t, path, TaskEdit, ask("number", "3").with("blocker", "4"))
	did(t, path, TaskEdit, ask("number", "4", "state", "done"))
	did(t, path, TaskEdit, ask("number", "5", "because", "the client"))
	listed := Of(TaskList, did(t, path, TaskList, ask("unblocked", "true"))).(Tasks)
	if got := rowsListed(listed); !slices.Equal(got, []int{1, 3, 4}) {
		t.Errorf("unblocked keeps %v, want 1, 3 and 4", got)
	}
	did(t, path, ViewNew, ask("name", "Free", "unblocked", "true"))
	if views := Of(ViewList, did(t, path, ViewList, ask())).(Views); !views.Views[0].Holds.Unblocked || views.Views[0].Keeps != 3 {
		t.Errorf("the view holds %+v", views.Views[0])
	}
	if got := rowsListed(Of(TaskList, did(t, path, TaskList, ask("view", "Free"))).(Tasks)); !slices.Equal(got, []int{1, 3, 4}) {
		t.Errorf("the view keeps %v", got)
	}
}
