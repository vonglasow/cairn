package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"cairn/internal/engine"
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

func TestTui_opensTheWorkspaceNamedAndOpensAnotherInItsPlace(t *testing.T) {
	was := openTui
	t.Cleanup(func() { openTui = was })
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, name := range []string{"work", "home"} {
		if r := run(t, "workspace", "add", name, filepath.Join(home, name)); r.Code != exitOK {
			t.Fatalf("workspace add %s = %d (%s)", name, r.Code, r.Stderr)
		}
	}
	var said string
	openTui = func(_ *cobra.Command, model tea.Model) error {
		batch, _ := model.Init()().(tea.BatchMsg)
		model, _ = model.Update(batch[len(batch)-1]())
		model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
		model, cmd := model.Update(tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl})
		model, _ = model.Update(cmd())
		model, cmd = model.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
		model, _ = model.Update(cmd())
		said = model.View().Content
		return nil
	}
	if r := run(t, "tui", "--workspace", "home"); r.Code != exitOK {
		t.Fatalf("board exit = %d (%s)", r.Code, r.Stderr)
	}
	if !strings.Contains(said, "opened work in place of home") {
		t.Errorf("the board says:\n%s", said)
	}
}
