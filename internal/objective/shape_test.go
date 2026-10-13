package objective_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

func TestOfObjective_carriesEveryFieldAScriptReads(t *testing.T) {
	deadline, err := domain.ParseDeadline("2026-10-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	estimate, err := domain.ParseEstimate("5h")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	var a objective.AsTold
	a.Numbered(5)
	a.Titled("Ship it")
	a.Standing(objective.Active)
	a.Carries(domain.Normal)
	a.Described("in full")
	a.Labelled("launch")
	a.Labelled("team")
	a.LabelsTold()
	a.Due(deadline)
	a.Estimated(estimate)
	built := a.Objective()

	encoded, err := json.Marshal(objective.OfObjective(built))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(encoded, &doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	want := map[string]any{
		"number":      float64(5),
		"title":       "Ship it",
		"description": "in full",
		"state":       string(objective.Active),
		"priority":    string(domain.Normal),
		"deadline":    "2026-10-01",
		"estimate":    "5h",
	}
	for field, expected := range want {
		if doc[field] != expected {
			t.Errorf("%s = %v, want %v", field, doc[field], expected)
		}
	}
	labels, ok := doc["labels"].([]any)
	if !ok || len(labels) != 2 || labels[0] != "launch" || labels[1] != "team" {
		t.Errorf("labels = %v, want [launch team]", doc["labels"])
	}
}

func TestOfObjective_writesAnEmptyListRatherThanNothingAtAll(t *testing.T) {
	built, err := objective.NewObjective("Ship it")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	encoded, err := json.Marshal(objective.OfObjective(built))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := string(encoded); !strings.Contains(got, `"labels":[]`) {
		t.Errorf("labels should be an empty list, got %s", got)
	}
}

func TestShape_everyFieldSaysWhatItHolds(t *testing.T) {
	for _, shape := range []any{objective.Shape{}} {
		of := reflect.TypeOf(shape)
		for i := range of.NumField() {
			field := of.Field(i)
			if field.Anonymous {
				continue
			}
			if field.Tag.Get("jsonschema") == "" {
				t.Errorf("%s.%s says nothing of what it holds", of.Name(), field.Name)
			}
		}
	}
}
