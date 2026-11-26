package engine

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"cairn/internal/engine/agent"
)

type processes struct{}

func (processes) Alive(pid int, _ string) bool {
	return alive(pid)
}

func (processes) Terminal(pid int) agent.Terminal {
	t := terminalAmong(ancestry(pid))
	t.TTY = ttyOf(pid)
	return t
}

func terminalAmong(ancestors []string) agent.Terminal {
	for _, command := range ancestors {
		if name, ok := terminalNamed(command); ok {
			return agent.Terminal{Program: name, Reachable: name == "iTerm2" || name == "Terminal" || name == "kitty" || name == "Ghostty" && runtime.GOOS == "darwin"}
		}
	}
	return agent.Terminal{}
}

func terminalNamed(command string) (string, bool) {
	for _, known := range []struct{ in, name string }{
		{"iTerm", "iTerm2"}, {"Terminal.app", "Terminal"}, {"ghostty", "Ghostty"}, {"WezTerm", "WezTerm"},
		{"kitty", "kitty"}, {"Alacritty", "Alacritty"}, {"Code Helper", "VS Code"}, {"gnome-terminal", "GNOME Terminal"}, {"konsole", "Konsole"},
		{"WindowsTerminal.exe", "Windows Terminal"}, {"Code.exe", "VS Code"}, {"wezterm-gui.exe", "WezTerm"}, {`\WindowsApps\Claude_`, "Claude Desktop"}, {`\AnthropicClaude\`, "Claude Desktop"},
	} {
		if strings.Contains(command, known.in) {
			return known.name, true
		}
	}
	return "", false
}

const reachITerm = `on run argv
	tell application "iTerm2"
		repeat with w in windows
			repeat with t in tabs of w
				repeat with s in sessions of t
					if tty of s is item 1 of argv then
						select w
						tell t to select
						tell s to select
						activate
						return
					end if
				end repeat
			end repeat
		end repeat
	end tell
	error "no session of iTerm2 runs on " & item 1 of argv
end run`

const reachTerminal = `on run argv
	tell application "Terminal"
		repeat with w in windows
			repeat with t in tabs of w
				if tty of t is item 1 of argv then
					set selected of t to true
					set index of w to 1
					set frontmost of w to true
					activate
					return
				end if
			end repeat
		end repeat
	end tell
	error "no tab of Terminal runs on " & item 1 of argv
end run`

const reachGhostty = `on run argv
	tell application "Ghostty"
		set found to every terminal whose tty is item 1 of argv
		if (count of found) is 0 then error "no terminal of Ghostty runs on " & item 1 of argv
		focus (item 1 of found)
		activate
	end tell
end run`

const startGhostty = `on run argv
	tell application "Ghostty"
		activate
		set made to new surface configuration
		set initial working directory of made to item 1 of argv
		set initial input of made to (item 2 of argv) & linefeed
		if (count of windows) is 0 then
			new window with configuration made
		else
			new tab in front window with configuration made
		end if
	end tell
end run`

const startITerm = `on run argv
	tell application "iTerm2"
		activate
		if (count of windows) is 0 then
			create window with default profile
		else
			tell current window to create tab with default profile
		end if
		tell current session of current window to write text (item 1 of argv)
	end tell
end run`

const startTerminal = `on run argv
	tell application "Terminal"
		activate
		do script (item 1 of argv)
	end tell
end run`

func (processes) Reach(t agent.Terminal) error {
	switch t.Program {
	case "iTerm2":
		return exec.Command("osascript", "-e", reachITerm, "/dev/"+t.TTY).Run()
	case "Terminal":
		return exec.Command("osascript", "-e", reachTerminal, "/dev/"+t.TTY).Run()
	case "Ghostty":
		return exec.Command("osascript", "-e", reachGhostty, "/dev/"+t.TTY).Run()
	case "kitty":
		return exec.Command("kitty", "@", "focus-window", "--match", "pid:"+strconv.Itoa(t.PID)).Run()
	}
	return fmt.Errorf("cairn cannot bring %s to the front", t.Program)
}

func (processes) Start(program, folder, command string) error {
	var c *exec.Cmd
	switch program {
	case "iTerm2":
		return exec.Command("osascript", "-e", startITerm, "cd "+agent.Quoted(folder)+" && "+command).Run()
	case "Terminal":
		return exec.Command("osascript", "-e", startTerminal, "cd "+agent.Quoted(folder)+" && "+command).Run()
	case "Ghostty":
		if runtime.GOOS == "darwin" {
			return exec.Command("osascript", "-e", startGhostty, folder, command).Run()
		}
		c = exec.Command("ghostty", "--working-directory="+folder, "-e", "sh", "-c", command)
	case "kitty":
		c = exec.Command("kitty", "--detach", "--directory", folder, "sh", "-c", command)
	case "gnome-terminal":
		c = exec.Command("gnome-terminal", "--working-directory="+folder, "--", "sh", "-c", command)
	default:
		return fmt.Errorf("cairn cannot start a session in %s", program)
	}
	if err := c.Start(); err != nil {
		return err
	}
	return c.Process.Release()
}

func (processes) Terminals() []string {
	installed := []string{}
	for _, t := range []struct{ name, system, at string }{
		{"iTerm2", "darwin", "/Applications/iTerm.app"},
		{"Terminal", "darwin", "/System/Applications/Utilities/Terminal.app"},
		{"Ghostty", "darwin", "/Applications/Ghostty.app"},
		{"Ghostty", "linux", "ghostty"},
		{"kitty", "linux", "kitty"},
		{"gnome-terminal", "linux", "gnome-terminal"},
	} {
		if t.system != runtime.GOOS {
			continue
		}
		_, err := os.Stat(t.at)
		if t.system == "linux" {
			_, err = exec.LookPath(t.at)
		}
		if err == nil {
			installed = append(installed, t.name)
		}
	}
	return installed
}
