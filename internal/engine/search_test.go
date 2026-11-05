package engine

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTaskSearch_anyKeepsATaskCarryingOneOfWhatIsNamed(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Port T010")
	add(t, e, "Merge T013 and T022")
	add(t, e, "Write the spec")

	answer, err := TaskSearch.Run(On{Engine: e}, Input{
		Given:  map[string]string{"any": "true"},
		Listed: map[string][]string{"named": {"T022", "T010", "T013", "T999", "T010"}},
	})
	if err != nil {
		t.Fatalf("TaskSearch: %v", err)
	}
	written, err := json.Marshal(Of(TaskSearch, answer))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := len(answer.Matched); got != 2 {
		t.Errorf("%d tasks, want 2", got)
	}
	if want := `"matched":{"T010":[1],"T013":[2],"T022":[2],"T999":[]}`; !strings.Contains(string(written), want) {
		t.Errorf("it wrote %s, want %s", written, want)
	}
}

func TestTaskSearch_anyIsNarrowedLikeEverySearch(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Port T010")
	add(t, e, "Merge T013")
	if _, err := TaskEdit.Run(On{Engine: e}, Input{Given: map[string]string{"number": "2", "state": "todo"}}); err != nil {
		t.Fatalf("TaskEdit: %v", err)
	}

	answer, err := TaskSearch.Run(On{Engine: e}, Input{
		Given:  map[string]string{"any": "true", "state": "todo"},
		Listed: map[string][]string{"named": {"T010", "T013"}},
	})
	if err != nil {
		t.Fatalf("TaskSearch: %v", err)
	}
	if len(answer.Matched) != 1 || answer.Terms["T013"][0] != 2 || len(answer.Terms["T010"]) != 0 {
		t.Errorf("it kept %d tasks and said %v, want task 2 alone, under T013", len(answer.Matched), answer.Terms)
	}
}

func TestTaskSearch_everythingNamedSaysNoTerm(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Port T010")

	answer, err := TaskSearch.Run(On{Engine: e}, Input{Listed: map[string][]string{"named": {"T010"}}})
	if err != nil {
		t.Fatalf("TaskSearch: %v", err)
	}
	written, err := json.Marshal(Of(TaskSearch, answer))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(written), `"matched"`) {
		t.Errorf("it wrote %s, want nothing said of which term matched", written)
	}
}
