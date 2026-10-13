package task

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
)

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
func (r *told) Archived()          { r.said = append(r.said, "Archived()") }
func (r *told) NotArchived()       { r.said = append(r.said, "NotArchived()") }
func (r *told) Described(s string) { r.said = append(r.said, fmt.Sprintf("Described(%q)", s)) }
func (r *told) NotDescribed()      { r.said = append(r.said, "NotDescribed()") }
func (r *told) Labelled(s string)  { r.said = append(r.said, fmt.Sprintf("Labelled(%q)", s)) }
func (r *told) NotLabelled()       { r.said = append(r.said, "NotLabelled()") }
func (r *told) LabelsTold()        { r.said = append(r.said, "LabelsTold()") }
func (r *told) BlockedBy(n int)    { r.said = append(r.said, fmt.Sprintf("BlockedBy(%d)", n)) }
func (r *told) NotBlocked()        { r.said = append(r.said, "NotBlocked()") }
func (r *told) BlockersTold()      { r.said = append(r.said, "BlockersTold()") }
func (r *told) Because(s string)   { r.said = append(r.said, fmt.Sprintf("Because(%q)", s)) }
func (r *told) NotBecause()        { r.said = append(r.said, "NotBecause()") }
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
func (r *told) ObjectivesTold()       { r.said = append(r.said, "ObjectivesTold()") }

func TestTell_saysEveryFactATaskCarries(t *testing.T) {
	deadline, err := domain.ParseDeadline("2026-01-02")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}

	full := Task{
		number: 3, parent: 1, objectives: []int{2}, title: "Write the spec",
		description: "Say what it does", state: domain.Current, priority: domain.High,
		labels: []string{"backend", "urgent"}, blockers: []int{4, 5},
		because: "the client's answer", deadline: deadline,
		estimate: 90 * time.Minute, timeSpent: 30 * time.Minute,
	}
	got := &told{}
	full.Tell(got)

	want := []string{
		`Numbered(3)`, `Titled("Write the spec")`, `InColumn("current")`,
		`NotArchived()`, `Carries("high")`, `Described("Say what it does")`,
		`Labelled("backend")`, `Labelled("urgent")`, `LabelsTold()`,
		`BlockedBy(4)`, `BlockedBy(5)`, `BlockersTold()`,
		`Because("the client's answer")`,
		fmt.Sprintf("Due(%s)", deadline), fmt.Sprintf("Estimated(%s)", 90*time.Minute),
		fmt.Sprintf("Spent(%s)", 30*time.Minute), `Under(1)`, `Serves(2)`, `ObjectivesTold()`,
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
		`NotArchived()`, `Carries("normal")`, `NotDescribed()`, `NotLabelled()`, `LabelsTold()`,
		`NotBlocked()`, `BlockersTold()`, `NotBecause()`,
		`NotDue()`, `NotEstimated()`, `NothingSpent()`, `UnderNothing()`,
		`ServesNothing()`, `ObjectivesTold()`,
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
	if len(got.objectives) != 0 {
		t.Errorf("Objectives = %v, want none", got.objectives)
	}
}

func TestNewTask_refusesATitleThatIsNotOneLine(t *testing.T) {
	for _, title := range []string{"two\nlines", "a\ttab", "carriage\rreturn", "bell\a", "\aleading",
		"line\u2028separator", "para\u2029separator"} {
		if _, err := NewTask(title); !errors.Is(err, domain.ErrTitleNotOneLine) {
			t.Errorf("NewTask(%q) err = %v, want ErrTitleNotOneLine", title, err)
		}
	}
}

func TestTask_refusesToCarrySubTasksWhenItIsOneItself(t *testing.T) {
	subTask := Task{number: 2, parent: 1}
	if err := subTask.MayCarrySubTasks(); !errors.Is(err, ErrParentIsSubTask) {
		t.Errorf("err = %v, want ErrParentIsSubTask", err)
	}
}

func TestTask_refusesToServeAnObjectiveWhileHangingUnderAParent(t *testing.T) {
	for _, c := range []struct {
		name string
		task Task
		want error
	}{
		{"a parent and an objective at once", Task{number: 2, parent: 1, objectives: []int{3}}, ErrSubTaskServesNoObjective},
		{"an objective alone", Task{number: 2, objectives: []int{3}}, nil},
		{"a parent alone", Task{number: 2, parent: 1}, nil},
	} {
		if err := c.task.MayServeAnObjective(); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
}

func TestTask_refusesToHangUnderItself(t *testing.T) {
	self := Task{number: 1, parent: 1}
	if err := self.MayHangUnderItsParent(); !errors.Is(err, ErrParentIsSelf) {
		t.Errorf("err = %v, want ErrParentIsSelf", err)
	}
}

func TestNewTask_keepsAnOrdinaryTitle(t *testing.T) {
	for _, title := range []string{"Write  the spec", "café — 100%", "🎯 ship it"} {
		if _, err := NewTask(title); err != nil {
			t.Errorf("NewTask(%q): %v", title, err)
		}
	}
}

func TestMayFinish_refusesWhileSomethingUnderItIsOpen(t *testing.T) {
	open := func(state domain.State) Task {
		var a AsTold
		a.Numbered(2)
		a.Titled("Write the spec")
		a.Under(1)
		a.InColumn(state)
		return a.Task()
	}
	done := open(domain.Done)
	for _, tc := range []struct {
		what    string
		hanging []Task
		wantErr bool
	}{
		{"nothing under it", nil, false},
		{"one still in the backlog", []Task{open(domain.Backlog)}, true},
		{"one being worked on", []Task{open(domain.Current)}, true},
		{"one blocked", []Task{open(domain.Blocked)}, true},
		{"one paused", []Task{open(domain.Paused)}, true},
		{"the only one done", []Task{done}, false},
		{"one done and one open", []Task{done, open(domain.Todo)}, true},
		{"both done", []Task{done, done}, false},
	} {
		err := MayFinish(1, tc.hanging)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: MayFinish = %v, want an error: %v", tc.what, err, tc.wantErr)
		}
		if tc.wantErr && !errors.Is(err, ErrSubTaskStillOpen) {
			t.Errorf("%s: MayFinish = %v, want it to be ErrSubTaskStillOpen", tc.what, err)
		}
	}
}

func TestMayFinish_namesWhatIsInTheWay(t *testing.T) {
	var a AsTold
	a.Numbered(7)
	a.Titled("Write the spec")
	a.Under(1)
	a.InColumn(domain.Todo)
	err := MayFinish(1, []Task{a.Task()})
	if err == nil || !strings.Contains(err.Error(), "7") {
		t.Errorf("MayFinish = %v, and says nothing of task 7", err)
	}
}

func TestMayCarryItsColumn_holdsASubTaskToDoneOrNot(t *testing.T) {
	under := func(state domain.State) Task {
		var a AsTold
		a.Numbered(2)
		a.Titled("Write the tests")
		a.Under(1)
		a.InColumn(state)
		return a.Task()
	}
	root := func(state domain.State) Task {
		var a AsTold
		a.Numbered(1)
		a.Titled("Ship it")
		a.InColumn(state)
		return a.Task()
	}
	for _, tc := range []struct {
		what    string
		task    Task
		wantErr bool
	}{
		{"a sub-task done", under(domain.Done), false},
		{"a sub-task not done", under(domain.Todo), false},
		{"a sub-task in the backlog", under(domain.Backlog), true},
		{"a sub-task being worked on", under(domain.Current), true},
		{"a sub-task blocked", under(domain.Blocked), true},
		{"a sub-task watching", under(domain.Watching), true},
		{"a sub-task paused", under(domain.Paused), true},
		{"a task in any column it likes", root(domain.Current), false},
		{"a task blocked", root(domain.Blocked), false},
	} {
		err := tc.task.MayCarryItsColumn()
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: MayCarryItsColumn = %v, want an error: %v", tc.what, err, tc.wantErr)
		}
		if tc.wantErr && !errors.Is(err, ErrSubTaskHasNoColumn) {
			t.Errorf("%s: %v, want ErrSubTaskHasNoColumn", tc.what, err)
		}
	}
}

func TestSettled_keepsDoneAndDropsEveryOtherColumn(t *testing.T) {
	made := func(parent int, state domain.State) Task {
		var a AsTold
		a.Numbered(2)
		a.Titled("Write the tests")
		if parent != NoParent {
			a.Under(parent)
		}
		a.InColumn(state)
		return a.Task()
	}
	for _, tc := range []struct {
		what string
		task Task
		want domain.State
	}{
		{"a sub-task being worked on", made(1, domain.Current), domain.Todo},
		{"a sub-task in the backlog", made(1, domain.Backlog), domain.Todo},
		{"a sub-task blocked", made(1, domain.Blocked), domain.Todo},
		{"a sub-task already done", made(1, domain.Done), domain.Done},
		{"a sub-task already todo", made(1, domain.Todo), domain.Todo},
		{"a task hanging under nothing is left alone", made(NoParent, domain.Current), domain.Current},
		{"a task hanging under nothing, done", made(NoParent, domain.Done), domain.Done},
	} {
		var told AsTold
		tc.task.Settled().Tell(&told)
		if got := told.Task().state; got != tc.want {
			t.Errorf("%s: Settled = %q, want %q", tc.what, got, tc.want)
		}
	}
}

func TestArchived_isCarriedBesideTheColumn(t *testing.T) {
	made := func(state domain.State, archived bool) Task {
		var a AsTold
		a.Numbered(1)
		a.Titled("Migration Jenkins")
		a.InColumn(state)
		if archived {
			a.Archived()
		} else {
			a.NotArchived()
		}
		return a.Task()
	}
	for _, tc := range []struct {
		what  string
		task  Task
		state domain.State
		want  bool
	}{
		{"done and archived", made(domain.Done, true), domain.Done, true},
		{"done and not", made(domain.Done, false), domain.Done, false},
		{"never done and archived", made(domain.Backlog, true), domain.Backlog, true},
	} {
		var back AsTold
		tc.task.Tell(&back)
		got := back.Task()
		var read told
		got.Tell(&read)
		want := "Archived()"
		if !tc.want {
			want = "NotArchived()"
		}
		if !slices.Contains(read.said, want) {
			t.Errorf("%s: told %v, want it to say %s", tc.what, read.said, want)
		}
		if said := fmt.Sprintf("InColumn(%q)", tc.state); !slices.Contains(read.said, said) {
			t.Errorf("%s: told %v, want %s — archiving is not a column", tc.what, read.said, said)
		}
	}
}

func TestMayArchive_refusesWhatIsAlreadyArchived(t *testing.T) {
	var a AsTold
	a.Numbered(1)
	a.Titled("Migration Jenkins")
	a.Archived()
	if err := a.Task().MayArchive(); !errors.Is(err, ErrAlreadyArchived) {
		t.Errorf("MayArchive = %v, want it to say the task is already archived", err)
	}
	var open AsTold
	open.Numbered(2)
	open.Titled("Spike Argo")
	open.NotArchived()
	if err := open.Task().MayArchive(); err != nil {
		t.Errorf("MayArchive on one that is not = %v", err)
	}
}

func TestMayArchive_refusesAPartOnItsOwn(t *testing.T) {
	var a AsTold
	a.Numbered(2)
	a.Titled("Préparation")
	a.Under(1)
	a.NotArchived()
	err := a.Task().MayArchive()
	if !errors.Is(err, ErrPartGoesWithItsWork) {
		t.Errorf("MayArchive = %v, want it to say the part goes with its work", err)
	}
	if err == nil || !strings.Contains(err.Error(), "task 1") {
		t.Errorf("MayArchive = %v, want it to name the work it belongs to", err)
	}
}

func TestMayUnarchive_refusesWhatIsOnTheBoard(t *testing.T) {
	var here AsTold
	here.Numbered(1)
	here.Titled("Migration Jenkins")
	here.NotArchived()
	if err := here.Task().MayUnarchive(); !errors.Is(err, ErrNotArchived) {
		t.Errorf("MayUnarchive = %v, want it to say the task is on the board", err)
	}
	var away AsTold
	away.Numbered(2)
	away.Titled("Spike Argo")
	away.Archived()
	if err := away.Task().MayUnarchive(); err != nil {
		t.Errorf("MayUnarchive on an archived task = %v, want it allowed", err)
	}
}
