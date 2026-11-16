package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/backup"
	"cairn/internal/gui/geom"
	"errors"
	"testing"
)

var errNoBackup = errors.New("the backups cannot be reached")

func TestBackingUpSaysWhatWentWrong(t *testing.T) {
	f := newFake()
	f.fail = errNoBackup
	w := New(f, 1200, 800)
	if got := w.backingUp(); !errors.Is(got.Wrong(), errNoBackup) || got.Note() != "" {
		t.Errorf("a backup refused says %v and tells %q", got.Wrong(), got.Note())
	}
	if got := w.listingBackups(); !errors.Is(got.Wrong(), errNoBackup) || got.Screen != board.OnTheBoard {
		t.Errorf("a listing refused says %v, on %v", got.Wrong(), got.Screen)
	}
	w.Backups = []backup.Backup{{Number: 3}}
	if got := w.puttingBack(); !errors.Is(got.Wrong(), errNoBackup) || got.Screen != board.OnTheBoard {
		t.Errorf("putting back refused says %v, on %v", got.Wrong(), got.Screen)
	}
}

func TestNoBackupYetIsSaid(t *testing.T) {
	got := New(newFake(), 1200, 800).listingBackups()
	if got.Screen != board.OnTheBoard || got.Note() != "no backup yet: `⌘B` makes one" {
		t.Errorf("no backup left the window on %v, telling %q", got.Screen, got.Note())
	}
}

func TestTheBackupsAreWalkedBothWays(t *testing.T) {
	w := Window{Board: board.Board{BackupsState: board.BackupsState{Backups: []backup.Backup{{Number: 2}, {Number: 1}}}, Screen: board.OnBackups}}
	if got := w.onBackups("up").BackupOn; got != 0 {
		t.Errorf("up from the first is on %d", got)
	}
	down := w.onBackups("j").onBackups("down")
	if got := down.BackupOn; got != 1 {
		t.Errorf("down past the last is on %d", got)
	}
	if got := down.onBackups("k").BackupOn; got != 0 {
		t.Errorf("k is on %d", got)
	}
	if got := w.onBackups("esc").Screen; got != board.OnTheBoard {
		t.Errorf("esc left the window on %v", got)
	}
	if got := w.onBackups("enter"); got.Screen != board.OnAsking || got.Question != board.AboutPuttingBack {
		t.Errorf("enter left the window on %v asking %v", got.Screen, got.Question)
	}
	if got := w.onBackups("enter").Choose("Cancel"); got.Screen != board.OnTheBoard {
		t.Errorf("cancelling the question left the window on %v", got.Screen)
	}
}

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

func TestTheBackupsAnswerThePointer(t *testing.T) {
	w := Window{Board: board.Board{BackupsState: board.BackupsState{Backups: []backup.Backup{{Number: 2}, {Number: 1}}}, Screen: board.OnBackups}, width: 1200, height: 800}
	row := ListRowAt(w.BackupsBox(), 1)
	asked := w.Click(geom.Point{X: row.X + 5, Y: row.Y + 5})
	if asked.Screen != board.OnAsking || asked.Question != board.AboutPuttingBack || asked.BackupOn != 1 {
		t.Errorf("a click on the second backup left the window on %v asking %v about %d", asked.Screen, asked.Question, asked.BackupOn)
	}
	if got := w.Click(geom.Point{X: 2, Y: 2}); got.Screen != board.OnTheBoard {
		t.Errorf("a click off the list left the window on %v", got.Screen)
	}
}

func TestTheWheelStaysInsideAPanel(t *testing.T) {
	var many []board.Card
	for i := 1; i <= 12; i++ {
		many = append(many, board.Card{Number: i, Title: "Task", State: "todo"})
	}
	w := holding(newFake(), many...)
	l := w.Layout()
	over := geom.Point{X: l.Column[1].X + l.Column[1].W/2, Y: l.Column[1].Y + l.Column[1].H/2}
	for _, screen := range []board.Screen{board.OnCalendar, board.OnMenu, board.OnBackups} {
		v := w
		v.Screen = screen
		if got := v.Scroll(over, geom.CardHeight*3); got.scroll != w.scroll {
			t.Errorf("the wheel over %v scrolled the board to %v", screen, got.scroll)
		}
	}
	backups := make([]backup.Backup, 40)
	for i := range backups {
		backups[i] = backup.Backup{Number: 40 - i}
	}
	w.Screen, w.Backups = board.OnBackups, backups
	box := w.BackupsBox()
	in := geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
	if got := w.Scroll(in, 2*lineHigh).BackupsFrom; got != 2 {
		t.Errorf("two notches down the list starts at %d", got)
	}
	if got := w.Scroll(in, 999*lineHigh).BackupsFrom; got != len(backups)-w.BackupsRoom() {
		t.Errorf("turned past its end the list starts at %d", got)
	}
	if got := w.Scroll(in, -lineHigh).BackupsFrom; got != 0 {
		t.Errorf("turned up past its start the list starts at %d", got)
	}
	if got := w.Scroll(geom.Point{X: 1, Y: 1}, 2*lineHigh).BackupsFrom; got != 0 {
		t.Errorf("the wheel off the list moved it to %d", got)
	}
}
