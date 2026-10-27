package debug

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Machine is what the machine is, a fact a line: its system, and on Linux and BSD its display session, desktop, input methods and keyboard, on a Mac and on Windows their version; then its language and its terminal. A fact nobody set reads unset, a tool that is not there not found: the machine that misbehaves is rarely the one every tool is on.
func Machine(goos, goarch string, getenv func(string) string, ran func(string, ...string) string) [][2]string {
	env := func(name string) string {
		if said := getenv(name); said != "" {
			return said
		}
		return "unset"
	}
	tool := func(name string, args ...string) string {
		if said := strings.Join(strings.Fields(ran(name, args...)), " "); said != "" {
			return said
		}
		return "not found"
	}
	facts := [][2]string{{"system", goos + "/" + goarch}}
	switch goos {
	case "darwin":
		facts = append(facts, [2]string{"macos", tool("sw_vers", "-productVersion")})
	case "windows":
		facts = append(facts, [2]string{"windows", tool("cmd", "/c", "ver")})
	case "linux", "freebsd", "openbsd", "netbsd":
		facts = append(facts,
			[2]string{"distribution", distribution()},
			[2]string{"session", env("XDG_SESSION_TYPE")},
			[2]string{"desktop", env("XDG_CURRENT_DESKTOP")},
			[2]string{"wayland", env("WAYLAND_DISPLAY")},
			[2]string{"x11", env("DISPLAY")},
			[2]string{"gtk im", env("GTK_IM_MODULE")},
			[2]string{"qt im", env("QT_IM_MODULE")},
			[2]string{"xmodifiers", env("XMODIFIERS")},
			[2]string{"ibus", tool("pgrep", "-x", "ibus-daemon")},
			[2]string{"fcitx", tool("pgrep", "-f", "fcitx")},
			[2]string{"keyboard", tool("setxkbmap", "-query")},
		)
	}
	language := getenv("LC_ALL")
	if language == "" {
		language = env("LANG")
	}
	return append(facts,
		[2]string{"language", language},
		[2]string{"terminal", env("TERM")},
		[2]string{"colours", env("COLORTERM")},
		[2]string{"terminal app", env("TERM_PROGRAM")},
	)
}

func distribution() string {
	read, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "not found"
	}
	for _, line := range strings.Split(string(read), "\n") {
		if name, ok := strings.CutPrefix(line, "PRETTY_NAME="); ok {
			return strings.Trim(name, `"`)
		}
	}
	return "not found"
}

// Ran runs a tool for what it says, given two seconds, and answers nothing when it is not there or says nothing.
func Ran(name string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	said, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		return ""
	}
	return string(said)
}
