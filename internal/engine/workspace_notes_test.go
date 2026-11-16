package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cairn/internal/engine/workspace"
	"cairn/internal/store/files"
)

func TestWorkspaceAdd_namesTheNotesItIsGiven(t *testing.T) {
	h := homeHolding(t, "", nil)
	in := adding("play", "~/p")
	in.Given["notes"] = "~/vault/play"
	answer, err := WorkspaceAdd.Run(workedOn{Home: h}, in)
	if err != nil {
		t.Fatal(err)
	}
	if want := "[workspaces]\nplay = { folder = \"~/p\", notes = \"~/vault/play\" }\n"; fileRead(t, files.FileIn(h.Dir)) != want {
		t.Errorf("the file reads %q, want %q", fileRead(t, files.FileIn(h.Dir)), want)
	}
	if answer.WorkspaceAdded.Notes != filepath.Join(h.Dir, "vault", "play") {
		t.Errorf("the notes are said at %s", answer.WorkspaceAdded.Notes)
	}
	if info, err := os.Stat(filepath.Join(h.Dir, "vault", "play")); err != nil || !info.IsDir() {
		t.Errorf("the notes folder: %v", err)
	}
	if _, err := os.Stat(filepath.Join(h.Dir, "p", "notes")); err == nil {
		t.Error("a notes folder was made in the workspace's folder too")
	}
}

func TestWorkspaceAdd_refusesNotesHoldingTheBoard(t *testing.T) {
	for _, notes := range []string{"~/p", "~", "notes"} {
		h := homeHolding(t, "", nil)
		in := adding("play", "~/p")
		in.Given["notes"] = notes
		if _, err := WorkspaceAdd.Run(workedOn{Home: h}, in); err == nil {
			t.Errorf("notes %q are taken", notes)
		}
		if fileRead(t, files.FileIn(h.Dir)) != "" {
			t.Errorf("notes %q refused wrote %q", notes, fileRead(t, files.FileIn(h.Dir)))
		}
	}
	h := homeHolding(t, "", nil)
	in := adding("play", "~/p")
	in.Given["notes"] = "~/p"
	if _, err := WorkspaceAdd.Run(workedOn{Home: h}, in); !errors.Is(err, workspace.ErrNotesHoldTheBoard) {
		t.Errorf("notes holding the board give %v", err)
	}
}

func TestConfigSet_theNotesOfAWorkspaceAreItsOwn(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	got, err := setting(h, ConfigSet, map[string]string{"key": "notes", "value": "~/vault/w"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "[workspaces]\ndefault = \"work\"\nwork = { folder = \"~/w\", notes = \"~/vault/w\" }\nhome = \"~/h\"\n"; fileRead(t, files.FileIn(h.Dir)) != want {
		t.Errorf("the file reads %q, want %q", fileRead(t, files.FileIn(h.Dir)), want)
	}
	if c := got.ConfigChanged; c == nil || c.File != files.FileIn(h.Dir) || c.Value != filepath.Join(h.Dir, "vault", "w") {
		t.Errorf("the answer is %+v", got.ConfigChanged)
	}
	if _, err := setting(h, ConfigUnset, map[string]string{"key": "notes"}); err != nil {
		t.Fatal(err)
	}
	if fileRead(t, files.FileIn(h.Dir)) != twoWorkspaces {
		t.Errorf("unset, the file reads %q", fileRead(t, files.FileIn(h.Dir)))
	}
}

func TestConfigSet_refusesNotesNoWorkspaceOwnsOrThatHoldTheBoard(t *testing.T) {
	for _, c := range []struct {
		said  string
		given map[string]string
		want  error
	}{
		{twoWorkspaces, map[string]string{"key": "notes", "value": "~/w"}, workspace.ErrNotesHoldTheBoard},
		{twoWorkspaces, map[string]string{"key": "notes", "value": "~/v", "global": "true"}, workspace.ErrAnsweredKey},
		{"", map[string]string{"key": "notes", "value": "~/v"}, workspace.ErrAnsweredKey},
	} {
		h := homeHolding(t, c.said, nil, "w", "h")
		if _, err := setting(h, ConfigSet, c.given); !errors.Is(err, c.want) {
			t.Errorf("%v gives %v, want %v", c.given, err, c.want)
		}
		if fileRead(t, files.FileIn(h.Dir)) != c.said {
			t.Errorf("%v refused wrote %q", c.given, fileRead(t, files.FileIn(h.Dir)))
		}
	}
}

func TestNotesMove_movesTheNotesLeftWhereTheyWere(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	summary := filepath.Join(h.Dir, "w", "notes", "1", "summary.md")
	if err := os.MkdirAll(filepath.Dir(summary), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(summary, []byte("where it stands"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := setting(h, ConfigSet, map[string]string{"key": "notes", "value": "~/vault/w"}); err != nil {
		t.Fatal(err)
	}
	shown, err := setting(h, ConfigShow, nil)
	if err != nil {
		t.Fatal(err)
	}
	if left := shown.Config.NotesLeft; left == nil || left.Tasks != 1 || left.From != filepath.Join(h.Dir, "w", "notes") || left.To != filepath.Join(h.Dir, "vault", "w") {
		t.Fatalf("the notes left are said %+v", left)
	}
	if _, err := setting(h, NotesMove, nil); err != nil {
		t.Fatal(err)
	}
	if fileRead(t, filepath.Join(h.Dir, "vault", "w", "1", "summary.md")) != "where it stands" {
		t.Error("the notes were not moved")
	}
	if shown, _ := setting(h, ConfigShow, nil); shown.Config.NotesLeft != nil {
		t.Errorf("moved, the notes are still said left: %+v", shown.Config.NotesLeft)
	}
	if _, err := setting(h, NotesMove, nil); err == nil {
		t.Error("moving with nothing left is taken")
	}
}
