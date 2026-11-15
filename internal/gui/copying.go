package gui

import (
	"cairn/internal/board"
	"cairn/internal/build"
	"cairn/internal/gui/geom"
)

// CopyMark is the mark drawn before a path the window shows, and the path a click on it copies.
type CopyMark struct {
	At   geom.Box
	Said string
}

const (
	copySide = 12
	copyRoom = copySide + 6
)

func markBefore(line geom.Box, said string) CopyMark {
	return CopyMark{At: geom.Box{X: line.X, Y: line.Y + (line.H-copySide)/2, W: copySide, H: copySide}, Said: said}
}

// CopyMarks is every path the open panel shows with a mark before it to copy it: the places and the file of the configuration, where an open task's context lives, and where cairn is kept.
func (w Window) CopyMarks() []CopyMark {
	switch {
	case w.Screen == board.OnConfiguration && !w.ChooserOpen && !w.IntrosOpen:
		var marks []CopyMark
		region := SettingsRows(w.SettingsBox())
		for _, row := range w.SettingRows() {
			if copiedKey(row.Key.Name) && row.Fixed && region.In(geom.Point{X: row.Control.X, Y: row.Control.Y + row.Control.H/2}) {
				marks = append(marks, markBefore(row.Control, w.PlaceOf(row.Key.Name)))
			}
		}
		path := w.ScopePathAt()
		return append(marks, markBefore(geom.Box{X: path.X - copyRoom, Y: path.Y, H: path.H}, w.SettingsWritten()))
	case w.Screen == board.OnAsking && w.Question == board.AboutCairn:
		return []CopyMark{markBefore(w.SaidAt(len(AboutSays(build.Read()))-1), build.Home)}
	case w.Context() != "":
		path := w.contextPath()
		return []CopyMark{markBefore(geom.Box{X: path.X - copyRoom, Y: path.Y, H: path.H}, w.Context())}
	}
	return nil
}

func copiedKey(key string) bool { return key == "store" || key == "notes" }

func (w Window) copyingAt(p geom.Point) (Window, bool) {
	for _, mark := range w.CopyMarks() {
		if mark.At.In(p) {
			w.copying = mark.Said
			return w.telling("copied " + mark.Said), true
		}
	}
	return w, false
}
