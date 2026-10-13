package gui

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

type dialog struct {
	save, open string
	closed     bool
	broken     error
	failing    bool
}

type breaking struct{}

func (breaking) Read([]byte) (int, error)  { return 0, errBroke }
func (breaking) Write([]byte) (int, error) { return 0, errBroke }
func (breaking) Close() error              { return nil }

var errBroke = errors.New("the file broke")

func (d dialog) ChooseFile(...string) (io.ReadCloser, error) {
	if d.closed {
		return nil, ErrDeclined
	}
	if d.broken != nil {
		return nil, d.broken
	}
	if d.failing {
		return breaking{}, nil
	}
	return os.Open(d.open)
}

func (d dialog) CreateFile(string) (io.WriteCloser, error) {
	if d.closed {
		return nil, ErrDeclined
	}
	if d.broken != nil {
		return nil, d.broken
	}
	if d.failing {
		return breaking{}, nil
	}
	return os.Create(d.save)
}

func written(t *testing.T, held string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte(held), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFilingSavesWhereTheDialogSays(t *testing.T) {
	from := written(t, "the tasks")
	to := filepath.Join(t.TempDir(), "mine.json")
	got := Filing(FileAsk{Save: true, Name: "tasks.json", From: from}, dialog{save: to})
	if got.Err != nil || got.Declined || got.Called != to {
		t.Fatalf("filed %+v", got)
	}
	if held, _ := os.ReadFile(to); string(held) != "the tasks" {
		t.Errorf("the file holds %q", held)
	}
	if _, err := os.Stat(filepath.Dir(from)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the copy written from is still there: %v", err)
	}
	from = written(t, "the tasks")
	if got := Filing(FileAsk{Save: true, Name: "tasks.json", From: from}, dialog{closed: true}); !got.Declined || got.Err != nil {
		t.Errorf("a dialog closed filed %+v", got)
	}
	if _, err := os.Stat(from); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the copy written from outlived a closed dialog: %v", err)
	}
	broke := errors.New("the disk is full")
	if got := Filing(FileAsk{Save: true, From: written(t, "x")}, dialog{broken: broke}); !errors.Is(got.Err, broke) {
		t.Errorf("a dialog that failed filed %+v", got)
	}
	if got := Filing(FileAsk{Save: true, From: written(t, "x")}, dialog{failing: true}); !errors.Is(got.Err, errBroke) {
		t.Errorf("a file that broke as it was written filed %+v", got)
	}
}

func TestFilingOpensWhatTheDialogChose(t *testing.T) {
	chosen := written(t, "some tasks")
	got := Filing(FileAsk{}, dialog{open: chosen})
	if got.Err != nil || got.Declined || got.Called != chosen || got.Chosen == "" || got.Chosen == chosen {
		t.Fatalf("filed %+v", got)
	}
	if held, _ := os.ReadFile(got.Chosen); string(held) != "some tasks" {
		t.Errorf("the copy holds %q", held)
	}
	if got := Filing(FileAsk{}, dialog{closed: true}); !got.Declined || got.Chosen != "" {
		t.Errorf("a dialog closed filed %+v", got)
	}
	broke := errors.New("the file is gone")
	if got := Filing(FileAsk{}, dialog{broken: broke}); !errors.Is(got.Err, broke) {
		t.Errorf("a dialog that failed filed %+v", got)
	}
	if got := Filing(FileAsk{}, dialog{failing: true}); !errors.Is(got.Err, errBroke) || got.Chosen != "" {
		t.Errorf("a file that broke as it was read filed %+v", got)
	}
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "absent"))
	if got := Filing(FileAsk{}, dialog{open: chosen}); got.Err == nil || got.Chosen != "" {
		t.Errorf("with nowhere to copy it aside, filed %+v", got)
	}
}

func TestExportingAsksTheEngineThenTheDialog(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800).At(time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)).Key("cmd+e")
	in, ok := f.gave("task", "export")
	if !ok || filepath.Base(in.Given["to"]) != "tasks.json" {
		t.Fatalf("the engine was asked %+v", in)
	}
	ask, asked := w.FileAsked()
	if !asked || !ask.Save || ask.Name != "cairn-tasks-2026-10-12.json" || ask.From != in.Given["to"] {
		t.Errorf("the dialog asked for is %+v", ask)
	}
	if _, again := w.FileTaken().FileAsked(); again {
		t.Error("a dialog taken is asked for again")
	}
	saved := w.FileTaken().Filed(Filed{Ask: ask, Called: "/somewhere/mine.json"})
	if saved.Note() != "the tasks were written to /somewhere/mine.json" {
		t.Errorf("the window tells %q", saved.Note())
	}
	if got := w.FileTaken().Filed(Filed{Ask: ask, Declined: true}); got.Note() != "" || got.Wrong() != nil {
		t.Errorf("a dialog closed tells %q and %v", got.Note(), got.Wrong())
	}
	broke := errors.New("the disk is full")
	if got := w.FileTaken().Filed(Filed{Ask: ask, Err: broke}); !errors.Is(got.Wrong(), broke) {
		t.Errorf("a dialog that failed says %v", got.Wrong())
	}
	f.fail = broke
	if got := New(f, 1200, 800).Key("cmd+e"); !errors.Is(got.Wrong(), broke) {
		t.Errorf("an export refused says %v", got.Wrong())
	} else if _, asked := got.FileAsked(); asked {
		t.Error("an export refused still asked for the dialog")
	}
}

func TestImportingAddsTheTasksOfTheFileChosen(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800).Key("cmd+o")
	ask, asked := w.FileAsked()
	if !asked || ask.Save {
		t.Fatalf("the dialog asked for is %+v, %v", ask, asked)
	}
	copied := written(t, "some tasks")
	added := w.FileTaken().Filed(Filed{Ask: ask, Chosen: copied, Called: "/Downloads/tasks.json"})
	in, ok := f.gave("task", "import")
	if !ok || in.Given["from"] != copied || added.Screen() != OnTheBoard {
		t.Errorf("the engine was asked %+v and the window is on %v", in, added.Screen())
	}
	if added.Note() != "0 task(s) and 0 objective(s) were added from tasks.json" {
		t.Errorf("the window tells %q", added.Note())
	}
	if _, err := os.Stat(copied); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the copy outlived the import: %v", err)
	}
	f.failOp = errors.New("not an export")
	copied = written(t, "some tasks")
	refused := w.FileTaken().Filed(Filed{Ask: ask, Chosen: copied, Called: "x.json"})
	if !errors.Is(refused.Wrong(), f.failOp) {
		t.Errorf("an import refused says %v", refused.Wrong())
	}
	if _, err := os.Stat(copied); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the copy outlived a refusal: %v", err)
	}
}

func TestCarryingTheTasksIsOffered(t *testing.T) {
	w := New(newFake(), 1200, 800)
	if offers := w.BoardOffers(); !slices.Contains(offers, "Export") || !slices.Contains(offers, "Import") {
		t.Errorf("the bar offers %v", offers)
	}
	if Shortcut("export") != "cmd+e" || Shortcut("import") != "cmd+o" || Shortcut("new backup") != "cmd+b" || Shortcut("backups") != "shift+b" {
		t.Error("the shortcuts are not the keys")
	}
	if got := w.menuFor(); !slices.Equal(got, []string{"New", "Search", "Archived", "Help", "New backup", "Backups", "Export", "Import"}) {
		t.Errorf("the board's menu offers %v", got)
	}
	f := newFake()
	f.fail = errors.New("refused")
	if New(f, 1200, 800).Choose("New backup"); f.what("backup", "new") != 1 {
		t.Error("Back up made no backup")
	}
	if New(f, 1200, 800).Choose("Backups"); f.what("backup", "list") != 1 {
		t.Error("Backups listed nothing")
	}
	if _, asked := New(f, 1200, 800).Choose("Import").FileAsked(); !asked {
		t.Error("Import asked for no dialog")
	}
	if New(f, 1200, 800).Choose("Export"); f.what("task", "export") != 1 {
		t.Error("Export wrote nothing")
	}
}
