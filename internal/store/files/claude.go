package files

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ClaudeVariable is the variable Claude Code reads its folder from; without it, the folder is ~/.claude.
const ClaudeVariable = "CLAUDE_CONFIG_DIR"

// ClaudeSession is a session as Claude Code keeps it in its sessions folder, one file a process.
type ClaudeSession struct {
	PID             int    `json:"pid"`
	SessionID       string `json:"sessionId"`
	Cwd             string `json:"cwd"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	WaitingFor      string `json:"waitingFor"`
	StatusUpdatedAt int64  `json:"statusUpdatedAt"`
	ProcStart       string `json:"procStart"`
	Entrypoint      string `json:"entrypoint"`
}

// ClaudeSaid is what a session's transcript last said: its branch, its model, the tokens its context holds, the last tool it used, on what, the description Claude gave it, and when it was last written, in milliseconds.
type ClaudeSaid struct {
	Branch      string
	Model       string
	Tokens      int
	Tool        string
	On          string
	Description string
	WrittenAt   int64
}

// ClaudeSessions reads every session Claude Code keeps in its folder; a file it cannot read is passed over.
func ClaudeSessions(folder string) []ClaudeSession {
	found, _ := filepath.Glob(filepath.Join(folder, "sessions", "*.json"))
	out := []ClaudeSession{}
	for _, path := range found {
		body, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var s ClaudeSession
		if json.Unmarshal(body, &s) == nil && s.SessionID != "" {
			out = append(out, s)
		}
	}
	return out
}

type transcriptLine struct {
	Type      string `json:"type"`
	GitBranch string `json:"gitBranch"`
	Message   struct {
		Model string `json:"model"`
		Usage struct {
			Input    int `json:"input_tokens"`
			Read     int `json:"cache_read_input_tokens"`
			Creation int `json:"cache_creation_input_tokens"`
		} `json:"usage"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type toolUse struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Input struct {
		FilePath     string `json:"file_path"`
		NotebookPath string `json:"notebook_path"`
		Command      string `json:"command"`
		Pattern      string `json:"pattern"`
		Description  string `json:"description"`
	} `json:"input"`
}

// tail is how much of a transcript's end is read: its last answer is there, and a transcript runs to megabytes.
const tail = 1 << 20

// ClaudeTranscript reads what the transcript of a session last said, found under the projects folder by the session's id.
func ClaudeTranscript(folder, id string) ClaudeSaid {
	found, _ := filepath.Glob(filepath.Join(folder, "projects", "*", id+".jsonl"))
	if len(found) == 0 {
		return ClaudeSaid{}
	}
	f, err := os.Open(found[0])
	if err != nil {
		return ClaudeSaid{}
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return ClaudeSaid{}
	}
	from := max(info.Size()-tail, 0)
	end := make([]byte, info.Size()-from)
	if _, err := f.ReadAt(end, from); err != nil {
		return ClaudeSaid{}
	}
	said := ClaudeSaid{WrittenAt: info.ModTime().UnixMilli()}
	for _, line := range bytes.Split(end, []byte("\n")) {
		var l transcriptLine
		if json.Unmarshal(line, &l) != nil || l.Type != "assistant" {
			continue
		}
		said.Branch, said.Model = l.GitBranch, l.Message.Model
		said.Tokens = l.Message.Usage.Input + l.Message.Usage.Read + l.Message.Usage.Creation
		var uses []toolUse
		if json.Unmarshal(l.Message.Content, &uses) != nil {
			continue
		}
		for _, u := range uses {
			if u.Type == "tool_use" {
				said.Tool, said.On, said.Description = u.Name, firstOf(u.Input.FilePath, u.Input.NotebookPath, u.Input.Command, u.Input.Pattern), u.Input.Description
			}
		}
	}
	return said
}

func firstOf(said ...string) string {
	for _, one := range said {
		if one != "" {
			line, _, _ := strings.Cut(one, "\n")
			return line
		}
	}
	return ""
}

// SessionHeld is the task a session's open entry in the notes names, 0 when none.
func SessionHeld(notes, id string) int {
	n, _ := strconv.Atoi(filepath.Base(filepath.Dir(SessionEntry(notes, id))))
	return n
}

// OtherHolds says whether a session other than the one named holds the task, by an open entry in its notes.
func OtherHolds(notes string, n int, id string) bool {
	found, _ := filepath.Glob(filepath.Join(notes, strconv.Itoa(n), "*-session-*.md"))
	for _, path := range found {
		if !strings.HasSuffix(path, "-session-"+id+".md") && openEntry(path) {
			return true
		}
	}
	return false
}

// SessionEntry is the path of a session's open entry in the notes, nothing when none: the hook writes <day>-session-<id>.md in the task's folder, and "released" in it when the link closes.
func SessionEntry(notes, id string) string {
	found, _ := filepath.Glob(filepath.Join(notes, "*", "*-session-"+id+".md"))
	for _, path := range found {
		if !openEntry(path) {
			continue
		}
		if _, err := strconv.Atoi(filepath.Base(filepath.Dir(path))); err == nil {
			return path
		}
	}
	return ""
}

func openEntry(path string) bool {
	body, err := os.ReadFile(path)
	return err == nil && !strings.Contains(string(body), "\nreleased ") && !strings.HasPrefix(string(body), "released ")
}

// TieSession ties a session to a task as the hook does: its open entry is released at the time said, and an entry saying how to resume it is written in the task's notes, named by the day.
func TieSession(notes, id, folder string, n int, day, released string) error {
	if open := SessionEntry(notes, id); open != "" {
		f, err := os.OpenFile(open, os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(f, "released %s\n", released)
		if closed := f.Close(); err == nil {
			err = closed
		}
		if err != nil {
			return err
		}
	}
	dir := filepath.Join(notes, strconv.Itoa(n))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	entry := fmt.Sprintf("cd %s\nclaude --resume %s\n", folder, id)
	return os.WriteFile(filepath.Join(dir, day+"-session-"+id+".md"), []byte(entry), 0o600)
}
