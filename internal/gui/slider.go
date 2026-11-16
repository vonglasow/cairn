package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"strconv"

	"cairn/internal/store/files"
)

const (
	sliderLabel = 48
	sliderStep  = 5
)

// PercentShown is the percent a row of a percentage shows: where its slider is held, or its value.
func (w Window) PercentShown(row SettingRow) int {
	if w.sliding == row.Key.Name {
		return w.slidTo
	}
	n, _ := strconv.Atoi(row.Value)
	return n
}

func sliderTrack(control geom.Box) geom.Box {
	return geom.Box{X: control.X, Y: control.Y, W: control.W - sliderLabel - between, H: control.H}
}

func percentAt(track geom.Box, x int) int {
	span := files.MostGhost - files.LeastGhost
	at := files.LeastGhost + ((x-track.X)*span*2/max(1, track.W)/sliderStep+1)/2*sliderStep
	return min(files.MostGhost, max(files.LeastGhost, at))
}

func (w Window) pressingSlider(p geom.Point) (Window, bool) {
	if w.Screen != board.OnConfiguration || w.ChooserOpen || w.IntrosOpen {
		return w, false
	}
	for _, row := range w.SettingRows() {
		if row.Key.Kind == files.KindPercent && !row.Off && len(row.Parts) == 1 && row.Parts[0].In(p) {
			w.sliding, w.slidTo = row.Key.Name, percentAt(row.Parts[0], p.X)
			return w, true
		}
	}
	return w, false
}

func (w Window) slidingTo(p geom.Point) Window {
	for _, row := range w.SettingRows() {
		if row.Key.Name == w.sliding {
			w.slidTo = percentAt(row.Parts[0], p.X)
		}
	}
	return w
}

func (w Window) lettingGoOfSlider() Window {
	key, to := w.sliding, w.slidTo
	w.sliding, w.pressed = "", false
	return w.setting(key, strconv.Itoa(to))
}

func (w Window) steppingPercent(row SettingRow, by int) Window {
	n := min(files.MostGhost, max(files.LeastGhost, w.PercentShown(row)+by))
	return w.setting(row.Key.Name, strconv.Itoa(n))
}
