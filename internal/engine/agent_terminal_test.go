package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/engine/agent"
)

func TestTerminalAmong(t *testing.T) {
	for _, tt := range []struct {
		name      string
		ancestors []string
		want      agent.Terminal
	}{
		{"Windows Terminal", []string{"powershell.exe", "WindowsTerminal.exe", "explorer.exe"}, agent.Terminal{Program: "Windows Terminal"}},
		{"Claude Desktop from the Store", []string{`C:\Program Files\WindowsApps\Claude_6.28899.6.0_x64__pzs8sxrjxfjjc\app\Claude.exe`, "explorer.exe"}, agent.Terminal{Program: "Claude Desktop"}},
		{"Claude Desktop from its installer", []string{`C:\Users\me\AppData\Local\AnthropicClaude\app-1.2.3\claude.exe`, "explorer.exe"}, agent.Terminal{Program: "Claude Desktop"}},
		{"a Claude Code session above", []string{"bash.exe", "cmd.exe", `C:\Users\me\.local\bin\claude.exe`, "pwsh.exe", "WindowsTerminal.exe"}, agent.Terminal{Program: "Windows Terminal"}},
		{"VS Code on Windows", []string{"pwsh.exe", "Code.exe"}, agent.Terminal{Program: "VS Code"}},
		{"WezTerm on Windows", []string{"cmd.exe", "wezterm-gui.exe"}, agent.Terminal{Program: "WezTerm"}},
		{"Claude Code is no terminal", []string{"claude.exe", "cmd.exe", "explorer.exe"}, agent.Terminal{}},
		{"iTerm2", []string{"-zsh", "/Applications/iTerm.app/Contents/MacOS/iTerm2"}, agent.Terminal{Program: "iTerm2", Reachable: true}},
		{"nothing known", []string{"bash", "sshd"}, agent.Terminal{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := terminalAmong(tt.ancestors); got != tt.want {
				t.Errorf("terminalAmong(%q) = %+v, want %+v", tt.ancestors, got, tt.want)
			}
		})
	}
}

const ancestryVariable = "CAIRN_TEST_SAY_ANCESTRY"

func TestAncestryStartsAtTheParent(t *testing.T) {
	if os.Getenv(ancestryVariable) != "" {
		fmt.Println(strings.Join(ancestry(os.Getpid()), "\n"))
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(exe, "-test.run=^TestAncestryStartsAtTheParent$")
	child.Env = append(os.Environ(), ancestryVariable+"=1")
	said, err := child.Output()
	if err != nil {
		t.Fatal(err)
	}
	first, _, _ := strings.Cut(string(said), "\n")
	if want := strings.TrimSuffix(filepath.Base(exe), ".exe"); !strings.Contains(first, want) {
		t.Errorf("the child's first ancestor is %q, want its parent %q", first, want)
	}
}
