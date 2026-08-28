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

	"github.com/google/jsonschema-go/jsonschema"

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
func (f *taskFacts) BlockedBy(int)           {}
func (f *taskFacts) NotBlocked()             {}
func (f *taskFacts) BlockersTold()           {}
func (f *taskFacts) Because(string)          {}
func (f *taskFacts) NotBecause()             {}
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
func (f *taskFacts) ObjectivesTold()         {}

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

// The shape a tool says its answer takes, from what the operation declared.
func TestAnswerSchemaFor_readsTheShapeOffTheTypeItWasGiven(t *testing.T) {
	one := answerSchemaFor(engine.Operation{Answers: engine.Deleted{}})
	shape, ok := one.(*jsonschema.Schema)
	if !ok {
		t.Fatalf("one shape gave %T, want a schema", one)
	}
	if shape.Type != "object" {
		t.Errorf("Deleted{} reads as %q, want an object", shape.Type)
	}
	if shape.AnyOf != nil {
		t.Error("one shape is wrapped in an either-of, and there is nothing to choose between")
	}
	if _, described := shape.Properties["deleted"]; !described {
		t.Errorf("the shape does not name the field it carries: %v", shape.Properties)
	}
}

// A listing describes an object naming what it carries, and the fields it
// knows besides its rows. Pinned here and not only in the scenario: this is the
// shape a client validates against, and the scenario reads the description
// through a running server while this reads it off the type.
func TestAnswerSchemaFor_namesWhatAListingCarriesBesidesItsRows(t *testing.T) {
	listing := answerSchemaFor(engine.Operation{Answers: engine.Tasks{}})
	shape, ok := listing.(*jsonschema.Schema)
	if !ok {
		t.Fatalf("a listing gave %T, want a schema", listing)
	}
	if shape.Type != "object" {
		t.Errorf("a listing reads as %q, want an object", shape.Type)
	}
	for _, named := range []string{"tasks", "in_focus"} {
		if _, described := shape.Properties[named]; !described {
			t.Errorf("the shape does not name %q: %v", named, shape.Properties)
		}
	}
}

// An operation that declares nothing describes nothing, rather than describing
// something empty a client would try to hold its answer to.
func TestAnswerSchemaFor_describesNothingWhenNothingWasDeclared(t *testing.T) {
	if got := answerSchemaFor(engine.Operation{}); got != nil {
		t.Errorf("answerSchemaFor = %v, want nothing described", got)
	}
}

// Every tool offered carries the shape of its answer. Asserted here as well as
// in the scenarios: a scenario runs the server, and this runs the description on
// its own, so a tool that stopped describing itself fails in the package that
// builds it.
func TestToolFor_everyToolCarriesTheShapeOfItsAnswer(t *testing.T) {
	silent := []string{}
	for _, op := range engine.All() {
		if toolFor(op).OutputSchema == nil {
			silent = append(silent, op.Name())
		}
	}
	if len(silent) > 0 {
		t.Errorf("%d tools carry no shape: %v", len(silent), silent)
	}
}

// A shape that cannot be read off its type describes nothing rather than
// stopping the server: a tool nobody can describe is better than a server that
// will not start, and it is a programming mistake rather than something a caller
// did.
func TestAnswerSchemaFor_describesNothingWhenTheTypeCannotBeRead(t *testing.T) {
	if got := answerSchemaFor(engine.Operation{
		Answers: make(chan int),
	}); got != nil {
		t.Errorf("answerSchemaFor on a shape no schema can describe = %v, want nothing", got)
	}
}
