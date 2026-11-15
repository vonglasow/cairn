package engine

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

func listedPage(t *testing.T, e *Engine, given map[string]string) (Answer, []int) {
	t.Helper()
	answer, err := TaskList.Run(workedOn{Engine: e}, Input{Given: given})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	var kept []int
	for _, one := range append(slices.Clone(answer.Tasks), answer.Matched...) {
		kept = append(kept, one.Carried().Number)
	}
	return answer, kept
}

func listingJSON(t *testing.T, a Answer) string {
	t.Helper()
	out, err := json.Marshal(Of(TaskList, a))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return string(out)
}

func TestTaskList_aPageOfNoTaskIsRefused(t *testing.T) {
	e := openEngine(t)
	for _, given := range []map[string]string{{"limit": "0"}, {"limit": "-1"}, {"offset": "-1"}} {
		_, err := TaskList.Run(workedOn{Engine: e}, Input{Given: given})
		if !errors.Is(err, errMisread) {
			t.Errorf("%v: err = %v, want it misread", given, err)
		}
	}
}

func TestTaskList_countsByStateInTheBoardsOrder(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	add(t, e, "Second")
	add(t, e, "Third")
	for number, state := range map[string]string{"1": "done", "3": "todo"} {
		if _, err := TaskEdit.Run(workedOn{Engine: e}, Input{Given: map[string]string{"number": number, "state": state}}); err != nil {
			t.Fatalf("TaskEdit: %v", err)
		}
	}

	answer, kept := listedPage(t, e, map[string]string{"count": "state"})
	if len(kept) != 0 {
		t.Errorf("counting kept %v, want no row", kept)
	}
	said := listingJSON(t, answer)
	if !strings.Contains(said, `"counts":{"backlog":1,"done":1,"todo":1}`) || !strings.Contains(said, `"total":3`) || !strings.Contains(said, `"tasks":[]`) {
		t.Errorf("it wrote %s, want one in each of backlog, todo and done, of 3, and no row", said)
	}
}

func TestTaskList_countsByLabelAmongWhatIsKept(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	add(t, e, "Second")
	add(t, e, "Third")
	for number, labels := range map[string][]string{"1": {"team", "spec"}, "2": {"team"}} {
		if _, err := TaskEdit.Run(workedOn{Engine: e}, Input{Given: map[string]string{"number": number}, Listed: map[string][]string{"label": labels}}); err != nil {
			t.Fatalf("TaskEdit: %v", err)
		}
	}

	answer, _ := listedPage(t, e, map[string]string{"count": "label"})
	if said := listingJSON(t, answer); !strings.Contains(said, `"counts":{"spec":1,"team":2}`) || !strings.Contains(said, `"total":3`) {
		t.Errorf("it wrote %s, want spec 1 and team 2, of 3", said)
	}

	answer, err := TaskList.Run(workedOn{Engine: e}, Input{Given: map[string]string{"count": "label"}, Listed: map[string][]string{"number": {"2", "9"}}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	if said := listingJSON(t, answer); !strings.Contains(said, `"counts":{"team":1}`) || !strings.Contains(said, `"unknown":[9]`) {
		t.Errorf("by number it wrote %s, want team 1 and 9 unknown", said)
	}
}

func TestTaskList_countingRefusesWhatShapesARow(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	for _, in := range []Input{
		{Given: map[string]string{"count": "state", "limit": "2"}},
		{Given: map[string]string{"count": "state", "offset": "0"}},
		{Given: map[string]string{"count": "state"}, Listed: map[string][]string{"fields": {"number"}}},
	} {
		_, err := TaskList.Run(workedOn{Engine: e}, in)
		if !errors.Is(err, errMisread) {
			t.Errorf("%v: err = %v, want it misread", in, err)
		}
	}
	_, err := TaskList.Run(workedOn{Engine: e}, Input{Given: map[string]string{"count": "priority"}})
	if !errors.Is(err, errNotOneOf) || !strings.Contains(err.Error(), "state, label") {
		t.Errorf("err = %v, want it not one of state, label", err)
	}
}

func searchedPage(t *testing.T, e *Engine, in Input) (Answer, []int) {
	t.Helper()
	answer, err := TaskSearch.Run(workedOn{Engine: e}, in)
	if err != nil {
		t.Fatalf("TaskSearch: %v", err)
	}
	var kept []int
	for _, one := range answer.Matched {
		kept = append(kept, one.Carried().Number)
	}
	return answer, kept
}

func TestTaskSearch_aPageIsCutFromWhatTheSearchKept(t *testing.T) {
	e := openEngine(t)
	for _, title := range []string{"Write the spec", "Ship it", "Review the spec", "Print the spec"} {
		add(t, e, title)
	}

	answer, kept := searchedPage(t, e, Input{Given: map[string]string{"limit": "1", "offset": "1"}, Listed: map[string][]string{"named": {"spec"}}})
	if !slices.Equal(kept, []int{3}) {
		t.Errorf("it kept %v, want 3", kept)
	}
	if said, _ := json.Marshal(Of(TaskSearch, answer)); !strings.Contains(string(said), `"total":3`) {
		t.Errorf("it wrote %s, want a total of 3", said)
	}

	answer, kept = searchedPage(t, e, Input{Given: map[string]string{"offset": "5"}, Listed: map[string][]string{"named": {"spec"}}})
	if said, _ := json.Marshal(Of(TaskSearch, answer)); len(kept) != 0 || !strings.Contains(string(said), `"total":3`) {
		t.Errorf("past the last it kept %v and wrote %s, want none of 3", kept, said)
	}

	answer, _ = searchedPage(t, e, Input{Listed: map[string][]string{"named": {"spec"}}})
	if said, _ := json.Marshal(Of(TaskSearch, answer)); strings.Contains(string(said), `"total"`) {
		t.Errorf("it wrote %s, want no total without a page", said)
	}
}

func TestTaskSearch_aPageSaysWhatMatchedOnlyOfItsRows(t *testing.T) {
	e := openEngine(t)
	for _, title := range []string{"Port T010", "Merge T013", "Close T010"} {
		add(t, e, title)
	}

	answer, kept := searchedPage(t, e, Input{Given: map[string]string{"any": "true", "limit": "1", "offset": "1"}, Listed: map[string][]string{"named": {"T010", "T013"}}})
	if !slices.Equal(kept, []int{2}) {
		t.Errorf("it kept %v, want 2", kept)
	}
	said, _ := json.Marshal(Of(TaskSearch, answer))
	if !strings.Contains(string(said), `"matched":{"T010":[],"T013":[2]}`) || !strings.Contains(string(said), `"total":3`) {
		t.Errorf("it wrote %s, want T013 on task 2, T010 on none of the page, of 3", said)
	}
}

func TestTaskSearch_aPageOfNoTaskIsRefused(t *testing.T) {
	e := openEngine(t)
	for _, given := range []map[string]string{{"limit": "0"}, {"limit": "-1"}, {"offset": "-1"}} {
		_, err := TaskSearch.Run(workedOn{Engine: e}, Input{Given: given, Listed: map[string][]string{"named": {"spec"}}})
		if !errors.Is(err, errMisread) {
			t.Errorf("%v: err = %v, want it misread", given, err)
		}
	}
}
