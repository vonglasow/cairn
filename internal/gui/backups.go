package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/backup"
	"cairn/internal/engine/calendar"
	"cairn/internal/gui/geom"
	"fmt"
	"strconv"

	"cairn/internal/engine"
)

func (w Window) telling(said string) Window {
	w.Telling, w.ToldFor = said, 0
	return w
}

func (w Window) backingUp() Window {
	made, err := w.work.Ask(engine.BackupNew, engine.Input{})
	if err != nil {
		return w.failing(err)
	}
	return w.telling(fmt.Sprintf("backup %d is %s", made.Backup.Number, made.Backup.Path))
}

func backupSays(one backup.Backup) string {
	return "backup " + strconv.Itoa(one.Number) + "   " + calendar.FormatMoment(one.Made) + "   " + backup.FormatSize(one.Size)
}

func (w Window) listingBackups() Window {
	kept, err := w.work.Ask(engine.BackupList, engine.Input{})
	if err != nil {
		return w.failing(err)
	}
	if len(kept.Backups) == 0 {
		return w.telling("no backup yet: `⌘B` makes one")
	}
	w.Backups, w.BackupOn, w.BackupsFrom, w.Screen = kept.Backups, 0, 0, board.OnBackups
	return w
}

func (w Window) onBackups(name string) Window {
	switch name {
	case "down", "j":
		w.BackupOn = min(len(w.Backups)-1, w.BackupOn+1)
		w.BackupsFrom = max(w.BackupsFrom, w.BackupOn-w.BackupsRoom()+1)
	case "up", "k":
		w.BackupOn = max(0, w.BackupOn-1)
		w.BackupsFrom = min(w.BackupsFrom, w.BackupOn)
	case "enter":
		w.Screen, w.Question = board.OnAsking, board.AboutPuttingBack
	case "esc":
		w.Screen = board.OnTheBoard
	}
	return w
}

func (w Window) puttingBack() Window {
	number := w.Backups[w.BackupOn].Number
	w.Screen, w.Question = board.OnTheBoard, board.AboutDeleting
	if _, err := w.work.Ask(engine.BackupRestore, engine.Agreed(board.Numbered(number))); err != nil {
		return w.failing(err)
	}
	return w.Look().telling(fmt.Sprintf("backup %d was restored", number))
}

func (w Window) scrollingBackups(p geom.Point, by int) Window {
	if !w.BackupsBox().In(p) {
		return w
	}
	lines, left := turning(w.spareChoosing, by)
	w.spareChoosing = left
	w.BackupsFrom = max(0, min(w.BackupsFrom+lines, len(w.Backups)-w.BackupsRoom()))
	return w
}

// BackupsRoom is how many backups the list shows at once: the lines between its heading and its offers.
func (w Window) BackupsRoom() int { return listRoom(w.BackupsBox()) }

func listRoom(box geom.Box) int {
	return max(1, (PanelOfferAt(box, 0).Y-padding-FirstRow(box)-between)/menuRow)
}

func (w Window) clickingBackups(p geom.Point) Window {
	box := w.BackupsBox()
	for i := w.BackupsFrom; i < min(len(w.Backups), w.BackupsFrom+w.BackupsRoom()); i++ {
		if ListRowAt(box, i-w.BackupsFrom).In(p) {
			w.BackupOn = i
			w.Screen, w.Question = board.OnAsking, board.AboutPuttingBack
			return w
		}
	}
	if !box.In(p) {
		w.Screen = board.OnTheBoard
	}
	return w
}

// ListRowAt is where the i-th line of a list panel in sight is written: a backup, a label.
func ListRowAt(box geom.Box, i int) geom.Box {
	return MenuRowAt(geom.Box{X: box.X, Y: FirstRow(box), W: box.W}, i)
}
