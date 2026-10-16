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
	if len(got.steps) != 0 {
		t.Errorf("Steps = %v, want none: a task is split into steps only when told to", got.steps)
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

func TestNewTask_keepsAnOrdinaryTitle(t *testing.T) {
	for _, title := range []string{"Write  the spec", "café — 100%", "🎯 ship it"} {
		if _, err := NewTask(title); err != nil {
			t.Errorf("NewTask(%q): %v", title, err)
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
