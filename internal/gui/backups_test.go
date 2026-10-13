package gui

import (
	"errors"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

var errNoBackup = errors.New("the backups cannot be reached")

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

func TestNoBackupYetIsSaid(t *testing.T) {
	got := New(newFake(), 1200, 800).listingBackups()
	if got.Screen() != OnTheBoard || got.Note() != "no backup yet: ⌘B makes one" {
		t.Errorf("no backup left the window on %v, telling %q", got.Screen(), got.Note())
	}
}

func TestTheBackupsAreWalkedBothWays(t *testing.T) {
	w := Window{screen: OnBackups, backupsState: backupsState{backups: []domain.Backup{{Number: 2}, {Number: 1}}}}
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

func TestTheBackupsAreDrawn(t *testing.T) {
	made := time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC)
	w := Window{screen: OnBackups, width: 1200, height: 800, backupsState: backupsState{backups: []domain.Backup{{Number: 2, Made: made}, {Number: 1, Made: made}}}}
	shot := drawn(t, w)
	box := w.BackupsBox()
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

func TestTheBackupsAreLaidOut(t *testing.T) {
	box := Box{X: 100, Y: 50, W: 400, H: 300}
	if got, want := BackupRowAt(box, 1), (Box{X: 106, Y: 100 + between + menuRow, W: 388, H: menuRow}); got != want {
		t.Errorf("the second backup is at %+v, want %+v", got, want)
	}
}

func TestRestoringAsksWhatItCosts(t *testing.T) {
	w := Window{screen: OnAsking, question: aboutPuttingBack, backupsState: backupsState{backups: []domain.Backup{{Number: 7}}}}
	asks := w.Asks()
	if len(asks) != 2 || asks[0].Means != "A new backup of what the store holds now is made first." {
		t.Errorf("the question says %+v", asks)
	}
	if got := Heads(w); got != "Restore backup 7?" {
		t.Errorf("the question is headed %q", got)
	}
}

func TestTheBackupsAnswerThePointer(t *testing.T) {
	w := Window{screen: OnBackups, width: 1200, height: 800, backupsState: backupsState{backups: []domain.Backup{{Number: 2}, {Number: 1}}}}
	row := BackupRowAt(w.BackupsBox(), 1)
	asked := w.Click(Point{X: row.X + 5, Y: row.Y + 5})
	if asked.Screen() != OnAsking || asked.question != aboutPuttingBack || asked.BackupOn() != 1 {
		t.Errorf("a click on the second backup left the window on %v asking %v about %d", asked.Screen(), asked.question, asked.BackupOn())
	}
	if got := w.Click(Point{X: 2, Y: 2}); got.Screen() != OnTheBoard {
		t.Errorf("a click off the list left the window on %v", got.Screen())
	}
}

func TestManyBackupsScroll(t *testing.T) {
	many := make([]domain.Backup, 40)
	for i := range many {
		many[i] = domain.Backup{Number: 40 - i}
	}
	w := Window{screen: OnBackups, width: 1200, height: 800, backupsState: backupsState{backups: many}}
	room := w.BackupsRoom()
	if room < 1 || room >= len(many) {
		t.Fatalf("the list has room for %d of 40", room)
	}
	for range room + 3 {
		w = w.onBackups("down")
	}
	if from := w.BackupsFrom(); w.BackupOn() < from || w.BackupOn() >= from+room {
		t.Errorf("backup %d is out of sight, the list starting at %d with room for %d", w.BackupOn(), from, room)
	}
	for range 3 {
		w = w.onBackups("up")
	}
	if from := w.BackupsFrom(); w.BackupOn() < from {
		t.Errorf("walking back up left backup %d above the list, which starts at %d", w.BackupOn(), from)
	}
}

func TestTheBackupsListIsMeasuredAndClickedWhereItIsDrawn(t *testing.T) {
	many := make([]domain.Backup, 40)
	for i := range many {
		many[i] = domain.Backup{Number: 40 - i}
	}
	w := Window{screen: OnBackups, width: 1200, height: 800, backupsState: backupsState{backups: many, backupsFrom: 5}}
	if got := w.BackupsRoom(); got != 19 {
		t.Errorf("the list has room for %d, want 19", got)
	}
	first := BackupRowAt(w.BackupsBox(), 0)
	if got := w.Click(Point{X: first.X + 5, Y: first.Y + 5}).BackupOn(); got != 5 {
		t.Errorf("a click on the first line drawn took backup %d, want the sixth, 5", got)
	}
	last := BackupRowAt(w.BackupsBox(), 18)
	if got := w.Click(Point{X: last.X + 5, Y: last.Y + 5}).BackupOn(); got != 23 {
		t.Errorf("a click on the last line drawn took backup %d, want 23", got)
	}
	under := BackupRowAt(w.BackupsBox(), 19)
	if got := w.Click(Point{X: under.X + 5, Y: under.Y + 5}); got.Screen() != OnBackups {
		t.Errorf("a click under the last line drawn left the window on %v", got.Screen())
	}
}

func TestAScrolledBackupsListLightsTheRightLine(t *testing.T) {
	many := make([]domain.Backup, 20)
	for i := range many {
		many[i] = domain.Backup{Number: 20 - i}
	}
	w := Window{screen: OnBackups, width: 1200, height: 800, backupsState: backupsState{backups: many, backupsFrom: 5, backupOn: 6}}
	shot := drawn(t, w)
	corner := func(i int) Box {
		row := BackupRowAt(w.BackupsBox(), i)
		return Box{X: row.X + 1, Y: row.Y + 1, W: 1, H: 1}
	}
	if ink(shot, corner(1), Dark().Back) != 0 || ink(shot, corner(0), Dark().Panel) != 0 {
		t.Error("the second line drawn, backup 6, is not the one lit")
	}
}

func TestTheWheelStaysInsideAPanel(t *testing.T) {
	var many []engine.Card
	for i := 1; i <= 12; i++ {
		many = append(many, engine.Card{Number: i, Title: "Task", State: "todo"})
	}
	w := holding(newFake(), many...)
	l := w.Layout()
	over := Point{X: l.Column[1].X + l.Column[1].W/2, Y: l.Column[1].Y + l.Column[1].H/2}
	for _, screen := range []Screen{OnCalendar, OnMenu, OnBackups} {
		v := w
		v.screen = screen
		if got := v.Scroll(over, CardHeight*3); got.scroll != w.scroll {
			t.Errorf("the wheel over %v scrolled the board to %v", screen, got.scroll)
		}
	}
	backups := make([]domain.Backup, 40)
	for i := range backups {
		backups[i] = domain.Backup{Number: 40 - i}
	}
	w.screen, w.backups = OnBackups, backups
	box := w.BackupsBox()
	in := Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
	if got := w.Scroll(in, 2*lineHigh).BackupsFrom(); got != 2 {
		t.Errorf("two notches down the list starts at %d", got)
	}
	if got := w.Scroll(in, 999*lineHigh).BackupsFrom(); got != len(backups)-w.BackupsRoom() {
		t.Errorf("turned past its end the list starts at %d", got)
	}
	if got := w.Scroll(in, -lineHigh).BackupsFrom(); got != 0 {
		t.Errorf("turned up past its start the list starts at %d", got)
	}
	if got := w.Scroll(Point{X: 1, Y: 1}, 2*lineHigh).BackupsFrom(); got != 0 {
		t.Errorf("the wheel off the list moved it to %d", got)
	}
}
