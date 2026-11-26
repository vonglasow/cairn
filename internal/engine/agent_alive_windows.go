package engine

import "syscall"

const (
	processQueryLimitedInformation = 0x1000
	stillActive                    = 259
)

func alive(pid int) bool {
	h, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	var code uint32
	return syscall.GetExitCodeProcess(h, &code) == nil && code == stillActive
}
