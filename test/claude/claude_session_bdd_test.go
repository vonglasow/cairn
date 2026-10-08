package claude

// Executable acceptance scenarios for bdd/claude-session.feature: the Claude
// Code hook of plugin/claude, driven the way Claude Code drives it — one JSON
// document on standard input per event — against the real cairn binary on a
// store and notes of the scenario's own.
//
// It sits two levels under the root, as every package holding scenarios does,
// and in no package of cairn's: the plugin knows cairn only through its
// command line, and so do these steps.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"cairn/test"
)

// idleSeconds is the idle delay the hook is given here, and settle how long
// past it a step waits for the hook's wait to have answered.
const (
	idleSeconds = 1
	settle      = 1500 * time.Millisecond
)

var (
	built     sync.Once
	binaries  string
	buildFail error
)

// cairnOnPath builds cairn once for the whole package, into a directory the
// hook finds it in through PATH, the way it finds it on a person's machine.
func cairnOnPath(t *testing.T) string {
	built.Do(func() {
		binaries, buildFail = os.MkdirTemp("", "cairn-claude-")
		if buildFail != nil {
			return
		}
		out, err := exec.Command("go", "build", "-o", filepath.Join(binaries, "cairn"), "../../cmd/cairn").CombinedOutput()
		if err != nil {
			buildFail = fmt.Errorf("building cairn: %w\n%s", err, out)
		}
	})
	if buildFail != nil {
		t.Fatal(buildFail)
	}
	return binaries
}

// sessionWorld is one scenario's machine: its store, its notes, its home, and
// what the hook last answered.
type sessionWorld struct {
	t      *testing.T
	env    []string
	notes  string
	tasks  int
	said   string
	failed error
}

func (w *sessionWorld) reset() {
	home := w.t.TempDir()
	w.notes = filepath.Join(home, "notes")
	w.tasks = 0
	w.said, w.failed = "", nil
	w.env = append(os.Environ(),
		"HOME="+home,
		"PATH="+cairnOnPath(w.t)+string(os.PathListSeparator)+os.Getenv("PATH"),
		"CAIRN_STORE="+filepath.Join(home, "cairn.db"),
		"CAIRN_NOTES="+w.notes,
		"CAIRN_IDLE_SECONDS="+strconv.Itoa(idleSeconds),
	)
}

// cairn runs the real command line, as the person or the MCP server would.
func (w *sessionWorld) cairn(args ...string) (string, error) {
	// The binary built here, by its path: a name would be looked for on this
	// process's PATH, where a cairn installed on the machine answers instead.
	run := exec.Command(filepath.Join(cairnOnPath(w.t), "cairn"), args...)
	run.Env = w.env
	out, err := run.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("cairn %v: %w\n%s", args, err, out)
	}
	return string(out), nil
}

// hook hands the hook one event, and keeps what it wrote and how it ended.
func (w *sessionWorld) hook(event map[string]any) {
	if _, ok := event["cwd"]; !ok {
		event["cwd"] = "/work/project"
	}
	input, _ := json.Marshal(event)
	run := exec.Command("bash", filepath.Join("..", "..", "plugin", "claude", "hooks", "cairn-session.sh"))
	run.Env = w.env
	run.Stdin = strings.NewReader(string(input))
	// The idle wait outlives the hook on purpose; its output is closed, but a
	// bound keeps a child that held a pipe from hanging the run.
	run.WaitDelay = time.Second
	out, err := run.CombinedOutput()
	w.said, w.failed = string(out), err
}

func (w *sessionWorld) toolUse(session, tool string, input map[string]any, cwd string) {
	event := map[string]any{
		"hook_event_name": "PostToolUse",
		"session_id":      session,
		"tool_name":       "mcp__cairn__" + tool,
		"tool_input":      input,
	}
	if cwd != "" {
		event["cwd"] = cwd
	}
	w.hook(event)
}

func (w *sessionWorld) event(session, name string) {
	w.hook(map[string]any{"hook_event_name": name, "session_id": session})
}

// aTask adds tasks until the one numbered n exists.
func (w *sessionWorld) aTask(n int) error {
	for w.tasks < n {
		w.tasks++
		if _, err := w.cairn("task", "add", "Task "+strconv.Itoa(w.tasks)); err != nil {
			return err
		}
	}
	return nil
}

// holds ties a session to a task without the hook — the entry written by hand
// and the focus taken by cairn — so that what a scenario asserts is never the
// precondition coming back through the code under test.
func (w *sessionWorld) holds(session string, n int) error {
	if err := w.aTask(n); err != nil {
		return err
	}
	dir := filepath.Join(w.notes, strconv.Itoa(n))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entry := filepath.Join(dir, time.Now().Format("2006-01-02")+"-session-"+session+".md")
	if err := os.WriteFile(entry, []byte("cd /work/project\nclaude --resume "+session+"\n"), 0o644); err != nil {
		return err
	}
	_, err := w.cairn("task", "focus", strconv.Itoa(n))
	return err
}

// entry is what a session's entry in a task's notes says, found by its
// session and not by its day, so a scenario run across midnight still finds it.
func (w *sessionWorld) entry(session string, n int) (string, error) {
	found, _ := filepath.Glob(filepath.Join(w.notes, strconv.Itoa(n), "*-session-"+session+".md"))
	if len(found) != 1 {
		return "", fmt.Errorf("the notes of task %d hold %d entries for session %q, want one", n, len(found), session)
	}
	said, err := os.ReadFile(found[0])
	return string(said), err
}

func released(entry string) bool {
	for _, line := range strings.Split(entry, "\n") {
		if strings.HasPrefix(line, "released") {
			return true
		}
	}
	return false
}

func (w *sessionWorld) openEntry(session string, n int) error {
	said, err := w.entry(session, n)
	if err != nil {
		return err
	}
	if released(said) {
		return fmt.Errorf("the entry of session %q in the notes of task %d was released:\n%s", session, n, said)
	}
	return nil
}

func (w *sessionWorld) releasedEntry(session string, n int) error {
	said, err := w.entry(session, n)
	if err != nil {
		return err
	}
	if !released(said) {
		return fmt.Errorf("the entry of session %q in the notes of task %d is still open:\n%s", session, n, said)
	}
	return nil
}

func (w *sessionWorld) inFocus() ([]int, error) {
	out, err := w.cairn("task", "focus", "--json")
	if err != nil {
		return nil, err
	}
	var focus struct {
		InFocus []int `json:"in_focus"`
	}
	if err := json.Unmarshal([]byte(out), &focus); err != nil {
		return nil, fmt.Errorf("reading the focus: %w\n%s", err, out)
	}
	return focus.InFocus, nil
}

func (w *sessionWorld) focused(n int, want bool) error {
	focus, err := w.inFocus()
	if err != nil {
		return err
	}
	if slices.Contains(focus, n) != want {
		return fmt.Errorf("task %d in focus is %v, want %v — in focus: %v", n, !want, want, focus)
	}
	return nil
}

func (w *sessionWorld) quietAndFine() error {
	if w.failed != nil {
		return fmt.Errorf("the hook failed: %w\n%s", w.failed, w.said)
	}
	if w.said != "" {
		return fmt.Errorf("the hook said %q, want nothing", w.said)
	}
	return nil
}

func initSession(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := &sessionWorld{t: t}
		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			w.reset()
			return ctx, nil
		})

		sc.Given(`^a task numbered (\d+)$`, w.aTask)
		sc.Given(`^a task numbered (\d+) in focus$`, func(n int) error {
			if err := w.aTask(n); err != nil {
				return err
			}
			_, err := w.cairn("task", "focus", strconv.Itoa(n))
			return err
		})
		sc.Given(`^session "([^"]*)" holds task (\d+)$`, w.holds)

		sc.When(`^session "([^"]*)", working in "([^"]*)", puts task (\d+) in focus$`, func(session, cwd string, n int) error {
			if _, err := w.cairn("task", "focus", strconv.Itoa(n)); err != nil {
				return err
			}
			w.toolUse(session, "task_focus", map[string]any{"number": n}, cwd)
			return nil
		})
		sc.When(`^session "([^"]*)" puts task (\d+) in focus$`, func(session string, n int) error {
			if _, err := w.cairn("task", "focus", strconv.Itoa(n)); err != nil {
				return err
			}
			w.toolUse(session, "task_focus", map[string]any{"number": n}, "")
			return nil
		})
		sc.When(`^session "([^"]*)" lets go of task (\d+)$`, func(session string, n int) error {
			if _, err := w.cairn("task", "unfocus", strconv.Itoa(n)); err != nil {
				return err
			}
			w.toolUse(session, "task_unfocus", map[string]any{"number": n}, "")
			return nil
		})
		sc.When(`^session "([^"]*)" asks which tasks are in focus$`, func(session string) error {
			w.toolUse(session, "task_focus", map[string]any{}, "")
			return nil
		})
		sc.When(`^session "([^"]*)" puts task (\d+), which does not exist, in focus$`, func(session string, n int) error {
			if _, err := w.cairn("task", "focus", strconv.Itoa(n)); err == nil {
				return fmt.Errorf("cairn put task %d in focus, which the scenario says does not exist", n)
			}
			w.toolUse(session, "task_focus", map[string]any{"number": n}, "")
			return nil
		})
		sc.When(`^session "([^"]*)" lets go of every task at once$`, func(session string) error {
			if _, err := w.cairn("task", "unfocus"); err != nil {
				return err
			}
			w.toolUse(session, "task_unfocus", map[string]any{}, "")
			return nil
		})
		sc.When(`^session "([^"]*)", tied to no task, receives a message$`, func(session string) error {
			w.event(session, "UserPromptSubmit")
			return nil
		})
		sc.When(`^session "([^"]*)", tied to no task, answers$`, func(session string) error {
			w.event(session, "Stop")
			time.Sleep(time.Duration(idleSeconds)*time.Second + settle)
			return nil
		})
		sc.When(`^session "([^"]*)", tied to no task, ends$`, func(session string) error {
			w.event(session, "SessionEnd")
			return nil
		})
		sc.When(`^session "([^"]*)" ends$`, func(session string) error {
			w.event(session, "SessionEnd")
			return nil
		})
		sc.When(`^session "([^"]*)" answers and no message follows within the idle delay$`, func(session string) error {
			w.event(session, "Stop")
			time.Sleep(time.Duration(idleSeconds)*time.Second + settle)
			return nil
		})
		sc.When(`^session "([^"]*)" answers and a message follows within the idle delay$`, func(session string) error {
			w.event(session, "Stop")
			w.event(session, "UserPromptSubmit")
			return nil
		})
		sc.When(`^a message comes in session "([^"]*)"$`, func(session string) error {
			w.event(session, "UserPromptSubmit")
			return nil
		})

		sc.Then(`^the notes of task (\d+) hold an open entry for session "([^"]*)"$`, func(n int, session string) error {
			return w.openEntry(session, n)
		})
		sc.Then(`^that entry says "([^"]*)" and "([^"]*)"$`, func(first, second string) error {
			found, _ := filepath.Glob(filepath.Join(w.notes, "*", "*-session-*.md"))
			if len(found) != 1 {
				return fmt.Errorf("the notes hold %d entries, want one", len(found))
			}
			said, err := os.ReadFile(found[0])
			if err != nil {
				return err
			}
			for _, want := range []string{first, second} {
				if !strings.Contains(string(said), want) {
					return fmt.Errorf("the entry says nothing of %q:\n%s", want, said)
				}
			}
			return nil
		})
		sc.Then(`^the hook said nothing and succeeded$`, w.quietAndFine)
		sc.Then(`^the entry of session "([^"]*)" in the notes of task (\d+) says it was released$`, func(session string, n int) error {
			return w.releasedEntry(session, n)
		})
		sc.Then(`^the entry of session "([^"]*)" in the notes of task (\d+) is still open$`, func(session string, n int) error {
			return w.openEntry(session, n)
		})
		sc.Then(`^after a message in session "([^"]*)", task (\d+) is not in focus$`, func(session string, n int) error {
			w.event(session, "UserPromptSubmit")
			return w.focused(n, false)
		})
		sc.Then(`^no notes hold an entry for session "([^"]*)"$`, func(session string) error {
			found, _ := filepath.Glob(filepath.Join(w.notes, "*", "*-session-"+session+".md"))
			if len(found) != 0 {
				return fmt.Errorf("the notes hold entries for session %q: %v", session, found)
			}
			return nil
		})
		sc.Then(`^task (\d+) is in focus$`, func(n int) error { return w.focused(n, true) })
		sc.Then(`^task (\d+) is not in focus$`, func(n int) error { return w.focused(n, false) })
		sc.Then(`^once the idle delay has passed, task (\d+) is in focus$`, func(n int) error {
			time.Sleep(time.Duration(idleSeconds)*time.Second + settle)
			return w.focused(n, true)
		})
	}
}

func TestAcceptance_ClaudeSession(t *testing.T) {
	if test.Suite(t, "claude-session.feature", initSession(t), "").Run() != 0 {
		t.Fatal("the Claude session's scenarios did not all pass")
	}
}
