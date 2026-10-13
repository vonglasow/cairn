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

func leaveConsole() {
	var processes [2]uint32
	attached, _, _ := consoleProcessList.Call(uintptr(unsafe.Pointer(&processes[0])), 2)
	if ownsConsole(uint32(attached)) {
		freeConsole.Call()
	}
}
