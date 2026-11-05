package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
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

func run(t *testing.T, args ...string) test.Result {
	t.Helper()
	var out, err bytes.Buffer
	code := cli.Main(context.Background(), args, strings.NewReader(""), &out, &err)
	return test.Result{Stdout: out.String(), Stderr: err.String(), Code: code}
}

func runAcceptance(t *testing.T, tag string) {
	t.Helper()
	if test.Suite(t, "mcp.feature", initMCP(t), tag).Run() != 0 {
		t.Fatalf("acceptance %s failed", tag)
	}
}

type world struct {
	*test.World
	session    *sdk.ClientSession
	tools      []*sdk.Tool
	result     *sdk.CallToolResult
	callErr    error
	calledTool string
	number     int
	onDisk     string
	told       chan struct{}
}

func newWorld(t *testing.T) *world {
	return &world{World: test.NewWorld(t, run)}
}

func (w *world) startedServe(extra ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

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

func (w *world) startedServeAndGreetedIt() error {
	return w.startedServe()
}

func (w *world) aCairnServerStartedWithAccess(access string) error {
	return w.startedServe("--access", access)
}

func (w *world) aCairnServerStartedWithNoAccessGiven() error {
	return w.startedServe()
}

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

func (w *world) iAskedItToAddATaskTitled(title string) error {
	return w.callTool("task_new", map[string]any{"title": title})
}

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

func (w *world) iCall(tool string) error {
	_ = w.callTool(tool, nil)
	return nil
}

func (w *world) iCallTaskAddWithTheTitle(title string) error {
	_ = w.callTool("task_new", map[string]any{"title": title})
	return nil
}

func (w *world) iCallTaskAddWithTheTitleAndTheLabels(title, first, second string) error {
	_ = w.callTool("task_new", map[string]any{"title": title, "label": []any{first, second}})
	return nil
}

func (w *world) iCallTaskAddWithTheTitleAndTheLink(title string, number int) error {
	_ = w.callTool("task_new", map[string]any{"title": title, "link": []any{strconv.Itoa(number)}})
	return nil
}

func (w *world) iCallTaskSearchLookingFor(term string) error {
	_ = w.callTool("task_search", map[string]any{"named": []any{term}})
	return nil
}

func (w *world) iCallTaskSearchLookingForNothing() error {
	_ = w.callTool("task_search", map[string]any{"named": []any{}})
	return nil
}

func (w *world) iCallNamingAnArgumentItDoesNotTake(tool string) error {
	_ = w.callTool(tool, map[string]any{"numbre": 1})
	return nil
}

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

func remarshalTo(from any, into any) error {
	raw, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, into)
}

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

func (w *world) theAnswerHoldsToItsShape() error {
	if w.result == nil {
		return fmt.Errorf("nothing came back to measure")
	}
	if w.result.StructuredContent == nil {
		return fmt.Errorf("the answer carries nothing for a machine")
	}
	name := w.calledTool
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

func (w *world) theHistorySaysItCameThroughTheServer(n int) error {
	w.Run(fmt.Sprintf("cairn trace show %d", n))
	return test.WantContains(w.Out.Stdout, "through mcp)", "the history")
}

func (w *world) iCallWithTheNumber(tool string, n int) error {
	w.number = n
	_ = w.callTool(tool, map[string]any{"number": n})
	return nil
}

func (w *world) iAddedATaskTitledWith(title, what, value string) error {
	flag := map[string]string{"deadline": "--deadline", "description": "--desc", "state": "--state"}[what]
	return w.Seed(fmt.Sprintf(`cairn task new "%s" %s "%s"`, title, flag, value))
}

func (w *world) iCallTaskEditWithTheNumberAndTheDeadlineAsNothing(n int) error {
	_ = w.callTool("task_edit", map[string]any{"number": n, "deadline": nil})
	return nil
}

func (w *world) itAnswersThatItFailed() error {
	if w.callErr != nil {
		return fmt.Errorf("the call itself failed: %w, want an answer marked as a failure", w.callErr)
	}
	if w.result == nil || !w.result.IsError {
		return fmt.Errorf("the call answered without being marked as a failure")
	}
	return nil
}

func (w *world) theAnswerSays(what string) error {
	switch what {
	case "what it needed":
		what = "named"
	case "which name it could not read":
		what = "numbre"
	}
	return test.WantContains(textOf(w.result), what, "the answer")
}

func (w *world) theCallItselfFails() error {
	if w.callErr == nil {
		return fmt.Errorf("the call succeeded, want it to fail before cairn ever saw it")
	}
	return nil
}

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

func textOf(res *sdk.CallToolResult) string {
	if res == nil {
		return ""
	}
	var said strings.Builder
	for _, one := range res.Content {
		if tc, ok := one.(*sdk.TextContent); ok {
			said.WriteString(tc.Text)
		}
	}
	return said.String()
}

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

func (w *world) iCallAskingForTheFields(tool, first, second, third string) error {
	_ = w.callTool(tool, map[string]any{"fields": []any{first, second, third}})
	return nil
}

func (w *world) iCallTaskSearchLookingForAskingForTheFields(term, first, second, third string) error {
	_ = w.callTool("task_search", map[string]any{"named": []any{term}, "fields": []any{first, second, third}})
	return nil
}

func (w *world) iCallTaskSearchLookingForAnyOf(first, second, third, fourth string) error {
	_ = w.callTool("task_search", map[string]any{"named": []any{first, second, third, fourth}, "any": true})
	return nil
}

func (w *world) iCallTaskSearchLookingForAnyOfAskingForTheFields(term, other, first, second, third string) error {
	_ = w.callTool("task_search", map[string]any{"named": []any{term, other}, "any": true, "fields": []any{first, second, third}})
	return nil
}

func (w *world) whatEachTermMatched() (map[string]any, error) {
	if w.callErr != nil {
		return nil, fmt.Errorf("the call failed: %w", w.callErr)
	}
	listing, ok := w.result.StructuredContent.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("structured content is not a listing: %#v", w.result.StructuredContent)
	}
	matched, _ := listing["matched"].(map[string]any)
	return matched, nil
}

func (w *world) itSaysTheTermMatched(term, which string) error {
	matched, err := w.whatEachTermMatched()
	if err != nil {
		return err
	}
	numbers, said := matched[term].([]any)
	if !said {
		return fmt.Errorf("the answer does not say what %q matched: %v", term, matched)
	}
	want := []any{}
	if which != "no task" {
		n, _ := strconv.Atoi(strings.TrimPrefix(which, "task "))
		want = []any{float64(n)}
	}
	if !slices.Equal(numbers, want) {
		return fmt.Errorf("%q matched %v, want %v", term, numbers, want)
	}
	return nil
}

func (w *world) itSaysNothingOfWhichTermMatched() error {
	matched, err := w.whatEachTermMatched()
	if err != nil {
		return err
	}
	if matched != nil {
		return fmt.Errorf("the answer says what each term matched: %v", matched)
	}
	return nil
}

func (w *world) iCallTaskListAskingForTheNumbers(numbers ...string) error {
	_ = w.callTool("task_list", map[string]any{"number": numbers})
	return nil
}

func (w *world) iCallTaskListAskingForTheNumbersArchivedOnly(first, second, third string) error {
	_ = w.callTool("task_list", map[string]any{"number": []string{first, second, third}, "archived": true})
	return nil
}

func (w *world) iCallTaskListAskingForTheNumbersAndTheFields(first, second, one, two, three string) error {
	_ = w.callTool("task_list", map[string]any{"number": []string{first, second}, "fields": []string{one, two, three}})
	return nil
}

func (w *world) iCallTaskListAskingForTheNumbersWrittenAsNumbers(first, second, third int) error {
	_ = w.callTool("task_list", map[string]any{"number": []int{first, second, third}})
	return nil
}

func (w *world) iCallTaskListAskingForTheNumberAndWhatWasSent(first int, sent string) error {
	_ = w.callTool("task_list", map[string]any{"number": json.RawMessage(fmt.Sprintf("[%d,%s]", first, sent))})
	return nil
}

func (w *world) iCallTaskNewBlockedAndLinkedWrittenAsNumbers(title string, blocker, link int) error {
	_ = w.callTool("task_new", map[string]any{"title": title, "blocker": []int{blocker}, "link": []int{link}})
	return nil
}

func (w *world) itAnswersWithATaskBlockedAndLinked(blocker, link int) error {
	doc, err := w.answeredTask()
	if err != nil {
		return err
	}
	if w.result.IsError {
		return fmt.Errorf("it was refused: %s", textOf(w.result))
	}
	blockedBy, _ := doc["blocked_by"].([]any)
	linked, _ := doc["linked"].([]any)
	if !slices.Equal(blockedBy, []any{float64(blocker)}) || !slices.Equal(linked, []any{float64(link)}) {
		return fmt.Errorf("blocked by %v and linked to %v, want %d and %d", doc["blocked_by"], doc["linked"], blocker, link)
	}
	return nil
}

func (w *world) iCallTaskArchiveWithTheNumberAndAlsoWrittenAsANumber(first, also int) error {
	_ = w.callTool("task_archive", map[string]any{"number": first, "also": []int{also}})
	return nil
}

func (w *world) iCallTaskEditWithTheLabelWrittenAsANumber(number, label int) error {
	_ = w.callTool("task_edit", map[string]any{"number": number, "label": []int{label}})
	return nil
}

func (w *world) everyListOfNumbersSaysItTakesWholeNumbers() error {
	seen := 0
	for _, op := range engine.All() {
		tool := w.tool(op.Name())
		if tool == nil {
			continue
		}
		var input struct {
			Properties map[string]struct {
				Items struct {
					Type any `json:"type"`
				} `json:"items"`
			} `json:"properties"`
		}
		if err := remarshalTo(tool.InputSchema, &input); err != nil {
			return fmt.Errorf("%s: %w", tool.Name, err)
		}
		for _, arg := range op.Takes {
			if arg.Kind != engine.List || arg.Of == "" {
				continue
			}
			seen++
			said := input.Properties[arg.Name].Items.Type
			if types, _ := said.([]any); !slices.Contains(types, any("integer")) {
				return fmt.Errorf("%s's %q takes %v, want whole numbers among them", tool.Name, arg.Name, said)
			}
		}
	}
	if seen == 0 {
		return fmt.Errorf("no tool takes a list of numbers")
	}
	return nil
}

func (w *world) itAnswersWithTheTasks(first, second string) error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	rows, err := carriedUnder(w.result.StructuredContent, "tasks")
	if err != nil {
		return err
	}
	want := []string{first}
	if second != "" {
		want = append(want, second)
	}
	got := make([]string, 0, len(rows))
	for _, one := range rows {
		row, _ := one.(map[string]any)
		got = append(got, fmt.Sprint(row["number"]))
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("it answered the tasks %v, want %v", got, want)
	}
	return nil
}

func (w *world) listingSaid() (map[string]any, error) {
	if w.callErr != nil {
		return nil, fmt.Errorf("the call failed: %w", w.callErr)
	}
	listing, ok := w.result.StructuredContent.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("structured content is not a listing: %#v", w.result.StructuredContent)
	}
	return listing, nil
}

func (w *world) noTaskSaid() ([]any, error) {
	listing, err := w.listingSaid()
	if err != nil {
		return nil, err
	}
	unknown, _ := listing["unknown"].([]any)
	return unknown, nil
}

func (w *world) iCallTaskListAskingForAPage(limit, offset int) error {
	_ = w.callTool("task_list", map[string]any{"limit": limit, "offset": offset})
	return nil
}

func (w *world) iCallTaskListInAskingForAPage(state string, limit, offset int) error {
	_ = w.callTool("task_list", map[string]any{"state": state, "limit": limit, "offset": offset})
	return nil
}

func (w *world) itSaysHowManyInAll(n int) error {
	listing, err := w.listingSaid()
	if err != nil {
		return err
	}
	if listing["total"] != float64(n) {
		return fmt.Errorf("it says %v in all, want %d", listing["total"], n)
	}
	return nil
}

func (w *world) itSaysNothingOfHowManyInAll() error {
	listing, err := w.listingSaid()
	if err != nil {
		return err
	}
	if total, said := listing["total"]; said {
		return fmt.Errorf("it says %v in all, want nothing said", total)
	}
	return nil
}

func (w *world) iCallTaskListCountingBy(by string) error {
	_ = w.callTool("task_list", map[string]any{"count": by})
	return nil
}

func (w *world) iCallTaskListCountingByWithAsWell(by, argument string) error {
	asked := map[string]map[string]any{
		"field":  {"fields": []string{"number"}},
		"limit":  {"limit": 2},
		"offset": {"offset": 1},
	}[argument]
	asked["count"] = by
	_ = w.callTool("task_list", asked)
	return nil
}

func (w *world) itCounts(n int, under string) error {
	listing, err := w.listingSaid()
	if err != nil {
		return err
	}
	counts, _ := listing["counts"].(map[string]any)
	if counts[under] != float64(n) {
		return fmt.Errorf("it counts %v under %q, want %d: %v", counts[under], under, n, counts)
	}
	return nil
}

func (w *world) itCountsNothingUnder(under string) error {
	listing, err := w.listingSaid()
	if err != nil {
		return err
	}
	counts, _ := listing["counts"].(map[string]any)
	if counted, said := counts[under]; said {
		return fmt.Errorf("it counts %v under %q, want it left out", counted, under)
	}
	return nil
}

func (w *world) itSaysThereIsNoTask(n int) error {
	unknown, err := w.noTaskSaid()
	if err != nil {
		return err
	}
	if !slices.Equal(unknown, []any{float64(n)}) {
		return fmt.Errorf("it says %v is no task, want %d alone", unknown, n)
	}
	return nil
}

func (w *world) itSaysEveryNumberIsATask() error {
	unknown, err := w.noTaskSaid()
	if err != nil {
		return err
	}
	if unknown != nil {
		return fmt.Errorf("it says %v is no task", unknown)
	}
	return nil
}

func (w *world) iCallAskingForNoField(tool string) error {
	_ = w.callTool(tool, map[string]any{"fields": []any{}})
	return nil
}

func (w *world) everyRowCarriesOnly(first, second, third string) error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	rows, err := carriedUnder(w.result.StructuredContent, "tasks")
	if err != nil {
		return err
	}
	for _, one := range rows {
		row, _ := one.(map[string]any)
		named := slices.Sorted(maps.Keys(row))
		if want := []string{first, second, third}; !slices.Equal(named, slices.Sorted(slices.Values(want))) {
			return fmt.Errorf("a row carries %v, want %v alone", named, want)
		}
	}
	return nil
}

func (w *world) everyFieldARowCarries() ([]string, error) {
	if len(w.tools) == 0 {
		if err := w.iAskItWhatItCanDo(); err != nil {
			return nil, err
		}
	}
	shape, err := w.schemaOf("task_list")
	if err != nil {
		return nil, err
	}
	var listing struct {
		Properties struct {
			Tasks struct {
				Items struct {
					Properties map[string]any `json:"properties"`
				} `json:"items"`
			} `json:"tasks"`
		} `json:"properties"`
	}
	if err := remarshalTo(shape, &listing); err != nil {
		return nil, err
	}
	fields := slices.Sorted(maps.Keys(listing.Properties.Tasks.Items.Properties))
	if len(fields) == 0 {
		return nil, fmt.Errorf("task_list says a row carries nothing: %v", shape)
	}
	return fields, nil
}

func (w *world) everyRowCarriesEveryField() error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	rows, err := carriedUnder(w.result.StructuredContent, "tasks")
	if err != nil {
		return err
	}
	fields, err := w.everyFieldARowCarries()
	if err != nil {
		return err
	}
	for _, one := range rows {
		row, _ := one.(map[string]any)
		if named := slices.Sorted(maps.Keys(row)); !slices.Equal(named, fields) {
			return fmt.Errorf("a row carries %v, want every one of %v", named, fields)
		}
	}
	return nil
}

func (w *world) theAnswerNamesEveryField() error {
	fields, err := w.everyFieldARowCarries()
	if err != nil {
		return err
	}
	said := textOf(w.result)
	for _, field := range fields {
		if !strings.Contains(said, field) {
			return fmt.Errorf("the answer %q does not name %q", said, field)
		}
	}
	return nil
}

func (w *world) namesEveryFieldARowCarries(tool string) error {
	fields, err := w.everyFieldARowCarries()
	if err != nil {
		return err
	}
	for _, t := range w.tools {
		if t.Name != tool {
			continue
		}
		var input struct {
			Properties map[string]struct {
				Description string `json:"description"`
				Items       struct {
					Enum []string `json:"enum"`
				} `json:"items"`
			} `json:"properties"`
		}
		if err := remarshalTo(t.InputSchema, &input); err != nil {
			return err
		}
		asked, ok := input.Properties["fields"]
		if !ok {
			return fmt.Errorf("%s takes no fields to ask for", tool)
		}
		if named := slices.Sorted(slices.Values(asked.Items.Enum)); !slices.Equal(named, fields) {
			return fmt.Errorf("%s accepts the fields %v, want %v", tool, named, fields)
		}
		for _, field := range fields {
			if !strings.Contains(asked.Description, field) {
				return fmt.Errorf("%s says %q of its fields, and does not name %q", tool, asked.Description, field)
			}
		}
		return nil
	}
	return fmt.Errorf("no tool named %s was offered", tool)
}

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

func (w *world) itAnswersWithoutAnError() error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
	}
	return nil
}

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

func (w *world) answeredTask() (map[string]any, error) {
	if w.callErr != nil {
		return nil, fmt.Errorf("the call failed: %w", w.callErr)
	}
	doc, ok := w.result.StructuredContent.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("structured content is not a task: %#v", w.result.StructuredContent)
	}
	return doc, nil
}

func (w *world) itAnswersWithATaskCarryingLabels(n int) error {
	doc, err := w.answeredTask()
	if err != nil {
		return err
	}
	if labels, _ := doc["labels"].([]any); len(labels) != n {
		return fmt.Errorf("labels = %v, want %d of them", doc["labels"], n)
	}
	return nil
}

func (w *world) itsAnswerSaysTheLabelsAdded(want string) error {
	doc, err := w.answeredTask()
	if err != nil {
		return err
	}
	added, _ := doc["added"].(map[string]any)
	labels, _ := added["labels"].([]any)
	said := make([]string, 0, len(labels))
	for _, l := range labels {
		said = append(said, fmt.Sprint(l))
	}
	if got := strings.Join(said, ", "); got != want {
		return fmt.Errorf("the labels added are %q, want %q: %#v", got, want, doc["added"])
	}
	return nil
}

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
	title, err := schemaProperty(w.tool("task_new"), "title")
	if err != nil {
		return err
	}
	if _, has := title["enum"]; has {
		return fmt.Errorf(`"title" names an enum: %#v, want none — it has no closed set`, title)
	}
	return nil
}

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

func (w *world) tool(name string) *sdk.Tool {
	for _, t := range w.tools {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func (w *world) itOffers(name string) error {
	if w.tool(name) == nil {
		return fmt.Errorf("tool %q was not offered", name)
	}
	return nil
}

func (w *world) itDoesNotOffer(name string) error {
	if w.tool(name) != nil {
		return fmt.Errorf("tool %q was offered, want it left out", name)
	}
	return nil
}

func (w *world) itOffersAToolForEveryCommand() error {
	for _, op := range engine.All() {
		if w.tool(op.Name()) == nil {
			return fmt.Errorf("no tool offered for command %q", op.Line())
		}
	}
	return nil
}

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

func (w *world) everyToolIsMarkedAsReachingNothingOutsideCairn() error {
	for _, t := range w.tools {
		if t.Annotations == nil || t.Annotations.OpenWorldHint == nil || *t.Annotations.OpenWorldHint {
			return fmt.Errorf("tool %q is not marked as reaching nothing outside cairn", t.Name)
		}
	}
	return nil
}

func (w *world) startedServeAndGreetedItWithoutWaiting() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	greeting := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"acceptance","version":"0"}}}`

	var out, errOut bytes.Buffer
	code := cli.Main(ctx, []string{"mcp", "serve"}, strings.NewReader(greeting+"\n"), &out, &errOut)
	w.Out = test.Result{Stdout: out.String(), Stderr: errOut.String(), Code: code}
	return nil
}

func (w *world) itStopsAndSaysNothingWentWrong() error {
	if w.Out.Code != 0 {
		return fmt.Errorf("exit %d, want 0", w.Out.Code)
	}
	if w.Out.Stderr != "" {
		return fmt.Errorf("stderr = %q, want empty", w.Out.Stderr)
	}
	return nil
}

func versionField(t *testing.T) string {
	t.Helper()
	first, _, _ := strings.Cut(run(t, "--version").Stdout, "\n")
	_, version, _ := strings.Cut(first, " ")
	return version
}

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

func (w *world) itTellsTheAgentWhereTheWorkGoes() error {
	if w.session == nil {
		return fmt.Errorf("no session: greet it first")
	}
	told := w.session.InitializeResult().Instructions
	for _, want := range []string{"what was done or decided", "goes in its notes", "never in its description"} {
		if !strings.Contains(told, want) {
			return fmt.Errorf("instructions = %q, want them to say %q", told, want)
		}
	}
	return nil
}

func (w *world) taskEditsDescriptionPointsAtTheNotes() error {
	desc, err := schemaProperty(w.tool("task_edit"), "desc")
	if err != nil {
		return err
	}
	if said, _ := desc["description"].(string); !strings.Contains(said, "what was done or decided goes in its notes") {
		return fmt.Errorf("desc says %q, want it to say where the work on a task goes", said)
	}
	return nil
}

func initMCP(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.Register(sc)
		w.registerHandOver(sc)
		sc.Step("^I start `cairn mcp serve` and greet it$", w.startedServeAndGreetedIt)
		sc.Step("^it answers with its name and its version$", w.itAnswersWithItsNameAndVersion)
		sc.Step("^I start `cairn mcp serve` and greet it without waiting$", w.startedServeAndGreetedItWithoutWaiting)
		sc.Step("^it stops, and says nothing went wrong$", w.itStopsAndSaysNothingWentWrong)
		sc.Step("^it tells the agent that what was done or decided on a task goes in its notes, never in its description$", w.itTellsTheAgentWhereTheWorkGoes)
		sc.Step("^the description `task_edit` takes says what was done or decided on a task goes in its notes$", w.taskEditsDescriptionPointsAtTheNotes)
		sc.Step("^a cairn server started with `--access (read|write|full)`$", w.aCairnServerStartedWithAccess)
		sc.Step("^a cairn server started with no access given$", w.aCairnServerStartedWithNoAccessGiven)
		sc.Step(`^a configuration that turns the page "([^"]*)" off$`, func(page string) error {
			return w.WriteConfig(fmt.Sprintf("[pages]\n%s = false\n", page))
		})
		sc.Step(`^a configuration that turns every page off$`, func() error {
			return w.WriteConfig("[pages]\ntasks = false\nobjectives = false\ntime = false\n")
		})
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
		sc.Step("^I call `task_new` with the title \"([^\"]*)\" and the link (\\d+)$", w.iCallTaskAddWithTheTitleAndTheLink)
		sc.Step("^I call `task_search` looking for \"([^\"]*)\"$", w.iCallTaskSearchLookingFor)
		sc.Step("^I call `task_search` looking for nothing$", w.iCallTaskSearchLookingForNothing)
		sc.Step("^it answers with a task carrying both labels$", w.itAnswersWithATaskCarryingBothLabels)
		sc.Step("^I call `task_edit` with the number (\\d+) and the label \"([^\"]*)\"$", func(n int, label string) error {
			w.number = n
			_ = w.callTool("task_edit", map[string]any{"number": n, "label": []string{label}})
			return nil
		})
		sc.Step(`^I added a task titled "([^"]*)" with the labels "([^"]*)" and "([^"]*)"$`, func(title, first, second string) error {
			return w.Seed(fmt.Sprintf(`cairn task new "%s" --label "%s" --label "%s"`, title, first, second))
		})
		sc.Step("^I call `task_edit` with the number (\\d+), replacing \"([^\"]*)\" by \"([^\"]*)\"$", func(n int, old, by string) error {
			w.number = n
			_ = w.callTool("task_edit", map[string]any{"number": n, "replace": old, "by": by})
			return nil
		})
		sc.Step(`^it answers with a task carrying (\d+) labels$`, w.itAnswersWithATaskCarryingLabels)
		sc.Step(`^its answer says the labels added are "([^"]*)"$`, w.itsAnswerSaysTheLabelsAdded)
		sc.Step("^I call `([a-z_]+)` with the number (\\d+)$", w.iCallWithTheNumber)
		sc.Step(`^it answers with (\d+) tasks?$`, w.itAnswersWithNTasks)
		sc.Step(`^no row of it carries a description$`, w.noRowCarriesADescription)
		sc.Step("^I call `([a-z_]+)` asking for the fields \"([^\"]*)\", \"([^\"]*)\" and \"([^\"]*)\"$", w.iCallAskingForTheFields)
		sc.Step("^I call `task_search` looking for \"([^\"]*)\", asking for the fields \"([^\"]*)\", \"([^\"]*)\" and \"([^\"]*)\"$", w.iCallTaskSearchLookingForAskingForTheFields)
		sc.Step("^I call `task_search` looking for any of \"([^\"]*)\", \"([^\"]*)\", \"([^\"]*)\" and \"([^\"]*)\"$", w.iCallTaskSearchLookingForAnyOf)
		sc.Step("^I call `task_search` looking for any of \"([^\"]*)\" and \"([^\"]*)\", asking for the fields \"([^\"]*)\", \"([^\"]*)\" and \"([^\"]*)\"$", w.iCallTaskSearchLookingForAnyOfAskingForTheFields)
		sc.Step(`^it says "([^"]*)" matched (task \d+|no task)$`, w.itSaysTheTermMatched)
		sc.Step(`^it says nothing of which term matched$`, w.itSaysNothingOfWhichTermMatched)
		sc.Step("^I call `task_list` asking for the numbers (\\d+), (\\d+) and (\\d+)$", func(first, second, third string) error {
			return w.iCallTaskListAskingForTheNumbers(first, second, third)
		})
		sc.Step("^I call `task_list` asking for the numbers (\\d+), (\\d+) and (\\d+), archived only$", w.iCallTaskListAskingForTheNumbersArchivedOnly)
		sc.Step("^I call `task_list` asking for the numbers (\\d+) and (\\d+), and for the fields \"([^\"]*)\", \"([^\"]*)\" and \"([^\"]*)\"$", w.iCallTaskListAskingForTheNumbersAndTheFields)
		sc.Step("^I call `task_list` asking for the numbers (\\d+), (\\d+) and (\\d+), written as numbers$", w.iCallTaskListAskingForTheNumbersWrittenAsNumbers)
		sc.Step("^I call `task_list` asking for the numbers (\\d+) and (-?[\\d.]+|true|\"[^\"]*\")$", w.iCallTaskListAskingForTheNumberAndWhatWasSent)
		sc.Step("^I call `task_new` with the title \"([^\"]*)\", blocked by (\\d+) and linked to (\\d+), written as numbers$", w.iCallTaskNewBlockedAndLinkedWrittenAsNumbers)
		sc.Step(`^it answers with a task blocked by task (\d+) and linked to task (\d+)$`, w.itAnswersWithATaskBlockedAndLinked)
		sc.Step("^I call `task_archive` with the number (\\d+) and also (\\d+), written as a number$", w.iCallTaskArchiveWithTheNumberAndAlsoWrittenAsANumber)
		sc.Step("^I call `task_edit` with the number (\\d+) and the label (\\d+), written as a number$", w.iCallTaskEditWithTheLabelWrittenAsANumber)
		sc.Step(`^every list of task or objective numbers it takes says it takes whole numbers$`, w.everyListOfNumbersSaysItTakesWholeNumbers)
		sc.Step("^I call `task_list` asking for no number$", func() error { return w.iCallTaskListAskingForTheNumbers([]string{}...) })
		sc.Step(`^it answers with the tasks? (\d+)(?: and (\d+))?$`, w.itAnswersWithTheTasks)
		sc.Step(`^it says there is no task (\d+)$`, w.itSaysThereIsNoTask)
		sc.Step(`^it says every number asked for is a task$`, w.itSaysEveryNumberIsATask)
		sc.Step("^I call `task_list` asking for (-?\\d+) tasks? after the first (-?\\d+)$", w.iCallTaskListAskingForAPage)
		sc.Step("^I call `task_list` in \"([^\"]*)\", asking for (\\d+) tasks? after the first (\\d+)$", w.iCallTaskListInAskingForAPage)
		sc.Step(`^it says there (?:are|is) (\d+) tasks? in all$`, w.itSaysHowManyInAll)
		sc.Step(`^it says nothing of how many there are in all$`, w.itSaysNothingOfHowManyInAll)
		sc.Step("^I call `task_list` counting by \"([^\"]*)\"$", w.iCallTaskListCountingBy)
		sc.Step("^I call `task_list` counting by \"([^\"]*)\", with a (field|limit|offset) as well$", w.iCallTaskListCountingByWithAsWell)
		sc.Step(`^it counts (\d+) under "([^"]*)"$`, w.itCounts)
		sc.Step(`^it counts nothing under "([^"]*)"$`, w.itCountsNothingUnder)
		sc.Step("^I call `([a-z_]+)` asking for no field$", w.iCallAskingForNoField)
		sc.Step(`^every row of it carries "([^"]*)", "([^"]*)" and "([^"]*)", and nothing else$`, w.everyRowCarriesOnly)
		sc.Step(`^every row of it carries every field a row can carry$`, w.everyRowCarriesEveryField)
		sc.Step(`^the answer names every field a row can carry$`, w.theAnswerNamesEveryField)
		sc.Step("^`([a-z_]+)` names every field a row can carry$", w.namesEveryFieldARowCarries)
		sc.Step(`^it answers without an error$`, w.itAnswersWithoutAnError)
		sc.Step(`^it answers with the task titled "([^"]*)"$`, w.itAnswersWithTheTaskTitled)
		sc.Step(`^it answers with the card a person would have read$`, w.itAnswersWithTheCardAPersonWouldHaveRead)
		sc.Step("^it answers with the shape `--json` would have printed$", w.itAnswersWithTheShapeJSONWouldHavePrinted)
		sc.Step("^`task_new` names every state cairn accepts$", w.taskAddNamesEveryStateCairnAccepts)
		sc.Step("^it says what each thing it takes is for$", w.itSaysWhatEachThingItTakesIsFor)
		sc.Step(`^I added a task titled "([^"]*)" with the (deadline|description|state) "([^"]*)"$`, w.iAddedATaskTitledWith)
		sc.Step("^I call `task_edit` with the number (\\d+) and the priority \"([^\"]*)\"$",
			w.iCallTaskEditWithThePriority)
		sc.Step("^the trace of task (\\d+) says the change was made by the agent that asked$", func(n int) error {
			w.Run(fmt.Sprintf("cairn trace show %d", n))
			return test.WantContains(w.Out.Stdout, "(by acceptance, through mcp)", "the history")
		})
		sc.Step("^the trace of task (\\d+) says the change came through the server$",
			w.theHistorySaysItCameThroughTheServer)
		sc.Step("^I call `task_edit` with the number (\\d+) and the deadline as nothing$",
			w.iCallTaskEditWithTheNumberAndTheDeadlineAsNothing)
		sc.Step(`^the task has no deadline$`, w.theTaskHasNoDeadline)
		sc.Step(`^it is still titled "([^"]*)"$`, w.itAnswersWithTheTaskTitled)
		sc.Step(`^it answers that it failed$`, w.itAnswersThatItFailed)
		sc.Step(`^it says the refusal is "([^"]*)", because "([^"]*)"$`, w.itSaysTheRefusalIs)
		sc.Step(`^the answer says (.+)$`, w.theAnswerSays)
		sc.Step(`^the answer does not say "([^"]*)"$`, func(what string) error {
			if strings.Contains(textOf(w.result), what) {
				return fmt.Errorf("the answer says %q:\n%s", what, textOf(w.result))
			}
			return nil
		})
		sc.Step(`^the call itself fails$`, w.theCallItselfFails)
		sc.Step(`^a store from an older cairn, holding a task titled "([^"]*)"$`,
			w.AStoreFromAnOlderCairn)
	}
}

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

func TestAcceptance_MCP(t *testing.T) {
	for _, tag := range scenariosOf(t, "mcp.feature") {
		t.Run(strings.TrimPrefix(tag, "@"), func(t *testing.T) {
			runAcceptance(t, tag)
		})
	}
}
