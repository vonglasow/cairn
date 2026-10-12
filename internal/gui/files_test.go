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

// dialog is the system's file dialog as a test plays it: a file to save to,
// a file to open, or closed without either.
type dialog struct {
	save, open string
	closed     bool
	broken     error
	// failing is a file that breaks once it is read or written.
	failing bool
}

// breaking is a file that fails the first read or write.
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

// written is a file of the test's own, holding what is given, in a folder of
// its own the way an export's is.
func written(t *testing.T, held string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, []byte(held), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// A store written out goes to the file chosen in the dialog, under the name
// the dialog gave it, and the copy it was written from goes.
func TestFilingSavesWhereTheDialogSays(t *testing.T) {
	from := written(t, "the store")
	to := filepath.Join(t.TempDir(), "mine.db")
	got := Filing(FileAsk{Save: true, Name: "cairn.db", From: from}, dialog{save: to})
	if got.Err != nil || got.Declined || got.Called != to {
		t.Fatalf("filed %+v", got)
	}
	if held, _ := os.ReadFile(to); string(held) != "the store" {
		t.Errorf("the file holds %q", held)
	}
	if _, err := os.Stat(filepath.Dir(from)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the copy written from is still there: %v", err)
	}
	from = written(t, "the store")
	if got := Filing(FileAsk{Save: true, Name: "cairn.db", From: from}, dialog{closed: true}); !got.Declined || got.Err != nil {
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

// A store brought in is copied out of the file chosen, into one of the
// window's own: the dialog's file may only be read while it is open.
func TestFilingOpensWhatTheDialogChose(t *testing.T) {
	chosen := written(t, "another store")
	got := Filing(FileAsk{}, dialog{open: chosen})
	if got.Err != nil || got.Declined || got.Called != chosen || got.Chosen == "" || got.Chosen == chosen {
		t.Fatalf("filed %+v", got)
	}
	if held, _ := os.ReadFile(got.Chosen); string(held) != "another store" {
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

// ⌘E writes the store out through the engine first, then asks the loop for
// the dialog, naming the file after the day.
func TestExportingAsksTheEngineThenTheDialog(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800).At(time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)).Key("cmd+e")
	in, ok := f.gave("backup", "export")
	if !ok || filepath.Base(in.Given["to"]) != "cairn.db" {
		t.Fatalf("the engine was asked %+v", in)
	}
	ask, asked := w.FileAsked()
	if !asked || !ask.Save || ask.Name != "cairn-2026-10-12.db" || ask.From != in.Given["to"] {
		t.Errorf("the dialog asked for is %+v", ask)
	}
	if _, again := w.FileTaken().FileAsked(); again {
		t.Error("a dialog taken is asked for again")
	}
	saved := w.FileTaken().Filed(Filed{Ask: ask, Called: "/somewhere/mine.db"})
	if saved.Note() != "the store was written to /somewhere/mine.db" {
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

// ⌘O asks for the dialog; the file chosen is put in place only once agreed
// to, and the copy of it goes either way.
func TestImportingAsksBeforeReplacingTheStore(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800).Key("cmd+o")
	ask, asked := w.FileAsked()
	if !asked || ask.Save {
		t.Fatalf("the dialog asked for is %+v, %v", ask, asked)
	}
	copied := written(t, "another store")
	chosen := w.FileTaken().Filed(Filed{Ask: ask, Chosen: copied, Called: "/Downloads/cairn.db"})
	if chosen.Screen() != OnAsking || Heads(chosen) != "Bring in cairn.db?" {
		t.Fatalf("the window is on %v headed %q", chosen.Screen(), Heads(chosen))
	}
	if asks := chosen.Asks(); len(asks) != 2 || asks[0].Means != "It takes the place of the store, a new backup of which is made first." {
		t.Errorf("the question says %+v", asks)
	}
	if left := chosen.Choose("Cancel"); left.Screen() != OnTheBoard || f.what("backup", "import") != 0 {
		t.Errorf("cancelling left the window on %v", left.Screen())
	}
	if _, err := os.Stat(copied); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the copy outlived a cancel: %v", err)
	}
	copied = written(t, "another store")
	chosen = w.FileTaken().Filed(Filed{Ask: ask, Chosen: copied, Called: "/Downloads/cairn.db"})
	agreed := chosen.Key("enter")
	in, ok := f.gave("backup", "import")
	if !ok || in.Given["from"] != copied || agreed.Screen() != OnTheBoard {
		t.Errorf("agreeing asked %+v and left the window on %v", in, agreed.Screen())
	}
	if agreed.Note() != "the store now holds what cairn.db held" {
		t.Errorf("the window tells %q", agreed.Note())
	}
	if _, err := os.Stat(copied); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the copy outlived the import: %v", err)
	}
	f.failOp = errors.New("not a cairn store")
	copied = written(t, "another store")
	refused := w.FileTaken().Filed(Filed{Ask: ask, Chosen: copied, Called: "x.db"}).Key("enter")
	if !errors.Is(refused.Wrong(), f.failOp) {
		t.Errorf("an import refused says %v", refused.Wrong())
	}
}

// Both are offered where a person looks for them: on the bar of a board with
// nothing chosen, and on a right click on the board, beside backing up.
func TestCarryingTheStoreIsOffered(t *testing.T) {
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
	// Refused, so the empty answer of the fake is never read: what is
	// counted is that each was asked for.
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
	if New(f, 1200, 800).Choose("Export"); f.what("backup", "export") != 1 {
		t.Error("Export wrote nothing")
	}
}
