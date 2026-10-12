//go:build windows

package gui

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32           = windows.NewLazySystemDLL("kernel32.dll")
	consoleProcessList = kernel32.NewProc("GetConsoleProcessList")
	freeConsole        = kernel32.NewProc("FreeConsole")
)

// leaveConsole lets go of a console the window was started on alone, so a
// window opened from the Explorer has no console behind it. cairn is a console
// program, which the command line needs; the window is the one that has none.
func leaveConsole() {
	var processes [2]uint32
	attached, _, _ := consoleProcessList.Call(uintptr(unsafe.Pointer(&processes[0])), 2)
	if ownsConsole(uint32(attached)) {
		freeConsole.Call()
	}
}
