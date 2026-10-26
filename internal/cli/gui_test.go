package cli

import (
	"cairn/internal/engine"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuiOpensTheStoreAndHandsItToTheWindow(t *testing.T) {
	store := storeFor(t)
	t.Setenv("CAIRN_STORE", store)
	var readable error
	handed := false
	was := openGui
	openGui = func(work engine.Work, _ engine.Pages, draft, sizes string, width, height int) error {
		handed = true
		if sizes != "" {
			t.Errorf("with nothing configured, the window keeps its sizes at %q", sizes)
		}
		_, readable = work.Version()
		if draft != store+".draft.json" {
			t.Errorf("the window keeps its draft at %q, not beside the store", draft)
		}
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
	t.Setenv("CAIRN_STORE", storeFor(t))
	was := openGui
	openGui = func(engine.Work, engine.Pages, string, string, int, int) error { return errors.New("no display here") }
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

func TestGuiActsThroughTheWindow(t *testing.T) {
	store := storeFor(t)
	t.Setenv("CAIRN_STORE", store)
	t.Setenv("CAIRN_AUTHOR", "sam")
	was := openGui
	openGui = func(work engine.Work, _ engine.Pages, _, _ string, _, _ int) error {
		_, err := work.Ask(engine.TaskNew, engine.Input{Given: map[string]string{"title": "Write the spec"}})
		return err
	}
	t.Cleanup(func() { openGui = was })
	if got := run(t, "gui"); got.Code != 0 {
		t.Fatalf("gui exited %d: %s", got.Code, got.Stderr)
	}
	shown := run(t, "trace", "show", "1")
	if !strings.Contains(shown.Stdout, "(by sam, through gui)") {
		t.Errorf("an act from the window reads:\n%s", shown.Stdout)
	}
}

func TestGuiKeepsTheSizesBesideTheStoreWhenTheConfigurationSaysSo(t *testing.T) {
	store := storeFor(t)
	t.Setenv("CAIRN_STORE", store)
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".config", "cairn", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[window]\nkeep_sizes = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var kept string
	was := openGui
	openGui = func(_ engine.Work, _ engine.Pages, _, sizes string, _, _ int) error {
		kept = sizes
		return nil
	}
	t.Cleanup(func() { openGui = was })
	if got := run(t, "gui"); got.Code != 0 {
		t.Fatalf("gui exited %d: %s", got.Code, got.Stderr)
	}
	if kept != store+".sizes.json" {
		t.Errorf("the window keeps its sizes at %q, not beside the store", kept)
	}
}
