//go:build !windows && !(darwin && !ios)

package gui

import "gioui.org/io/event"

// installMenu has nothing to install where the system gives a window no menu
// bar of its own: the button at the end of the tabs stands in for one.
func installMenu(event.Event, chan<- string, func()) bool { return false }
