package mcp

// Plumbing tests for handlerFor: they prove the handler a tool is offered
// with really reaches engine.Do, and nothing more — what a call answers, how
// a refusal comes back, and what happens when the store is behind are
// bdd/mcp.feature's next rule, not this file's.

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/task"
)

// taskFacts is a task's membrane for what a test needs to read back off it.
// A local helper, not a second way to build a task: it never builds one, it
// only reads what Tell says.
type taskFacts struct {
	title string
}

func (f *taskFacts) Numbered(int)            {}
func (f *taskFacts) Titled(s string)         { f.title = s }
func (f *taskFacts) InColumn(domain.State)   {}
func (f *taskFacts) Carries(domain.Priority) {}
func (f *taskFacts) Described(string)        {}
func (f *taskFacts) NotDescribed()           {}
func (f *taskFacts) Labelled(string)         {}
func (f *taskFacts) NotLabelled()            {}
func (f *taskFacts) LabelsTold()             {}
func (f *taskFacts) Due(time.Time)           {}
func (f *taskFacts) NotDue()                 {}
func (f *taskFacts) Estimated(time.Duration) {}
func (f *taskFacts) NotEstimated()           {}
func (f *taskFacts) Spent(time.Duration)     {}
func (f *taskFacts) NothingSpent()           {}
func (f *taskFacts) Under(int)               {}
func (f *taskFacts) UnderNothing()           {}
func (f *taskFacts) Serves(int)              {}
func (f *taskFacts) ServesNothing()          {}

// factsOf reads back what a test needs from a task.
func factsOf(t task.Task) taskFacts {
	var f taskFacts
	t.Tell(&f)
	return f
}

// storeAt gives a path to a real, up-to-date store, closed — the same shape
// internal/action's own storeAt gives, since this package needs the same
// thing to prove handlerFor reaches engine.Do at all.
func storeAt(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := engine.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return path
}

func TestHandlerFor_reachesActionDoWithWhatWasNamed(t *testing.T) {
	path := storeAt(t)
	handle := handlerFor(engine.TaskAdd, path)
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{Arguments: []byte(`{"title":"From the handler"}`)}}

	if _, err := handle(context.Background(), req); err != nil {
		t.Fatalf("handle: %v", err)
	}

	answer, err := engine.Do(engine.TaskList, path, engine.Input{})
	if err != nil {
		t.Fatalf("Do(TaskList): %v", err)
	}
	if len(answer.Tasks) != 1 || factsOf(answer.Tasks[0]).title != "From the handler" {
		t.Fatalf("Tasks = %v, want one titled %q", answer.Tasks, "From the handler")
	}
}

func TestHandlerFor_readsNoArgumentsAsNothingNamed(t *testing.T) {
	path := storeAt(t)
	handle := handlerFor(engine.TaskAdd, path)
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{}}

	// Empty arguments must reach engine.Do as nothing named, and come back as
	// an answer marked IsError, not a protocol error: a refusal is something
	// cairn understood and turned down, and the agent has to be able to read
	// it as one.
	res, err := handle(context.Background(), req)
	if err != nil {
		t.Fatalf("handle = %v, want the refusal answered rather than returned as an error", err)
	}
	if !res.IsError {
		t.Fatal("IsError = false, want the refusal marked as a failure")
	}
	text, ok := res.Content[0].(*sdk.TextContent)
	if !ok || !strings.Contains(text.Text, domain.ErrTitleRequired.Error()) {
		t.Fatalf("content = %v, want it to carry %q", res.Content, domain.ErrTitleRequired)
	}
}

func TestHandlerFor_refusesArgumentsItCannotRead(t *testing.T) {
	path := storeAt(t)
	handle := handlerFor(engine.TaskAdd, path)
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{Arguments: []byte(`not json`)}}

	if _, err := handle(context.Background(), req); err == nil {
		t.Fatal("handle = nil, want an error reading arguments that are not valid JSON")
	}
}
