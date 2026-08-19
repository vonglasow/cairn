package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"cairn/internal/domain"
)

// A machine reading the answer has to loop over the labels. A task carrying
// none must therefore hand back an empty list, never a null: null is not a
// list, and a reader that trusted the field would break on the first task
// nobody labelled.
func TestForWire_writesAnEmptyListRatherThanNothingAtAll(t *testing.T) {
	encoded, err := json.Marshal(forWire(spec(t)))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(encoded), `"labels":[]`) {
		t.Errorf("labels should be an empty list, got %s", encoded)
	}
}

func TestForWire_writesTheLabelsATaskCarries(t *testing.T) {
	task := spec(t)
	task.Labels = []string{"spec", "team"}

	encoded, err := json.Marshal(forWire(task))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(encoded), `"labels":["spec","team"]`) {
		t.Errorf("labels should carry what the task carries, got %s", encoded)
	}
}

func TestForWire_writesEveryAttributeAsText(t *testing.T) {
	task := spec(t)
	var err error
	if task.Deadline, err = domain.ParseDeadline("2026-09-01"); err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	if task.Estimate, err = domain.ParseEstimate("90m"); err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}

	wire := forWire(task)
	if wire.Deadline != "2026-09-01" {
		t.Errorf("Deadline = %q, want 2026-09-01", wire.Deadline)
	}
	if wire.Estimate != "1h30m" {
		t.Errorf("Estimate = %q, want 1h30m", wire.Estimate)
	}
}
