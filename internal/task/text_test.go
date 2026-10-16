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
	a := From(built)
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
	a := From(built)
	a.Described("the one")
	a.Labelled("spec")
	a.Labelled("team")
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
	got := Text(spec(t), domain.Progress{Steps: domain.Tally{Total: 4, Done: 1}})
	if !strings.Contains(got, "1 of 4 steps are done") {
		t.Errorf("text is missing the progress line:\n%s", got)
	}
	if strings.Contains(got, "0 of 0 tasks") {
		t.Errorf("text should not count tasks it does not have:\n%s", got)
	}
}

func TestText_countsTasksApartFromSteps(t *testing.T) {
	got := Text(spec(t), domain.Progress{Tasks: domain.Tally{Total: 2, Done: 2}})
	if !strings.Contains(got, "2 of 2 tasks are done") {
		t.Errorf("text is missing the progress line:\n%s", got)
	}
	if strings.Contains(got, "0 of 0 steps") {
		t.Errorf("text should not count steps it does not have:\n%s", got)
	}
}

func TestLine_showsTheNumberAndTheTitle(t *testing.T) {
	if got := Line(spec(t)); !strings.Contains(got, "1") || !strings.Contains(got, "Write the spec") {
		t.Errorf("line = %q, want the number and the title", got)
	}
}

func listed(number int, title string) Task {
	var a AsTold
	a.Numbered(number)
	a.Titled(title)
	return a.Task()
}

func TestLines_writesNothingForNoTasks(t *testing.T) {
	if got := Lines(nil, Marks{}); got != "" {
		t.Errorf("Lines(nil, Marks{}) = %q, want nothing", got)
	}
}

func TestLines_marksTheTasksInFocus(t *testing.T) {
	tasks := []Task{
		listed(1, "Ship it"),
		listed(2, "Write the tests"),
		listed(3, "Something else"),
	}
	for _, inFocus := range []int{1, 2, 3} {
		got := Lines(tasks, Marks{InFocus: []int{inFocus}})
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

func TestLines_marksEveryTaskInFocus(t *testing.T) {
	tasks := []Task{
		listed(1, "Ship it"),
		listed(2, "Write the tests"),
		listed(3, "Something else"),
	}
	got := Lines(tasks, Marks{InFocus: []int{1, 3}})
	if marks := strings.Count(got, InFocusMark); marks != 2 {
		t.Errorf("with two in focus, %d lines are marked, want two:\n%s", marks, got)
	}
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "Write the tests") && strings.Contains(line, InFocusMark) {
			t.Errorf("a task that is not in focus is marked: %q", line)
		}
	}
}

func TestLines_marksNothingWhenNothingIsInFocus(t *testing.T) {
	got := Lines([]Task{listed(1, "Ship it")}, Marks{})
	if strings.Contains(got, InFocusMark) {
		t.Errorf("nothing is in focus, yet a line is marked:\n%s", got)
	}
}

func TestText_carriesTheAttributesMostTasksDoNot(t *testing.T) {
	built, err := NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	a := From(built)
	a.Numbered(2)
	a.Spent(90 * time.Minute)
	a.Serves(3)

	got := Text(a.Task(), domain.Progress{})
	for _, want := range []string{"time spent", "1h30m", "objective"} {
		if !strings.Contains(got, want) {
			t.Errorf("the card does not carry %q:\n%s", want, got)
		}
	}
}

func TestStepsText_keepsTheOrderItWasGiven(t *testing.T) {
	got := StepsText([]Step{{Title: "Write the docs"}, {Title: "Write the tests"}})
	if docs, tests := strings.Index(got, "Write the docs"), strings.Index(got, "Write the tests"); docs > tests {
		t.Errorf("StepsText reordered them:\n%s", got)
	}
}

func TestText_saysWhenATaskIsOffTheBoard(t *testing.T) {
	made := func(archived bool) Task {
		var a AsTold
		a.Numbered(1)
		a.Titled("Migration Jenkins")
		a.InColumn(domain.Done)
		if archived {
			a.Archived()
		}
		return a.Task()
	}
	if got := Text(made(true), domain.Progress{}); !strings.Contains(got, "archived") {
		t.Errorf("Text of an archived task =\n%s\nwant it to say so", got)
	}
	if got := Text(made(false), domain.Progress{}); strings.Contains(got, "archived") {
		t.Errorf("Text of one on the board =\n%s\nwant it to say nothing of archiving", got)
	}
}

func TestALineCarriesEnoughToDecideOn(t *testing.T) {
	var a AsTold
	a.Numbered(12)
	a.Titled("Write the spec")
	a.InColumn(domain.Current)
	a.Carries(domain.High)
	a.Due(time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC))
	a.Labelled("backend")
	a.Labelled("sprint-172")
	line := Line(a.Task())
	for _, said := range []string{"Write the spec", "current", "high", "2026-10-09", "backend", "sprint-172"} {
		if !strings.Contains(line, said) {
			t.Errorf("the line does not say %q: %q", said, line)
		}
	}
}
