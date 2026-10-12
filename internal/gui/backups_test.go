package gui

import (
	"errors"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/reading"
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

// A click on a backup asks to put it back; a click off the list closes it.
func TestTheBackupsAnswerThePointer(t *testing.T) {
	w := Window{screen: OnBackups, width: 1200, height: 800, backups: []domain.Backup{{Number: 2}, {Number: 1}}}
	row := BackupRowAt(ChoosingOver(1200, 800), 1)
	asked := w.Click(Point{X: row.X + 5, Y: row.Y + 5})
	if asked.Screen() != OnAsking || asked.question != aboutPuttingBack || asked.BackupOn() != 1 {
		t.Errorf("a click on the second backup left the window on %v asking %v about %d", asked.Screen(), asked.question, asked.BackupOn())
	}
	if got := w.Click(Point{X: 2, Y: 2}); got.Screen() != OnTheBoard {
		t.Errorf("a click off the list left the window on %v", got.Screen())
	}
}

// Many backups scroll: the one the keyboard is on is always among the lines
// drawn, and the list never starts past the last screenful.
func TestManyBackupsScroll(t *testing.T) {
	many := make([]domain.Backup, 40)
	for i := range many {
		many[i] = domain.Backup{Number: 40 - i}
	}
	w := Window{screen: OnBackups, width: 1200, height: 800, backups: many}
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

// The list's room, to the line: 387 between its first line and its offers,
// thirteen lines of 28. And a click on a list scrolled down takes the backup
// drawn there, not the one that would be there unscrolled.
func TestTheBackupsListIsMeasuredAndClickedWhereItIsDrawn(t *testing.T) {
	many := make([]domain.Backup, 40)
	for i := range many {
		many[i] = domain.Backup{Number: 40 - i}
	}
	w := Window{screen: OnBackups, width: 1200, height: 800, backups: many, backupsFrom: 5}
	if got := w.BackupsRoom(); got != 13 {
		t.Errorf("the list has room for %d, want 13", got)
	}
	first := BackupRowAt(ChoosingOver(1200, 800), 0)
	if got := w.Click(Point{X: first.X + 5, Y: first.Y + 5}).BackupOn(); got != 5 {
		t.Errorf("a click on the first line drawn took backup %d, want the sixth, 5", got)
	}
	last := BackupRowAt(ChoosingOver(1200, 800), 12)
	if got := w.Click(Point{X: last.X + 5, Y: last.Y + 5}).BackupOn(); got != 17 {
		t.Errorf("a click on the last line drawn took backup %d, want 17", got)
	}
	// Under the last line drawn, inside the panel, there is nothing to take.
	under := BackupRowAt(ChoosingOver(1200, 800), 13)
	if got := w.Click(Point{X: under.X + 5, Y: under.Y + 5}); got.Screen() != OnBackups {
		t.Errorf("a click under the last line drawn left the window on %v", got.Screen())
	}
}

// Scrolled, the line lit is the one the keyboard is on, wherever it is drawn.
func TestAScrolledBackupsListLightsTheRightLine(t *testing.T) {
	many := make([]domain.Backup, 20)
	for i := range many {
		many[i] = domain.Backup{Number: 20 - i}
	}
	w := Window{screen: OnBackups, width: 1200, height: 800, backups: many, backupsFrom: 5, backupOn: 6}
	shot := drawn(t, w)
	corner := func(i int) Box {
		row := BackupRowAt(ChoosingOver(1200, 800), i)
		return Box{X: row.X + 1, Y: row.Y + 1, W: 1, H: 1}
	}
	if ink(shot, corner(1), Dark().Back) != 0 || ink(shot, corner(0), Dark().Panel) != 0 {
		t.Error("the second line drawn, backup 6, is not the one lit")
	}
}

// A panel over the board stops the wheel: over the calendar, the menu or the
// list of backups it used to scroll the column hidden behind them. Over the
// list of backups it scrolls the list, a line a notch, and stops at its ends.
func TestTheWheelStaysInsideAPanel(t *testing.T) {
	var many []reading.Card
	for i := 1; i <= 12; i++ {
		many = append(many, reading.Card{Number: i, Title: "Task", State: "todo"})
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
	box := ChoosingOver(w.width, w.height)
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
