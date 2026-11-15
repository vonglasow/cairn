package engine

import (
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
