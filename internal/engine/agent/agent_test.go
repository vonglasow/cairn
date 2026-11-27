package agent

import (
	"reflect"
	"testing"
)

func TestCommands(t *testing.T) {
	for _, tt := range []struct {
		name string
		got  string
		want string
	}{
		{"a folder is one word whatever it holds", ResumeOf("darwin", "/w/it's; rm -rf ~", "a3f9c1d2-1"), `cd '/w/it'\''s; rm -rf ~' && claude --resume a3f9c1d2-1`},
		{"an id that is not a plain word is quoted", ResumeOf("linux", "/w", "a b"), `cd '/w' && claude --resume 'a b'`},
		{"on Windows a folder is one word to PowerShell", ResumeOf("windows", `C:\w\it's; rm ~`, "a3f9c1d2-1"), `cd 'C:\w\it''s; rm ~'; claude --resume a3f9c1d2-1`},
		{"on Windows an id that is not a plain word is quoted for PowerShell", ResumeOf("windows", `C:\w`, "it's"), `cd 'C:\w'; claude --resume 'it''s'`},
		{"a session started on a task is asked to work on it", CommandFor(12), `claude 'let'\''s work on cairn#12'`},
		{"a session started on no task is claude alone", CommandFor(0), "claude"},
	} {
		if tt.got != tt.want {
			t.Errorf("%s: %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestProgramOf(t *testing.T) {
	for entrypoint, want := range map[string]string{"": "", "cli": "", "claude-vscode": "VS Code", "claude-desktop": "Claude Desktop", "sdk-ts": "sdk-ts"} {
		if got := ProgramOf(entrypoint); got != want {
			t.Errorf("%q: %q, want %q", entrypoint, got, want)
		}
	}
}

func TestStateOf(t *testing.T) {
	for _, tt := range []struct {
		status, waitingFor string
		want               State
	}{
		{"busy", "", Working},
		{"shell", "", Working},
		{"idle", "", Idle},
		{"idle", "permission", Waiting},
	} {
		if got := StateOf(tt.status, tt.waitingFor); got != tt.want {
			t.Errorf("%s/%s: %s, want %s", tt.status, tt.waitingFor, got, tt.want)
		}
	}
}

func TestTexts(t *testing.T) {
	listed := Agents{Agents: []Agent{
		{Short: "a3f9c1", State: Waiting, Task: &Held{Number: 1, Title: "Spec"}, Folder: "/w/cairn", Waiting: "permission"},
		{Short: "c04d77", State: Idle, Folder: "/w/ops"},
	}}
	for _, tt := range []struct {
		name, got, want string
	}{
		{"a listing, a line a session", AgentsText(listed), "a3f9c1  waiting  #1 Spec  /w/cairn  waits for permission\nc04d77  idle     no task  /w/ops\n"},
		{"no session", AgentsText(Agents{}), "no Claude session is running\n"},
		{"the hook's answer is printed as it is", DecisionText(Decision{Hook: []byte(`{"a":1}`)}), "{\"a\":1}\n"},
		{"an answer given", DecisionText(Decision{Session: "a3f9c1d2-1", Answer: Allow}), "a3f9c1 answered allow\n"},
		{"no answer prints nothing", DecisionText(Decision{}), ""},
		{"a session started", StartedText(Started{Terminal: "Ghostty", Command: "claude", Folder: "/w"}), "Ghostty runs claude in /w\n"},
		{"the terminals installed", TerminalsText(Terminals{Terminals: []string{"Terminal", "Ghostty"}}), "Terminal\nGhostty\n"},
		{"no terminal installed", TerminalsText(Terminals{}), "no terminal cairn starts a session in is installed\n"},
	} {
		if tt.got != tt.want {
			t.Errorf("%s: %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestPendingOf_saysWhatTheToolActsOn(t *testing.T) {
	for input, want := range map[string]string{
		`{"command":"make test\nmake lint"}`:    "make test",
		`{"file_path":"/w/main.go"}`:            "/w/main.go",
		`{"limit":3,"pattern":"restartPolicy"}`: "restartPolicy",
		`{"limit":3}`:                           "",
		`[1]`:                                   "",
	} {
		if got := PendingOf(Request{Tool: "Bash", Input: []byte(input)}).On; got != want {
			t.Errorf("%s: %q, want %q", input, got, want)
		}
	}
}

func TestShortOf(t *testing.T) {
	if got := ShortOf("a3f9"); got != "a3f9" {
		t.Errorf("%q", got)
	}
}

func TestModelSaid(t *testing.T) {
	for id, want := range map[string]string{
		"claude-opus-5-5":           "Opus 5.5",
		"claude-sonnet-5-5":         "Sonnet 5.5",
		"claude-haiku-4-5-20251001": "Haiku 4.5",
		"claude-opus-4-1":           "Opus 4.1",
		"claude-3-5-sonnet":         "claude-3-5-sonnet",
		"gpt-5":                     "gpt-5",
		"":                          "",
	} {
		if got := ModelSaid(id); got != want {
			t.Errorf("%q: %q, want %q", id, got, want)
		}
	}
}

func TestAskOf_carriesTheQuestionsAsked(t *testing.T) {
	input := `{"questions":[{"question":"Which store?","header":"Store","options":[{"label":"Postgres","description":"the server"},{"label":"SQLite"}],"multiSelect":true}]}`
	got := AskOf(PendingOf(Request{Tool: AskUserQuestion, Input: []byte(input)}))
	want := []Question{{Question: "Which store?", Header: "Store", Options: []Option{{Label: "Postgres", Description: "the server"}, {Label: "SQLite"}}, Multi: true}}
	if got.Tool != AskUserQuestion || got.On != "" || got.Always || !reflect.DeepEqual(got.Questions, want) {
		t.Errorf("%+v", got)
	}
}

func TestDoingOf(t *testing.T) {
	for _, tt := range []struct{ tool, on, description, want string }{
		{"Bash", "cd /Users/me/work && go test ./...", "Run the engine tests", "Run the engine tests"},
		{"Bash", "cd /Users/me/work && go test ./...", "", "Runs go"},
		{"Bash", "/usr/bin/make test", "", "Runs make"},
		{"Read", "/Users/me/work/internal/gui/card.go", "", "Reads card.go"},
		{"Edit", "/Users/me/work/internal/gui/card.go", "", "Edits card.go"},
		{"Write", "/Users/me/work/card.go", "", "Writes card.go"},
		{"NotebookEdit", "/Users/me/work/look.ipynb", "", "Edits look.ipynb"},
		{"Grep", "restartPolicy", "", "Searches restartPolicy"},
		{"Glob", "**/*.go", "", "Searches **/*.go"},
		{"WebFetch", "https://example.com/a", "", "Reads the web"},
		{"WebSearch", "gio layout", "", "Reads the web"},
		{"Task", "", "", "Runs a subagent"},
		{"Agent", "", "", "Runs a subagent"},
		{"mcp__azure-devops__pipelines_build", "", "", "Uses azure-devops pipelines_build"},
		{"SendMessage", "", "", "SendMessage"},
		{"", "", "", ""},
	} {
		if got := DoingOf(tt.tool, tt.on, tt.description); got != tt.want {
			t.Errorf("%s %q: %q, want %q", tt.tool, tt.on, got, tt.want)
		}
	}
}

func TestContextOf_countsAgainstTheLargerWindowPastTheSmaller(t *testing.T) {
	for tokens, want := range map[int]int{100000: 50, 200000: 100, 231000: 23, 2000000: 200} {
		if got := ContextOf(tokens); got != want {
			t.Errorf("%d tokens: %d%%, want %d%%", tokens, got, want)
		}
	}
}

func TestProjectOf(t *testing.T) {
	for folder, want := range map[string]string{
		"/Users/me/work/cairn": "cairn",
		"/Users/me/work/shared-devops/.claude/worktrees/feature-85621-registry-pvc-size":     "shared-devops",
		"/Users/me/work/shared-devops/.claude/worktrees/feature-85621-registry-pvc-size/k8s": "shared-devops",
		"/Users/me/.claude": ".claude",
	} {
		if got := ProjectOf(folder); got != want {
			t.Errorf("%s: %q, want %q", folder, got, want)
		}
	}
}

func TestContextLevelOf(t *testing.T) {
	for percent, want := range map[int]ContextLevel{0: ContextRoomy, 69: ContextRoomy, 70: ContextHigh, 84: ContextHigh, 85: ContextFull, 100: ContextFull} {
		if got := ContextLevelOf(percent); got != want {
			t.Errorf("%d%%: %v, want %v", percent, got, want)
		}
	}
}
