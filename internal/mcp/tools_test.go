package mcp

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

func factsOf(t task.Task) taskFacts {
	var f taskFacts
	t.Tell(&f)
	return f
}

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
	handle := handlerFor(engine.TaskNew, homeFor(t, path))
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
	handle := handlerFor(engine.TaskNew, homeFor(t, path))
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{}}

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
	handle := handlerFor(engine.TaskNew, homeFor(t, storeAt(t)))
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{Arguments: []byte(`not json`)}}
	refusedAsNotUnderstood(t, handle, req, "")
}

func TestHandlerFor_refusesAListHandedAsSomethingThatIsNotOne(t *testing.T) {
	handle := handlerFor(engine.TaskNew, homeFor(t, storeAt(t)))
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{
		Arguments: []byte(`{"title":"From the handler","label":"spec"}`),
	}}
	refusedAsNotUnderstood(t, handle, req, `"label"`)
}

func TestHandlerFor_refusesAListElementThatIsNotAString(t *testing.T) {
	handle := handlerFor(engine.TaskNew, homeFor(t, storeAt(t)))
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{
		Arguments: []byte(`{"title":"From the handler","label":[1,true]}`),
	}}
	refusedAsNotUnderstood(t, handle, req, `"label"`)
}

func refusedAsNotUnderstood(t *testing.T, handle sdk.ToolHandler, req *sdk.CallToolRequest, naming string) {
	t.Helper()
	result, err := handle(context.Background(), req)
	if err != nil || result == nil || !result.IsError {
		t.Fatalf("handle = %+v, %v; want a refusal answered", result, err)
	}
	if text := result.Content[0].(*sdk.TextContent).Text; !strings.Contains(text, naming) {
		t.Errorf("text = %q, want it to name %s", text, naming)
	}
	if said, ok := result.Meta[refusalKey].(engine.Refusal); !ok || said.Kind != engine.KindUsage {
		t.Errorf("_meta = %#v, want a refusal of kind usage", result.Meta)
	}
}

func TestHandlerFor_emptiesAListNamedAsNull(t *testing.T) {
	path := storeAt(t)
	added, err := engine.Do(engine.TaskNew, path, engine.Input{
		Given:  map[string]string{"title": "From the handler"},
		Listed: map[string][]string{"label": {"spec"}},
	})
	if err != nil {
		t.Fatalf("Do(TaskNew): %v", err)
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

func TestAnswerSchemaFor_describesNothingWhenNothingWasDeclared(t *testing.T) {
	if got := answerSchemaFor(engine.Operation{}); got != nil {
		t.Errorf("answerSchemaFor = %v, want nothing described", got)
	}
}

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

func TestAnswerSchemaFor_describesNothingWhenTheTypeCannotBeRead(t *testing.T) {
	if got := answerSchemaFor(engine.Operation{
		Answers: make(chan int),
	}); got != nil {
		t.Errorf("answerSchemaFor on a shape no schema can describe = %v, want nothing", got)
	}
}

func TestToolFor_saysWhatACallCosts(t *testing.T) {
	for _, c := range []struct {
		op                    engine.Operation
		destructive, repeated bool
	}{
		{engine.TaskList, false, true},
		{engine.TaskNew, false, false},
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

func TestSchemaFor_takesNothingItDoesNotName(t *testing.T) {
	schema := schemaFor(engine.TaskNew)
	if schema["additionalProperties"] != false {
		t.Errorf("additionalProperties = %v, want false", schema["additionalProperties"])
	}
	label := schema["properties"].(map[string]any)["label"].(map[string]any)
	if items, _ := label["items"].(map[string]any); items["type"] != "string" {
		t.Errorf("label items = %v, want strings", label["items"])
	}
}

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

func TestHandlerFor_runsNothingTheClientGaveUpOn(t *testing.T) {
	path := storeAt(t)
	handle := handlerFor(engine.TaskNew, homeFor(t, path))
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
