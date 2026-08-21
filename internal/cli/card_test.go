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
	got := card(spec(t), domain.Progress{})
	if first := strings.SplitN(got, "\n", 2)[0]; !strings.Contains(first, "1") || !strings.Contains(first, "Write the spec") {
		t.Errorf("first line = %q, want the number and the title", first)
	}
}

func TestCard_namesEveryAttributeEvenTheEmptyOnes(t *testing.T) {
	got := card(spec(t), domain.Progress{})
	for _, key := range []string{"state", "priority", "description", "labels", "deadline", "estimate", "time spent"} {
		if !strings.Contains(got, key) {
			t.Errorf("card is missing %q:\n%s", key, got)
		}
	}
}

func TestCard_showsTheDefaultsATaskStartsWith(t *testing.T) {
	got := card(spec(t), domain.Progress{})
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

	got := card(task, domain.Progress{})
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

// listed builds a task with a number, a title and a parent, for the listing
// tests below.
func listed(number int, title string, parent int) domain.Task {
	return domain.Task{Number: number, Title: title, Parent: parent}
}

func TestTree_indentsWhatHangsUnderATask(t *testing.T) {
	got := tree([]domain.Task{
		listed(1, "Ship it", domain.NoParent),
		listed(2, "Write the tests", 1),
	})
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2:\n%s", len(lines), got)
	}
	if strings.HasPrefix(lines[0], " ") {
		t.Errorf("a task hanging under nothing should not be indented: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], " ") {
		t.Errorf("a sub-task should be indented: %q", lines[1])
	}
}

func TestTree_putsEverythingHangingUnderATaskRightAfterIt(t *testing.T) {
	got := tree([]domain.Task{
		listed(1, "Ship it", domain.NoParent),
		listed(2, "Write the tests", 1),
		listed(3, "Something else", domain.NoParent),
	})
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3:\n%s", len(lines), got)
	}
	if !strings.Contains(lines[1], "Write the tests") {
		t.Errorf("the sub-task should follow its parent, got %q", lines[1])
	}
	if strings.HasPrefix(lines[2], " ") {
		t.Errorf("a task hanging under nothing should not be indented: %q", lines[2])
	}
}

func TestTree_keepsSeveralSubTasksInTheirOrder(t *testing.T) {
	got := tree([]domain.Task{
		listed(1, "Ship it", domain.NoParent),
		listed(2, "Write the tests", 1),
		listed(3, "Write the docs", 1),
	})
	if at2, at3 := strings.Index(got, "#2"), strings.Index(got, "#3"); at2 > at3 {
		t.Errorf("sub-tasks should keep their order:\n%s", got)
	}
}

func TestTree_writesNothingForNoTasks(t *testing.T) {
	if got := tree(nil); got != "" {
		t.Errorf("tree(nil) = %q, want nothing", got)
	}
}
