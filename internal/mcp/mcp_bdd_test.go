package mcp_test

// Executable acceptance scenarios for bdd/mcp.feature. The Gherkin is the
// contract; this file is only the wiring between it and the real server.
//
// It is `package mcp_test` and not `package mcp`, and that is forced rather
// than stylistic: `mcp serve` is a command, so internal/cli imports this
// package, and a scenario file in package mcp that reached the command line
// would close the cycle. An external test package may import both, and still
// counts towards this package's coverage.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/jsonschema-go/jsonschema"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/cli"
	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/test"
)

// run executes the real command line in-process, through the same Main the
// binary runs. A scenario about an agent still has to be able to reach the
// terminal: "what was done at the terminal is there on the next call" is a
// claim about the disk, and nothing but the terminal can put it there.
//
// An empty reader stands in for standard input: no command exercised here
// reads it.
func run(t *testing.T, args ...string) test.Result {
	t.Helper()
	var out, err bytes.Buffer
	code := cli.Main(context.Background(), args, strings.NewReader(""), &out, &err)
	return test.Result{Stdout: out.String(), Stderr: err.String(), Code: code}
}

// runAcceptance runs the scenarios carrying tag from mcp.feature, and fails t
// on any of them.
func runAcceptance(t *testing.T, tag string) {
	t.Helper()
	if test.Suite(t, "mcp.feature", initMCP(t), tag).Run() != 0 {
		t.Fatalf("acceptance %s failed", tag)
	}
}

// world carries what this package's scenarios need beyond the shared world:
// the client session a greeting opened, so the next step can read what it
// answered, and the tools the last "what can it do" asked for.
type world struct {
	*test.World
	session *sdk.ClientSession
	tools   []*sdk.Tool
	// result and callErr are what the last `task_new`/`task_list`/`task_show`
	// call answered, kept apart because a call that fails still has to be
	// told apart from one that answered nothing.
	result  *sdk.CallToolResult
	callErr error
	// calledTool is the tool the last call named, so a step can ask what shape
	// that tool said its answer would take without being told twice.
	calledTool string
	// number is the argument the last `task_show` call was made with, kept so
	// a later step can ask the command line the same question.
	number int
}

// newWorld builds the state one scenario works against, reaching the command
// line through run — the same Main the binary calls.
func newWorld(t *testing.T) *world {
	return &world{World: test.NewWorld(t, run)}
}

// startedServe starts `cairn mcp serve extra...` over a pair of pipes,
// standing in for the standard input and output a real client would speak
// over, and connects an SDK client to it.
func (w *world) startedServe(extra ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// One pipe each way: the server reads what the client wrote, and the
	// client reads what the server wrote.
	clientToServer, writtenByClient := io.Pipe()
	writtenByServer, serverToClient := io.Pipe()
	closeAll := func() {
		_ = clientToServer.Close()
		_ = writtenByClient.Close()
		_ = writtenByServer.Close()
		_ = serverToClient.Close()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		cli.Main(ctx, append([]string{"mcp", "serve"}, extra...), clientToServer, serverToClient, io.Discard)
	}()

	// A pipe's Write blocks until read, and nothing about a context tells it
	// to give up on its own: force it closed once the deadline passes, so a
	// server that never answers fails the step instead of hanging the suite.
	go func() {
		<-ctx.Done()
		closeAll()
	}()

	client := sdk.NewClient(&sdk.Implementation{Name: "acceptance", Version: "0"}, nil)
	session, err := client.Connect(ctx, &sdk.IOTransport{Reader: writtenByServer, Writer: writtenByClient}, nil)
	if err != nil {
		cancel()
		<-done
		return err
	}
	w.session = session

	w.OnEnd(func() {
		_ = session.Close()
		cancel()
		<-done
	})
	return nil
}

// startedServeAndGreetedIt starts `cairn mcp serve` with no extra flags, and
// connects an SDK client to it.
func (w *world) startedServeAndGreetedIt() error {
	return w.startedServe()
}

// aCairnServerStartedWithAccess starts `cairn mcp serve --access <access>`,
// reusing the same machinery a plain greeting does.
func (w *world) aCairnServerStartedWithAccess(access string) error {
	return w.startedServe("--access", access)
}

// aCairnServerStartedWithNoAccessGiven starts `cairn mcp serve` with no
// `--access` flag at all, which is `write`, the default.
func (w *world) aCairnServerStartedWithNoAccessGiven() error {
	return w.startedServe()
}

// iAskItWhatItCanDo lists the tools the started server offers, and keeps
// them so the next step can read what was or was not among them.
func (w *world) iAskItWhatItCanDo() error {
	if w.session == nil {
		return fmt.Errorf("no session: start a server first")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := w.session.ListTools(ctx, nil)
	if err != nil {
		return err
	}
	w.tools = res.Tools
	return nil
}

// callTool calls the tool named name over the session a server was started
// with, keeping what it answered and what went wrong reaching it apart: one
// is the call's answer, the other is why there is none.
func (w *world) callTool(name string, args map[string]any) error {
	if w.session == nil {
		return fmt.Errorf("no session: start a server first")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w.calledTool = name
	w.result, w.callErr = w.session.CallTool(ctx, &sdk.CallToolParams{Name: name, Arguments: args})
	return w.callErr
}

// iAskedItToAddATaskTitled puts a task in the server's store through the
// tool itself, the way a precondition through the command line goes through
// `cairn task new`.
func (w *world) iAskedItToAddATaskTitled(title string) error {
	return w.callTool("task_new", map[string]any{"title": title})
}

// itSaysTheRefusalIs reads the kind and the reason a refusal carries beside
// its sentence, as a client receives them: under _meta, over the wire.
func (w *world) itSaysTheRefusalIs(kind, reason string) error {
	if w.result == nil {
		return fmt.Errorf("nothing came back")
	}
	said, ok := w.result.Meta["cairn/refusal"].(map[string]any)
	if !ok {
		return fmt.Errorf("_meta = %v, want the refusal under cairn/refusal", w.result.Meta)
	}
	if said["kind"] != kind || said["reason"] != reason {
		return fmt.Errorf("refusal = %v, want %s because %s", said, kind, reason)
	}
	return nil
}

// iCallTaskList calls `task_list`, keeping what it answered for the next
// step to read.
// A tool asked for by name and nothing else. One step for every tool that
// takes nothing rather than one step per tool: the name is what varies, and a
// step per tool would be the same three lines written as many times as cairn
// has commands.
func (w *world) iCall(tool string) error {
	_ = w.callTool(tool, nil)
	return nil
}

// iCallTaskAddWithTheTitle calls `task_new` with title, keeping what it
// answered for the next step to read.
func (w *world) iCallTaskAddWithTheTitle(title string) error {
	_ = w.callTool("task_new", map[string]any{"title": title})
	return nil
}

// iCallTaskAddWithTheTitleAndTheLabels calls `task_new` with title and a
// label list carrying both labels, the way a client sends a JSON array
// rather than the string that array would read as when joined.
func (w *world) iCallTaskAddWithTheTitleAndTheLabels(title, first, second string) error {
	_ = w.callTool("task_new", map[string]any{"title": title, "label": []any{first, second}})
	return nil
}

// iCallTaskSearchLookingFor calls `task_search` with a one-element label
// list, keeping what it answered for the next step to read.
func (w *world) iCallTaskSearchLookingFor(term string) error {
	_ = w.callTool("task_search", map[string]any{"named": []any{term}})
	return nil
}

// iCallTaskSearchLookingForNothing calls `task_search` with the list of
// things to look for present and empty — which is not the same call as leaving
// it out, and was answered as if everything had matched.
func (w *world) iCallTaskSearchLookingForNothing() error {
	_ = w.callTool("task_search", map[string]any{"named": []any{}})
	return nil
}

// iCallNamingAnArgumentItDoesNotTake calls tool with a name among its
// arguments it does not recognise — the way a typo such as "numbre" for
// "number" would arrive over the wire.
func (w *world) iCallNamingAnArgumentItDoesNotTake(tool string) error {
	_ = w.callTool(tool, map[string]any{"numbre": 1})
	return nil
}

// iCallTaskShowWithTheNumber calls `task_show` with n, keeping both what it
// answered and n itself, so a later step can ask the command line the same
// question.
// The same, for every tool whose one argument is a number — a task to show, a
// backup to put back, a change to apply. The number is kept so a later step can
// say which one it was about.
// iCallTaskEditWithThePriority changes a task's priority through the tool, so
// what lands in the history was done by an agent and not by the command line.
// schemaOf reads the shape a tool says its answer takes.
func (w *world) schemaOf(name string) (map[string]any, error) {
	for _, t := range w.tools {
		if t.Name != name {
			continue
		}
		if t.OutputSchema == nil {
			return nil, fmt.Errorf("%s says nothing about the shape of its answer", name)
		}
		var m map[string]any
		if err := remarshalTo(t.OutputSchema, &m); err != nil {
			return nil, fmt.Errorf("%s: its shape cannot be read: %w", name, err)
		}
		return m, nil
	}
	return nil, fmt.Errorf("no tool named %s was offered", name)
}

// remarshalTo moves a value through JSON, which is how a schema declared as a
// Go value is read back as the document a client receives.
func remarshalTo(from any, into any) error {
	raw, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, into)
}

// everyToolDescribesItsAnswer fails unless every tool offered says what shape
// its answer takes. Every one, because a client cannot guess for the one that
// stays silent.
func (w *world) everyToolDescribesItsAnswer() error {
	silent := []string{}
	for _, t := range w.tools {
		if _, err := w.schemaOf(t.Name); err != nil {
			silent = append(silent, t.Name)
		}
	}
	if len(silent) > 0 {
		return fmt.Errorf("%d of %d tools say nothing about their answer: %v",
			len(silent), len(w.tools), silent)
	}
	return nil
}

// everyToolAnswersAnObject fails unless every tool says its answer is an
// object, at the top of its shape.
//
// That is not cairn's own rule and it is not the SDK's — the SDK says a schema
// may describe any JSON value. It is what a client enforces, and a client that
// rejects one tool's shape rejects the whole list and offers none of them.
// Measured, at the cost of a regression: ten tools describing a list took the
// server from twenty-six tools to zero, and nothing in this suite noticed,
// because every check here asks cairn about cairn. This one asks what a client
// will accept.
func (w *world) everyToolAnswersAnObject() error {
	refused := []string{}
	for _, t := range w.tools {
		shape, err := w.schemaOf(t.Name)
		if err != nil {
			return err
		}
		if shape["type"] != "object" {
			refused = append(refused, fmt.Sprintf("%s says %v", t.Name, shape["type"]))
		}
	}
	if len(refused) > 0 {
		return fmt.Errorf("%d of %d tools describe something a client will refuse: %v",
			len(refused), len(w.tools), refused)
	}
	return nil
}

// saysItsAnswerCarries fails unless the named tool's shape names the given
// field. What a tool answers is an object, so what it carries is a property of
// it — and naming one is how a scenario says which object it is without
// spelling out every field.
func (w *world) saysItsAnswerCarries(name, field string) error {
	shape, err := w.schemaOf(name)
	if err != nil {
		return err
	}
	properties, ok := shape["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("%s names no fields at all: %v", name, shape)
	}
	if _, carried := properties[field]; !carried {
		named := make([]string, 0, len(properties))
		for k := range properties {
			named = append(named, k)
		}
		sort.Strings(named)
		return fmt.Errorf("%s carries %v, and not %q", name, named, field)
	}
	return nil
}

// theAnswerHoldsToItsShape measures what a tool answered against the shape it
// said it would take.
//
// The shape does the judging, rather than a step reading it field by field: a
// step that checked the shape by hand would be a second reading of it, and the
// two would eventually disagree.
func (w *world) theAnswerHoldsToItsShape() error {
	if w.result == nil {
		return fmt.Errorf("nothing came back to measure")
	}
	if w.result.StructuredContent == nil {
		return fmt.Errorf("the answer carries nothing for a machine")
	}
	name := w.calledTool
	// Asked for here rather than left to a step before it: what this measures is
	// the answer against the description, and a scenario about the answer should
	// not have to remember to ask for the descriptions first.
	if len(w.tools) == 0 {
		if err := w.iAskItWhatItCanDo(); err != nil {
			return err
		}
	}
	var declared *jsonschema.Schema
	for _, t := range w.tools {
		if t.Name == name {
			if err := remarshalTo(t.OutputSchema, &declared); err != nil {
				return fmt.Errorf("%s: its shape cannot be read: %w", name, err)
			}
		}
	}
	if declared == nil {
		return fmt.Errorf("%s says nothing about the shape of its answer", name)
	}
	resolved, err := declared.Resolve(nil)
	if err != nil {
		return fmt.Errorf("%s: its shape is not a usable schema: %w", name, err)
	}
	var crossed any
	if err := remarshalTo(w.result.StructuredContent, &crossed); err != nil {
		return err
	}
	if err := resolved.Validate(crossed); err != nil {
		return fmt.Errorf("%s answered something its own shape refuses: %w", name, err)
	}
	return nil
}

// itCarriesNothingToMeasure fails unless a refusal came back with nothing for a
// machine: a refusal is not an answer of the shape a tool describes, and
// measuring it against one would make every refusal look like a violation.
func (w *world) itCarriesNothingToMeasure() error {
	if w.result == nil {
		return fmt.Errorf("nothing came back")
	}
	if w.result.StructuredContent != nil {
		return fmt.Errorf("a refusal carries %v for a machine to measure",
			w.result.StructuredContent)
	}
	return nil
}

func (w *world) iCallTaskEditWithThePriority(n int, priority string) error {
	w.number = n
	_ = w.callTool("task_edit", map[string]any{"number": n, "priority": priority})
	return nil
}

// theHistorySaysItCameThroughTheServer asks the command line what happened to a
// task and looks for the surface on the line. Asked through the command line on
// purpose: what an agent did has to be visible to the person reading, which is
// the whole point of recording it.
func (w *world) theHistorySaysItCameThroughTheServer(n int) error {
	w.Run(fmt.Sprintf("cairn trace show %d", n))
	return test.WantContains(w.Out.Stdout, "(mcp)", "the history")
}

func (w *world) iCallWithTheNumber(tool string, n int) error {
	w.number = n
	_ = w.callTool(tool, map[string]any{"number": n})
	return nil
}

// iAddedATaskTitledWith puts a task in the store through the real
// command, carrying a deadline or a description — a precondition set through the command line
// rather than the tool under test, so a scenario that later empties the
// deadline through the tool proves something the tool did, not something the
// precondition already did.
func (w *world) iAddedATaskTitledWith(title, what, value string) error {
	flag := map[string]string{"deadline": "--deadline", "description": "--desc"}[what]
	return w.Seed(fmt.Sprintf(`cairn task new "%s" %s "%s"`, title, flag, value))
}

// iCallTaskEditWithTheNumberAndTheDeadlineAsNothing calls `task_edit`,
// naming the deadline as JSON null rather than leaving it out: only null
// asks for it to be emptied, and only sending the literal value proves the
// tool tells the two apart.
func (w *world) iCallTaskEditWithTheNumberAndTheDeadlineAsNothing(n int) error {
	_ = w.callTool("task_edit", map[string]any{"number": n, "deadline": nil})
	return nil
}

// itAnswersThatItFailed fails unless the last call reached cairn and came
// back marked as a failure, rather than failing before cairn ever saw it.
func (w *world) itAnswersThatItFailed() error {
	if w.callErr != nil {
		return fmt.Errorf("the call itself failed: %w, want an answer marked as a failure", w.callErr)
	}
	if w.result == nil || !w.result.IsError {
		return fmt.Errorf("the call answered without being marked as a failure")
	}
	return nil
}

// theAnswerSays fails unless the last call's text content mentions what.
// theAnswerSays reads a scenario's words as what they mean, not as a substring
// to go looking for. Most of them are the reason itself and are found as
// written; a few name a thing the reason has to carry rather than quoting it,
// and those are translated here — otherwise a scenario's prose would dictate
// the wording of cairn's own refusals, which is the tail wagging the dog.
func (w *world) theAnswerSays(what string) error {
	switch what {
	case "what it needed":
		what = "named"
	case "which name it could not read":
		what = "numbre"
	}
	return test.WantContains(textOf(w.result), what, "the answer")
}

// theCallItselfFails fails unless reaching the tool failed before cairn ever
// saw the call — a protocol error, not an answer cairn gave.
func (w *world) theCallItselfFails() error {
	if w.callErr == nil {
		return fmt.Errorf("the call succeeded, want it to fail before cairn ever saw it")
	}
	return nil
}

// theTaskHasNoDeadline fails unless the last call's structured content names
// a task with an empty deadline.
func (w *world) theTaskHasNoDeadline() error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	doc, ok := w.result.StructuredContent.(map[string]any)
	if !ok {
		return fmt.Errorf("structured content is not a task: %#v", w.result.StructuredContent)
	}
	if doc["deadline"] != "" {
		return fmt.Errorf("deadline = %v, want none", doc["deadline"])
	}
	return nil
}

// textOf reads the card a call answered with off its content, the way a
// person reading it would.
func textOf(res *sdk.CallToolResult) string {
	if res == nil || len(res.Content) == 0 {
		return ""
	}
	if tc, ok := res.Content[0].(*sdk.TextContent); ok {
		return tc.Text
	}
	return ""
}

// itAnswersWithNTasks fails unless the last call answered with exactly n
// tasks, read off its structured content rather than counted in a string.
func (w *world) itAnswersWithNTasks(n int) error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	tasks, err := carriedUnder(w.result.StructuredContent, "tasks")
	if err != nil {
		return err
	}
	if len(tasks) != n {
		return fmt.Errorf("%d tasks, want %d", len(tasks), n)
	}
	return nil
}

// noRowCarriesADescription is what keeps a listing within what an agent can
// read: measured on a real board, 105 tasks came to 106 000 characters, past
// what a client hands an agent in one answer.
func (w *world) noRowCarriesADescription() error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	tasks, err := carriedUnder(w.result.StructuredContent, "tasks")
	if err != nil {
		return err
	}
	for _, one := range tasks {
		if row, _ := one.(map[string]any); row != nil {
			if _, carried := row["description"]; carried {
				return fmt.Errorf("task %v carries its description in the listing", row["number"])
			}
		}
	}
	return nil
}

// carriedUnder reads the rows of a listing out of the object naming them. A
// listing crosses as an object rather than a bare list, so nothing reads the
// structured content of one as a list any more.
func carriedUnder(content any, named string) ([]any, error) {
	listing, ok := content.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("structured content is not a listing: %#v", content)
	}
	rows, ok := listing[named].([]any)
	if !ok {
		return nil, fmt.Errorf("the listing carries no %q: %#v", named, listing)
	}
	return rows, nil
}

// itAnswersWithoutAnError fails unless the last call reported nothing wrong.
func (w *world) itAnswersWithoutAnError() error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	return nil
}

// itAnswersWithTheTaskTitled fails unless the last call's structured content
// names a task titled title.
func (w *world) itAnswersWithTheTaskTitled(title string) error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	doc, ok := w.result.StructuredContent.(map[string]any)
	if !ok {
		return fmt.Errorf("structured content is not a task: %#v", w.result.StructuredContent)
	}
	if doc["title"] != title {
		return fmt.Errorf("title = %v, want %q", doc["title"], title)
	}
	return nil
}

// itAnswersWithATaskCarryingBothLabels fails unless the last call's
// structured content names a task whose labels are exactly "spec" and "team".
func (w *world) itAnswersWithATaskCarryingBothLabels() error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	doc, ok := w.result.StructuredContent.(map[string]any)
	if !ok {
		return fmt.Errorf("structured content is not a task: %#v", w.result.StructuredContent)
	}
	labels, ok := doc["labels"].([]any)
	if !ok {
		return fmt.Errorf("labels = %#v, want a list of two", doc["labels"])
	}
	want := map[string]bool{"spec": true, "team": true}
	if len(labels) != len(want) {
		return fmt.Errorf("labels = %v, want %v", labels, want)
	}
	for _, l := range labels {
		s, ok := l.(string)
		if !ok || !want[s] {
			return fmt.Errorf("labels = %v, want %v", labels, want)
		}
	}
	return nil
}

// itAnswersWithTheCardAPersonWouldHaveRead fails unless the last call's
// content is exactly what `cairn task show <number>` prints.
func (w *world) itAnswersWithTheCardAPersonWouldHaveRead() error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	w.Run(fmt.Sprintf("cairn task show %d", w.number))
	want := w.Out.Stdout
	if got := textOf(w.result); got != want {
		return fmt.Errorf("content = %q, want %q (what the terminal printed)", got, want)
	}
	return nil
}

// itAnswersWithTheShapeJSONWouldHavePrinted fails unless the last call's
// structured content is exactly what `cairn task show <number> --json`
// prints.
func (w *world) itAnswersWithTheShapeJSONWouldHavePrinted() error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	w.Run(fmt.Sprintf("cairn task show %d --json", w.number))
	var want any
	if err := json.Unmarshal([]byte(w.Out.Stdout), &want); err != nil {
		return err
	}
	if got := w.result.StructuredContent; !reflect.DeepEqual(got, want) {
		return fmt.Errorf("structured content = %#v, want %#v (what --json printed)", got, want)
	}
	return nil
}

// schemaProperty reads one property out of tool's input schema, the way a
// client would read it off the wire rather than off the map schemaFor built.
func schemaProperty(tool *sdk.Tool, name string) (map[string]any, error) {
	if tool == nil {
		return nil, fmt.Errorf("tool was not offered")
	}
	schema, ok := tool.InputSchema.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("input schema is not an object: %#v", tool.InputSchema)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no properties: %#v", schema)
	}
	prop, ok := properties[name].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema names no %q: %#v", name, properties)
	}
	return prop, nil
}

// taskAddNamesEveryStateCairnAccepts fails unless task_new's "state" argument
// enumerates exactly the states domain.States holds, read off the schema
// rather than a literal list, so the scenario cannot drift from what cairn
// accepts.
func (w *world) taskAddNamesEveryStateCairnAccepts() error {
	prop, err := schemaProperty(w.tool("task_new"), "state")
	if err != nil {
		return err
	}
	enum, ok := prop["enum"].([]any)
	if !ok {
		return fmt.Errorf("%q names no enum: %#v", "state", prop)
	}
	if len(enum) != len(domain.States) {
		return fmt.Errorf("enum = %v, want the %d states cairn accepts", enum, len(domain.States))
	}
	for i, s := range domain.States {
		if enum[i] != string(s) {
			return fmt.Errorf("enum = %v, want %v", enum, domain.States)
		}
	}
	// "title" has no closed set of its own: an enum named here regardless
	// would forbid every value JSON schema's empty enum forbids, rather than
	// leaving it open the way a plain string is meant to be.
	title, err := schemaProperty(w.tool("task_new"), "title")
	if err != nil {
		return err
	}
	if _, has := title["enum"]; has {
		return fmt.Errorf(`"title" names an enum: %#v, want none — it has no closed set`, title)
	}
	return nil
}

// itSaysWhatEachThingItTakesIsFor fails unless every argument task_new's
// schema names carries a non-empty description.
func (w *world) itSaysWhatEachThingItTakesIsFor() error {
	tool := w.tool("task_new")
	if tool == nil {
		return fmt.Errorf("tool %q was not offered", "task_new")
	}
	schema, ok := tool.InputSchema.(map[string]any)
	if !ok {
		return fmt.Errorf("input schema is not an object: %#v", tool.InputSchema)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("schema has no properties: %#v", schema)
	}
	for name, raw := range properties {
		prop, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("%q is not an object: %#v", name, raw)
		}
		if means, ok := prop["description"].(string); !ok || means == "" {
			return fmt.Errorf("%q says nothing of what it is for: %#v", name, prop)
		}
	}
	return nil
}

// tool finds the tool named name among what the last "ask it what it can do"
// step read, or nil when it was not among them.
func (w *world) tool(name string) *sdk.Tool {
	for _, t := range w.tools {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// itOffers fails unless the last tool listing named this tool.
func (w *world) itOffers(name string) error {
	if w.tool(name) == nil {
		return fmt.Errorf("tool %q was not offered", name)
	}
	return nil
}

// itDoesNotOffer fails unless the last tool listing left this tool out.
func (w *world) itDoesNotOffer(name string) error {
	if w.tool(name) != nil {
		return fmt.Errorf("tool %q was offered, want it left out", name)
	}
	return nil
}

// itOffersAToolForEveryCommand fails unless every command cairn has, named by
// engine.All(), was among the last tool listing.
func (w *world) itOffersAToolForEveryCommand() error {
	for _, op := range engine.All() {
		if w.tool(op.Name()) == nil {
			return fmt.Errorf("no tool offered for command %q", op.Line())
		}
	}
	return nil
}

// eachOneIsNamedForTheCommandItRuns fails unless every tool the last listing
// held is named after a real command, rather than under a name of its own.
func (w *world) eachOneIsNamedForTheCommandItRuns() error {
	names := map[string]bool{}
	for _, op := range engine.All() {
		names[op.Name()] = true
	}
	for _, t := range w.tools {
		if !names[t.Name] {
			return fmt.Errorf("tool %q is not named for any command cairn has", t.Name)
		}
	}
	return nil
}

// toolIsMarkedAsLosingWork fails unless name was offered and its
// destructiveHint says calling it can lose work already done.
func (w *world) toolIsMarkedAsLosingWork(name string) error {
	tool := w.tool(name)
	if tool == nil {
		return fmt.Errorf("tool %q was not offered", name)
	}
	if tool.Annotations == nil || tool.Annotations.DestructiveHint == nil || !*tool.Annotations.DestructiveHint {
		return fmt.Errorf("tool %q is not marked as losing work", name)
	}
	return nil
}

// toolIsMarkedAsReadOnly fails unless name was offered and its readOnlyHint
// says calling it only looks.
func (w *world) toolIsMarkedAsReadOnly(name string) error {
	tool := w.tool(name)
	if tool == nil {
		return fmt.Errorf("tool %q was not offered", name)
	}
	if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
		return fmt.Errorf("tool %q is not marked as read-only", name)
	}
	return nil
}

// everyToolIsMarkedAsReachingNothingOutsideCairn fails unless every tool the
// last listing held says, through openWorldHint, that it reaches nothing
// outside cairn.
func (w *world) everyToolIsMarkedAsReachingNothingOutsideCairn() error {
	for _, t := range w.tools {
		if t.Annotations == nil || t.Annotations.OpenWorldHint == nil || *t.Annotations.OpenWorldHint {
			return fmt.Errorf("tool %q is not marked as reaching nothing outside cairn", t.Name)
		}
	}
	return nil
}

// startedServeAndGreetedItWithoutWaiting starts `cairn mcp serve`, writes one
// greeting to its standard input and closes it before anything reads an
// answer. It does not use the SDK client: Connect blocks until the greeting
// is answered, and the whole point of the scenario is a client that never
// waits for one. cli.Main is run to completion and what it answered is kept
// in w.Out, the way every other step in this package reads it.
//
// The context bounds the call so a server that never returns cannot hang the
// suite; nothing here starts a goroutine that could outlive the scenario.
func (w *world) startedServeAndGreetedItWithoutWaiting() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	greeting := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"acceptance","version":"0"}}}`

	var out, errOut bytes.Buffer
	code := cli.Main(ctx, []string{"mcp", "serve"}, strings.NewReader(greeting+"\n"), &out, &errOut)
	w.Out = test.Result{Stdout: out.String(), Stderr: errOut.String(), Code: code}
	return nil
}

// itStopsAndSaysNothingWentWrong fails unless the last command line exited
// clean and left standard error empty: a reason printed on a successful exit
// is still a reason somebody will read and worry about.
func (w *world) itStopsAndSaysNothingWentWrong() error {
	if w.Out.Code != 0 {
		return fmt.Errorf("exit %d, want 0", w.Out.Code)
	}
	if w.Out.Stderr != "" {
		return fmt.Errorf("stderr = %q, want empty", w.Out.Stderr)
	}
	return nil
}

// versionField reads the version off the first line `cairn --version`
// prints: "cairn <version>".
func versionField(t *testing.T) string {
	t.Helper()
	first, _, _ := strings.Cut(run(t, "--version").Stdout, "\n")
	_, version, _ := strings.Cut(first, " ")
	return version
}

// itAnswersWithItsNameAndVersion fails unless the greeted session reports the
// name "cairn" and the version this very build reports elsewhere.
func (w *world) itAnswersWithItsNameAndVersion() error {
	if w.session == nil {
		return fmt.Errorf("no session: greet it first")
	}
	info := w.session.InitializeResult().ServerInfo
	if info.Name != "cairn" {
		return fmt.Errorf("name = %q, want %q", info.Name, "cairn")
	}
	if want := versionField(w.T); info.Version != want {
		return fmt.Errorf("version = %q, want %q, the one this build reports", info.Version, want)
	}
	return nil
}

func initMCP(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.Register(sc)
		sc.Step("^I start `cairn mcp serve` and greet it$", w.startedServeAndGreetedIt)
		sc.Step("^it answers with its name and its version$", w.itAnswersWithItsNameAndVersion)
		sc.Step("^I start `cairn mcp serve` and greet it without waiting$", w.startedServeAndGreetedItWithoutWaiting)
		sc.Step("^it stops, and says nothing went wrong$", w.itStopsAndSaysNothingWentWrong)
		sc.Step("^a cairn server started with `--access (read|write|full)`$", w.aCairnServerStartedWithAccess)
		sc.Step("^a cairn server started with no access given$", w.aCairnServerStartedWithNoAccessGiven)
		sc.Step("^I ask it what it can do$", w.iAskItWhatItCanDo)
		sc.Step("^every tool describes the shape of its answer$", w.everyToolDescribesItsAnswer)
		sc.Step("^every tool says it answers one object$", w.everyToolAnswersAnObject)
		sc.Step("^`([a-z_]+)` says its answer carries the ([a-z]+)$", w.saysItsAnswerCarries)
		sc.Step("^`([a-z_]+)` says its answer carries a ([a-z]+)$", w.saysItsAnswerCarries)
		sc.Step("^its answer holds to the shape the tool described$", w.theAnswerHoldsToItsShape)
		sc.Step("^it carries nothing for a machine to measure$", w.itCarriesNothingToMeasure)
		sc.Step("^it offers `([^`]*)`$", w.itOffers)
		sc.Step("^it does not offer `([^`]*)`$", w.itDoesNotOffer)
		sc.Step("^it offers a tool for every command cairn has$", w.itOffersAToolForEveryCommand)
		sc.Step("^each one is named for the command it runs$", w.eachOneIsNamedForTheCommandItRuns)
		sc.Step("^`([^`]*)` is marked as losing work$", w.toolIsMarkedAsLosingWork)
		sc.Step("^`([^`]*)` is marked as read-only$", w.toolIsMarkedAsReadOnly)
		sc.Step("^every tool is marked as reaching nothing outside cairn$", w.everyToolIsMarkedAsReachingNothingOutsideCairn)
		sc.Step(`^I asked it to add a task titled "([^"]*)"$`, w.iAskedItToAddATaskTitled)
		sc.Step("^I call `([a-z_]+)`$", w.iCall)
		sc.Step("^I call `([a-z_]+)` with nothing at all$", w.iCall)
		sc.Step("^I call `([a-z_]+)` naming an argument it does not take$", w.iCallNamingAnArgumentItDoesNotTake)
		sc.Step("^I call `task_new` with the title \"([^\"]*)\"$", w.iCallTaskAddWithTheTitle)
		sc.Step("^I call `task_new` with the title \"([^\"]*)\" and the labels \"([^\"]*)\" and \"([^\"]*)\"$",
			w.iCallTaskAddWithTheTitleAndTheLabels)
		sc.Step("^I call `task_search` looking for \"([^\"]*)\"$", w.iCallTaskSearchLookingFor)
		sc.Step("^I call `task_search` looking for nothing$", w.iCallTaskSearchLookingForNothing)
		sc.Step("^it answers with a task carrying both labels$", w.itAnswersWithATaskCarryingBothLabels)
		sc.Step("^I call `([a-z_]+)` with the number (\\d+)$", w.iCallWithTheNumber)
		sc.Step(`^it answers with (\d+) tasks?$`, w.itAnswersWithNTasks)
		sc.Step(`^no row of it carries a description$`, w.noRowCarriesADescription)
		sc.Step(`^it answers without an error$`, w.itAnswersWithoutAnError)
		sc.Step(`^it answers with the task titled "([^"]*)"$`, w.itAnswersWithTheTaskTitled)
		sc.Step(`^it answers with the card a person would have read$`, w.itAnswersWithTheCardAPersonWouldHaveRead)
		sc.Step("^it answers with the shape `--json` would have printed$", w.itAnswersWithTheShapeJSONWouldHavePrinted)
		sc.Step("^`task_new` names every state cairn accepts$", w.taskAddNamesEveryStateCairnAccepts)
		sc.Step("^it says what each thing it takes is for$", w.itSaysWhatEachThingItTakesIsFor)
		sc.Step(`^I added a task titled "([^"]*)" with the (deadline|description) "([^"]*)"$`, w.iAddedATaskTitledWith)
		sc.Step("^I call `task_edit` with the number (\\d+) and the priority \"([^\"]*)\"$",
			w.iCallTaskEditWithThePriority)
		sc.Step("^the trace of task (\\d+) says the change came through the server$",
			w.theHistorySaysItCameThroughTheServer)
		sc.Step("^I call `task_edit` with the number (\\d+) and the deadline as nothing$",
			w.iCallTaskEditWithTheNumberAndTheDeadlineAsNothing)
		sc.Step(`^the task has no deadline$`, w.theTaskHasNoDeadline)
		sc.Step(`^it is still titled "([^"]*)"$`, w.itAnswersWithTheTaskTitled)
		sc.Step(`^it answers that it failed$`, w.itAnswersThatItFailed)
		sc.Step(`^it says the refusal is "([^"]*)", because "([^"]*)"$`, w.itSaysTheRefusalIs)
		sc.Step(`^the answer says (.+)$`, w.theAnswerSays)
		sc.Step(`^the call itself fails$`, w.theCallItselfFails)
		sc.Step(`^a store from an older cairn, holding a task titled "([^"]*)"$`,
			w.AStoreFromAnOlderCairn)
	}
}

// scenariosOf reads the tag of every scenario in a feature file.
//
// Read rather than listed, and that is the whole point. This package used to
// carry one Go test per tag, and a scenario added to the feature without its
// test did not run at all — measured: the suite stayed green while the scenario
// never executed, and it took refusing to believe an impossible pass to notice.
// Taking the tags from the feature means a scenario added there runs, named
// after itself, with nothing to keep in step.
func scenariosOf(t *testing.T, feature string) []string {
	t.Helper()
	read, err := os.ReadFile(filepath.Join("..", "..", "bdd", feature))
	if err != nil {
		t.Fatalf("reading %s: %v", feature, err)
	}
	tags := []string{}
	for _, line := range strings.Split(string(read), "\n") {
		if tag := strings.TrimSpace(line); strings.HasPrefix(tag, "@") {
			tags = append(tags, tag)
		}
	}
	if len(tags) == 0 {
		t.Fatalf("%s names no scenario", feature)
	}
	return tags
}

// TestAcceptance_MCP runs every scenario of bdd/mcp.feature, each as a subtest
// named after it, so a failure says which scenario failed without a Go function
// per scenario to forget to write.
func TestAcceptance_MCP(t *testing.T) {
	for _, tag := range scenariosOf(t, "mcp.feature") {
		t.Run(strings.TrimPrefix(tag, "@"), func(t *testing.T) {
			runAcceptance(t, tag)
		})
	}
}
