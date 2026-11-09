package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/cucumber/godog"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/test"
)

type serving struct {
	session *sdk.ClientSession
	result  *sdk.CallToolResult
}

func (w *workspaces) registerServer(sc *godog.ScenarioContext) {
	s := &serving{}
	sc.Step("^a cairn server started with `cairn ([^`]*)`$", func(line string) error {
		return w.startServer(s, test.SplitArgs(line))
	})
	sc.Step(`^I ask the server to add a task titled "([^"]*)"$`, func(title string) error {
		return s.call("task_new", map[string]any{"title": title})
	})
	sc.Step("^I ask the server for `([a-z_]+)`$", func(tool string) error {
		return s.call(tool, map[string]any{})
	})
	sc.Step(`^the server says the workspace "([^"]*)" answers, because "([^"]*)"$`, func(name, because string) error {
		said := s.text()
		if err := test.WantContains(said, "workspace  "+name+"\n", "the server's answer"); err != nil {
			return err
		}
		return test.WantContains(said, "because    "+because+"\n", "the server's answer")
	})
	sc.Step(`^the server says "([^"]*)"$`, func(want string) error {
		return test.WantContains(s.text(), want, "the server's answer")
	})
	sc.Step(`^the server's structured answer says the workspace is "([^"]*)"$`, func(name string) error {
		var shown struct {
			Workspace string `json:"workspace"`
			Switch    string `json:"switch"`
		}
		said, err := json.Marshal(s.result.StructuredContent)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(said, &shown); err != nil {
			return err
		}
		if shown.Workspace != name || shown.Switch == "" {
			return fmt.Errorf("the structured answer says %s, want the workspace %q and how to switch", said, name)
		}
		return nil
	})
	sc.Step(`^no store was written to "([^"]*)"$`, func(place string) error {
		if _, err := os.Stat(w.expand(place)); !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w, want no store there", place, err)
		}
		return nil
	})
	sc.Step(`^it does not say "([^"]*)"$`, func(said string) error {
		if strings.Contains(w.Out.Stdout, said) {
			return fmt.Errorf("the answer says %q:\n%s", said, w.Out.Stdout)
		}
		return nil
	})
}

func (w *workspaces) startServer(s *serving, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	clientToServer, writtenByClient := io.Pipe()
	writtenByServer, serverToClient := io.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		Main(ctx, args, clientToServer, serverToClient, io.Discard)
	}()
	client := sdk.NewClient(&sdk.Implementation{Name: "acceptance", Version: "0"}, nil)
	session, err := client.Connect(ctx, &sdk.IOTransport{Reader: writtenByServer, Writer: writtenByClient}, nil)
	stop := func() {
		if session != nil {
			_ = session.Close()
		}
		cancel()
		_ = writtenByClient.Close()
		_ = clientToServer.Close()
		<-done
	}
	if err != nil {
		stop()
		return err
	}
	s.session = session
	w.OnEnd(stop)
	return nil
}

func (s *serving) call(tool string, args map[string]any) error {
	if s.session == nil {
		return fmt.Errorf("no server: start one first")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := s.session.CallTool(ctx, &sdk.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		return err
	}
	s.result = result
	if result.IsError {
		return fmt.Errorf("%s refused: %s", tool, s.text())
	}
	return nil
}

func (s *serving) text() string {
	if s.result == nil {
		return ""
	}
	var said strings.Builder
	for _, part := range s.result.Content {
		if text, ok := part.(*sdk.TextContent); ok {
			said.WriteString(text.Text)
		}
	}
	return said.String()
}
