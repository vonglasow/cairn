package engine

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"cairn/internal/engine/agent"
	"cairn/internal/engine/task"
	"cairn/internal/store/files"
)

// AsksVariable names the folder the permissions waiting on the window are kept in, cairn-asks in XDG_RUNTIME_DIR, or .cache/cairn/asks in the home, when it is not set; none with no home.
const AsksVariable = "CAIRN_ASKS"

const (
	listenFor = 10 * time.Second
	polled    = 250 * time.Millisecond
)

type clock struct {
	now   func() time.Time
	sleep func(time.Duration)
}

var ticking = clock{now: time.Now, sleep: time.Sleep}

var sessionArgument = Argument{Name: "session", Kind: Text, Positional: true, Required: true, Means: "the session, by its id or its first 6 characters"}

// AgentGo brings the terminal a running session runs in to the front.
var AgentGo = Operation{
	Thing:   "agent",
	Verb:    "go",
	Short:   "Bring the terminal of a Claude session to the front",
	Does:    "Bring the terminal a running Claude session runs in to the front: iTerm2, Terminal and Ghostty on macOS, kitty on Linux. Any other terminal, or an editor, is refused with the command that resumes the session in a terminal of one's own.",
	Example: "cairn agent go a3f9c1",
	Level:   write,
	Needs:   noStore,
	Again:   ChangesNothingMore,
	Answers: agent.Agent{},
	Takes:   []Argument{sessionArgument},
	Run: func(on workedOn, in Input) (Answer, error) {
		a, err := runningNamed(on.Home, in.Given["session"])
		if err != nil {
			return Answer{}, err
		}
		if !a.Terminal.Reachable {
			return Answer{}, Refusing("out-of-reach", fmt.Sprintf("cairn cannot bring %s to the front: %s resumes the session", cmp.Or(a.Terminal.Program, "its terminal"), a.Resume))
		}
		if err := machine.Reach(a.Terminal); err != nil {
			return Answer{}, err
		}
		return Answer{Agent: &a}, nil
	},
}

// AgentStart starts claude in a new tab or window of a terminal, in a folder, asked to work on a task when one is named.
var AgentStart = Operation{
	Thing:   "agent",
	Verb:    "start",
	Short:   "Start a Claude session in a terminal",
	Does:    "Start claude in a new tab or window of a terminal installed, in a folder. With a task, claude is asked to work on it, and the cairn skill ties the session to it.",
	Example: "cairn agent start --folder ~/code/cairn --terminal iTerm2 --task 12",
	Level:   write,
	Needs:   aStore,
	Again:   doesItAgain,
	Answers: agent.Started{},
	Takes: []Argument{
		{Name: "folder", Kind: Text, Required: true, Means: "the folder the session works in"},
		{Name: "terminal", Kind: Text, Required: true, OneOf: []string{"iTerm2", "Terminal", "Ghostty", "kitty", "gnome-terminal"}, Means: "the terminal it runs in"},
		{Name: "task", Kind: Number, Of: "a task", Means: "the task it is asked to work on"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		program := in.Given["terminal"]
		if installed := machine.Terminals(); !slices.Contains(installed, program) {
			return Answer{}, Refusing("terminal-not-installed", fmt.Sprintf("%s is not installed; installed: %s", program, strings.Join(installed, ", ")))
		}
		folder, err := filepath.Abs(in.Given["folder"])
		if err != nil {
			return Answer{}, err
		}
		if info, err := os.Stat(folder); err != nil || !info.IsDir() {
			return Answer{}, Refusing("not-a-folder", folder+" is not a folder")
		}
		n, named := in.whole("task")
		if named {
			if _, err := on.Engine.show(n); err != nil {
				return Answer{}, taskNamed(n, err)
			}
		}
		started := agent.Started{Terminal: program, Folder: folder, Command: agent.CommandFor(n)}
		if err := machine.Start(program, folder, started.Command); err != nil {
			return Answer{}, err
		}
		return Answer{Started: &started}, nil
	},
}

// AgentTerminals names the terminals installed that a session can be started in.
var AgentTerminals = Operation{
	Thing:   "agent",
	Verb:    "terminals",
	Short:   "Name the terminals a Claude session can be started in",
	Does:    "Name the terminals installed on this machine that agent start opens a session in: iTerm2, Terminal and Ghostty on macOS, Ghostty, kitty and gnome-terminal on Linux.",
	Example: "cairn agent terminals",
	Level:   Read,
	Needs:   noStore,
	Answers: agent.Terminals{},
	Run: func(workedOn, Input) (Answer, error) {
		return Answer{Terminals: &agent.Terminals{Terminals: machine.Terminals()}}, nil
	},
}

// AgentTie ties a running session to a task, or to a task made for it, and puts that task in focus.
var AgentTie = Operation{
	Thing: "agent",
	Verb:  "tie",
	Short: "Tie a Claude session to a task",
	Does:  "Tie a running Claude session to a task, as its hook does when it focuses one: the entry of the task it held is released, an entry saying how to resume it is written in the task's notes, and the task goes in focus. With a title, the task is made first.",
	Example: `cairn agent tie a3f9c1 --task 12
cairn agent tie a3f9c1 --title "Look into the pod restarts"`,
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: Task{},
	Takes: []Argument{
		sessionArgument,
		{Name: "task", Kind: Number, Of: "a task", Means: "the task to tie it to"},
		{Name: "title", Kind: Text, Means: "the title of a task made to tie it to"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		a, err := runningNamed(on.Home, in.Given["session"])
		if err != nil {
			return Answer{}, err
		}
		n, named := in.whole("task")
		title, titled := in.Given["title"]
		if named == titled {
			return Answer{}, Refusing("tie-to-what", "agent tie takes --task or --title, one of them")
		}
		if titled {
			added, err := on.Engine.add(title, Changes{})
			if err != nil {
				return Answer{}, task.WithTypedTitle(err, title)
			}
			n = added.Carried().Number
		}
		t, err := on.Engine.putInFocus(n)
		if err != nil {
			return Answer{}, taskNamed(n, err)
		}
		now := ticking.now()
		was := files.SessionHeld(on.Notes, a.ID)
		if err := files.TieSession(on.Notes, a.ID, a.Folder, n, now.Local().Format(time.DateOnly), now.UTC().Format(time.RFC3339)); err != nil {
			return Answer{}, err
		}
		if was != 0 && was != n && !files.OtherHolds(on.Notes, was, a.ID) {
			if err := on.Engine.letGoOf(was); err != nil {
				return Answer{}, err
			}
		}
		p, err := on.Engine.taskProgress(n)
		return Answer{Task: &t, Progress: p, Context: on.contextOf(n)}, err
	},
}

func taskNamed(n int, err error) error {
	if errors.Is(err, task.ErrNoSuchTask) {
		return fmt.Errorf("task %d: %w", n, err)
	}
	return err
}

// AgentAsk is the hook of a permission Claude Code asks: it waits for the window to answer while one listens, and answers nothing when none does, so the terminal asks.
var AgentAsk = Operation{
	Thing:   "agent",
	Verb:    "ask",
	Short:   "Ask the window for a permission, or the answers to questions, a Claude session asks",
	Does:    "Run by the PermissionRequest hook of the Claude Code plugin, and by its PreToolUse hook when Claude asks questions, with the request on standard input. While agents.permissions is on and a window listens, the request waits there for an answer — allow, deny, or allow for good in the project's or the user's settings, or an option chosen for each question — and the answer is printed for Claude Code. Otherwise, or once the window stops listening, nothing is printed and the terminal asks.",
	Example: "cairn agent ask --request - < request.json",
	Level:   Read,
	Needs:   noStore,
	Answers: agent.Decision{},
	Takes:   []Argument{{Name: "request", Kind: Text, Required: true, Means: "the PermissionRequest, or the PreToolUse of AskUserQuestion, Claude Code hands its hook, as JSON"}},
	Run: func(on workedOn, in Input) (Answer, error) {
		nothing := Answer{Decision: &agent.Decision{}}
		read, err := configuredAt(on.Home, on.Path)
		if err != nil || !read.Agents().Permissions {
			return nothing, nil
		}
		var r agent.Request
		if err := json.Unmarshal([]byte(in.Given["request"]), &r); err != nil {
			return Answer{}, Misread(fmt.Errorf("the request is not JSON: %w", err))
		}
		asks := asksFolder(on.Home)
		listen := listeningAt(asks, on.Path)
		if asks == "" || !inAFolder(r.Session) || !listening(listen) {
			return nothing, nil
		}
		pending, err := json.Marshal(agent.PendingOf(r))
		if err != nil {
			return Answer{}, err
		}
		asked, answered := askAt(asks, r.Session), answerAt(asks, r.Session)
		_ = os.Remove(answered)
		if err := os.WriteFile(asked, pending, 0o600); err != nil {
			return Answer{}, err
		}
		defer func() { _, _ = os.Remove(asked), os.Remove(answered) }()
		for listening(listen) {
			if said, err := os.ReadFile(answered); err == nil {
				answer := strings.TrimSpace(string(said))
				hook, err := agent.HookSaying(answer, r)
				if err != nil {
					return Answer{}, err
				}
				return Answer{Decision: &agent.Decision{Session: r.Session, Answer: answer, Hook: hook}}, nil
			}
			ticking.sleep(polled)
		}
		return nothing, nil
	},
}

// AgentAnswer answers the permission, or the questions, a session asks the window.
var AgentAnswer = Operation{
	Thing:   "agent",
	Verb:    "answer",
	Short:   "Answer the permission, or the questions, a Claude session asks",
	Does:    "Answer the permission a Claude session asks the window: allow, deny, always-project to allow it for good in the project's settings, or always-user in the user's. Answer the questions it asks with --choice, once for each question in their order: the label of the option chosen, the labels joined by commas where a question takes several.",
	Example: "cairn agent answer a3f9c1d2-… allow\ncairn agent answer a3f9c1d2-… --choice Postgres --choice \"Tests, Docs\"",
	Level:   write,
	Needs:   noStore,
	Again:   ChangesNothingMore,
	Answers: agent.Decision{},
	Takes: []Argument{
		{Name: "session", Kind: Text, Positional: true, Required: true, Means: "the session asking, by its id"},
		{Name: "answer", Kind: Text, Positional: true, OneOf: agent.Answers, Means: "allow, deny, always-project or always-user, for a permission"},
		{Name: "choice", Kind: List, Of: "a label", Means: "for questions: the option chosen for each, in their order, its labels joined by commas where it takes several"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		id, answer := in.Given["session"], in.Given["answer"]
		asks := asksFolder(on.Home)
		pending, ok := pendingOf(asks, id)
		if asks == "" || !ok {
			return Answer{}, Refusing("asks-nothing", id+" asks nothing of the window")
		}
		if len(pending.Questions) > 0 {
			said, err := agent.Answered(pending.Questions, in.Listed["choice"])
			if err != nil {
				return Answer{}, Misread(err)
			}
			answer = said
		} else if !slices.Contains(agent.Answers, answer) {
			return Answer{}, Misread(fmt.Errorf("%q is not an answer: %s", answer, strings.Join(agent.Answers, ", ")))
		}
		if err := os.WriteFile(answerAt(asks, id), []byte(answer), 0o600); err != nil {
			return Answer{}, err
		}
		return Answer{Decision: &agent.Decision{Session: id, Answer: answer}}, nil
	},
}

func asksFolder(h Home) string {
	if said, ok := h.lookup()(AsksVariable); ok && said != "" {
		return said
	}
	if run, ok := h.lookup()("XDG_RUNTIME_DIR"); ok && filepath.IsAbs(run) {
		return filepath.Join(run, "cairn-asks")
	}
	if h.Dir == "" {
		return ""
	}
	return filepath.Join(h.Dir, ".cache", "cairn", "asks")
}

func listeningAt(asks, store string) string {
	sum := sha256.Sum256([]byte(store))
	return filepath.Join(asks, "listening-"+hex.EncodeToString(sum[:8]))
}

func listening(at string) bool {
	info, err := os.Stat(at)
	return err == nil && ticking.now().Sub(info.ModTime()) < listenFor
}

func listened(at string) error {
	if err := os.MkdirAll(filepath.Dir(at), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(at, nil, 0o600); err != nil {
		return err
	}
	now := ticking.now()
	return os.Chtimes(at, now, now)
}

const listeningKept = time.Hour

func swept(asks, claude string) error {
	if asks == "" {
		return nil
	}
	running := map[string]bool{}
	for _, s := range files.ClaudeSessions(claude) {
		running[s.SessionID] = machine.Alive(s.PID, s.ProcStart)
	}
	found, err := os.ReadDir(asks)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, f := range found {
		name := f.Name()
		id := strings.TrimSuffix(strings.TrimSuffix(name, ".json"), ".answer")
		stale := id != name && !running[id]
		if strings.HasPrefix(name, "listening-") {
			info, err := f.Info()
			stale = err == nil && ticking.now().Sub(info.ModTime()) > listeningKept
		}
		if stale {
			if err := os.Remove(filepath.Join(asks, name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}
	}
	return nil
}

func inAFolder(id string) bool { return id != "" && id != "." && id != ".." && filepath.Base(id) == id }

func askAt(asks, id string) string { return filepath.Join(asks, id+".json") }

func answerAt(asks, id string) string { return filepath.Join(asks, id+".answer") }

func pendingOf(asks, id string) (agent.Pending, bool) {
	if !inAFolder(id) {
		return agent.Pending{}, false
	}
	body, err := os.ReadFile(askAt(asks, id))
	var p agent.Pending
	return p, err == nil && json.Unmarshal(body, &p) == nil
}

func configuredAt(h Home, store string) (files.Places, error) {
	read, err := files.Read(h.Dir, h.lookup())
	if err != nil {
		return files.Places{}, err
	}
	return read.WithSettings(files.SettingsBeside(store))
}
