package gui

import (
	"os/exec"
	"runtime"
	"strings"
)

// Notice is a desktop notification the window wants sent: what it is about, and the line under it.
type Notice struct{ Title, Said string }

var appleQuoted = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

// NoticeCommand is the command that has the system show a notice: osascript on macOS, notify-send elsewhere.
func NoticeCommand(goos string, n Notice) (string, []string) {
	if goos == "darwin" {
		return "osascript", []string{"-e", `display notification "` + appleQuoted.Replace(n.Said) + `" with title "` + appleQuoted.Replace(n.Title) + `"`}
	}
	return "notify-send", []string{n.Title, n.Said}
}

// Notify starts the command that shows a notice, and does not wait for it.
func Notify(n Notice) error {
	name, args := NoticeCommand(runtime.GOOS, n)
	return started(exec.Command(name, args...))
}
