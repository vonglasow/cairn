package engine

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/engine/agent"
	"cairn/internal/engine/item"
	"cairn/internal/store/files"
)

type machineSaying map[int]agent.Terminal

func (m machineSaying) Alive(pid int, _ string) bool { _, ok := m[pid]; return ok }

func (m machineSaying) Terminal(pid int) agent.Terminal { return m[pid] }

func (m machineSaying) Reach(agent.Terminal) error { return nil }

func (m machineSaying) Start(string, string, string) error { return nil }

func (m machineSaying) Terminals() []string { return nil }

type machineActing struct {
	machineSaying
	installed []string
	reached   []agent.Terminal
	started   []string
}

func (m *machineActing) Reach(t agent.Terminal) error {
	m.reached = append(m.reached, t)
	return nil
}

func (m *machineActing) Start(program, folder, command string) error {
	m.started = append(m.started, program+" | "+folder+" | "+command)
	return nil
}

func (m *machineActing) Terminals() []string { return m.installed }

func claudeAt(t *testing.T, sessions map[string]string, transcripts map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range sessions {
		writeUnder(t, filepath.Join(dir, "sessions", name), body)
	}
	for name, body := range transcripts {
		writeUnder(t, filepath.Join(dir, "projects", "p", name), body)
	}
	return dir
}

func session(pid int, id, cwd, more string) string {
	return fmt.Sprintf(`{"pid":%d,"sessionId":%q,"cwd":%q,"status":"busy","procStart":"x"%s}`, pid, id, cwd, more)
}

func writeUnder(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func agentsListed(t *testing.T, store, notes, claude string, m agent.Machine) agent.Agents {
	t.Helper()
	was := machine
	machine = m
	t.Cleanup(func() { machine = was })
	answer, err := doThere(AgentList, Places{Store: store, Notes: notes, Home: agentHome(claude, "", "")}, Input{}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	return *answer.Agents
}

func agentHome(claude, asks, dir string) Home {
	return Home{Dir: dir, Lookup: func(k string) (string, bool) {
		switch k {
		case ClaudeVariable:
			return claude, true
		case AsksVariable:
			return asks, asks != ""
		}
		return "", false
	}}
}

func TestAgentList(t *testing.T) {
	const (
		busy    = `{"pid":11,"sessionId":"aaaaaaaa-1","cwd":"/work/cairn","name":"cairn-a0","status":"busy","statusUpdatedAt":1795542308925,"procStart":"x"}`
		waiting = `{"pid":12,"sessionId":"bbbbbbbb-2","cwd":"/work/ops","name":"ops","status":"busy","waitingFor":"permission","statusUpdatedAt":1795542308925,"procStart":"x"}`
		idle    = `{"pid":13,"sessionId":"cccccccc-3","cwd":"/work/cairn","name":"idle","status":"idle","statusUpdatedAt":1795542308925,"procStart":"x"}`
		dead    = `{"pid":14,"sessionId":"dddddddd-4","cwd":"/work/cairn","status":"busy","procStart":"x"}`
		shell   = `{"pid":15,"sessionId":"eeeeeeee-5","cwd":"/work/cairn","status":"shell","procStart":"x"}`
		ran     = `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"cd /work && go test ./...","description":"Run the engine tests"}}]}}`
		said    = `{"type":"assistant","gitBranch":"feat/agents","message":{"model":"claude-opus-5-5","usage":{"input_tokens":2,"cache_read_input_tokens":99998,"cache_creation_input_tokens":0},"content":[{"type":"tool_use","name":"Edit","input":{"file_path":"/work/cairn/card.go"}}]}}`
	)
	alive := machineSaying{11: {Program: "iTerm2", TTY: "ttys004"}, 12: {}, 13: {}, 15: {}}
	store := storeAt(t, "Add the agents page")
	notes := t.TempDir()
	writeUnder(t, filepath.Join(notes, "1", "2026-11-25-session-aaaaaaaa-1.md"), "cd /work/cairn\nclaude --resume aaaaaaaa-1\n")
	writeUnder(t, filepath.Join(notes, "1", "2026-11-24-session-cccccccc-3.md"), "cd /work/cairn\nclaude --resume cccccccc-3\nreleased 2026-11-24T10:00:00Z\n")
	claude := claudeAt(t,
		map[string]string{"11.json": busy, "12.json": waiting, "13.json": idle, "14.json": dead, "15.json": shell},
		map[string]string{"aaaaaaaa-1.jsonl": "{}\n" + said + "\n", "bbbbbbbb-2.jsonl": ran + "\n"})

	got := agentsListed(t, store, notes, claude, alive)
	byID := map[string]agent.Agent{}
	for _, a := range got.Agents {
		byID[a.ID] = a
	}

	tests := []struct {
		name  string
		id    string
		check func(agent.Agent) bool
	}{
		{"a busy session is working", "aaaaaaaa-1", func(a agent.Agent) bool { return a.State == agent.Working }},
		{"its project is its folder's name", "aaaaaaaa-1", func(a agent.Agent) bool { return a.Project == "cairn" }},
		{"its short id is its first 6 characters", "aaaaaaaa-1", func(a agent.Agent) bool { return a.Short == "aaaaaa" }},
		{"a session whose entry is open holds its task", "aaaaaaaa-1", func(a agent.Agent) bool {
			return a.Task != nil && a.Task.Number == 1 && a.Task.Title == "Add the agents page"
		}},
		{"its transcript gives the branch, the model and what it does", "aaaaaaaa-1", func(a agent.Agent) bool {
			return a.Branch == "feat/agents" && a.Model == "claude-opus-5-5" && a.Doing == "Edits card.go"
		}},
		{"what it does is said by the description Claude gave its command", "bbbbbbbb-2", func(a agent.Agent) bool { return a.Doing == "Run the engine tests" }},
		{"its context is the tokens of its last answer, against the window", "aaaaaaaa-1", func(a agent.Agent) bool {
			return a.Tokens == 100000 && a.Context == 50
		}},
		{"its terminal is the machine's", "aaaaaaaa-1", func(a agent.Agent) bool { return a.Terminal.Program == "iTerm2" && a.Terminal.TTY == "ttys004" }},
		{"a session waiting says for what", "bbbbbbbb-2", func(a agent.Agent) bool { return a.State == agent.Waiting && a.Waiting == "permission" }},
		{"an idle session is idle", "cccccccc-3", func(a agent.Agent) bool { return a.State == agent.Idle }},
		{"a released entry holds no task", "cccccccc-3", func(a agent.Agent) bool { return a.Task == nil }},
		{"a session running a shell command is working", "eeeeeeee-5", func(a agent.Agent) bool { return a.State == agent.Working }},
		{"it says the command that resumes it", "bbbbbbbb-2", func(a agent.Agent) bool { return a.Resume == "cd '/work/ops' && claude --resume bbbbbbbb-2" }},
		{"its terminal carries its process", "aaaaaaaa-1", func(a agent.Agent) bool { return a.Terminal.PID == 11 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, ok := byID[tt.id]
			if !ok {
				t.Fatalf("%s is not listed: %+v", tt.id, got.Agents)
			}
			if !tt.check(a) {
				t.Fatalf("%+v", a)
			}
		})
	}
	t.Run("a session whose process is dead is not listed", func(t *testing.T) {
		if _, ok := byID["dddddddd-4"]; ok {
			t.Fatal("listed")
		}
	})
	t.Run("the count waiting is what the tab shows", func(t *testing.T) {
		if got.Waiting != 1 {
			t.Fatalf("waiting %d", got.Waiting)
		}
	})
}

func TestAgentList_leavesOutASessionHoldingNoTaskWhoseFolderAnswersFromAnotherStore(t *testing.T) {
	was := machine
	machine = machineSaying{11: {}, 12: {}}
	t.Cleanup(func() { machine = was })
	store, notes := storeAt(t, "Add the agents page"), t.TempDir()
	writeUnder(t, filepath.Join(notes, "1", "2026-11-25-session-aaaaaaaa-1.md"), "cd /elsewhere\nclaude --resume aaaaaaaa-1\n")
	claude := claudeAt(t, map[string]string{"11.json": session(11, "aaaaaaaa-1", "/elsewhere", ""), "12.json": session(12, "bbbbbbbb-2", "/elsewhere", "")}, nil)
	answer, err := doThere(AgentList, Places{Store: store, Notes: notes, Home: agentHome(claude, "", t.TempDir())}, Input{}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	if got := answer.Agents.Agents; len(got) != 1 || got[0].ID != "aaaaaaaa-1" {
		t.Fatalf("listed %+v", got)
	}
}

func TestAgentGo(t *testing.T) {
	claude := claudeAt(t, map[string]string{
		"11.json": session(11, "aaaaaaaa-1", "/work/cairn", ""),
		"12.json": session(12, "bbbbbbbb-2", "/work/my ops", ""),
		"13.json": session(13, "cccccccc-3", "/work/cairn", `,"entrypoint":"claude-vscode"`),
	}, nil)
	tests := []struct {
		name    string
		session string
		reached []agent.Terminal
		refused string
	}{
		{"a session in a terminal it reaches is brought to the front", "aaaaaaaa-1", []agent.Terminal{{Program: "iTerm2", TTY: "ttys004", PID: 11, Reachable: true}}, ""},
		{"a session is named by its short id too", "aaaaaa", []agent.Terminal{{Program: "iTerm2", TTY: "ttys004", PID: 11, Reachable: true}}, ""},
		{"a terminal it cannot reach says the command to copy", "bbbbbbbb-2", nil, "cd '/work/my ops' && claude --resume bbbbbbbb-2"},
		{"a session started from an editor is not reached", "cccccccc-3", nil, "VS Code"},
		{"a session not running is refused", "dddddddd-4", nil, "no Claude session"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &machineActing{machineSaying: machineSaying{11: {Program: "iTerm2", TTY: "ttys004", Reachable: true}, 12: {Program: "GNOME Terminal"}, 13: {Program: "iTerm2", Reachable: true}}}
			was := machine
			machine = m
			t.Cleanup(func() { machine = was })
			answer, err := doThere(AgentGo, Places{Store: storeAt(t), Home: agentHome(claude, "", "")}, Input{Given: map[string]string{"session": tt.session}}, Acting{})
			if tt.refused != "" {
				if err == nil || !strings.Contains(err.Error(), tt.refused) || RefusalOf(err).Kind != KindRefused {
					t.Fatalf("err %v", err)
				}
			} else if err != nil || answer.Agent == nil || answer.Agent.ID != "aaaaaaaa-1" {
				t.Fatalf("answer %+v, err %v", answer.Agent, err)
			}
			if fmt.Sprint(m.reached) != fmt.Sprint(tt.reached) {
				t.Errorf("reached %+v", m.reached)
			}
		})
	}
}

func TestAgentStart(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "it's here")
	if err := os.MkdirAll(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		given   map[string]string
		started string
		refused string
	}{
		{"claude starts in the terminal in the folder", map[string]string{"folder": folder, "terminal": "iTerm2"}, "iTerm2 | " + folder + " | claude", ""},
		{"with a task it is asked to work on it", map[string]string{"folder": folder, "terminal": "kitty", "task": "1"}, "kitty | " + folder + ` | claude 'let'\''s work on cairn#1'`, ""},
		{"claude starts in Ghostty", map[string]string{"folder": folder, "terminal": "Ghostty"}, "Ghostty | " + folder + " | claude", ""},
		{"a terminal not installed is refused", map[string]string{"folder": folder, "terminal": "Terminal"}, "", "not installed"},
		{"a folder that is not there is refused", map[string]string{"folder": filepath.Join(folder, "gone"), "terminal": "iTerm2"}, "", "gone"},
		{"a task that is not there is refused", map[string]string{"folder": folder, "terminal": "iTerm2", "task": "9"}, "", "9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &machineActing{installed: []string{"iTerm2", "kitty", "Ghostty"}}
			was := machine
			machine = m
			t.Cleanup(func() { machine = was })
			answer, err := doThere(AgentStart, Places{Store: storeAt(t, "Add the agents page")}, Input{Given: tt.given}, Acting{})
			if tt.refused != "" {
				if err == nil || !strings.Contains(err.Error(), tt.refused) || len(m.started) != 0 {
					t.Fatalf("err %v, started %v", err, m.started)
				}
				return
			}
			if err != nil || len(m.started) != 1 || m.started[0] != tt.started || answer.Started == nil {
				t.Fatalf("started %v, err %v", m.started, err)
			}
		})
	}
}

func TestAgentTerminals_answersTheTerminalsInstalled(t *testing.T) {
	was := machine
	machine = &machineActing{installed: []string{"iTerm2", "Terminal"}}
	t.Cleanup(func() { machine = was })
	answer, err := doThere(AgentTerminals, Places{}, Input{}, Acting{})
	if err != nil || fmt.Sprint(answer.Terminals.Terminals) != "[iTerm2 Terminal]" {
		t.Fatalf("%+v %v", answer.Terminals, err)
	}
}

func clockAt(t *testing.T, at time.Time, sleep func(time.Duration)) {
	t.Helper()
	was := ticking
	ticking = clock{now: func() time.Time { return at }, sleep: sleep}
	t.Cleanup(func() { ticking = was })
}

func TestAgentTie(t *testing.T) {
	at := time.Date(2026, 11, 25, 9, 30, 0, 0, time.UTC)
	tests := []struct {
		name    string
		given   map[string]string
		tied    int
		refused string
		shared  bool
		focus   string
	}{
		{"a session is tied to a task, and the task it held leaves the focus", map[string]string{"session": "aaaaaaaa-1", "task": "2"}, 2, "", false, "[2]"},
		{"the task it held stays in focus while another session holds it", map[string]string{"session": "aaaaaaaa-1", "task": "2"}, 2, "", true, "[1 2]"},
		{"a session is tied to a task made for it", map[string]string{"session": "aaaaaaaa-1", "title": "Look into the pods"}, 3, "", false, "[3]"},
		{"a session not running is refused", map[string]string{"session": "dddddddd-4", "task": "2"}, 0, "no Claude session", false, ""},
		{"a task or a title is named", map[string]string{"session": "aaaaaaaa-1"}, 0, "--task or --title", false, ""},
		{"a task that is not there is refused", map[string]string{"session": "aaaaaaaa-1", "task": "9"}, 0, "9", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			was := machine
			machine = machineSaying{11: {}}
			t.Cleanup(func() { machine = was })
			clockAt(t, at, nil)
			store, notes := storeAt(t, "Add the agents page", "Fix the overrun"), t.TempDir()
			old := filepath.Join(notes, "1", "2026-11-24-session-aaaaaaaa-1.md")
			writeUnder(t, old, "cd /work/cairn\nclaude --resume aaaaaaaa-1\n")
			if tt.shared {
				writeUnder(t, filepath.Join(notes, "1", "2026-11-24-session-bbbbbbbb-2.md"), "cd /work/cairn\nclaude --resume bbbbbbbb-2\n")
			}
			if _, err := doThere(TaskFocus, Places{Store: store}, Input{Given: map[string]string{"number": "1"}}, Acting{}); err != nil {
				t.Fatal(err)
			}
			claude := claudeAt(t, map[string]string{"11.json": session(11, "aaaaaaaa-1", "/work/cairn", "")}, nil)
			answer, err := doThere(AgentTie, Places{Store: store, Notes: notes, Home: agentHome(claude, "", "")}, Input{Given: tt.given}, Acting{})
			if tt.refused != "" {
				if err == nil || !strings.Contains(err.Error(), tt.refused) {
					t.Fatalf("err %v", err)
				}
				return
			}
			if err != nil || answer.Task == nil || answer.Task.Carried().Number != tt.tied || answer.Task.Carried().State != item.Current {
				t.Fatalf("answer %v, err %v", answer.Task, err)
			}
			if body, _ := os.ReadFile(old); string(body) != "cd /work/cairn\nclaude --resume aaaaaaaa-1\nreleased 2026-11-25T09:30:00Z\n" {
				t.Errorf("the old entry reads %q", body)
			}
			entry := filepath.Join(notes, fmt.Sprint(tt.tied), at.Local().Format(time.DateOnly)+"-session-aaaaaaaa-1.md")
			if body, _ := os.ReadFile(entry); string(body) != "cd /work/cairn\nclaude --resume aaaaaaaa-1\n" {
				t.Errorf("the new entry reads %q", body)
			}
			focus, err := doThere(TaskFocus, Places{Store: store}, Input{}, Acting{})
			if err != nil || fmt.Sprint(*focus.InFocus) != tt.focus {
				t.Errorf("in focus %v, err %v", focus.InFocus, err)
			}
		})
	}
}

const permissionAsked = `{"session_id":"aaaaaaaa-1","hook_event_name":"PermissionRequest","tool_name":"Bash","tool_input":{"command":"rm -rf node_modules","description":"clean"},"permission_suggestions":[{"type":"addRules","rules":[{"toolName":"Bash","ruleContent":"rm -rf node_modules"}],"behavior":"allow","destination":"localSettings"}]}`

const questionAsked = `{"session_id":"aaaaaaaa-1","hook_event_name":"PreToolUse","tool_name":"AskUserQuestion","tool_input":{"questions":[{"question":"Which store?","header":"Store","options":[{"label":"Postgres","description":"the server"},{"label":"SQLite","description":"a file"}],"multiSelect":false},{"question":"What to add?","header":"Add","options":[{"label":"Tests"},{"label":"Docs"}],"multiSelect":true}]}}`

func permissionsOn(t *testing.T, on bool) string {
	t.Helper()
	home := t.TempDir()
	writeUnder(t, filepath.Join(home, ".config", "cairn", "config.toml"), fmt.Sprintf("[agents]\npermissions = %t\n", on))
	return home
}

func TestAgentAsk(t *testing.T) {
	at := time.Date(2026, 11, 25, 9, 30, 0, 0, time.UTC)
	tests := []struct {
		name      string
		on        bool
		listening time.Duration
		answer    string
		stops     bool
		want      string
		request   string
	}{
		{"with permissions off nothing is answered", false, time.Second, "allow", false, "", ""},
		{"with no window listening nothing is answered", true, -1, "allow", false, "", ""},
		{"with a window listened to long ago nothing is answered", true, 11 * time.Second, "allow", false, "", ""},
		{"an allow answers allow", true, time.Second, "allow", false, `{"hookSpecificOutput":{"decision":{"behavior":"allow"},"hookEventName":"PermissionRequest"}}`, ""},
		{"a deny answers deny", true, time.Second, "deny", false, `{"hookSpecificOutput":{"decision":{"behavior":"deny","message":"Denied from the cairn window"},"hookEventName":"PermissionRequest"}}`, ""},
		{"always for the project writes the suggestions to its settings", true, time.Second, "always-project", false,
			`{"hookSpecificOutput":{"decision":{"behavior":"allow","updatedPermissions":[{"behavior":"allow","destination":"projectSettings","rules":[{"ruleContent":"rm -rf node_modules","toolName":"Bash"}],"type":"addRules"}]},"hookEventName":"PermissionRequest"}}`, ""},
		{"always for the user writes them to the user's", true, time.Second, "always-user", false,
			`{"hookSpecificOutput":{"decision":{"behavior":"allow","updatedPermissions":[{"behavior":"allow","destination":"userSettings","rules":[{"ruleContent":"rm -rf node_modules","toolName":"Bash"}],"type":"addRules"}]},"hookEventName":"PermissionRequest"}}`, ""},
		{"a window that stops listening leaves the terminal to ask", true, time.Second, "", true, "", ""},
		{"questions answered are given back with the questions", true, time.Second, `{"What to add?":"Tests, Docs","Which store?":"Postgres"}`, false,
			`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","updatedInput":{"answers":{"What to add?":"Tests, Docs","Which store?":"Postgres"},"questions":[{"header":"Store","multiSelect":false,"options":[{"description":"the server","label":"Postgres"},{"description":"a file","label":"SQLite"}],"question":"Which store?"},{"header":"Add","multiSelect":true,"options":[{"label":"Tests"},{"label":"Docs"}],"question":"What to add?"}]}}}`, questionAsked},
		{"questions with permissions off leave the terminal to ask", false, time.Second, "{}", false, "", questionAsked},
		{"questions a window stops listening for leave the terminal to ask", true, time.Second, "", true, "", questionAsked},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asks, store := t.TempDir(), storeAt(t)
			listen := listeningAt(asks, store)
			if tt.listening >= 0 {
				writeUnder(t, listen, "")
				if err := os.Chtimes(listen, at.Add(-tt.listening), at.Add(-tt.listening)); err != nil {
					t.Fatal(err)
				}
			}
			now := at
			slept := 0
			clockAt(t, now, nil)
			ticking.now = func() time.Time { return now }
			ticking.sleep = func(time.Duration) {
				slept++
				if _, err := os.Stat(filepath.Join(asks, "aaaaaaaa-1.json")); err != nil {
					t.Fatalf("no ask is pending while it waits: %v", err)
				}
				if tt.stops {
					now = now.Add(10 * time.Second)
					return
				}
				writeUnder(t, filepath.Join(asks, "aaaaaaaa-1.answer"), tt.answer)
			}
			request := cmp.Or(tt.request, permissionAsked)
			answer, err := doThere(AgentAsk, Places{Store: store, Home: agentHome("", asks, permissionsOn(t, tt.on))}, Input{Given: map[string]string{"request": request}}, Acting{})
			if err != nil || answer.Decision == nil {
				t.Fatalf("err %v", err)
			}
			if string(answer.Decision.Hook) != tt.want {
				t.Fatalf("hook %s", answer.Decision.Hook)
			}
			for _, left := range []string{"aaaaaaaa-1.json", "aaaaaaaa-1.answer"} {
				if _, err := os.Stat(filepath.Join(asks, left)); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("%s is left: %v", left, err)
				}
			}
			if tt.want == "" && !tt.stops && slept != 0 {
				t.Errorf("waited %d times though nobody answers", slept)
			}
		})
	}
}

const asksBash = `{"tool_name":"Bash","on":"make"}`

var asksQuestions = func() string {
	var r agent.Request
	_ = json.Unmarshal([]byte(questionAsked), &r)
	said, _ := json.Marshal(agent.PendingOf(r))
	return string(said)
}()

func TestAgentAnswer(t *testing.T) {
	tests := []struct {
		name    string
		pending string
		given   map[string]string
		chosen  []string
		refused string
		want    string
	}{
		{"an answer is written for the session asking", asksBash, map[string]string{"session": "aaaaaaaa-1", "answer": "deny"}, nil, "", "deny"},
		{"a session asking nothing is refused", "", map[string]string{"session": "aaaaaaaa-1", "answer": "deny"}, nil, "asks nothing", ""},
		{"an answer it does not take is refused", asksBash, map[string]string{"session": "aaaaaaaa-1", "answer": "maybe"}, nil, "maybe", ""},
		{"a permission is not answered by choices", asksBash, map[string]string{"session": "aaaaaaaa-1"}, []string{"Postgres"}, "allow, deny", ""},
		{"a session named out of the folder is refused", asksBash, map[string]string{"session": "../aaaaaaaa-1", "answer": "deny"}, nil, "asks nothing", ""},
		{"questions are answered a choice each, in their order", asksQuestions, map[string]string{"session": "aaaaaaaa-1"}, []string{"Postgres", "Tests, Docs"}, "",
			`{"What to add?":"Tests, Docs","Which store?":"Postgres"}`},
		{"a choice for each question is wanted", asksQuestions, map[string]string{"session": "aaaaaaaa-1"}, []string{"Postgres"}, "2 questions", ""},
		{"a choice that is not an option is refused", asksQuestions, map[string]string{"session": "aaaaaaaa-1"}, []string{"Mongo", "Tests"}, "Mongo", ""},
		{"one choice only where one is taken", asksQuestions, map[string]string{"session": "aaaaaaaa-1"}, []string{"Postgres, SQLite", "Tests"}, "Postgres, SQLite", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asks := t.TempDir()
			if tt.pending != "" {
				writeUnder(t, filepath.Join(asks, "aaaaaaaa-1.json"), tt.pending)
			}
			_, err := doThere(AgentAnswer, Places{Home: agentHome("", asks, "")}, Input{Given: tt.given, Listed: map[string][]string{"choice": tt.chosen}}, Acting{})
			if tt.refused != "" {
				if err == nil || !strings.Contains(err.Error(), tt.refused) {
					t.Fatalf("err %v", err)
				}
				return
			}
			if body, _ := os.ReadFile(filepath.Join(asks, "aaaaaaaa-1.answer")); err != nil || string(body) != tt.want {
				t.Fatalf("answer %q, err %v", body, err)
			}
		})
	}
}

func TestAgentList_saysThePermissionAskedAndKeepsTheWindowListening(t *testing.T) {
	was := machine
	machine = machineSaying{11: {}, 12: {}}
	t.Cleanup(func() { machine = was })
	at := time.Date(2026, 11, 25, 9, 30, 0, 0, time.UTC)
	clockAt(t, at, nil)
	asks, store, notes := t.TempDir(), storeAt(t, "Add the agents page"), t.TempDir()
	request := agent.Request{}
	if err := json.Unmarshal([]byte(permissionAsked), &request); err != nil {
		t.Fatal(err)
	}
	pending, _ := json.Marshal(agent.PendingOf(request))
	writeUnder(t, filepath.Join(asks, "aaaaaaaa-1.json"), string(pending))
	claude := claudeAt(t, map[string]string{"11.json": session(11, "aaaaaaaa-1", "/work", ""), "12.json": session(12, "bbbbbbbb-2", "/work", "")}, nil)
	answer, err := doThere(AgentList, Places{Store: store, Notes: notes, Home: agentHome(claude, asks, "")}, Input{Given: map[string]string{"listening": "true"}}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range answer.Agents.Agents {
		want := (*agent.Ask)(nil)
		if a.ID == "aaaaaaaa-1" {
			want = &agent.Ask{Tool: "Bash", On: "rm -rf node_modules", Always: true}
		}
		if fmt.Sprint(a.Asks) != fmt.Sprint(want) {
			t.Errorf("%s asks %+v", a.ID, a.Asks)
		}
	}
	info, err := os.Stat(listeningAt(asks, store))
	if err != nil || !info.ModTime().Equal(at) {
		t.Errorf("the window is not said to listen: %v", err)
	}
}

func TestAsksFolder(t *testing.T) {
	for _, tt := range []struct {
		name string
		set  map[string]string
		want string
	}{
		{"named by its variable", map[string]string{AsksVariable: "/asks"}, "/asks"},
		{"in the person's runtime folder, which no other person reads", map[string]string{"XDG_RUNTIME_DIR": "/run/user/1000"}, "/run/user/1000/cairn-asks"},
		{"a runtime folder not absolute is not used: the person's cache is", map[string]string{"XDG_RUNTIME_DIR": "run"}, "/home/ada/.cache/cairn/asks"},
		{"under the person's home, where a test's home keeps it", nil, "/home/ada/.cache/cairn/asks"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			home := Home{Dir: "/home/ada", Lookup: func(name string) (string, bool) { said, ok := tt.set[name]; return said, ok }}
			if got := asksFolder(home); got != tt.want {
				t.Errorf("%q, want %q", got, tt.want)
			}
		})
	}
	t.Run("none with no home, and nothing listens or asks there", func(t *testing.T) {
		if got := asksFolder(Home{}); got != "" {
			t.Errorf("%q", got)
		}
	})
}

func TestAgentStart_offersTheTerminalsItStartsIn(t *testing.T) {
	for _, arg := range AgentStart.Takes {
		if arg.Name == "terminal" && !slices.Equal(arg.OneOf, []string{"iTerm2", "Terminal", "Ghostty", "kitty", "gnome-terminal"}) {
			t.Errorf("offers %v", arg.OneOf)
		}
	}
}

func TestAgentList_suggestsATaskInFocusOrCurrentNoSessionHolds(t *testing.T) {
	tests := []struct {
		name    string
		focus   []string
		held    string
		branch  string
		want    int
		current []string
	}{
		{"the one task in current no session holds, out of focus", nil, "", "main", 3, []string{"3"}},
		{"of a task in focus and one in current, the one the branch names", []string{"1"}, "", "feat/3-pods", 3, []string{"3"}},
		{"the one task in focus no session holds", []string{"2"}, "", "main", 2, nil},
		{"none when the one in focus is held", []string{"2"}, "2", "main", 0, nil},
		{"of several, the one the branch names", []string{"1", "2", "3"}, "", "feat/2-overrun", 2, nil},
		{"of several, the one the branch names that no session holds", []string{"1", "2", "3"}, "1", "fix/3", 3, nil},
		{"none when the branch names none", []string{"1", "2"}, "", "feat/agents", 0, nil},
		{"a number is read whole", []string{"1", "2"}, "", "feat/12-x", 0, nil},
		{"none when nothing is in focus", nil, "", "feat/2", 0, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, notes := storeAt(t, "Add the agents page", "Fix the overrun", "Pods restart"), t.TempDir()
			for _, n := range tt.focus {
				if _, err := doThere(TaskFocus, Places{Store: store}, Input{Given: map[string]string{"number": n}}, Acting{}); err != nil {
					t.Fatal(err)
				}
			}
			for _, n := range tt.current {
				if _, err := doThere(TaskEdit, Places{Store: store}, Input{Given: map[string]string{"number": n, "state": "current"}}, Acting{}); err != nil {
					t.Fatal(err)
				}
			}
			sessions := map[string]string{"11.json": session(11, "aaaaaaaa-1", "/work/cairn", "")}
			if tt.held != "" {
				sessions["12.json"] = session(12, "bbbbbbbb-2", "/work/cairn", "")
				writeUnder(t, filepath.Join(notes, tt.held, "2026-11-25-session-bbbbbbbb-2.md"), "cd /work/cairn\nclaude --resume bbbbbbbb-2\n")
			}
			said := fmt.Sprintf(`{"type":"assistant","gitBranch":%q,"message":{"model":"claude-opus-5-5","content":[]}}`, tt.branch)
			claude := claudeAt(t, sessions, map[string]string{"aaaaaaaa-1.jsonl": said + "\n"})
			got := agentsListed(t, store, notes, claude, machineSaying{11: {}, 12: {}})
			for _, a := range got.Agents {
				if a.Task != nil && a.Suggested != nil {
					t.Errorf("%s holds a task and is suggested #%d", a.ID, a.Suggested.Number)
				}
				if a.ID != "aaaaaaaa-1" {
					continue
				}
				suggested := 0
				if a.Suggested != nil {
					suggested = a.Suggested.Number
				}
				if suggested != tt.want {
					t.Errorf("suggested #%d, want #%d", suggested, tt.want)
				}
			}
		})
	}
}

func TestAgentList_refusesASilenceThatIsNotADuration(t *testing.T) {
	home := t.TempDir()
	writeUnder(t, filepath.Join(home, ".config", "cairn", "config.toml"), "[agents]\nsilent = \"soon\"\n")
	_, err := doThere(AgentList, Places{Store: storeAt(t), Notes: t.TempDir(), Home: agentHome(t.TempDir(), "", home)}, Input{}, Acting{})
	if !errors.Is(err, files.ErrSilent) || !strings.Contains(err.Error(), "agents.silent") {
		t.Errorf("err %v", err)
	}
}

func TestAgentList_saysAWorkingSessionSilentPastTheTimeConfigured(t *testing.T) {
	at := time.Date(2026, 11, 26, 9, 30, 0, 0, time.UTC)
	tests := []struct {
		name    string
		config  string
		status  string
		written time.Duration
		silent  bool
	}{
		{"a working session written to 12m ago is silent, by default past 10m", "", "busy", 12 * time.Minute, true},
		{"one written to 5m ago is not", "", "busy", 5 * time.Minute, false},
		{"the time is the configuration's", "[agents]\nsilent = \"3m\"\n", "busy", 5 * time.Minute, true},
		{"an idle session is never silent", "", "idle", time.Hour, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			was := machine
			machine = machineSaying{11: {}}
			t.Cleanup(func() { machine = was })
			clockAt(t, at, nil)
			store, notes, home := storeAt(t, "Add the agents page"), t.TempDir(), t.TempDir()
			writeUnder(t, filepath.Join(home, ".config", "cairn", "config.toml"), tt.config)
			writeUnder(t, filepath.Join(notes, "1", "2026-11-26-session-aaaaaaaa-1.md"), "cd /work\nclaude --resume aaaaaaaa-1\n")
			claude := claudeAt(t, map[string]string{"11.json": fmt.Sprintf(`{"pid":11,"sessionId":"aaaaaaaa-1","cwd":"/work","status":%q,"procStart":"x"}`, tt.status)},
				map[string]string{"aaaaaaaa-1.jsonl": `{"type":"assistant","message":{"content":[]}}` + "\n"})
			written := at.Add(-tt.written)
			if err := os.Chtimes(filepath.Join(claude, "projects", "p", "aaaaaaaa-1.jsonl"), written, written); err != nil {
				t.Fatal(err)
			}
			answer, err := doThere(AgentList, Places{Store: store, Notes: notes, Home: agentHome(claude, "", home)}, Input{}, Acting{})
			if err != nil || len(answer.Agents.Agents) != 1 {
				t.Fatalf("listed %+v, %v", answer.Agents, err)
			}
			got := answer.Agents.Agents[0].Silent
			if tt.silent != !got.IsZero() || tt.silent && !got.Equal(written) {
				t.Errorf("silent since %v, want %v (%v)", got, written, tt.silent)
			}
		})
	}
}

func TestAgentList_sweepsWhatNoWindowNorSessionNeedsAnyMore(t *testing.T) {
	was := machine
	machine = machineSaying{11: {}}
	t.Cleanup(func() { machine = was })
	at := time.Date(2026, 11, 26, 9, 30, 0, 0, time.UTC)
	clockAt(t, at, nil)
	asks, store, notes := t.TempDir(), storeAt(t, "Add the agents page"), t.TempDir()
	written := func(name string, old time.Duration) string {
		path := filepath.Join(asks, name)
		writeUnder(t, path, "{}")
		if err := os.Chtimes(path, at.Add(-old), at.Add(-old)); err != nil {
			t.Fatal(err)
		}
		return path
	}
	kept := []string{written("listening-fresh", 5*time.Minute), written("aaaaaaaa-1.json", 3*time.Hour), written("aaaaaaaa-1.answer", 3*time.Hour)}
	gone := []string{written("listening-stale", 2*time.Hour), written("dddddddd-4.json", time.Minute), written("dddddddd-4.answer", time.Minute)}
	claude := claudeAt(t, map[string]string{"11.json": session(11, "aaaaaaaa-1", "/work", "")}, nil)
	if _, err := doThere(AgentList, Places{Store: store, Notes: notes, Home: agentHome(claude, asks, "")}, Input{}, Acting{}); err != nil {
		t.Fatal(err)
	}
	for _, path := range kept {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%s is gone: %v", filepath.Base(path), err)
		}
	}
	for _, path := range gone {
		if _, err := os.Stat(path); err == nil {
			t.Errorf("%s is still there", filepath.Base(path))
		}
	}
}
