package engine

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cairn/internal/engine/agent"
	"cairn/internal/store/files"
)

//go:embed fixtures/fixtures.json
var fixtures []byte

const fixturesMark = ".made-by-cairn-fixtures"

// errNotMadeByFixtures is returned when the folder named for the fixtures holds something cairn fixtures did not put there: it is never emptied.
var errNotMadeByFixtures = errors.New("the folder was not made by cairn fixtures, and is left alone")

// Fixtures says where the fixtures are, and whether they were made just now rather than opened as they were left.
type Fixtures struct {
	Store string
	Notes string
	Made  bool
}

// FixturesFile is the fixtures the binary carries, as a file task import reads.
func FixturesFile() []byte { return fixtures }

// FixturesFolder is where the fixtures are made when no folder is named: one of their own in the temporary folder, never the store in use.
func FixturesFolder() string { return filepath.Join(os.TempDir(), "cairn-fixtures") }

// MakeFixtures fills a store in the folder with the fixtures the binary carries, or with the tasks of the file named; a folder it made before is opened as it was left, unless it is to be made again.
func MakeFixtures(folder, from string, fresh bool) (Fixtures, error) {
	at := Fixtures{Store: filepath.Join(folder, "cairn.db"), Notes: filepath.Join(folder, "notes")}
	entries, err := os.ReadDir(folder)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Fixtures{}, err
	}
	_, marked := os.Stat(filepath.Join(folder, fixturesMark))
	if len(entries) > 0 && marked != nil {
		return Fixtures{}, fmt.Errorf("%s: %w", folder, errNotMadeByFixtures)
	}
	if len(entries) > 0 && !fresh {
		return at, agentFixtures(folder, at.Notes)
	}
	if err := os.RemoveAll(folder); err != nil {
		return Fixtures{}, err
	}
	if err := os.MkdirAll(folder, 0o700); err != nil {
		return Fixtures{}, err
	}
	if err := os.WriteFile(filepath.Join(folder, fixturesMark), nil, 0o600); err != nil {
		return Fixtures{}, err
	}
	if err := files.Kept(files.SettingsBeside(at.Store)).Keep([]byte("[series]\non = true\n\n[agents]\npermissions = true\n\n[workspace]\ncolour = \"peach\"\nname = \"fixtures\"\n")); err != nil {
		return Fixtures{}, err
	}
	if from == "" {
		from = filepath.Join(folder, "fixtures.json")
		if err := os.WriteFile(from, fixtures, 0o600); err != nil {
			return Fixtures{}, err
		}
	}
	if _, err := doThere(TaskImport, Places{Store: at.Store, Notes: at.Notes}, Input{Given: map[string]string{"from": from}}, Acting{Through: "fixtures"}); err != nil {
		return Fixtures{}, err
	}
	at.Made = true
	return at, agentFixtures(folder, at.Notes)
}

// AsksFixtures is the folder the fixtures' permissions wait on the window in, inside the fixtures' folder.
func AsksFixtures(folder string) string { return filepath.Join(folder, "asks") }

// ClaudeFixtures is the Claude folder the fixtures' sessions are kept in, inside the fixtures' folder.
func ClaudeFixtures(folder string) string { return filepath.Join(folder, "claude") }

// agentFixtures writes the Claude sessions a fixtures' window shows, under this process's pid so they run while it does: one waiting on a long branch, two working on one task — one of them in VS Code, the other silent with its context nearly full — one with no task the task in focus no session holds is suggested to, one asking questions, one idle.
func agentFixtures(folder, notes string) error {
	claude := ClaudeFixtures(folder)
	for _, made := range []string{claude, AsksFixtures(folder)} {
		if err := os.RemoveAll(made); err != nil {
			return err
		}
	}
	asked, err := json.Marshal(agent.Pending{Tool: "Bash", On: "make test", Suggestions: json.RawMessage(`[{"type":"addRules","rules":[{"toolName":"Bash","ruleContent":"make test"}],"behavior":"allow","destination":"localSettings"}]`)})
	if err != nil {
		return err
	}
	if err := written(askAt(AsksFixtures(folder), "a3f9c1d2-waiting"), asked); err != nil {
		return err
	}
	questions, err := json.Marshal(agent.PendingOf(agent.Request{Tool: agent.AskUserQuestion, Input: json.RawMessage(`{"questions":[` +
		`{"question":"Which store should the migration plan target first?","header":"Store","multiSelect":false,"options":[{"label":"Postgres","description":"The server the team already runs"},{"label":"SQLite","description":"One file beside the binary"}]},` +
		`{"question":"What should the plan cover?","header":"Plan","multiSelect":true,"options":[{"label":"Rollback","description":"Going back when a step fails"},{"label":"Backups","description":"A copy before each step"},{"label":"Timings","description":"How long each step takes"}]}]}`)}))
	if err != nil {
		return err
	}
	if err := written(askAt(AsksFixtures(folder), "e1b2c3d4-questions"), questions); err != nil {
		return err
	}
	now := time.Now()
	for i, s := range []struct {
		id, name, status, waiting, branch, model, tool, on, entry string
		task, tokens                                              int
		since                                                     time.Duration
	}{
		{"a3f9c1d2-waiting", "packaging", "busy", "permission", "feature/85621-registry-pvc-size-and-retention", "claude-opus-5-5", "Bash", "make test", "", 1, 144000, 2 * time.Minute},
		{"7be210aa-working", "overrun", "busy", "", "fix/overrun", "claude-sonnet-5-5", "Edit", "internal/gui/card.go", "", 2, 176000, 16 * time.Minute},
		{"5d0e44bb-working", "overrun-review", "busy", "", "fix/overrun", "claude-opus-5-5", "Read", "internal/board/card.go", "claude-vscode", 2, 18000, time.Minute},
		{"c04d77cc-no-task", "bootloader", "busy", "", "fix/60-bootloader", "claude-haiku-4-5-20251001", "Grep", "restartPolicy", "", 0, 16000, time.Minute},
		{"e1b2c3d4-questions", "migration-plan", "busy", "question", "docs/migration-plan", "claude-opus-5-5", "AskUserQuestion", "", "", 24, 51000, 3 * time.Minute},
		{"91e0aadd-idle", "helm-chart", "idle", "", "main", "claude-opus-5-5", "", "", "", 0, 108000, 14 * time.Minute},
	} {
		session, err := json.Marshal(map[string]any{
			"pid": os.Getpid(), "sessionId": s.id, "cwd": folder, "name": s.name, "status": s.status,
			"waitingFor": s.waiting, "entrypoint": s.entry, "statusUpdatedAt": now.Add(-s.since).UnixMilli(),
		})
		if err != nil {
			return err
		}
		if err := written(filepath.Join(claude, "sessions", fmt.Sprintf("%d.json", i)), session); err != nil {
			return err
		}
		said, err := json.Marshal(map[string]any{"type": "assistant", "gitBranch": s.branch, "message": map[string]any{
			"model": s.model, "usage": map[string]int{"input_tokens": s.tokens},
			"content": []map[string]any{{"type": "tool_use", "name": s.tool, "input": map[string]string{"file_path": s.on}}},
		}})
		if err != nil {
			return err
		}
		transcript := filepath.Join(claude, "projects", "fixtures", s.id+".jsonl")
		if err := written(transcript, said); err != nil {
			return err
		}
		if err := os.Chtimes(transcript, now.Add(-s.since), now.Add(-s.since)); err != nil {
			return err
		}
		if s.task != 0 {
			if err := files.TieSession(notes, s.id, folder, s.task, now.Format(time.DateOnly), now.UTC().Format(time.RFC3339)); err != nil {
				return err
			}
		}
	}
	return nil
}

func written(path string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}
