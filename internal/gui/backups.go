package gui

import (
	"fmt"
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

// Note is what the window last said that is not something gone wrong: where a backup lies, which one was put back.
func (w Window) Note() string { return w.told }

func (w Window) telling(said string) Window {
	w.told, w.toldFor = said, 0
	return w
}

func (w Window) backingUp() Window {
	made, err := w.work.Ask(engine.BackupNew, engine.Input{})
	if err != nil {
		return w.failing(err)
	}
	return w.telling(fmt.Sprintf("backup %d is %s", made.Backup.Number, made.Backup.Path))
}

// Backups is what the list of backups holds, as the engine listed them.
func (w Window) Backups() []domain.Backup { return w.backups }

type backupsState struct {
	backups     []domain.Backup
	backupOn    int
	backupsFrom int
}

func backupSays(one domain.Backup) string {
	return "backup " + strconv.Itoa(one.Number) + "   " + domain.FormatMoment(one.Made)
}

// BackupOn is the line of the list the keyboard is on.
func (w Window) BackupOn() int { return w.backupOn }

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

func (w Window) puttingBack() Window {
	number := w.backups[w.backupOn].Number
	w.screen, w.question = OnTheBoard, aboutDeleting
	if _, err := w.work.Ask(engine.BackupRestore, engine.Agreed(engine.Numbered(number))); err != nil {
		return w.failing(err)
	}
	return w.Look().telling(fmt.Sprintf("backup %d was restored", number))
}

// BackupsFrom is the first backup of the list in sight.
func (w Window) BackupsFrom() int { return w.backupsFrom }

func (w Window) scrollingBackups(p Point, by int) Window {
	if !w.BackupsBox().In(p) {
		return w
	}
	lines, left := turning(w.spareChoosing, by)
	w.spareChoosing = left
	w.backupsFrom = max(0, min(w.backupsFrom+lines, len(w.backups)-w.BackupsRoom()))
	return w
}

// BackupsRoom is how many backups the list shows at once: the lines between its heading and its offers.
func (w Window) BackupsRoom() int {
	box := w.BackupsBox()
	return max(1, (PanelOfferAt(box, 0).Y-padding-FirstRow(box)-between)/menuRow)
}

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
