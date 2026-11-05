//go:build !windows

package mcp

import (
	"os"
	"syscall"
)

func replaceItself(path string, payload []byte) error {
	return syscall.Exec(path, os.Args, append(os.Environ(), carriedVariable+"="+string(payload)))
}
