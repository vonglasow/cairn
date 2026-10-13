package gui

import "testing"

func TestTheWindowLetsGoOfAConsoleOfItsOwnAlone(t *testing.T) {
	for attached, want := range map[uint32]bool{0: false, 1: true, 2: false, 3: false} {
		if got := ownsConsole(attached); got != want {
			t.Errorf("with %d processes on the console, the window lets go of it: %v", attached, got)
		}
	}
}
