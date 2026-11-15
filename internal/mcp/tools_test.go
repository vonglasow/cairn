package mcp

import (
	"cairn/internal/engine/workspace"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/engine"
	"cairn/internal/store/files"
)

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
		if name == files.StoreVariable {
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
	said, ok := result.StructuredContent.(workspace.Config)
	if !ok || said.File != file || said.Notes.By != "the file" {
		t.Errorf("structured = %#v, want the file and the notes it named", result.StructuredContent)
	}
	text := result.Content[0].(*sdk.TextContent).Text
	if !strings.Contains(text, file) || !strings.Contains(text, "named by the file") {
		t.Errorf("text = %q, want the file and what named the notes", text)
	}
}

func TestHandlerFor_refusesArgumentsItCannotRead(t *testing.T) {
	handle := handlerFor(engine.TaskNew, homeFor(t, storeAt(t)))
	req := &sdk.CallToolRequest{Params: &sdk.CallToolParamsRaw{Arguments: []byte(`not json`)}}
	refusedAsNotUnderstood(t, handle, req, "")
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

func TestAnswerSchemaFor_describesNothingWhenNothingWasDeclared(t *testing.T) {
	if got := answerSchemaFor(engine.Operation{}); got != nil {
		t.Errorf("answerSchemaFor = %v, want nothing described", got)
	}
}

func TestAnswerSchemaFor_describesNothingWhenTheTypeCannotBeRead(t *testing.T) {
	if got := answerSchemaFor(engine.Operation{
		Answers: make(chan int),
	}); got != nil {
		t.Errorf("answerSchemaFor on a shape no schema can describe = %v, want nothing", got)
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
