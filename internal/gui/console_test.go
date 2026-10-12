package gui

import "testing"

// A console the window lets go of is one nobody else is attached to: the one
// Windows opened for a binary started from the Explorer. One shared with the
// shell it was typed in stays, and what the window says on it still reaches
// whoever typed it — a trace written with CAIRN_GUI_TRACE among it.
func TestTheWindowLetsGoOfAConsoleOfItsOwnAlone(t *testing.T) {
	for attached, want := range map[uint32]bool{0: false, 1: true, 2: false, 3: false} {
		if got := ownsConsole(attached); got != want {
			t.Errorf("with %d processes on the console, the window lets go of it: %v", attached, got)
		}
	}
}
