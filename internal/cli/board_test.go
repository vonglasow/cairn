package cli

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"cairn/internal/board"
)

// TestBoard_refusesGlyphsItCannotMakeSenseOf checks that `board --icons` names
// what was typed and says what it takes instead, rather than quietly drawing
// something nobody asked for.
func TestBoard_refusesGlyphsItCannotMakeSenseOf(t *testing.T) {
	r := run(t, "board", "--icons", "wingdings")
	if r.Code != exitRefused {
		t.Fatalf("board --icons wingdings exit = %d, want %d (%s)", r.Code, exitRefused, r.Stderr)
	}
	if !strings.Contains(strings.ToLower(r.Stderr), "wingdings") {
		t.Errorf("stderr = %q, want it to name what was typed", r.Stderr)
	}
	if !strings.Contains(strings.ToLower(r.Stderr), "nerdfont") {
		t.Errorf("stderr = %q, want it to say what it takes instead", r.Stderr)
	}
}

// TestBoard_takesNoArguments checks that a stray word is refused rather than
// passed over.
func TestBoard_takesNoArguments(t *testing.T) {
	r := run(t, "board", "tasks")
	if r.Code == exitOK {
		t.Fatalf("board with a stray argument exited %d", r.Code)
	}
}

// TestBoard_opensOverTheStore drives the command with the program stubbed out:
// what is worth checking here is that the board is built over the store this
// command was pointed at, and starting a real bubbletea program needs a
// terminal that a test has no business demanding.
func TestBoard_opensOverTheStore(t *testing.T) {
	was := openBoard
	t.Cleanup(func() { openBoard = was })

	var opened tea.Model
	openBoard = func(_ *cobra.Command, model tea.Model) error {
		opened = model
		return nil
	}
	if r := run(t, "board"); r.Code != exitOK {
		t.Fatalf("board exit = %d (%s)", r.Code, r.Stderr)
	}
	if _, ok := opened.(board.Board); !ok {
		t.Fatalf("what was opened is %T, want a board", opened)
	}
}

// TestBoard_saysWhatWentWrongWhileItWasOpen checks that a board that stopped
// on an error stops the command with it, rather than exiting zero on a screen
// that went away.
func TestBoard_saysWhatWentWrongWhileItWasOpen(t *testing.T) {
	was := openBoard
	t.Cleanup(func() { openBoard = was })

	openBoard = func(*cobra.Command, tea.Model) error {
		return errors.New("the terminal went away")
	}
	r := run(t, "board")
	if r.Code == exitOK {
		t.Fatal("a board that stopped on an error exited as though nothing had happened")
	}
	if !strings.Contains(strings.ToLower(r.Stderr), "terminal went away") {
		t.Errorf("stderr = %q, want it to carry what went wrong", r.Stderr)
	}
}

// TestBoard_refusesAStoreItCannotOpen checks that the store is opened before
// anything is drawn, so a store that is behind is said at the terminal rather
// than inside a screen nobody can read it in.
func TestBoard_refusesAStoreItCannotOpen(t *testing.T) {
	was := openBoard
	t.Cleanup(func() { openBoard = was })
	openBoard = func(*cobra.Command, tea.Model) error {
		t.Fatal("the board was opened over a store that could not be")
		return nil
	}
	t.Setenv("CAIRN_STORE", t.TempDir())
	if r := run(t, "board"); r.Code == exitOK {
		t.Fatalf("board over a directory exited %d", r.Code)
	}
}
