package objective

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
)

func TestObjectiveStates_areExactlyWhatParseObjectiveStateAccepts(t *testing.T) {
	for _, name := range []string{"active", "on-hold", "reached", "dropped"} {
		if !slices.Contains(ObjectiveStates, State(name)) {
			t.Errorf("ObjectiveStates is missing %q, though ParseObjectiveState accepts it", name)
		}
	}
	for _, s := range ObjectiveStates {
		if _, err := ParseObjectiveState(string(s)); err != nil {
			t.Errorf("ParseObjectiveState(%q): %v, want every entry of ObjectiveStates accepted", s, err)
		}
	}
	if slices.Contains(ObjectiveStates, State("todo")) {
		t.Error("ObjectiveStates should not hold todo: ParseObjectiveState refuses it")
	}
}

func TestAsTold_buildsAnObjectiveFromEveryFactItIsTold(t *testing.T) {
	deadline, err := domain.ParseDeadline("2026-01-02")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}

	want := Objective{
		number: 3, title: "Ship cairn 1.0", description: "the one that unblocks the rest",
		state: OnHold, priority: domain.High, labels: []string{"perso", "spec"},
		deadline: deadline, estimate: 90 * time.Minute,
	}

	a := &AsTold{}
	*a = From(want)
	got := a.Objective()

	if got.number != want.number {
		t.Errorf("Number = %d, want %d", got.number, want.number)
	}
	if got.title != want.title {
		t.Errorf("Title = %q, want %q", got.title, want.title)
	}
	if got.description != want.description {
		t.Errorf("Description = %q, want %q", got.description, want.description)
	}
	if got.state != want.state {
		t.Errorf("State = %q, want %q", got.state, want.state)
	}
	if got.priority != want.priority {
		t.Errorf("Priority = %q, want %q", got.priority, want.priority)
	}
	if len(got.labels) != len(want.labels) {
		t.Fatalf("Labels = %v, want %v", got.labels, want.labels)
	}
	for i := range want.labels {
		if got.labels[i] != want.labels[i] {
			t.Errorf("Labels[%d] = %q, want %q", i, got.labels[i], want.labels[i])
		}
	}
	if !got.deadline.Equal(want.deadline) {
		t.Errorf("Deadline = %v, want %v", got.deadline, want.deadline)
	}
	if got.estimate != want.estimate {
		t.Errorf("Estimate = %v, want %v", got.estimate, want.estimate)
	}
}

func TestNewObjective_refusesAnEmptyTitle(t *testing.T) {
	if _, err := NewObjective(""); !errors.Is(err, domain.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestNewObjective_refusesATitleOfSpaces(t *testing.T) {
	if _, err := NewObjective("   "); !errors.Is(err, domain.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired: spaces are not a title", err)
	}
}

func TestNewObjective_saysItIsAnObjectiveThatNeedsATitle(t *testing.T) {
	_, err := NewObjective("")
	if got := err.Error(); !strings.Contains(got, "an objective needs a title") {
		t.Errorf("err = %q, want it to name the objective", got)
	}
}

func TestNewObjective_trimsWhatSurroundsTheTitle(t *testing.T) {
	got, err := NewObjective("  Ship cairn 1.0  ")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	if got.title != "Ship cairn 1.0" {
		t.Errorf("Title = %q, want it trimmed", got.title)
	}
}

func TestNewObjective_keepsTheSpacesInsideTheTitle(t *testing.T) {
	got, err := NewObjective("Ship  cairn   1.0")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	if got.title != "Ship  cairn   1.0" {
		t.Errorf("Title = %q, want the inner spacing untouched", got.title)
	}
}

func TestNewObjective_startsActiveAtNormalPriority(t *testing.T) {
	got, err := NewObjective("Ship cairn 1.0")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	if got.state != Active {
		t.Errorf("State = %q, want %q", got.state, Active)
	}
	if got.priority != domain.Normal {
		t.Errorf("Priority = %q, want %q", got.priority, domain.Normal)
	}
}

func TestNewObjective_carriesNothingElse(t *testing.T) {
	got, err := NewObjective("Ship cairn 1.0")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	if got.number != 0 || got.description != "" || len(got.labels) != 0 {
		t.Errorf("%+v carries more than it was given", got)
	}
	if !got.deadline.IsZero() || got.estimate != 0 {
		t.Errorf("%+v carries a deadline or an estimate it was not given", got)
	}
}

func TestParseObjectiveState_readsEveryOne(t *testing.T) {
	for _, want := range []State{Active, OnHold, Reached, Dropped} {
		got, err := ParseObjectiveState(string(want))
		if err != nil {
			t.Errorf("ParseObjectiveState(%q): %v", want, err)
			continue
		}
		if got != want {
			t.Errorf("ParseObjectiveState(%q) = %q", want, got)
		}
	}
}

func TestParseObjectiveState_readsNothingAsActive(t *testing.T) {
	got, err := ParseObjectiveState("")
	if err != nil {
		t.Fatalf("ParseObjectiveState(%q): %v", "", err)
	}
	if got != Active {
		t.Errorf("ParseObjectiveState(%q) = %q, want %q", "", got, Active)
	}
}

func TestParseObjectiveState_refusesATasksState(t *testing.T) {
	for _, name := range []string{"backlog", "todo", "current", "done", "nope"} {
		if _, err := ParseObjectiveState(name); !errors.Is(err, ErrNotAnObjectiveState) {
			t.Errorf("ParseObjectiveState(%q) err = %v, want ErrNotAnObjectiveState", name, err)
		}
	}
}

func TestParseObjectiveState_saysWhatItWasGiven(t *testing.T) {
	_, err := ParseObjectiveState("todo")
	if got := err.Error(); !strings.Contains(got, `"todo"`) {
		t.Errorf("err = %q, want it to quote what it was given", got)
	}
}

func TestParseState_refusesAnObjectivesState(t *testing.T) {
	for _, name := range []string{"active", "on-hold", "reached", "dropped"} {
		if _, err := domain.ParseState(name); !errors.Is(err, domain.ErrNotAState) {
			t.Errorf("ParseState(%q) err = %v, want ErrNotAState", name, err)
		}
	}
}

func TestNewObjective_refusesATitleThatIsNotOneLine(t *testing.T) {
	if _, err := NewObjective("two\nlines"); !errors.Is(err, domain.ErrTitleNotOneLine) {
		t.Errorf("err = %v, want ErrTitleNotOneLine", err)
	}
}

func TestMayDelete_refusesAnObjectiveCarryingTasks(t *testing.T) {
	for _, c := range []struct {
		name    string
		serving int
		want    error
	}{
		{"tasks serve it", 2, ErrObjectiveCarriesTasks},
		{"nothing serves it", 0, nil},
	} {
		if err := MayDelete(1, c.serving); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
}
