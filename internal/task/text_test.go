package task

import (
	"strings"
	"testing"

	"cairn/internal/domain"
)

func spec(t *testing.T) Task {
	t.Helper()
	built, err := NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	var a AsTold
	built.Tell(&a)
	a.Numbered(1)
	return a.Task()
}

func TestText_leadsWithTheNumberAndTheTitle(t *testing.T) {
	got := Text(spec(t), domain.Progress{})
	if first := strings.SplitN(got, "\n", 2)[0]; !strings.Contains(first, "1") || !strings.Contains(first, "Write the spec") {
		t.Errorf("first line = %q, want the number and the title", first)
	}
}

func TestText_namesEveryAttributeEvenTheEmptyOnes(t *testing.T) {
	got := Text(spec(t), domain.Progress{})
	for _, key := range []string{"state", "priority", "description", "labels", "deadline", "estimate", "time spent"} {
		if !strings.Contains(got, key) {
			t.Errorf("text is missing %q:\n%s", key, got)
		}
	}
}

func TestText_showsTheDefaultsATaskStartsWith(t *testing.T) {
	got := Text(spec(t), domain.Progress{})
	if !strings.Contains(got, "backlog") || !strings.Contains(got, "normal") {
		t.Errorf("text should show backlog and normal:\n%s", got)
	}
}

func TestText_showsWhatTheTaskCarries(t *testing.T) {
	built := spec(t)
	deadline, err := domain.ParseDeadline("2026-09-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	estimate, err := domain.ParseEstimate("90m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	var a AsTold
	built.Tell(&a)
	a.Described("the one")
	a.Labelled("spec")
	a.Labelled("team")
	a.LabelsTold()
	a.Due(deadline)
	a.Estimated(estimate)
	built = a.Task()

	got := Text(built, domain.Progress{})
	for _, want := range []string{"the one", "spec, team", "2026-09-01", "1h30m"} {
		if !strings.Contains(got, want) {
			t.Errorf("text is missing %q:\n%s", want, got)
		}
	}
}

func TestText_showsHowFarAlongItIs(t *testing.T) {
	got := Text(spec(t), domain.Progress{SubTasks: domain.Tally{Total: 4, Done: 1}})
	if !strings.Contains(got, "1 of 4 sub-tasks are done") {
		t.Errorf("text is missing the progress line:\n%s", got)
	}
	if strings.Contains(got, "0 of 0 tasks") {
		t.Errorf("text should not count tasks it does not have:\n%s", got)
	}
}

// TestText_countsTasksApartFromSubTasks pins that the tasks tally is only
// said when there are tasks to say it about — the mirror of what
// TestText_showsHowFarAlongItIs pins for sub-tasks.
func TestText_countsTasksApartFromSubTasks(t *testing.T) {
	got := Text(spec(t), domain.Progress{Tasks: domain.Tally{Total: 2, Done: 2}})
	if !strings.Contains(got, "2 of 2 tasks are done") {
		t.Errorf("text is missing the progress line:\n%s", got)
	}
	if strings.Contains(got, "0 of 0 sub-tasks") {
		t.Errorf("text should not count sub-tasks it does not have:\n%s", got)
	}
}

func TestLine_showsTheNumberAndTheTitle(t *testing.T) {
	if got := Line(spec(t)); !strings.Contains(got, "1") || !strings.Contains(got, "Write the spec") {
		t.Errorf("line = %q, want the number and the title", got)
	}
}

// listed builds a task with a number, a title and a parent, for the listing
// tests below.
func listed(number int, title string, parent int) Task {
	var a AsTold
	a.Numbered(number)
	a.Titled(title)
	a.Under(parent)
	return a.Task()
}

func TestTree_indentsWhatHangsUnderATask(t *testing.T) {
	got := Tree([]Task{
		listed(1, "Ship it", NoParent),
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
	got := Tree([]Task{
		listed(1, "Ship it", NoParent),
		listed(2, "Write the tests", 1),
		listed(3, "Something else", NoParent),
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
	got := Tree([]Task{
		listed(1, "Ship it", NoParent),
		listed(2, "Write the tests", 1),
		listed(3, "Write the docs", 1),
	})
	if at2, at3 := strings.Index(got, "#2"), strings.Index(got, "#3"); at2 > at3 {
		t.Errorf("sub-tasks should keep their order:\n%s", got)
	}
}

func TestTree_writesNothingForNoTasks(t *testing.T) {
	if got := Tree(nil); got != "" {
		t.Errorf("Tree(nil) = %q, want nothing", got)
	}
}

func TestUnder_writesNothingForNoTasks(t *testing.T) {
	if got := Under(nil); got != "" {
		t.Errorf("Under(nil) = %q, want nothing", got)
	}
}

func TestUnder_listsWhatHangsUnderATask(t *testing.T) {
	got := Under([]Task{listed(2, "Write the tests", 1)})
	if !strings.Contains(got, "sub-tasks") || !strings.Contains(got, "Write the tests") {
		t.Errorf("Under is missing the sub-task:\n%s", got)
	}
}
