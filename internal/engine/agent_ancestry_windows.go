package engine

import (
	"cmp"
	"strings"
	"syscall"
	"unsafe"
)

type process struct {
	parent int
	exe    string
}

func ancestry(pid int) []string {
	table := processTable()
	var exes []string
	seen := map[int]bool{pid: true}
	for at := table[pid].parent; at != 0 && !seen[at]; at = table[at].parent {
		p, ok := table[at]
		if !ok {
			break
		}
		seen[at] = true
		if strings.EqualFold(p.exe, "claude.exe") {
			p.exe = cmp.Or(imageOf(at), p.exe)
		}
		exes = append(exes, p.exe)
	}
	return exes
}

var queryFullProcessImageName = syscall.NewLazyDLL("kernel32.dll").NewProc("QueryFullProcessImageNameW")

func imageOf(pid int) string {
	h, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return ""
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	path := make([]uint16, syscall.MAX_LONG_PATH)
	size := uint32(len(path))
	if ok, _, _ := queryFullProcessImageName.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&path[0])), uintptr(unsafe.Pointer(&size))); ok == 0 {
		return ""
	}
	return syscall.UTF16ToString(path[:size])
}

func processTable() map[int]process {
	table := map[int]process{}
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return table
	}
	defer func() { _ = syscall.CloseHandle(snapshot) }()
	entry := syscall.ProcessEntry32{Size: uint32(unsafe.Sizeof(syscall.ProcessEntry32{}))}
	for err = syscall.Process32First(snapshot, &entry); err == nil; err = syscall.Process32Next(snapshot, &entry) {
		table[int(entry.ProcessID)] = process{parent: int(entry.ParentProcessID), exe: syscall.UTF16ToString(entry.ExeFile[:])}
	}
	return table
}

func ttyOf(int) string {
	return ""
}
