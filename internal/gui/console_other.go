//go:build !windows

package gui

// leaveConsole has nothing to let go of where a program started from a
// desktop is given no console.
func leaveConsole() {}
