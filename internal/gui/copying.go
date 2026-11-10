package gui

import "cairn/internal/build"

// CopyMark is the mark drawn before a path the window shows, and the path a click on it copies.
type CopyMark struct {
	At   Box
	Said string
}

const (
	copySide = 12
	copyRoom = copySide + 6
)

func markBefore(line Box, said string) CopyMark {
	return CopyMark{At: Box{X: line.X, Y: line.Y + (line.H-copySide)/2, W: copySide, H: copySide}, Said: said}
}

// CopyMarks is every path the open panel shows with a mark before it to copy it: the places and the file of the configuration, where an open task's context lives, and where cairn is kept.
func (w Window) CopyMarks() []CopyMark {
	switch {
	case w.screen == OnConfiguration && !w.chooserOpen:
		var marks []CopyMark
		region := SettingsRows(w.SettingsBox())
		for _, row := range w.SettingRows() {
			if copiedKey(row.Key.Name) && row.Fixed && region.In(Point{X: row.Control.X, Y: row.Control.Y + row.Control.H/2}) {
				marks = append(marks, markBefore(row.Control, w.placeOf(row.Key.Name)))
			}
		}
		return marks
	case w.screen == OnAsking && w.question == aboutCairn:
		said := AboutSays(build.Read())
		box := w.QuestionBox()
		return []CopyMark{markBefore(RowIn(box, FirstRow(box)+(len(said)-1)*apart, lineHigh), build.Home)}
	case w.Context() != "":
		path := w.contextPath()
		return []CopyMark{markBefore(Box{X: path.X - copyRoom, Y: path.Y, H: path.H}, w.Context())}
	}
	return nil
}

func copiedKey(key string) bool { return key == "store" || key == "notes" }

func (w Window) placeOf(key string) string {
	if key == "notes" {
		return w.configured.Notes.At
	}
	return w.configured.Store.At
}

func (w Window) copyingAt(p Point) (Window, bool) {
	for _, mark := range w.CopyMarks() {
		if mark.At.In(p) {
			w.copying = mark.Said
			return w.telling("copied " + mark.Said), true
		}
	}
	return w, false
}
