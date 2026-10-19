package tui

import (
	"strings"
	"testing"

	"cairn/internal/engine"
)

func steppedTasks() *fake {
	f := newFake()
	f.cards = []engine.Card{
		{Number: 1, Title: "First", State: "todo", Steps: []engine.Part{{Number: 1, Title: "Draft", Done: true}, {Number: 2, Title: "Keep"}}},
		{Number: 2, Title: "Second", State: "current"},
	}
	return f
}

func openOnFirst(t *testing.T, f *fake) Board {
	t.Helper()
	b := hit(standing(f), "l", "i", "enter")
	if b.screen != OnDetail || b.task != 1 {
		t.Fatalf("the board is on %v with task %d open", b.screen, b.task)
	}
	return b
}

func says(b Board, want string) bool { return strings.Contains(strip(b.Draw()), want) }

func deletesAsked(f *fake) []engine.Input {
	var asked []engine.Input
	for _, a := range f.asked {
		if a.op.Thing == "step" && a.op.Verb == "delete" {
			asked = append(asked, a.in)
		}
	}
	return asked
}

func TestTheArrowsWalkTheStepsAsTheyAreShown(t *testing.T) {
	b := openOnFirst(t, steppedTasks())
	if says(b, "›") {
		t.Error("a step is chosen before any arrow")
	}
	for _, step := range []struct{ key, want string }{
		{"down", "› [ ] 2  Keep"}, {"j", "› [x] 1  Draft"}, {"down", "› [x] 1  Draft"},
		{"up", "› [ ] 2  Keep"}, {"k", ""},
	} {
		b = hit(b, step.key)
		if b.task != 1 {
			t.Fatalf("%s walked to task %d", step.key, b.task)
		}
		if step.want == "" && says(b, "›") || step.want != "" && !says(b, step.want) {
			t.Errorf("after %s the board says:\n%s", step.key, strip(b.Draw()))
		}
	}
}

func TestADeleteKeyAsksOnlyWithAStepChosen(t *testing.T) {
	b := openOnFirst(t, steppedTasks())
	for _, key := range []string{"backspace", "delete"} {
		if hit(b, key).asking {
			t.Errorf("%s with no step chosen asks something", key)
		}
	}
	for _, key := range []string{"left", "right", "h", "l"} {
		if moved := hit(b, "down", key); moved.task != 1 || !says(moved, "› [ ] 2  Keep") {
			t.Errorf("%s left the open task, or its step: task %d", key, moved.task)
		}
	}
	other := hit(b, "down", "esc", "l", "enter")
	if other.task != 2 || other.screen != OnDetail {
		t.Fatalf("the board opened task %d on %v", other.task, other.screen)
	}
	if hit(other, "down", "backspace").asking || says(hit(other, "down"), "›") {
		t.Error("a task with no steps chose one")
	}
	if back := hit(other, "esc", "h", "enter", "backspace"); back.task != 1 || back.asking {
		t.Error("the step chosen on a task was still chosen once it was opened again")
	}
}

func TestADeleteAgreedTakesTheChosenStepAway(t *testing.T) {
	f := steppedTasks()
	b := hit(openOnFirst(t, f), "down", "down")
	asked := hit(b, "delete")
	if !asked.asking || !says(asked, "delete step 1 of task 1, “Draft”?") {
		t.Fatalf("the delete key asked nothing, or not of Draft:\n%s", strip(asked.Draw()))
	}
	if kept := hit(asked, "esc"); kept.asking || len(deletesAsked(f)) != 0 {
		t.Error("esc deleted the step, or kept asking")
	}
	hit(asked, "enter")
	gone := deletesAsked(f)
	if len(gone) != 1 || gone[0].Given["number"] != "1" || gone[0].Given["step"] != "1" {
		t.Errorf("the engine was asked %+v", gone)
	}
}
