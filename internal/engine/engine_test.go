package engine

import (
	"errors"
	"path/filepath"
	"testing"

	"cairn/internal/domain"
)

func openEngine(t *testing.T) *Engine {
	t.Helper()
	e, err := Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = e.Close() })
	return e
}

func add(t *testing.T, e *Engine, title string) domain.Task {
	t.Helper()
	task, err := e.Add(title, Changes{})
	if err != nil {
		t.Fatalf("Add(%q): %v", title, err)
	}
	return task
}

func of[T any](v T) *T { return &v }

func TestEngine_numbersTheFirstTaskOne(t *testing.T) {
	if got := add(t, openEngine(t), "Write the spec").Number; got != 1 {
		t.Errorf("Number = %d, want 1", got)
	}
}

func TestEngine_givesTheNextNumberToTheNextTask(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	if got := add(t, e, "Second").Number; got != 2 {
		t.Errorf("Number = %d, want 2", got)
	}
}

func TestEngine_refusesATaskWithNoTitle(t *testing.T) {
	if _, err := openEngine(t).Add("  ", Changes{}); !errors.Is(err, domain.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestEngine_carriesTheAttributesItWasGiven(t *testing.T) {
	got, err := openEngine(t).Add("Write the spec", Changes{
		Description: of("the one"),
		State:       of("todo"),
		Priority:    of("high"),
		Labels:      of([]string{"Team", "spec"}),
		Deadline:    of("2026-09-01"),
		Estimate:    of("90m"),
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if got.Description != "the one" || got.State != domain.Todo || got.Priority != domain.High {
		t.Errorf("got %+v, want the attributes it was given", got)
	}
	if len(got.Labels) != 2 || got.Labels[0] != "spec" || got.Labels[1] != "team" {
		t.Errorf("Labels = %v, want [spec team]: lower case, sorted", got.Labels)
	}
	if domain.FormatDay(got.Deadline) != "2026-09-01" {
		t.Errorf("Deadline = %v", got.Deadline)
	}
	if domain.FormatDuration(got.Estimate) != "1h30m" {
		t.Errorf("Estimate = %v, want 1h30m", got.Estimate)
	}
}

func TestEngine_refusesAnAttributeItCannotMakeSenseOf(t *testing.T) {
	for name, c := range map[string]Changes{
		"priority": {Priority: of("urgent")},
		"state":    {State: of("someday")},
		"deadline": {Deadline: of("soon")},
		"estimate": {Estimate: of("ages")},
		"label":    {Labels: of([]string{"  "})},
	} {
		if _, err := openEngine(t).Add("Write the spec", c); err == nil {
			t.Errorf("Add with a bad %s should fail", name)
		}
	}
}

func TestEngine_listsTasksInTheOrderTheyWereAdded(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	add(t, e, "Second")

	got, err := e.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 || got[0].Title != "First" || got[1].Title != "Second" {
		t.Errorf("List = %v", got)
	}
}

func TestEngine_showsATaskByItsNumber(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")

	got, err := e.Show(kept.Number)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if got.Title != "Write the spec" {
		t.Errorf("Title = %q", got.Title)
	}
}

func TestEngine_saysThereIsNoSuchTask(t *testing.T) {
	if _, err := openEngine(t).Show(99); err == nil {
		t.Error("Show(99) should fail")
	}
}

func TestEngine_changesOnlyWhatWasNamed(t *testing.T) {
	e := openEngine(t)
	kept, err := e.Add("Write the spec", Changes{Description: of("the one")})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := e.Edit(kept.Number, Changes{Priority: of("high")})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got.Priority != domain.High {
		t.Errorf("Priority = %q, want high", got.Priority)
	}
	if got.Description != "the one" {
		t.Errorf("Description = %q, want it left alone", got.Description)
	}
	if got.Title != "Write the spec" {
		t.Errorf("Title = %q, want it left alone", got.Title)
	}
}

func TestEngine_clearsAnAttributeGivenEmpty(t *testing.T) {
	e := openEngine(t)
	kept, err := e.Add("Write the spec", Changes{
		Deadline: of("2026-09-01"), Priority: of("critical"), State: of("current"),
		Labels: of([]string{"spec", "team"}),
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := e.Edit(kept.Number, Changes{
		Deadline: of(""), Priority: of(""), State: of(""), Labels: of([]string{""}),
	})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if !got.Deadline.IsZero() {
		t.Errorf("Deadline = %v, want it cleared", got.Deadline)
	}
	if got.Priority != domain.Normal {
		t.Errorf("Priority = %q, want normal", got.Priority)
	}
	if got.State != domain.Backlog {
		t.Errorf("State = %q, want the backlog", got.State)
	}
	if len(got.Labels) != 0 {
		t.Errorf("Labels = %v, want none", got.Labels)
	}
}

func TestEngine_refusesToEmptyATitle(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")
	if _, err := e.Edit(kept.Number, Changes{Title: of("")}); !errors.Is(err, domain.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestEngine_saysWhenNothingWasNamedToChange(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")
	if _, err := e.Edit(kept.Number, Changes{}); !errors.Is(err, ErrNothingToChange) {
		t.Errorf("err = %v, want ErrNothingToChange", err)
	}
}

func TestEngine_doesNotRenumberATaskItChanges(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	second := add(t, e, "Second")

	got, err := e.Edit(second.Number, Changes{Title: of("Second, revised")})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got.Number != second.Number {
		t.Errorf("Number = %d, want %d", got.Number, second.Number)
	}
}

func TestEngine_deletesATask(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")
	if err := e.Delete(kept.Number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := e.Show(kept.Number); err == nil {
		t.Error("the task should be gone")
	}
}

func TestEngine_saysThereIsNoSuchTaskToDelete(t *testing.T) {
	if err := openEngine(t).Delete(99); err == nil {
		t.Error("Delete(99) should fail")
	}
}

func TestEngine_neverHandsOutANumberFreedByADeletion(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	second := add(t, e, "Second")
	if err := e.Delete(second.Number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got := add(t, e, "Third").Number; got != 3 {
		t.Errorf("Number = %d, want 3", got)
	}
}

func TestOpen_reportsAStoreItCannotUse(t *testing.T) {
	if _, err := Open(filepath.Join(t.TempDir(), "no", "such", "\x00", "cairn.db")); err == nil {
		t.Error("Open should report a store it cannot use")
	}
}

func TestEngine_reportsAStoreItCanNoLongerReach(t *testing.T) {
	e, err := Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	kept, err := e.Add("Write the spec", Changes{})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := e.Add("Nope", Changes{}); err == nil {
		t.Error("Add on a closed store should fail")
	}
	if _, err := e.List(); err == nil {
		t.Error("List on a closed store should fail")
	}
	if _, err := e.Edit(kept.Number, Changes{Title: of("Nope")}); err == nil {
		t.Error("Edit on a closed store should fail")
	}
}

func TestEngine_addsALabelToTheOnesAlreadyThere(t *testing.T) {
	e := openEngine(t)
	kept, err := e.Add("Write the spec", Changes{Labels: of([]string{"spec", "team"})})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := e.Edit(kept.Number, Changes{Labels: of([]string{"+urgent"})})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if len(got.Labels) != 3 {
		t.Errorf("Labels = %v, want three", got.Labels)
	}
}

func TestEngine_takesALabelAwayFromTheOnesAlreadyThere(t *testing.T) {
	e := openEngine(t)
	kept, err := e.Add("Write the spec", Changes{Labels: of([]string{"spec", "team"})})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := e.Edit(kept.Number, Changes{Labels: of([]string{"-spec"})})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if len(got.Labels) != 1 || got.Labels[0] != "team" {
		t.Errorf("Labels = %v, want [team]", got.Labels)
	}
}

func TestEngine_refusesASignedLabelAtCreation(t *testing.T) {
	_, err := openEngine(t).Add("Write the spec", Changes{Labels: of([]string{"+urgent"})})
	if !errors.Is(err, ErrLabelSignAtCreation) {
		t.Errorf("err = %v, want ErrLabelSignAtCreation: a new task has nothing to add to", err)
	}
}

func TestEngine_refusesToReplaceAndChangeLabelsAtOnce(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")
	if _, err := e.Edit(kept.Number, Changes{Labels: of([]string{"+urgent", "draft"})}); err == nil {
		t.Error("Edit should refuse to both replace and change the labels")
	}
}
