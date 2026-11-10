//go:build windows

package gui

import (
	"syscall"
	"unsafe"
)

const clientAreaAnimation = 0x1042

func lessMotion() bool {
	var on int32
	user := syscall.NewLazyDLL("user32.dll").NewProc("SystemParametersInfoW")
	if found, _, _ := user.Call(clientAreaAnimation, 0, uintptr(unsafe.Pointer(&on)), 0); found == 0 {
		return false
	}
	return on == 0
}
