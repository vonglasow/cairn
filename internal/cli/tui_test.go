package cli

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"cairn/internal/engine"
	"cairn/internal/tui"
)

// TestTui_refusesGlyphsItCannotMakeSenseOf checks that `tui --icons` names
// what was typed and says what it takes instead, rather than quietly drawing
// something nobody asked for.
func TestTui_refusesGlyphsItCannotMakeSenseOf(t *testing.T) {
	r := run(t, "tui", "--icons", "wingdings")
	if r.Code != exitRefused {
		t.Fatalf("tui --icons wingdings exit = %d, want %d (%s)", r.Code, exitRefused, r.Stderr)
	}
	if !strings.Contains(strings.ToLower(r.Stderr), "wingdings") {
		t.Errorf("stderr = %q, want it to name what was typed", r.Stderr)
	}
	if !strings.Contains(strings.ToLower(r.Stderr), "nerdfont") {
		t.Errorf("stderr = %q, want it to say what it takes instead", r.Stderr)
	}
}

// TestTui_takesNoArguments checks that a stray word is refused rather than
// passed over.
func TestTui_takesNoArguments(t *testing.T) {
	r := run(t, "tui", "tasks")
	if r.Code == exitOK {
		t.Fatalf("tui with a stray argument exited %d", r.Code)
	}
}

// TestTui_opensOverTheStore drives the command with the program stubbed out:
// what is worth checking here is that the board is built over the store this
// command was pointed at, and starting a real bubbletea program needs a
// terminal that a test has no business demanding.
func TestTui_opensOverTheStore(t *testing.T) {
	was := openTui
	t.Cleanup(func() { openTui = was })
	t.Setenv("HOME", t.TempDir())

	var opened tea.Model
	openTui = func(_ *cobra.Command, model tea.Model) error {
		opened = model
		return nil
	}
	if r := run(t, "tui"); r.Code != exitOK {
		t.Fatalf("board exit = %d (%s)", r.Code, r.Stderr)
	}
	if _, ok := opened.(tui.Board); !ok {
		t.Fatalf("what was opened is %T, want a board", opened)
	}
}

// TestTui_saysWhatWentWrongWhileItWasOpen checks that a board that stopped
// on an error stops the command with it, rather than exiting zero on a screen
// that went away.
func TestTui_saysWhatWentWrongWhileItWasOpen(t *testing.T) {
	was := openTui
	t.Cleanup(func() { openTui = was })
	t.Setenv("HOME", t.TempDir())

	openTui = func(*cobra.Command, tea.Model) error {
		return errors.New("the terminal went away")
	}
	r := run(t, "tui")
	if r.Code == exitOK {
		t.Fatal("a board that stopped on an error exited as though nothing had happened")
	}
	if !strings.Contains(strings.ToLower(r.Stderr), "terminal went away") {
		t.Errorf("stderr = %q, want it to carry what went wrong", r.Stderr)
	}
}

// TestTui_refusesAStoreItCannotOpen checks that the store is opened before
// anything is drawn, so a store that is behind is said at the terminal rather
// than inside a screen nobody can read it in.
func TestTui_refusesAStoreItCannotOpen(t *testing.T) {
	was := openTui
	t.Cleanup(func() { openTui = was })
	openTui = func(*cobra.Command, tea.Model) error {
		t.Fatal("the board was opened over a store that could not be")
		return nil
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_STORE", t.TempDir())
	if r := run(t, "tui"); r.Code == exitOK {
		t.Fatalf("board over a directory exited %d", r.Code)
	}
}

// The surface is opened by `cairn tui`, beside `cairn gui`. It was `cairn
// board` — a name that said what is drawn rather than where it is drawn, which
// left the two surfaces named by different kinds of word.
//
// Never asked of the bare command: it starts a program that reads a terminal,
// and a test that ran it would sit there waiting for a key nobody is going to
// press. Measured — the suite hung for ten minutes on exactly that.
func TestTui_isTheNameTheCommandGoesBy(t *testing.T) {
	if out := mustRun(t, "tui", "--help").Stdout; !strings.Contains(out, "cairn tui") {
		t.Errorf("`cairn tui --help` says:\n%s", out)
	}
	// The list of commands, and not `cairn board --help`: --help makes cobra
	// print the root's own usage and exit nought whatever word came before it,
	// so asking that way would be asking about cobra rather than about cairn.
	//
	// The first word of each line and not the whole of it: "board" still
	// appears in what both surfaces are for, because it is the thing they
	// draw. What was renamed is where it is drawn, not what it is.
	named := map[string]bool{}
	for _, line := range strings.Split(mustRun(t, "--help").Stdout, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 {
			named[fields[0]] = true
		}
	}
	if !named["tui"] {
		t.Error("cairn does not offer tui among its commands")
	}
	if named["board"] {
		t.Error("cairn still offers board among its commands")
	}
}

// The board is drawn in a terminal; asked for anywhere else, it says so
// rather than drawing escapes into a file or a pipe.
func TestNeedsATerminal_refusesWhatIsNotOne(t *testing.T) {
	var out bytes.Buffer
	err := needsATerminal(strings.NewReader(""), &out)
	if said := engine.RefusalOf(err); said.Reason != "not-a-terminal" {
		t.Errorf("needsATerminal = %v, want it refused as not a terminal", err)
	}
}

// A board asked for outside a terminal stops before a program is started on
// streams it cannot draw to.
func TestRunTui_stopsWhereThereIsNoTerminal(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&bytes.Buffer{})
	if err := runTui(cmd, nil); engine.RefusalOf(err).Reason != "not-a-terminal" {
		t.Errorf("runTui = %v, want it refused as not a terminal", err)
	}
}

// A file is not a terminal, and a character device is taken for one: the
// question asked is the one a terminal answers yes to.
func TestNeedsATerminal_readsWhatAFileIs(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	if err := needsATerminal(file, file); engine.RefusalOf(err).Reason != "not-a-terminal" {
		t.Errorf("a file: %v, want it refused", err)
	}
	device, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = device.Close() }()
	if err := needsATerminal(device, device); err != nil {
		t.Errorf("a character device: %v, want it taken", err)
	}
}
