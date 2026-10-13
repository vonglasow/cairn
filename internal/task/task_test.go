package task

import (
	"errors"
	"strings"
	"testing"

	"cairn/internal/domain"
)

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
		if got := tc.task.Settled().state; got != tc.want {
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
		got := tc.task.Carried()
		if got.Archived != tc.want || got.State != tc.state {
			t.Errorf("%s: carries %+v, want archived %v in %s — archiving is not a column", tc.what, got, tc.want, tc.state)
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
	if err := open.Task().MayArchive(); err != nil {
		t.Errorf("MayArchive on one that is not = %v", err)
	}
}

func TestMayArchive_refusesAPartOnItsOwn(t *testing.T) {
	var a AsTold
	a.Numbered(2)
	a.Titled("Préparation")
	a.Under(1)
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
