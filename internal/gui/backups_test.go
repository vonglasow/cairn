package gui

import (
	"errors"
	"testing"
	"time"

	"cairn/internal/domain"
)

var errNoBackup = errors.New("the backups cannot be reached")

// What the store refuses, backing up or listing or putting back, is said as
// anything gone wrong is, and the window stays where it was.
func TestBackingUpSaysWhatWentWrong(t *testing.T) {
	f := newFake()
	f.fail = errNoBackup
	w := New(f, 1200, 800)
	if got := w.backingUp(); !errors.Is(got.Wrong(), errNoBackup) || got.Note() != "" {
		t.Errorf("a backup refused says %v and tells %q", got.Wrong(), got.Note())
	}
	if got := w.listingBackups(); !errors.Is(got.Wrong(), errNoBackup) || got.Screen() != OnTheBoard {
		t.Errorf("a listing refused says %v, on %v", got.Wrong(), got.Screen())
	}
	w.backups = []domain.Backup{{Number: 3}}
	if got := w.puttingBack(); !errors.Is(got.Wrong(), errNoBackup) || got.Screen() != OnTheBoard {
		t.Errorf("putting back refused says %v, on %v", got.Wrong(), got.Screen())
	}
}

// With no backup at all there is nothing to list, and the window says how to
// make one rather than opening an empty list.
func TestNoBackupYetIsSaid(t *testing.T) {
	got := New(newFake(), 1200, 800).listingBackups()
	if got.Screen() != OnTheBoard || got.Note() != "no backup yet: ⌘B makes one" {
		t.Errorf("no backup left the window on %v, telling %q", got.Screen(), got.Note())
	}
}

// The list is walked both ways and stops at its ends; esc closes it.
func TestTheBackupsAreWalkedBothWays(t *testing.T) {
	w := Window{screen: OnBackups, backups: []domain.Backup{{Number: 2}, {Number: 1}}}
	if got := w.onBackups("up").BackupOn(); got != 0 {
		t.Errorf("up from the first is on %d", got)
	}
	down := w.onBackups("j").onBackups("down")
	if got := down.BackupOn(); got != 1 {
		t.Errorf("down past the last is on %d", got)
	}
	if got := down.onBackups("k").BackupOn(); got != 0 {
		t.Errorf("k is on %d", got)
	}
	if got := w.onBackups("esc").Screen(); got != OnTheBoard {
		t.Errorf("esc left the window on %v", got)
	}
	if got := w.onBackups("enter"); got.Screen() != OnAsking || got.question != aboutPuttingBack {
		t.Errorf("enter left the window on %v asking %v", got.Screen(), got.question)
	}
	if got := w.onBackups("enter").Choose("Cancel"); got.Screen() != OnTheBoard {
		t.Errorf("cancelling the question left the window on %v", got.Screen())
	}
}

// What the window tells goes on its own, as what went wrong does.
func TestWhatIsToldGoes(t *testing.T) {
	w := Window{}.telling("backup 1 is here")
	for range Lingers() - 1 {
		w = w.aged()
	}
	if w.Note() == "" {
		t.Fatal("what was told went too soon")
	}
	if w = w.aged(); w.Note() != "" {
		t.Errorf("what was told is still there: %q", w.Note())
	}
}

// The list is drawn, a backup to a line, the one the keyboard is on lit; and
// what the window tells is drawn above the bar.
func TestTheBackupsAreDrawn(t *testing.T) {
	made := time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC)
	w := Window{screen: OnBackups, width: 1200, height: 800, backups: []domain.Backup{{Number: 2, Made: made}, {Number: 1, Made: made}}}
	shot := drawn(t, w)
	box := ChoosingOver(1200, 800)
	corner := func(i int) Box {
		row := BackupRowAt(box, i)
		return Box{X: row.X + 1, Y: row.Y + 1, W: 1, H: 1}
	}
	if ink(shot, corner(0), Dark().Back) != 0 || ink(shot, corner(1), Dark().Panel) != 0 {
		t.Error("only the line the keyboard is on is lit")
	}
	told := Window{width: 1200, height: 800}.telling("backup 1 is here")
	if got := ink(drawn(t, told), WrongBox(told.Layout()), Dark().Back); got == 0 {
		t.Error("what the window tells is not drawn")
	}
}

// The list's lines start under its heading, one menu row apart.
func TestTheBackupsAreLaidOut(t *testing.T) {
	box := Box{X: 100, Y: 50, W: 400, H: 300}
	if got, want := BackupRowAt(box, 1), (Box{X: 106, Y: 100 + between + menuRow, W: 388, H: menuRow}); got != want {
		t.Errorf("the second backup is at %+v, want %+v", got, want)
	}
}

// The question says what agreeing costs: nothing, since the store is backed
// up first — and names the backup it would put back.
func TestPuttingBackAsksWhatItCosts(t *testing.T) {
	w := Window{screen: OnAsking, question: aboutPuttingBack, backups: []domain.Backup{{Number: 7}}}
	asks := w.Asks()
	if len(asks) != 2 || asks[0].Means != "What the store holds now is backed up first." {
		t.Errorf("the question says %+v", asks)
	}
	if got := Heads(w); got != "Put backup 7 back?" {
		t.Errorf("the question is headed %q", got)
	}
}
