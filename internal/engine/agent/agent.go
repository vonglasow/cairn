// Package agent holds a Claude session running on the machine, as Claude Code itself keeps it, and what the window shows of it.
package agent

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// State is what a session is doing, as its card says it.
type State string

// The three states a session running is in.
const (
	Working State = "working"
	Waiting State = "waiting"
	Idle    State = "idle"
)

// Window and LargeWindow are how many tokens a session's context holds, the measure its context is a share of: the larger once it holds more than the smaller.
const (
	Window      = 200000
	LargeWindow = 1000000
)

// Machine is what the engine asks the machine about a session: whether its process still runs, the terminal it runs in, bringing that terminal to the front, starting a command in a terminal, and which terminals are installed.
type Machine interface {
	Alive(pid int, started string) bool
	Terminal(pid int) Terminal
	Reach(t Terminal) error
	Start(program, folder, command string) error
	Terminals() []string
}

// Terminal is the terminal a session runs in: the program, the tty, the session's process, and whether the window can bring it to the front.
type Terminal struct {
	Program   string `json:"program,omitempty"`
	TTY       string `json:"tty,omitempty"`
	PID       int    `json:"pid,omitempty"`
	Reachable bool   `json:"reachable"`
}

// Terminals is the terminals installed that a session can be started in.
type Terminals struct {
	Terminals []string `json:"terminals"`
}

// Started is a session started in a terminal: the terminal, the folder and the command it runs.
type Started struct {
	Terminal string `json:"terminal"`
	Folder   string `json:"folder"`
	Command  string `json:"command"`
}

// Ask is what a session waits on the window for: a permission — the tool, what it acts on, and whether it can be allowed for good — or the questions Claude asks.
type Ask struct {
	Tool      string     `json:"tool"`
	On        string     `json:"on,omitempty"`
	Always    bool       `json:"always" jsonschema:"whether Claude suggested a rule that allows it for good"`
	Questions []Question `json:"questions,omitempty" jsonschema:"the questions Claude asks, when it asks some"`
}

// AskUserQuestion is the tool Claude asks the person questions with.
const AskUserQuestion = "AskUserQuestion"

// Question is one question Claude asks: what it asks, a short header, the options, and whether several can be chosen.
type Question struct {
	Question string   `json:"question"`
	Header   string   `json:"header,omitempty"`
	Options  []Option `json:"options"`
	Multi    bool     `json:"multiSelect"`
}

// Option is one answer a question offers: its label, and what it means.
type Option struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// Held is the task a session holds: its number, title and the steps left open.
type Held struct {
	Number int      `json:"number"`
	Title  string   `json:"title"`
	Open   []string `json:"open,omitempty"`
}

// Agent is one Claude session running, as its card shows it.
type Agent struct {
	ID        string    `json:"id"`
	Short     string    `json:"short" jsonschema:"the first 6 characters of the id"`
	Name      string    `json:"name,omitempty"`
	State     State     `json:"state" jsonschema:"working, waiting or idle"`
	Since     time.Time `json:"since" jsonschema:"when it entered that state"`
	Waiting   string    `json:"waiting,omitempty" jsonschema:"what it waits for, while it waits"`
	Doing     string    `json:"doing,omitempty" jsonschema:"the last tool it used, and on what"`
	Folder    string    `json:"folder"`
	Project   string    `json:"project" jsonschema:"the repository it works in: its folder's name, a worktree's own repository's"`
	Task      *Held     `json:"task,omitempty" jsonschema:"the task it holds; left out when none"`
	Suggested *Held     `json:"suggested,omitempty" jsonschema:"the task in focus it may be working on, while it holds none"`
	Shared    int       `json:"shared" jsonschema:"how many other sessions hold the same task"`
	Terminal  Terminal  `json:"terminal"`
	Branch    string    `json:"branch,omitempty"`
	Model     string    `json:"model,omitempty"`
	Tokens    int       `json:"tokens" jsonschema:"the tokens its context holds"`
	Context   int       `json:"context" jsonschema:"the share of its context window used, in percent"`
	Silent    time.Time `json:"silent,omitzero" jsonschema:"while it works but its transcript is not written to for long, when it last was"`
	Resume    string    `json:"resume" jsonschema:"the command that resumes it in a terminal of one's own"`
	Asks      *Ask      `json:"asks,omitempty" jsonschema:"the permission it waits on the window for; left out when none"`
}

// Agents is how a listing of sessions crosses, with how many wait for the person.
type Agents struct {
	Agents  []Agent `json:"agents"`
	Waiting int     `json:"waiting" jsonschema:"how many sessions wait for the person"`
}

// StateOf reads a session's state from the status and what it waits for, as Claude Code writes them.
func StateOf(status, waitingFor string) State {
	switch {
	case waitingFor != "":
		return Waiting
	case status == "idle":
		return Idle
	}
	return Working
}

// ShortOf is a session's id cut to its first 6 characters.
func ShortOf(id string) string {
	if len(id) <= 6 {
		return id
	}
	return id[:6]
}

// ContextOf is the share of the window a number of tokens fills, in percent: of the larger window once they are more than the smaller holds.
func ContextOf(tokens int) int {
	if tokens > Window {
		return tokens * 100 / LargeWindow
	}
	return tokens * 100 / Window
}

// Counted sets how many other sessions hold each session's task, and how many wait.
func Counted(listed []Agent) Agents {
	holding := map[int]int{}
	for _, a := range listed {
		if a.Task != nil {
			holding[a.Task.Number]++
		}
	}
	out := Agents{Agents: listed}
	for i, a := range listed {
		if a.Task != nil {
			out.Agents[i].Shared = holding[a.Task.Number] - 1
		}
		if a.State == Waiting {
			out.Waiting++
		}
	}
	if out.Agents == nil {
		out.Agents = []Agent{}
	}
	return out
}

// AgentsText is a listing of sessions as a person reads it.
func AgentsText(a Agents) string {
	if len(a.Agents) == 0 {
		return "no Claude session is running\n"
	}
	var said strings.Builder
	for _, one := range a.Agents {
		task := "no task"
		if one.Task != nil {
			task = fmt.Sprintf("#%d %s", one.Task.Number, one.Task.Title)
		}
		line := fmt.Sprintf("%s  %-7s  %s  %s", one.Short, one.State, task, one.Folder)
		if one.Waiting != "" {
			line += "  waits for " + one.Waiting
		}
		said.WriteString(line + "\n")
	}
	return said.String()
}

// Quoted is a text as a shell reads it, one word whatever it holds.
func Quoted(said string) string { return "'" + strings.ReplaceAll(said, "'", `'\''`) + "'" }

func word(said string) string {
	if said != "" && strings.Trim(said, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_") == "" {
		return said
	}
	return Quoted(said)
}

// ResumeOf is the command that resumes a session in its folder, for a person to run in a terminal of their own.
func ResumeOf(folder, id string) string {
	return "cd " + Quoted(folder) + " && claude --resume " + word(id)
}

// CommandFor is the command that starts a session: claude, asked to work on the task when one is named.
func CommandFor(n int) string {
	if n == 0 {
		return "claude"
	}
	return "claude " + Quoted(fmt.Sprintf("let's work on cairn#%d", n))
}

// ProgramOf is the program a session started from somewhere other than a terminal runs in, read off the entrypoint Claude Code writes; nothing for a terminal.
func ProgramOf(entrypoint string) string {
	switch {
	case entrypoint == "" || entrypoint == "cli":
		return ""
	case strings.Contains(entrypoint, "vscode"):
		return "VS Code"
	case strings.Contains(entrypoint, "jetbrains"):
		return "JetBrains"
	case strings.Contains(entrypoint, "desktop"):
		return "Claude Desktop"
	}
	return entrypoint
}

var modelID = regexp.MustCompile(`^claude-(opus|sonnet|haiku)-(\d+)-(\d+)(?:-\d{8})?$`)

// ModelSaid is a model's id as a person reads it, "claude-opus-5-5" as "Opus 5.5"; an id it does not know as it is.
func ModelSaid(id string) string {
	named := modelID.FindStringSubmatch(id)
	if named == nil {
		return id
	}
	return strings.ToUpper(named[1][:1]) + named[1][1:] + " " + named[2] + "." + named[3]
}

var doneTo = map[string]string{"Read": "Reads ", "Edit": "Edits ", "Write": "Writes ", "NotebookEdit": "Edits "}

// DoingOf is the last tool a session used, said as a person reads it: the description Claude gave a command, else the program it runs, the name of a file it reads or writes, what it searches — never a whole path or a whole command.
func DoingOf(tool, on, description string) string {
	switch {
	case tool == "Bash" && description != "":
		return description
	case tool == "Bash":
		for _, part := range strings.Split(on, "&&") {
			if words := strings.Fields(part); len(words) > 0 && words[0] != "cd" {
				return "Runs " + filepath.Base(words[0])
			}
		}
		return tool
	case doneTo[tool] != "":
		return doneTo[tool] + filepath.Base(on)
	case tool == "Grep" || tool == "Glob":
		return "Searches " + on
	case tool == "WebFetch" || tool == "WebSearch":
		return "Reads the web"
	case tool == "Task" || tool == "Agent":
		return "Runs a subagent"
	}
	if server, name, ok := strings.Cut(strings.TrimPrefix(tool, "mcp__"), "__"); ok && strings.HasPrefix(tool, "mcp__") {
		return "Uses " + server + " " + name
	}
	return tool
}

// ProjectOf is the project a session works in: its folder's name, or, in a worktree Claude made, the name of the repository it was made in.
func ProjectOf(folder string) string {
	if repository, _, ok := strings.Cut(folder, "/.claude/worktrees/"); ok {
		return filepath.Base(repository)
	}
	return filepath.Base(folder)
}

// ContextLevel is how full a session's context is: roomy, high from 70%, nearly full from 85%.
type ContextLevel int

// The levels a context is at.
const (
	ContextRoomy ContextLevel = iota
	ContextHigh
	ContextFull
)

// ContextLevelOf is the level a context filled to percent is at.
func ContextLevelOf(percent int) ContextLevel {
	switch {
	case percent >= 85:
		return ContextFull
	case percent >= 70:
		return ContextHigh
	}
	return ContextRoomy
}
