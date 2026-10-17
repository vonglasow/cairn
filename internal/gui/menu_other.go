//go:build !windows && !(darwin && !ios)

package gui

import (
	"gioui.org/io/event"

	"cairn/internal/engine"
)

func installMenu(event.Event, engine.Pages, chan<- string, func()) bool { return false }
