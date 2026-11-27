package gui

import (
	"cairn/internal/board"
	"time"

	"cairn/internal/engine"
)

func viewedHome(width, height int) Window {
	w := homeWindow(width, height).At(time.Now())
	w.Tab = board.Tasks
	w.Snap.Views = []board.Viewing{
		{View: engine.View{Name: "Home", Place: 1, Holds: engine.Holds{Objective: []int{1}}}, Kept: []int{1, 2, 3}},
		{View: engine.View{Name: "Done", Place: 2, Holds: engine.Holds{State: "done"}}, Kept: []int{2}},
	}
	return w
}
