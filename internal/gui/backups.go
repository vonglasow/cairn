package gui

// Backing the store up, and putting a backup back, from the window. Both were
// the command line's, and a person at the window went to a terminal to do
// either. ⌘B backs it up and says where the backup lies — a file that can be
// copied anywhere, which is what exporting is; ⇧B lists the backups, and
// enter puts one back once agreed to, the store backed up first by the engine
// as it always is. A file from elsewhere is put back the same way, once it
// lies among the backups. See docs/gui.md.

import (
	"fmt"
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/reading"
)

// Note is what the window last said that is not something gone wrong: where a
// backup lies, which one was put back. It goes as what went wrong does.
func (w Window) Note() string { return w.told }

// telling says something, and starts its time on screen over.
func (w Window) telling(said string) Window {
	w.told, w.toldFor = said, 0
	return w
}

// backingUp backs the store up and says where the backup lies.
func (w Window) backingUp() Window {
	made, err := w.work.Ask(engine.BackupNew, engine.Input{})
	if err != nil {
		return w.failing(err)
	}
	return w.telling(fmt.Sprintf("backup %d is %s", made.Backup.Number, made.Backup.Path))
}

// Backups is what the list of backups holds, as the engine listed them.
func (w Window) Backups() []domain.Backup { return w.backups }

// backupsState is what the list of backups keeps: the backups, the line the
// keyboard is on, and the first line in sight.
type backupsState struct {
	backups     []domain.Backup
	backupOn    int
	backupsFrom int
}

// backupSays is how a line of the list of backups reads.
func backupSays(one domain.Backup) string {
	return "backup " + strconv.Itoa(one.Number) + "   " + domain.FormatMoment(one.Made)
}

// BackupOn is the line of the list the keyboard is on.
func (w Window) BackupOn() int { return w.backupOn }

// listingBackups opens the list of backups, or says there is none to list.
func (w Window) listingBackups() Window {
	kept, err := w.work.Ask(engine.BackupList, engine.Input{})
	if err != nil {
		return w.failing(err)
	}
	if len(kept.Backups) == 0 {
		return w.telling("no backup yet: ⌘B makes one")
	}
	w.backups, w.backupOn, w.backupsFrom, w.screen = kept.Backups, 0, 0, OnBackups
	return w
}

// onBackups is what a key does on the list: the arrows and j k walk it, enter
// asks before putting the backup back, esc closes it.
func (w Window) onBackups(name string) Window {
	switch name {
	case "down", "j":
		w.backupOn = min(len(w.backups)-1, w.backupOn+1)
		w.backupsFrom = max(w.backupsFrom, w.backupOn-w.BackupsRoom()+1)
	case "up", "k":
		w.backupOn = max(0, w.backupOn-1)
		w.backupsFrom = min(w.backupsFrom, w.backupOn)
	case "enter":
		w.screen, w.question = OnAsking, aboutPuttingBack
	case "esc":
		w.screen = OnTheBoard
	}
	return w
}

// puttingBack puts the backup the list was on back, and says so.
func (w Window) puttingBack() Window {
	number := w.backups[w.backupOn].Number
	w.screen, w.question = OnTheBoard, aboutDeleting
	if _, err := w.work.Ask(engine.BackupRestore, reading.Numbered(number)); err != nil {
		return w.failing(err)
	}
	return w.Look().telling(fmt.Sprintf("backup %d was put back", number))
}

// BackupsFrom is the first backup of the list in sight. Every backup put back
// sets another aside, so the list grows past its panel, and the line the
// keyboard is on is kept among the ones drawn.
func (w Window) BackupsFrom() int { return w.backupsFrom }

// scrollingBackups turns the list of backups under the pointer, a line a
// notch, between its first screenful and its last. The list is drawn in the
// panel a list of choices is, and the two are never open together, so the
// part of a notch left over is kept where the choices keep theirs.
func (w Window) scrollingBackups(p Point, by int) Window {
	if !w.BackupsBox().In(p) {
		return w
	}
	lines, left := turning(w.spareChoosing, by)
	w.spareChoosing = left
	w.backupsFrom = max(0, min(w.backupsFrom+lines, len(w.backups)-w.BackupsRoom()))
	return w
}

// BackupsRoom is how many backups the list shows at once: the lines between
// its heading and its offers.
func (w Window) BackupsRoom() int {
	box := w.BackupsBox()
	return max(1, (PanelOfferAt(box, 0).Y-padding-FirstRow(box)-between)/menuRow)
}

// clickingBackups asks to put back the backup clicked, and a click off the
// list closes it.
func (w Window) clickingBackups(p Point) Window {
	box := w.BackupsBox()
	for i := w.backupsFrom; i < min(len(w.backups), w.backupsFrom+w.BackupsRoom()); i++ {
		if BackupRowAt(box, i-w.backupsFrom).In(p) {
			w.backupOn = i
			w.screen, w.question = OnAsking, aboutPuttingBack
			return w
		}
	}
	if !box.In(p) {
		w.screen = OnTheBoard
	}
	return w
}

// BackupRowAt is where the i-th backup of the list is written.
func BackupRowAt(box Box, i int) Box { return MenuRowAt(Box{X: box.X, Y: FirstRow(box), W: box.W}, i) }
