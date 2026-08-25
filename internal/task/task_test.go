package task

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
)

// told is a fake Told: it records, in order, the facts it was told rather
// than doing anything with them. A test on it is a test on the order and the
// content of what Tell says, nothing else.
type told struct {
	said []string
}

func (r *told) Numbered(n int)  { r.said = append(r.said, fmt.Sprintf("Numbered(%d)", n)) }
func (r *told) Titled(s string) { r.said = append(r.said, fmt.Sprintf("Titled(%q)", s)) }
func (r *told) InColumn(s domain.State) {
	r.said = append(r.said, fmt.Sprintf("InColumn(%q)", s))
}
func (r *told) Carries(p domain.Priority) {
	r.said = append(r.said, fmt.Sprintf("Carries(%q)", p))
}
func (r *told) Described(s string) { r.said = append(r.said, fmt.Sprintf("Described(%q)", s)) }
func (r *told) NotDescribed()      { r.said = append(r.said, "NotDescribed()") }
func (r *told) Labelled(s string)  { r.said = append(r.said, fmt.Sprintf("Labelled(%q)", s)) }
func (r *told) NotLabelled()       { r.said = append(r.said, "NotLabelled()") }
func (r *told) LabelsTold()        { r.said = append(r.said, "LabelsTold()") }
func (r *told) Due(t time.Time)    { r.said = append(r.said, fmt.Sprintf("Due(%s)", t)) }
func (r *told) NotDue()            { r.said = append(r.said, "NotDue()") }
func (r *told) Estimated(d time.Duration) {
	r.said = append(r.said, fmt.Sprintf("Estimated(%s)", d))
}
func (r *told) NotEstimated()         { r.said = append(r.said, "NotEstimated()") }
func (r *told) Spent(d time.Duration) { r.said = append(r.said, fmt.Sprintf("Spent(%s)", d)) }
func (r *told) NothingSpent()         { r.said = append(r.said, "NothingSpent()") }
func (r *told) Under(n int)           { r.said = append(r.said, fmt.Sprintf("Under(%d)", n)) }
func (r *told) UnderNothing()         { r.said = append(r.said, "UnderNothing()") }
func (r *told) Serves(n int)          { r.said = append(r.said, fmt.Sprintf("Serves(%d)", n)) }
func (r *told) ServesNothing()        { r.said = append(r.said, "ServesNothing()") }

// TestTell_saysEveryFactATaskCarries checks that Tell says every fact a task
// carries, in the card's order, and says the absence of a fact rather than
// staying silent about it.
func TestTell_saysEveryFactATaskCarries(t *testing.T) {
	deadline, err := domain.ParseDeadline("2026-01-02")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}

	full := Task{
		number: 3, parent: 1, objective: 2, title: "Write the spec",
		description: "Say what it does", state: domain.Current, priority: domain.High,
		labels: []string{"backend", "urgent"}, deadline: deadline,
		estimate: 90 * time.Minute, timeSpent: 30 * time.Minute,
	}
	got := &told{}
	full.Tell(got)

	want := []string{
		`Numbered(3)`, `Titled("Write the spec")`, `InColumn("current")`,
		`Carries("high")`, `Described("Say what it does")`,
		`Labelled("backend")`, `Labelled("urgent")`, `LabelsTold()`,
		fmt.Sprintf("Due(%s)", deadline), fmt.Sprintf("Estimated(%s)", 90*time.Minute),
		fmt.Sprintf("Spent(%s)", 30*time.Minute), `Under(1)`, `Serves(2)`,
	}
	if len(got.said) != len(want) {
		t.Fatalf("said = %v, want %v", got.said, want)
	}
	for i := range want {
		if got.said[i] != want[i] {
			t.Errorf("said[%d] = %q, want %q", i, got.said[i], want[i])
		}
	}

	bare, err := NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	got = &told{}
	bare.Tell(got)

	wantBare := []string{
		`Numbered(0)`, `Titled("Write the spec")`, `InColumn("backlog")`,
		`Carries("normal")`, `NotDescribed()`, `NotLabelled()`, `LabelsTold()`,
		`NotDue()`, `NotEstimated()`, `NothingSpent()`, `UnderNothing()`,
		`ServesNothing()`,
	}
	if len(got.said) != len(wantBare) {
		t.Fatalf("said = %v, want %v", got.said, wantBare)
	}
	for i := range wantBare {
		if got.said[i] != wantBare[i] {
			t.Errorf("said[%d] = %q, want %q", i, got.said[i], wantBare[i])
		}
	}
}

func TestNewTask_refusesAnEmptyTitle(t *testing.T) {
	if _, err := NewTask(""); !errors.Is(err, domain.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestNewTask_refusesATitleOfSpaces(t *testing.T) {
	if _, err := NewTask("   "); !errors.Is(err, domain.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired: spaces are not a title", err)
	}
}

// The sentinel is shared with an objective, the wording is not: a person told
// "an objective needs a title" after typing `task add` learns nothing.
func TestNewTask_saysItIsATaskThatNeedsATitle(t *testing.T) {
	_, err := NewTask("")
	if got := err.Error(); !strings.Contains(got, "a task needs a title") {
		t.Errorf("err = %q, want it to name the task", got)
	}
}

func TestNewTask_trimsWhatSurroundsTheTitle(t *testing.T) {
	got, err := NewTask("  Write the spec  ")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if got.title != "Write the spec" {
		t.Errorf("Title = %q, want it trimmed", got.title)
	}
}

func TestNewTask_keepsTheSpacesInsideTheTitle(t *testing.T) {
	got, err := NewTask("Write  the   spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if got.title != "Write  the   spec" {
		t.Errorf("Title = %q, want the inner spacing untouched", got.title)
	}
}

func TestNewTask_takesATitleAsTyped(t *testing.T) {
	for _, title := range []string{"Écrire la spéc — enfin !", "90% done, right?", "🧗 climb", "--not-a-flag"} {
		got, err := NewTask(title)
		if err != nil {
			t.Errorf("NewTask(%q): %v", title, err)
			continue
		}
		if got.title != title {
			t.Errorf("Title = %q, want %q", got.title, title)
		}
	}
}

func TestNewTask_doesNotTruncateALongTitle(t *testing.T) {
	long := strings.Repeat("a", 5000)
	got, err := NewTask(long)
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if len(got.title) != 5000 {
		t.Errorf("len(Title) = %d, want 5000", len(got.title))
	}
}

func TestNewTask_startsInTheBacklogAtNormalPriority(t *testing.T) {
	got, err := NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if got.state != domain.Backlog {
		t.Errorf("State = %q, want the backlog", got.state)
	}
	if got.priority != domain.Normal {
		t.Errorf("Priority = %q, want normal", got.priority)
	}
}

func TestNewTask_carriesNothingElse(t *testing.T) {
	got, err := NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if got.description != "" {
		t.Errorf("Description = %q, want none", got.description)
	}
	if len(got.labels) != 0 {
		t.Errorf("Labels = %v, want none", got.labels)
	}
	if !got.deadline.IsZero() {
		t.Errorf("Deadline = %v, want none", got.deadline)
	}
	if got.estimate != 0 {
		t.Errorf("Estimate = %v, want none", got.estimate)
	}
	if got.timeSpent != 0 {
		t.Errorf("TimeSpent = %v, want none", got.timeSpent)
	}
	if got.number != 0 {
		t.Errorf("Number = %d, want none: numbering is the store's business", got.number)
	}
	if got.parent != NoParent {
		t.Errorf("Parent = %d, want none: a task hangs under nothing unless told to", got.parent)
	}
}

func TestNewTask_servesNoObjectiveUnlessItIsTold(t *testing.T) {
	got, err := NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if got.objective != NoObjective {
		t.Errorf("Objective = %d, want none", got.objective)
	}
}

// A title is one line. A line break turns one row of a listing into two, the
// second carrying no number and belonging to nothing; a tab walks through the
// tabwriter that lines a card up and shifts columns it has no business in.
// Cleaning it up quietly would mean cairn keeping a title other than the one
// that was typed, which is the thing it refuses to do everywhere else.
func TestNewTask_refusesATitleThatIsNotOneLine(t *testing.T) {
	for _, title := range []string{"two\nlines", "a\ttab", "carriage\rreturn", "bell\a", "\aleading",
		"line\u2028separator", "para\u2029separator"} {
		if _, err := NewTask(title); !errors.Is(err, domain.ErrTitleNotOneLine) {
			t.Errorf("NewTask(%q) err = %v, want ErrTitleNotOneLine", title, err)
		}
	}
}

// TestTask_refusesToCarrySubTasksWhenItIsOneItself checks the depth rule from
// the side of the would-be parent: only one level, so a sub-task cannot
// itself carry sub-tasks.
func TestTask_refusesToCarrySubTasksWhenItIsOneItself(t *testing.T) {
	subTask := Task{number: 2, parent: 1}
	if err := subTask.MayCarrySubTasks(); !errors.Is(err, ErrParentIsSubTask) {
		t.Errorf("err = %v, want ErrParentIsSubTask", err)
	}
}

// TestTask_refusesToServeAnObjectiveWhileHangingUnderAParent checks the rule
// seen from the task itself: a sub-task lives with the task it hangs under,
// and is never given an objective of its own — but carrying either alone is
// still fine, so the refusal is about the two together.
func TestTask_refusesToServeAnObjectiveWhileHangingUnderAParent(t *testing.T) {
	for _, c := range []struct {
		name string
		task Task
		want error
	}{
		{"a parent and an objective at once", Task{number: 2, parent: 1, objective: 3}, ErrSubTaskServesNoObjective},
		{"an objective alone", Task{number: 2, objective: 3}, nil},
		{"a parent alone", Task{number: 2, parent: 1}, nil},
	} {
		if err := c.task.MayServeAnObjective(); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
}

// TestTask_refusesToHangUnderItself checks the rule from the task about to
// hang: it cannot name itself as its own parent.
func TestTask_refusesToHangUnderItself(t *testing.T) {
	self := Task{number: 1, parent: 1}
	if err := self.MayHangUnderItsParent(); !errors.Is(err, ErrParentIsSelf) {
		t.Errorf("err = %v, want ErrParentIsSelf", err)
	}
}

// TestMayDelete_refusesATaskCarryingSubTasks checks the rule that keeps a
// task's sub-tasks from being left pointing at a number nothing carries —
// and that a task with nothing hanging under it is not caught by it.
func TestMayDelete_refusesATaskCarryingSubTasks(t *testing.T) {
	for _, c := range []struct {
		name    string
		hanging []Task
		want    error
	}{
		{"something hangs under it", []Task{{number: 2, parent: 1}}, ErrCarriesSubTasks},
		{"nothing hangs under it", nil, nil},
	} {
		if err := MayDelete(1, c.hanging); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
}

// The spaces inside a title are still spaces: only control characters go.
func TestNewTask_keepsAnOrdinaryTitle(t *testing.T) {
	for _, title := range []string{"Write  the spec", "café — 100%", "🎯 ship it"} {
		if _, err := NewTask(title); err != nil {
			t.Errorf("NewTask(%q): %v", title, err)
		}
	}
}
