//go:build !windows

package engine

import (
	"os/exec"
	"strconv"
	"strings"
)

func ancestry(pid int) []string {
	var commands []string
	for at := pid; at > 1; {
		out, err := exec.Command("ps", "-o", "ppid=,comm=", "-p", strconv.Itoa(at)).Output()
		if err != nil {
			break
		}
		ppid, comm, _ := strings.Cut(strings.TrimSpace(string(out)), " ")
		if at != pid {
			commands = append(commands, strings.TrimSpace(comm))
		}
		at, _ = strconv.Atoi(ppid)
	}
	return commands
}

func ttyOf(pid int) string {
	said, err := exec.Command("ps", "-o", "tty=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(said))
}
