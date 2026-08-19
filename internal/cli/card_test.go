package cli

import (
	"strings"
	"testing"

	"cairn/internal/domain"
)

func spec(t *testing.T) domain.Task {
	t.Helper()
	task, err := domain.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	task.Number = 1
	return task
}

func TestCard_leadsWithTheNumberAndTheTitle(t *testing.T) {
	got := card(spec(t))
	if first := strings.SplitN(got, "\n", 2)[0]; !strings.Contains(first, "1") || !strings.Contains(first, "Write the spec") {
		t.Errorf("first line = %q, want the number and the title", first)
	}
}

func TestCard_namesEveryAttributeEvenTheEmptyOnes(t *testing.T) {
	got := card(spec(t))
	for _, key := range []string{"state", "priority", "description", "labels", "deadline", "estimate", "time spent"} {
		if !strings.Contains(got, key) {
			t.Errorf("card is missing %q:\n%s", key, got)
		}
	}
}

func TestCard_showsTheDefaultsATaskStartsWith(t *testing.T) {
	got := card(spec(t))
	if !strings.Contains(got, "backlog") || !strings.Contains(got, "normal") {
		t.Errorf("card should show backlog and normal:\n%s", got)
	}
}

func TestCard_showsWhatTheTaskCarries(t *testing.T) {
	task := spec(t)
	task.Description = "the one"
	task.Labels = []string{"spec", "team"}
	var err error
	if task.Deadline, err = domain.ParseDeadline("2026-09-01"); err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	if task.Estimate, err = domain.ParseEstimate("90m"); err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}

	got := card(task)
	for _, want := range []string{"the one", "spec, team", "2026-09-01", "1h30m"} {
		if !strings.Contains(got, want) {
			t.Errorf("card is missing %q:\n%s", want, got)
		}
	}
}

func TestLine_showsTheNumberAndTheTitle(t *testing.T) {
	if got := line(spec(t)); !strings.Contains(got, "1") || !strings.Contains(got, "Write the spec") {
		t.Errorf("line = %q, want the number and the title", got)
	}
}

func TestNothing_marksAnAbsentValue(t *testing.T) {
	if got := orNothing(""); got != absent {
		t.Errorf("orNothing(\"\") = %q, want %q", got, absent)
	}
	if got := orNothing("something"); got != "something" {
		t.Errorf("orNothing = %q, want it untouched", got)
	}
}
