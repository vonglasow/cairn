package cli

import (
	"errors"
	"strings"
	"testing"

	"cairn/internal/reading"
)

// The `gui` command, without a display. What is worth testing is that it finds
// the store, opens it, and hands the window what it needs — a window that needs
// a screen is not something a test should need.

func TestGuiOpensTheStoreAndHandsItToTheWindow(t *testing.T) {
	// Asked while the command still holds it: the store is closed on the way
	// out, which is the point of opening it where it is needed.
	var readable error
	handed := false
	was := openGui
	openGui = func(work reading.Work, width, height int) error {
		handed = true
		_, readable = work.Version()
		if width <= 0 || height <= 0 {
			t.Errorf("the window was opened at %d by %d", width, height)
		}
		return nil
	}
	t.Cleanup(func() { openGui = was })

	got := run(t, "gui")
	if got.Code != 0 {
		t.Fatalf("gui exited %d: %s%s", got.Code, got.Stdout, got.Stderr)
	}
	if !handed {
		t.Fatal("the window was never opened")
	}
	if readable != nil {
		t.Errorf("the window was handed a store it cannot read: %v", readable)
	}
}

func TestGuiPassesOnWhatWentWrongOpeningTheWindow(t *testing.T) {
	was := openGui
	openGui = func(reading.Work, int, int) error { return errors.New("no display here") }
	t.Cleanup(func() { openGui = was })

	got := run(t, "gui")
	if got.Code == 0 {
		t.Fatal("a window that would not open exited 0")
	}
	if !strings.Contains(strings.ToLower(got.Stderr), "display") {
		t.Errorf("it said %q", got.Stderr)
	}
}

func TestGuiTakesNoArguments(t *testing.T) {
	got := run(t, "gui", "something")
	if got.Code == 0 {
		t.Fatal("gui took an argument it has no use for")
	}
	if !strings.Contains(got.Stderr, "gui") {
		t.Errorf("the refusal does not name the command: %q", got.Stderr)
	}
}
