package task

import (
	"strconv"
	"strings"
	"testing"
	"time"

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
	}, NothingInFocus)
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
	}, NothingInFocus)
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
	}, NothingInFocus)
	if at2, at3 := strings.Index(got, "#2"), strings.Index(got, "#3"); at2 > at3 {
		t.Errorf("sub-tasks should keep their order:\n%s", got)
	}
}

func TestTree_writesNothingForNoTasks(t *testing.T) {
	if got := Tree(nil, NothingInFocus); got != "" {
		t.Errorf("Tree(nil, NothingInFocus) = %q, want nothing", got)
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

// Tree marks the task being worked on, and marks nothing else. The second half
// is what carries the rule: one line marked proves nothing if every line is.
func TestTree_marksOnlyTheTaskInFocus(t *testing.T) {
	tasks := []Task{
		listed(1, "Ship it", NoParent),
		listed(2, "Write the tests", 1),
		listed(3, "Something else", NoParent),
	}
	for _, inFocus := range []int{1, 2, 3} {
		got := Tree(tasks, inFocus)
		if marks := strings.Count(got, InFocusMark); marks != 1 {
			t.Errorf("with %d in focus, %d lines are marked, want exactly one:\n%s",
				inFocus, marks, got)
		}
		for _, line := range strings.Split(got, "\n") {
			if !strings.Contains(line, InFocusMark) {
				continue
			}
			if want := "#" + strconv.Itoa(inFocus); !strings.Contains(line, want) {
				t.Errorf("with %d in focus, %q is the marked line", inFocus, line)
			}
		}
	}
}

func TestTree_marksNothingWhenNothingIsInFocus(t *testing.T) {
	got := Tree([]Task{listed(1, "Ship it", NoParent)}, NothingInFocus)
	if strings.Contains(got, InFocusMark) {
		t.Errorf("nothing is in focus, yet a line is marked:\n%s", got)
	}
}

// A card carrying what most tasks do not: a time already spent on it, a parent
// it hangs under, an objective it serves. Each is a line of the card and each
// has a membrane method of its own, reached only by a task that carries it.
func TestText_carriesTheAttributesMostTasksDoNot(t *testing.T) {
	built, err := NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	var a AsTold
	built.Tell(&a)
	a.Numbered(2)
	a.Spent(90 * time.Minute)
	a.Under(1)
	a.Serves(3)

	got := Text(a.Task(), domain.Progress{})
	for _, want := range []string{"time spent", "1h30m", "parent", "objective"} {
		if !strings.Contains(got, want) {
			t.Errorf("the card does not carry %q:\n%s", want, got)
		}
	}
}

// The two rules that say where a task may hang answer nothing when it may. The
// refusing side is asserted elsewhere; this is the side that says yes, and a
// rule that only ever refuses in a test is a rule nobody has seen allow.
func TestTheHangingRules_answerNothingWhenTheTaskMayHang(t *testing.T) {
	root := listed(1, "Ship it", NoParent)
	if err := root.MayCarrySubTasks(); err != nil {
		t.Errorf("a root task may carry sub-tasks, and it said %v", err)
	}
	under := listed(2, "Write the tests", 1)
	if err := under.MayHangUnderItsParent(); err != nil {
		t.Errorf("a task may hang under another, and it said %v", err)
	}
}
