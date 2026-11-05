package engine

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

func listedNumbers(t *testing.T, e *Engine, given map[string]string, numbers ...string) (Answer, []int) {
	t.Helper()
	answer, err := TaskList.Run(On{Engine: e}, Input{Given: given, Listed: map[string][]string{"number": numbers}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	var kept []int
	for _, one := range answer.Matched {
		kept = append(kept, one.Carried().Number)
	}
	return answer, kept
}

func TestTaskList_numbersAnswerOnlyThoseTasksAndSayWhichIsNone(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	add(t, e, "Second")
	add(t, e, "Third")

	answer, kept := listedNumbers(t, e, nil, "3", "9", "1", "3", "9", "12")
	if !slices.Equal(kept, []int{1, 3}) {
		t.Errorf("it kept %v, want 1 then 3", kept)
	}
	written, err := json.Marshal(Of(TaskList, answer))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if want := `"unknown":[9,12]`; !strings.Contains(string(written), want) {
		t.Errorf("it wrote %s, want %s", written, want)
	}
}

func TestTaskList_numbersReachAnArchivedTask(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	add(t, e, "Second")
	if err := e.archive(1); err != nil {
		t.Fatalf("archive: %v", err)
	}

	answer, kept := listedNumbers(t, e, nil, "1", "2")
	if !slices.Equal(kept, []int{1, 2}) {
		t.Errorf("it kept %v, want 1 and 2", kept)
	}
	written, err := json.Marshal(Of(TaskList, answer))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(written), `"unknown"`) {
		t.Errorf("it wrote %s, want nothing said unknown", written)
	}

	_, kept = listedNumbers(t, e, map[string]string{"archived": "true"}, "1", "2")
	if !slices.Equal(kept, []int{1}) {
		t.Errorf("asked for the archived ones, it kept %v, want 1", kept)
	}
}

func TestTaskList_numbersAreNarrowedLikeEveryListing(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	add(t, e, "Second")
	if _, err := TaskEdit.Run(On{Engine: e}, Input{Given: map[string]string{"number": "2", "state": "todo"}}); err != nil {
		t.Fatalf("TaskEdit: %v", err)
	}

	answer, kept := listedNumbers(t, e, map[string]string{"state": "todo"}, "1", "2")
	if !slices.Equal(kept, []int{2}) || len(answer.Unknown) != 0 {
		t.Errorf("it kept %v and said %v unknown, want 2 alone and nothing unknown", kept, answer.Unknown)
	}
}

func TestTaskList_numbersRefuseWhatIsNoNumber(t *testing.T) {
	e := openEngine(t)
	for _, numbers := range [][]string{{}, {"first"}, {"0"}, {"-2"}} {
		_, err := TaskList.Run(On{Engine: e}, Input{Listed: map[string][]string{"number": numbers}})
		if !errors.Is(err, errMisread) {
			t.Errorf("%q: err = %v, want it misread", numbers, err)
		}
	}
}
