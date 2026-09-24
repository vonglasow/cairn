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

func TestTree_writesNothingForNoTasks(t *testing.T) {
	if got := Tree(nil, Marks{}); got != "" {
		t.Errorf("Tree(nil, Marks{}) = %q, want nothing", got)
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

// Tree marks the tasks being worked on, and marks nothing else. The second
// half is what carries the rule: one line marked proves nothing if every line
// is.
func TestTree_marksTheTasksInFocus(t *testing.T) {
	tasks := []Task{
		listed(1, "Ship it", NoParent),
		listed(2, "Write the tests", 1),
		listed(3, "Something else", NoParent),
	}
	// The work, and not what hangs under it: a listing no longer draws a
	// sub-task, so one in focus has no line here to be marked on.
	if got := Tree(tasks, Marks{InFocus: []int{2}}); strings.Contains(got, InFocusMark) {
		t.Errorf("a sub-task in focus marked a listing it is not in:\n%s", got)
	}
	for _, inFocus := range []int{1, 3} {
		got := Tree(tasks, Marks{InFocus: []int{inFocus}})
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

// Several tasks in focus are all marked, and a task that is not in focus is
// not — several can be under way at once, one per agent session.
func TestTree_marksEveryTaskInFocus(t *testing.T) {
	tasks := []Task{
		listed(1, "Ship it", NoParent),
		listed(2, "Write the tests", 1),
		listed(3, "Something else", NoParent),
	}
	got := Tree(tasks, Marks{InFocus: []int{1, 3}})
	if marks := strings.Count(got, InFocusMark); marks != 2 {
		t.Errorf("with two in focus, %d lines are marked, want two:\n%s", marks, got)
	}
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "Write the tests") && strings.Contains(line, InFocusMark) {
			t.Errorf("a task that is not in focus is marked: %q", line)
		}
	}
}

func TestTree_marksNothingWhenNothingIsInFocus(t *testing.T) {
	got := Tree([]Task{listed(1, "Ship it", NoParent)}, Marks{})
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

// The section under a task says which of them are finished. It is the section
// that answers "is the whole of it done", and one that named them without
// saying so made somebody open each in turn to find out.
func TestUnder_namesEachOneAsDoneOrNot(t *testing.T) {
	made := func(number int, title string, state domain.State) Task {
		var a AsTold
		a.Numbered(number)
		a.Titled(title)
		a.Under(1)
		a.InColumn(state)
		return a.Task()
	}
	got := Under([]Task{
		made(2, "Write the tests", domain.Done),
		made(3, "Write the docs", domain.Todo),
	})
	for _, want := range []string{DoneMark + "  #2  Write the tests", "#3  Write the docs"} {
		if !strings.Contains(got, want) {
			t.Errorf("Under =\n%s\nwant a line reading %q", got, want)
		}
	}
	if strings.Contains(got, DoneMark+"  #3") {
		t.Errorf("Under =\n%s\nand task 3 is not done", got)
	}
}

// A listing names the work, and a sub-task is part of one. It used to draw
// them indented under their parent; what hangs under a task is now read where
// the task is read, which is the one place it answers a question.
func TestTree_leavesSubTasksOut(t *testing.T) {
	made := func(number int, title string, parent int) Task {
		var a AsTold
		a.Numbered(number)
		a.Titled(title)
		if parent != NoParent {
			a.Under(parent)
		}
		return a.Task()
	}
	got := Tree([]Task{
		made(1, "Ship it", NoParent),
		made(2, "Write the tests", 1),
		made(3, "Something else", NoParent),
	}, Marks{})
	for _, want := range []string{"#1  Ship it", "#3  Something else"} {
		if !strings.Contains(got, want) {
			t.Errorf("Tree =\n%s\nwant it to name %q", got, want)
		}
	}
	if strings.Contains(got, "Write the tests") {
		t.Errorf("Tree =\n%s\nand a sub-task is not laid out in a listing", got)
	}
}

// What hangs under a task keeps the order it was given, which is the order the
// engine holds it in. A section that sorted it again would be a second place
// deciding where a part sits.
func TestUnder_keepsTheOrderItWasGiven(t *testing.T) {
	made := func(number int, title string) Task {
		var a AsTold
		a.Numbered(number)
		a.Titled(title)
		a.Under(1)
		return a.Task()
	}
	got := Under([]Task{made(3, "Write the docs"), made(2, "Write the tests")})
	if at3, at2 := strings.Index(got, "#3"), strings.Index(got, "#2"); at3 > at2 {
		t.Errorf("Under reordered them:\n%s", got)
	}
}
