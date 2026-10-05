package mcp

// Plumbing tests for handlerFor: they prove the handler a tool is offered
// with really reaches engine.Do, and nothing more — what a call answers, how
// a refusal comes back, and what happens when the store is behind are
// bdd/mcp.feature's next rule, not this file's.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/jsonschema-go/jsonschema"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/config"
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
func (f *taskFacts) Archived()               {}
func (f *taskFacts) NotArchived()            {}
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

// homeFor is a home of this test's own whose environment names the store at
// path — the way a server is started with CAIRN_STORE set.
func homeFor(t *testing.T, path string) engine.Home {
	t.Helper()
	dir := t.TempDir()
	return engine.Home{Dir: dir, Lookup: func(name string) (string, bool) {
		if name == config.StoreVariable {
			return path, true
		}
		return "", false
	}}
}

// config_show is answered by the engine, so an agent is told the file a
// person would be told, in both renderings. It used to be read by the command
// line alone, and an agent was answered an object with every field empty.
func TestHandlerFor_showsTheConfigurationToAnAgent(t *testing.T) {
	home := homeFor(t, storeAt(t))
	dir := filepath.Join(home.Dir, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(file, []byte("notes = \"~/vault\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	handle := handlerFor(engine.ConfigShow, home)
	result, err := handle(context.Background(), &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{}})
	if err != nil || result.IsError {
		t.Fatalf("handle = %+v, %v", result, err)
	}
	said, ok := result.StructuredContent.(engine.Config)
	if !ok || said.File != file || said.Notes.By != "the file" {
		t.Errorf("structured = %#v, want the file and the notes it named", result.StructuredContent)
	}
	text := result.Content[0].(*sdk.TextContent).Text
	if !strings.Contains(text, file) || !strings.Contains(text, "named by the file") {
		t.Errorf("text = %q, want the file and what named the notes", text)
	}
}

// A number is read as the number it is. It went through fmt.Sprint on a
// float, so a million arrived as "1e+06" and was refused as not a number —
// rather than answered as a task nobody numbered that high.
func TestHandlerFor_readsALargeNumberAsTheNumberItIs(t *testing.T) {
	handle := handlerFor(engine.TaskShow, homeFor(t, storeAt(t)))
	result, err := handle(context.Background(), &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{Arguments: []byte(`{"number": 1000000}`)}})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	text := result.Content[0].(*sdk.TextContent).Text
	if strings.Contains(text, "e+06") || !strings.Contains(text, "1000000") {
		t.Errorf("the answer reads %q, want it about task 1000000", text)
	}
}

func TestHandlerFor_reachesActionDoWithWhatWasNamed(t *testing.T) {
	path := storeAt(t)
	handle := handlerFor(engine.TaskAdd, homeFor(t, path))
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
	handle := handlerFor(engine.TaskAdd, homeFor(t, path))
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
	handle := handlerFor(engine.TaskAdd, homeFor(t, path))
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
	handle := handlerFor(engine.TaskAdd, homeFor(t, path))
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
	handle := handlerFor(engine.TaskAdd, homeFor(t, path))
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

	handle := handlerFor(engine.TaskEdit, homeFor(t, path))
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

// What a client is told a call costs: whether it only looks, whether calling it
// twice does more than calling it once, and whether it can replace what was
// written — an edit empties a description as surely as a deletion does.
func TestToolFor_saysWhatACallCosts(t *testing.T) {
	for _, c := range []struct {
		op                    engine.Operation
		destructive, repeated bool
	}{
		{engine.TaskList, false, true},
		{engine.TaskAdd, false, false},
		{engine.TaskEdit, true, true},
		{engine.TaskLabel, true, true},
		{engine.TimesheetAdd, false, false},
		{engine.TaskDelete, true, true},
		{engine.MigrationRollback, true, false},
	} {
		said := toolFor(c.op).Annotations
		if *said.DestructiveHint != c.destructive {
			t.Errorf("%s: destructiveHint = %v, want %v", c.op.Name(), *said.DestructiveHint, c.destructive)
		}
		if said.IdempotentHint != c.repeated {
			t.Errorf("%s: idempotentHint = %v, want %v", c.op.Name(), said.IdempotentHint, c.repeated)
		}
		if said.Title != c.op.Short {
			t.Errorf("%s: title = %q, want the operation's own short sentence", c.op.Name(), said.Title)
		}
	}
}

// A client reads the schema before it calls: a name the operation does not
// take, or a list holding something other than words, is refused there.
func TestSchemaFor_takesNothingItDoesNotName(t *testing.T) {
	schema := schemaFor(engine.TaskAdd)
	if schema["additionalProperties"] != false {
		t.Errorf("additionalProperties = %v, want false", schema["additionalProperties"])
	}
	label := schema["properties"].(map[string]any)["label"].(map[string]any)
	if items, _ := label["items"].(map[string]any); items["type"] != "string" {
		t.Errorf("label items = %v, want strings", label["items"])
	}
}

// A refusal carries what kind it is and which, so an agent branches on words
// that do not change when the sentence is reworded.
func TestHandlerFor_answersARefusalWithItsKindAndReason(t *testing.T) {
	handle := handlerFor(engine.TaskShow, homeFor(t, storeAt(t)))
	result, err := handle(context.Background(), &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{Arguments: []byte(`{"number": 4}`)}})
	if err != nil || !result.IsError {
		t.Fatalf("handle = %+v, %v; want a refusal", result, err)
	}
	said, ok := result.Meta[refusalKey].(engine.Refusal)
	if !ok || said.Kind != engine.KindRefused || said.Reason != "no-such-task" {
		t.Errorf("_meta = %#v, want refused/no-such-task", result.Meta)
	}
	if result.StructuredContent != nil {
		t.Errorf("structured = %#v, want nothing a client would measure against the answer's shape", result.StructuredContent)
	}
}

// A call the client gave up on before it ran is not run.
func TestHandlerFor_runsNothingTheClientGaveUpOn(t *testing.T) {
	path := storeAt(t)
	handle := handlerFor(engine.TaskAdd, homeFor(t, path))
	gone, cancel := context.WithCancel(context.Background())
	cancel()
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{Arguments: []byte(`{"title":"Too late"}`)}}
	if _, err := handle(gone, req); err == nil {
		t.Error("handle = nil, want the cancellation said")
	}
	answer, err := engine.Do(engine.TaskList, path, engine.Input{})
	if err != nil {
		t.Fatalf("Do(TaskList): %v", err)
	}
	if len(answer.Tasks) != 0 {
		t.Errorf("%d tasks, want none: the call was given up on", len(answer.Tasks))
	}
}
