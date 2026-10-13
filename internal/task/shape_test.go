package task_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/task"
)

func TestOfTask_carriesEveryFieldAScriptReads(t *testing.T) {
	deadline, err := domain.ParseDeadline("2026-09-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	estimate, err := domain.ParseEstimate("90m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	spent, err := domain.ParseEstimate("30m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	var a task.AsTold
	a.Numbered(1)
	a.Under(2)
	a.Serves(3)
	a.ObjectivesTold()
	a.Titled("Write the spec")
	a.Described("in full")
	a.InColumn(domain.Todo)
	a.Carries(domain.High)
	a.Labelled("spec")
	a.Labelled("team")
	a.LabelsTold()
	a.Due(deadline)
	a.Estimated(estimate)
	a.Spent(spent)
	a.BlockedBy(4)
	a.BlockedBy(5)
	a.BlockersTold()
	a.Because("the client's answer")
	built := a.Task()

	encoded, err := json.Marshal(task.OfTask(built))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(encoded, &doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	want := map[string]any{
		"number": float64(1),
		"parent": float64(2),

		"title":       "Write the spec",
		"description": "in full",
		"state":       string(domain.Todo),
		"priority":    string(domain.High),
		"deadline":    "2026-09-01",
		"estimate":    "1h30m",
		"time_spent":  "30m",
	}
	for field, expected := range want {
		if doc[field] != expected {
			t.Errorf("%s = %v, want %v", field, doc[field], expected)
		}
	}
	labels, ok := doc["labels"].([]any)
	if !ok || len(labels) != 2 || labels[0] != "spec" || labels[1] != "team" {
		t.Errorf("labels = %v, want [spec team]", doc["labels"])
	}
	objectives, ok := doc["objectives"].([]any)
	if !ok || len(objectives) != 1 || objectives[0] != float64(3) {
		t.Errorf("objectives = %v, want [3]", doc["objectives"])
	}
	blockers, ok := doc["blocked_by"].([]any)
	if !ok || len(blockers) != 2 || blockers[0] != float64(4) || blockers[1] != float64(5) {
		t.Errorf("blocked_by = %v, want [4 5]", doc["blocked_by"])
	}
	if doc["because"] != "the client's answer" {
		t.Errorf("because = %v, want the words it was given", doc["because"])
	}
}

func TestOfTask_writesAnEmptyListRatherThanNothingAtAll(t *testing.T) {
	built, err := task.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	encoded, err := json.Marshal(task.OfTask(built))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	for _, want := range []string{`"labels":[]`, `"blocked_by":[]`, `"objectives":[]`} {
		if got := string(encoded); !strings.Contains(got, want) {
			t.Errorf("%s should be there, got %s", want, got)
		}
	}
}

func TestOfTask_keepsWhatATaskSaysApartFromItsFacts(t *testing.T) {
	var a task.AsTold
	a.Numbered(1)
	a.Titled("Write the spec")
	a.Described("in full")
	shape := task.OfTask(a.Task())
	if shape.Title != "Write the spec" || shape.Description != "in full" {
		t.Errorf("the facts say %q and the task says %q", shape.Title, shape.Description)
	}
}

func TestShape_everyFieldSaysWhatItHolds(t *testing.T) {
	for _, shape := range []any{task.Shape{}, task.Facts{}} {
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
