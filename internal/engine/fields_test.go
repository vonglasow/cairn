package engine

import (
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strings"
	"testing"
)

func rowsAsJSON(t *testing.T, op Operation, a Answer) []map[string]any {
	t.Helper()
	written, err := json.Marshal(Of(op, a))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var listing struct {
		Tasks []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal(written, &listing); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	return listing.Tasks
}

func TestTaskList_carriesOnlyTheFieldsAskedFor(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")

	answer, err := TaskList.Run(On{Engine: e}, Input{Listed: map[string][]string{"fields": {"title", "number", "title"}}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	rows := rowsAsJSON(t, TaskList, answer)
	if len(rows) != 1 {
		t.Fatalf("%d rows, want 1", len(rows))
	}
	if named := slices.Sorted(maps.Keys(rows[0])); !slices.Equal(named, []string{"number", "title"}) {
		t.Errorf("the row carries %v, want number and title alone", named)
	}
}

func TestTaskList_writesTheFieldsInTheOrderARowCarriesThem(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")

	answer, err := TaskList.Run(On{Engine: e}, Input{Listed: map[string][]string{"fields": {"state", "number"}}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	written, err := json.Marshal(Of(TaskList, answer))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(written), `{"number":1,"state":"backlog"}`) {
		t.Errorf("it wrote %s, want the number before the state", written)
	}
}

func TestTaskSearch_carriesOnlyTheFieldsAskedFor(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")

	answer, err := TaskSearch.Run(On{Engine: e}, Input{Listed: map[string][]string{"named": {"spec"}, "fields": {"described"}}})
	if err != nil {
		t.Fatalf("TaskSearch: %v", err)
	}
	rows := rowsAsJSON(t, TaskSearch, answer)
	if len(rows) != 1 || len(rows[0]) != 1 || rows[0]["described"] != false {
		t.Errorf("it carried %v, want one row saying only whether it is described", rows)
	}
}

func TestTaskList_withoutFieldsCarriesEveryOne(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")

	answer, err := TaskList.Run(On{Engine: e}, Input{})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	rows := rowsAsJSON(t, TaskList, answer)
	if named := slices.Sorted(maps.Keys(rows[0])); !slices.Equal(named, slices.Sorted(slices.Values(rowFields()))) {
		t.Errorf("the row carries %v, want every field", named)
	}
}

func TestTaskList_refusesAFieldARowDoesNotCarry(t *testing.T) {
	e := openEngine(t)

	for _, op := range []Operation{TaskList, TaskSearch} {
		_, err := op.Run(On{Engine: e}, Input{Listed: map[string][]string{"named": {"spec"}, "fields": {"number", "description"}}})
		if !errors.Is(err, ErrNotAField) {
			t.Fatalf("%s: err = %v, want ErrNotAField", op.Name(), err)
		}
		if !strings.Contains(err.Error(), `"description"`) || !strings.Contains(err.Error(), strings.Join(rowFields(), ", ")) {
			t.Errorf("%s: %q does not name what was asked and what a row carries", op.Name(), err)
		}
	}
}

func TestTaskList_refusesAnEmptyListOfFields(t *testing.T) {
	e := openEngine(t)

	_, err := TaskList.Run(On{Engine: e}, Input{Listed: map[string][]string{"fields": {}}})
	if err == nil || RefusalOf(err).Kind != KindUsage {
		t.Fatalf("err = %v, want it not understood", err)
	}
	if !strings.Contains(err.Error(), strings.Join(rowFields(), ", ")) {
		t.Errorf("%q does not name what a row carries", err)
	}
}

func TestRowFields_areWhatARowCarries(t *testing.T) {
	want := []string{"number", "title", "state", "archived", "priority", "labels", "blocked_by", "linked",
		"because", "deadline", "estimate", "time_spent", "objectives", "described"}
	if got := rowFields(); !slices.Equal(got, want) {
		t.Errorf("rowFields() = %v, want %v", got, want)
	}
}
