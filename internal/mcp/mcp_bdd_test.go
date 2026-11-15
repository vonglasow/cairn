package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/cli"
	"cairn/internal/engine"
	"cairn/test"
)

func run(t *testing.T, h engine.Home, args ...string) test.Result {
	t.Helper()
	var out, err bytes.Buffer
	code := cli.MainIn(context.Background(), h, args, strings.NewReader(""), &out, &err)
	return test.Result{Stdout: out.String(), Stderr: err.String(), Code: code}
}

func runAcceptance(t *testing.T, tag string) {
	t.Helper()
	t.Parallel()
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
	ctx, cancel := context.WithCancel(context.Background())

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
		cli.MainIn(ctx, w.Home(), append([]string{"mcp", "serve"}, extra...), clientToServer, serverToClient, io.Discard)
	}()

	go func() {
		<-ctx.Done()
		closeAll()
	}()

	connecting, connected := context.WithTimeout(ctx, 5*time.Second)
	session, err := w.client().Connect(connecting, &sdk.IOTransport{Reader: writtenByServer, Writer: writtenByClient}, beforeSubscriptions)
	connected()
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

func (w *world) iCall(tool string) error {
	_ = w.callTool(tool, nil)
	return nil
}

func (w *world) iCallTaskAddWithTheTitle(title string) error {
	_ = w.callTool("task_new", map[string]any{"title": title})
	return nil
}

func (w *world) iCallWithTheNumber(tool string, n int) error {
	w.number = n
	_ = w.callTool(tool, map[string]any{"number": n})
	return nil
}

func (w *world) iAddedATaskTitledWith(title, what, value string) error {
	flag := map[string]string{"deadline": "--deadline", "description": "--desc", "state": "--state", "objective": "--objective"}[what]
	return w.Seed(fmt.Sprintf(`cairn task new "%s" %s "%s"`, title, flag, value))
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

func (w *world) itAnswersWithoutAnError() error {
	if w.callErr != nil {
		return fmt.Errorf("the call failed: %w", w.callErr)
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

func initMCP(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.Register(sc)
		w.registerHandOver(sc)
		sc.Step("^I start `cairn mcp serve` and greet it$", w.startedServeAndGreetedIt)
		sc.Step("^it tells the agent that what was done or decided on a task goes in its notes, never in its description$", w.itTellsTheAgentWhereTheWorkGoes)
		sc.Step("^a cairn server started with `--access (read|write|full)`$", w.aCairnServerStartedWithAccess)
		sc.Step("^a cairn server started with no access given$", w.aCairnServerStartedWithNoAccessGiven)
		sc.Step(`^a configuration that turns the page "([^"]*)" off$`, func(page string) error {
			return w.WriteConfig(fmt.Sprintf("[pages]\n%s = false\n", page))
		})
		sc.Step("^I ask it what it can do$", w.iAskItWhatItCanDo)
		sc.Step("^it offers `([^`]*)`$", w.itOffers)
		sc.Step("^it does not offer `([^`]*)`$", w.itDoesNotOffer)
		sc.Step("^`([^`]*)` is marked as losing work$", w.toolIsMarkedAsLosingWork)
		sc.Step("^`([^`]*)` is marked as read-only$", w.toolIsMarkedAsReadOnly)
		sc.Step("^I call `([a-z_]+)`$", w.iCall)
		sc.Step("^I call `task_new` with the title \"([^\"]*)\"$", w.iCallTaskAddWithTheTitle)
		sc.Step(`^I edited task (\d+) with the blocker "([^"]*)"$`, func(n int, blocker string) error {
			return w.Seed(fmt.Sprintf(`cairn task edit %d --blocker %s`, n, blocker))
		})
		sc.Step(`^I added a task titled "([^"]*)" with the labels "([^"]*)" and "([^"]*)"$`, func(title, first, second string) error {
			return w.Seed(fmt.Sprintf(`cairn task new "%s" --label "%s" --label "%s"`, title, first, second))
		})
		sc.Step("^I call `([a-z_]+)` with the number (\\d+)$", w.iCallWithTheNumber)
		// A search for one term with a page: `task_search` looking for "spec", asking for 1 task after the first 1 -> "spec", 1, 1.
		// A search for any of two terms with a page: any of "T010" and "T013", asking for 1 task after the first 1 -> "T010", "T013", 1, 1.
		sc.Step(`^it answers without an error$`, w.itAnswersWithoutAnError)
		sc.Step(`^I added a task titled "([^"]*)" with the (deadline|description|state|objective) "([^"]*)"$`, w.iAddedATaskTitledWith)
		sc.Step("^I call `([a-z_]+)` with (\\{.*\\})$", w.iCallWithJSON)
		sc.Step(`^the task has no deadline$`, w.theTaskHasNoDeadline)
		sc.Step(`^it answers that it failed$`, w.itAnswersThatItFailed)
		sc.Step(`^the answer says (.+)$`, w.theAnswerSays)
		sc.Step(`^the answer does not say "([^"]*)"$`, func(what string) error {
			if strings.Contains(textOf(w.result), what) {
				return fmt.Errorf("the answer says %q:\n%s", what, textOf(w.result))
			}
			return nil
		})
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

func (w *world) iCallWithJSON(tool, sent string) error {
	args := map[string]any{}
	if err := json.Unmarshal([]byte(sent), &args); err != nil {
		return err
	}
	_ = w.callTool(tool, args)
	return nil
}
