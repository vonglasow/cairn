package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/task"
)

func spec(t *testing.T) task.Task {
	t.Helper()
	built, err := task.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	var a task.AsTold
	built.Tell(&a)
	a.Numbered(1)
	return a.Task()
}

func TestForWire_writesAnEmptyListRatherThanNothingAtAll(t *testing.T) {
	encoded, err := json.Marshal(engine.OfTask(spec(t), domain.Progress{}))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(encoded), `"labels":[]`) {
		t.Errorf("labels should be an empty list, got %s", encoded)
	}
}

func TestForWire_writesTheLabelsATaskCarries(t *testing.T) {
	built := spec(t)
	var a task.AsTold
	built.Tell(&a)
	a.Labelled("spec")
	a.Labelled("team")
	a.LabelsTold()
	built = a.Task()

	encoded, err := json.Marshal(engine.OfTask(built, domain.Progress{}))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(encoded), `"labels":["spec","team"]`) {
		t.Errorf("labels should carry what the task carries, got %s", encoded)
	}
}

func TestForWire_writesEveryAttributeAsText(t *testing.T) {
	built := spec(t)
	deadline, err := domain.ParseDeadline("2026-09-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	estimate, err := domain.ParseEstimate("90m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	var a task.AsTold
	built.Tell(&a)
	a.Due(deadline)
	a.Estimated(estimate)
	built = a.Task()

	wire := engine.OfTask(built, domain.Progress{})
	if wire.Deadline != "2026-09-01" {
		t.Errorf("Deadline = %q, want 2026-09-01", wire.Deadline)
	}
	if wire.Estimate != "1h30m" {
		t.Errorf("Estimate = %q, want 1h30m", wire.Estimate)
	}
}
