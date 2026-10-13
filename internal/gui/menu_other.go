//go:build !windows && !(darwin && !ios)

package gui

import "gioui.org/io/event"

func installMenu(event.Event, chan<- string, func()) bool { return false }
