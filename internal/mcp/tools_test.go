package mcp

// Plumbing tests for handlerFor: they prove the handler a tool is offered
// with really reaches engine.Do, and nothing more — what a call answers, how
// a refusal comes back, and what happens when the store is behind are
// bdd/mcp.feature's next rule, not this file's.

import (
	"context"
	"fmt"
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
	// it as one. "title" is required and was never named, which is now
	// engine.Operation.Check's own refusal rather than the empty-title one
	// domain raises once Run is reached — the same call is still refused, and
	// the reason still names what was missing.
	res, err := handle(context.Background(), req)
	if err != nil {
		t.Fatalf("handle = %v, want the refusal answered rather than returned as an error", err)
	}
	if !res.IsError {
		t.Fatal("IsError = false, want the refusal marked as a failure")
	}
	text, ok := res.Content[0].(*sdk.TextContent)
	if !ok || !strings.Contains(text.Text, `"title"`) {
		t.Fatalf("content = %v, want it to name %q, the argument that was never given", res.Content, "title")
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

// TestHandlerFor_refusesAListHandedAsSomethingThatIsNotOne pins the decision
// this slice made for {"label": "spec"} instead of {"label": ["spec"]}: a
// bare string is not the one-element list it might have meant, so it is
// refused by name rather than guessed at.
func TestHandlerFor_refusesAListHandedAsSomethingThatIsNotOne(t *testing.T) {
	path := storeAt(t)
	handle := handlerFor(engine.TaskAdd, path)
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{
		Arguments: []byte(`{"title":"From the handler","label":"spec"}`),
	}}

	_, err := handle(context.Background(), req)
	if err == nil {
		t.Fatal(`handle = nil, want an error: "label" was handed a string, not a list`)
	}
	if !strings.Contains(err.Error(), `"label"`) {
		t.Fatalf("err = %v, want it to name %q", err, "label")
	}
}

// TestHandlerFor_refusesAListElementThatIsNotAString pins the decision this
// slice made for {"label": [1, true]}: an element that is not a string is
// refused by name rather than turned into one.
func TestHandlerFor_refusesAListElementThatIsNotAString(t *testing.T) {
	path := storeAt(t)
	handle := handlerFor(engine.TaskAdd, path)
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{
		Arguments: []byte(`{"title":"From the handler","label":[1,true]}`),
	}}

	_, err := handle(context.Background(), req)
	if err == nil {
		t.Fatal(`handle = nil, want an error: an element of "label" is not a string`)
	}
	if !strings.Contains(err.Error(), `"label"`) {
		t.Fatalf("err = %v, want it to name %q", err, "label")
	}
}

// TestHandlerFor_emptiesAListNamedAsNull proves JSON null empties a Kind:
// List argument the same way it empties any other kind: a name present and
// holding nothing, not the same request as leaving it out. The label is
// seeded through engine.Do directly, not through the handler under test, so
// the precondition and the assertion do not travel the same path and cancel
// out.
func TestHandlerFor_emptiesAListNamedAsNull(t *testing.T) {
	path := storeAt(t)
	added, err := engine.Do(engine.TaskAdd, path, engine.Input{
		Given:  map[string]string{"title": "From the handler"},
		Listed: map[string][]string{"label": {"spec"}},
	})
	if err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}
	number := task.OfTask(*added.Task).Number

	handle := handlerFor(engine.TaskEdit, path)
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{
		Arguments: []byte(fmt.Sprintf(`{"number":%d,"label":null}`, number)),
	}}
	if _, err := handle(context.Background(), req); err != nil {
		t.Fatalf("handle: %v", err)
	}

	shown, err := engine.Do(engine.TaskShow, path, engine.Input{Given: map[string]string{"number": fmt.Sprint(number)}})
	if err != nil {
		t.Fatalf("Do(TaskShow): %v", err)
	}
	if labels := task.OfTask(*shown.Task).Labels; len(labels) != 0 {
		t.Fatalf("labels = %v, want none", labels)
	}
}
