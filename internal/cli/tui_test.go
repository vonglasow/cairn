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

func TestTui_takesNoArguments(t *testing.T) {
	r := run(t, "tui", "tasks")
	if r.Code == exitOK {
		t.Fatalf("tui with a stray argument exited %d", r.Code)
	}
}

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

func TestTui_isTheNameTheCommandGoesBy(t *testing.T) {
	if out := mustRun(t, "tui", "--help").Stdout; !strings.Contains(out, "cairn tui") {
		t.Errorf("`cairn tui --help` says:\n%s", out)
	}
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

func TestNeedsATerminal_refusesWhatIsNotOne(t *testing.T) {
	var out bytes.Buffer
	err := needsATerminal(strings.NewReader(""), &out)
	if said := engine.RefusalOf(err); said.Reason != "not-a-terminal" {
		t.Errorf("needsATerminal = %v, want it refused as not a terminal", err)
	}
}

func TestRunTui_stopsWhereThereIsNoTerminal(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&bytes.Buffer{})
	if err := runTui(cmd, nil); engine.RefusalOf(err).Reason != "not-a-terminal" {
		t.Errorf("runTui = %v, want it refused as not a terminal", err)
	}
}

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

func TestTui_writesItsKeysDownInDebugMode(t *testing.T) {
	was := openTui
	t.Cleanup(func() { openTui = was })
	t.Setenv("HOME", t.TempDir())
	store := storeFor(t)
	t.Setenv("CAIRN_STORE", store)
	t.Setenv("CAIRN_DEBUG", "1")

	openTui = func(_ *cobra.Command, model tea.Model) error {
		_, _ = model.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
		return nil
	}
	if r := run(t, "tui"); r.Code != exitOK {
		t.Fatalf("board exit = %d (%s)", r.Code, r.Stderr)
	}
	entries, err := os.ReadDir(engine.Places{Store: store}.Debug())
	if err != nil || len(entries) != 1 {
		t.Fatalf("the debug folder holds %v, %v", entries, err)
	}
	read, err := os.ReadFile(engine.Places{Store: store}.Debug() + "/" + entries[0].Name())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(read, []byte("board key ?")) {
		t.Errorf("the key pressed is not written down:\n%s", read)
	}
}
