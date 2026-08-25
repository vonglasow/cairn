package objective

import (
	"errors"
	"fmt"
	"slices"
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
func (r *told) Standing(s State) {
	r.said = append(r.said, fmt.Sprintf("Standing(%q)", s))
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
func (r *told) NotEstimated() { r.said = append(r.said, "NotEstimated()") }

// TestTell_saysEveryFactAnObjectiveCarries checks that Tell says every fact an
// objective carries, in the card's order, and says the absence of a fact
// rather than staying silent about it.
func TestTell_saysEveryFactAnObjectiveCarries(t *testing.T) {
	deadline, err := domain.ParseDeadline("2026-01-02")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}

	full := Objective{
		number: 3, title: "Ship cairn 1.0", description: "the one that unblocks the rest",
		state: OnHold, priority: domain.High, labels: []string{"perso", "spec"},
		deadline: deadline, estimate: 90 * time.Minute,
	}
	got := &told{}
	full.Tell(got)

	want := []string{
		`Numbered(3)`, `Titled("Ship cairn 1.0")`, `Standing("on-hold")`,
		`Carries("high")`, `Described("the one that unblocks the rest")`,
		`Labelled("perso")`, `Labelled("spec")`, `LabelsTold()`,
		fmt.Sprintf("Due(%s)", deadline), fmt.Sprintf("Estimated(%s)", 90*time.Minute),
	}
	if len(got.said) != len(want) {
		t.Fatalf("said = %v, want %v", got.said, want)
	}
	for i := range want {
		if got.said[i] != want[i] {
			t.Errorf("said[%d] = %q, want %q", i, got.said[i], want[i])
		}
	}

	bare, err := NewObjective("Ship cairn 1.0")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	got = &told{}
	bare.Tell(got)

	wantBare := []string{
		`Numbered(0)`, `Titled("Ship cairn 1.0")`, `Standing("active")`,
		`Carries("normal")`, `NotDescribed()`, `NotLabelled()`, `LabelsTold()`,
		`NotDue()`, `NotEstimated()`,
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

// Same proof as for States and Priorities: the list and the parser cannot
// say different things about what an objective's state is.
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

// TestAsTold_buildsAnObjectiveFromEveryFactItIsTold checks that an objective
// told to a builder, fact by fact, is handed back whole: this is the same
// membrane Tell uses, the other way round, so what goes in has to come out
// unchanged.
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
	want.Tell(a)
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

// The four tests below each poison the field first, through the fact that
// carries a value, before telling AsTold the fact says there is none. A
// builder that dropped the absence on the floor would still hold the
// poisoned value afterwards.

func TestAsTold_notDescribedMeansNoDescription(t *testing.T) {
	a := &AsTold{}
	a.Described("the one that unblocks the rest")
	a.NotDescribed()
	if got := a.Objective().description; got != "" {
		t.Errorf("Description = %q, want none", got)
	}
}

func TestAsTold_notDueMeansNoDeadline(t *testing.T) {
	a := &AsTold{}
	a.Due(time.Now())
	a.NotDue()
	if got := a.Objective().deadline; !got.IsZero() {
		t.Errorf("Deadline = %v, want none", got)
	}
}

func TestAsTold_notEstimatedMeansNoEstimate(t *testing.T) {
	a := &AsTold{}
	a.Estimated(90 * time.Minute)
	a.NotEstimated()
	if got := a.Objective().estimate; got != 0 {
		t.Errorf("Estimate = %v, want none", got)
	}
}

func TestAsTold_notLabelledMeansNoLabels(t *testing.T) {
	a := &AsTold{}
	a.Labelled("spec")
	a.NotLabelled()
	a.LabelsTold()
	if got := a.Objective().labels; len(got) != 0 {
		t.Errorf("Labels = %v, want none", got)
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

// The sentinel is shared with a task, the wording is not: a person told "a
// task needs a title" after typing `objective add` learns nothing. The task
// side of that sharing is task.TestNewTask_saysItIsATaskThatNeedsATitle.
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

// An objective has no column, so it cannot be put in one. A task's state given
// to an objective is the mistake most likely to be made, and the one a shared
// parser would have accepted in silence.
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

// The other way round: a column is still a column, and the states an
// objective can be in are not among them.
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

// TestMayDelete_refusesAnObjectiveCarryingTasks checks the rule that keeps an
// objective's tasks from being left pointing at a number nothing carries —
// and that an objective nothing serves is not caught by it.
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
