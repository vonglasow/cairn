package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewTask_refusesAnEmptyTitle(t *testing.T) {
	if _, err := NewTask(""); !errors.Is(err, ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestNewTask_refusesATitleOfSpaces(t *testing.T) {
	if _, err := NewTask("   "); !errors.Is(err, ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired: spaces are not a title", err)
	}
}

func TestNewTask_trimsWhatSurroundsTheTitle(t *testing.T) {
	got, err := NewTask("  Write the spec  ")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if got.Title != "Write the spec" {
		t.Errorf("Title = %q, want it trimmed", got.Title)
	}
}

func TestNewTask_keepsTheSpacesInsideTheTitle(t *testing.T) {
	got, err := NewTask("Write  the   spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if got.Title != "Write  the   spec" {
		t.Errorf("Title = %q, want the inner spacing untouched", got.Title)
	}
}

func TestNewTask_takesATitleAsTyped(t *testing.T) {
	for _, title := range []string{"Écrire la spéc — enfin !", "90% done, right?", "🧗 climb", "--not-a-flag"} {
		got, err := NewTask(title)
		if err != nil {
			t.Errorf("NewTask(%q): %v", title, err)
			continue
		}
		if got.Title != title {
			t.Errorf("Title = %q, want %q", got.Title, title)
		}
	}
}

func TestNewTask_doesNotTruncateALongTitle(t *testing.T) {
	long := strings.Repeat("a", 5000)
	got, err := NewTask(long)
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if len(got.Title) != 5000 {
		t.Errorf("len(Title) = %d, want 5000", len(got.Title))
	}
}

func TestNewTask_startsInTheBacklogAtNormalPriority(t *testing.T) {
	got, err := NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if got.State != Backlog {
		t.Errorf("State = %q, want the backlog", got.State)
	}
	if got.Priority != Normal {
		t.Errorf("Priority = %q, want normal", got.Priority)
	}
}

func TestNewTask_carriesNothingElse(t *testing.T) {
	got, err := NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if got.Description != "" {
		t.Errorf("Description = %q, want none", got.Description)
	}
	if len(got.Labels) != 0 {
		t.Errorf("Labels = %v, want none", got.Labels)
	}
	if !got.Deadline.IsZero() {
		t.Errorf("Deadline = %v, want none", got.Deadline)
	}
	if got.Estimate != 0 {
		t.Errorf("Estimate = %v, want none", got.Estimate)
	}
	if got.TimeSpent != 0 {
		t.Errorf("TimeSpent = %v, want none", got.TimeSpent)
	}
	if got.Number != 0 {
		t.Errorf("Number = %d, want none: numbering is the store's business", got.Number)
	}
	if got.Parent != NoParent {
		t.Errorf("Parent = %d, want none: a task hangs under nothing unless told to", got.Parent)
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
		if _, err := NewTask(title); !errors.Is(err, ErrTitleNotOneLine) {
			t.Errorf("NewTask(%q) err = %v, want ErrTitleNotOneLine", title, err)
		}
	}
}

func TestNewObjective_refusesATitleThatIsNotOneLine(t *testing.T) {
	if _, err := NewObjective("two\nlines"); !errors.Is(err, ErrTitleNotOneLine) {
		t.Errorf("err = %v, want ErrTitleNotOneLine", err)
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
