package mcp

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	rdebug "runtime/debug"
	"slices"
	"sync"

	"cairn/internal/debug"
	"cairn/internal/engine"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server is what a client sees when it greets `cairn mcp serve`: a name, a version, and how far it is allowed to reach.
type Server struct {
	Name    string
	Version string
	Access  engine.Level
	Home    engine.Home
	Noted   *debug.Log
	Binary  string
}

const whereTheWorkGoes = "A task's description says what the task is; what was done or decided on it goes in its notes, never in its description. " +
	"Its notes are the directory task_show names as its context: a dated entry, YYYY-MM-DD-<subject>.md, for each piece of work, " +
	"and summary.md, rewritten whole, for where the task stands."

// Serve builds an SDK server announcing s.Name and s.Version, and runs it over in and out, one message a line, answering from the store it started on; between two requests, a cairn installed at s.Binary takes the session over, store included, and the tools follow the pages the configuration turns on.
func Serve(ctx context.Context, s Server, in io.Reader, out io.Writer) error {
	before, resumed, err := takenOver()
	if err != nil {
		return fmt.Errorf("the session handed over by the cairn this one replaced cannot be read: %w", err)
	}
	home := engine.Keep(s.Home)
	if before.Kept != nil {
		home.Kept = before.Kept
	}
	server := sdk.NewServer(&sdk.Implementation{Name: s.Name, Version: s.Version}, &sdk.ServerOptions{Instructions: whereTheWorkGoes})
	offers := &offering{server: server, access: s.Access, home: home}
	offers.refresh()
	offers.heard()
	greeted := &remembered{}
	server.AddReceivingMiddleware(greeted.remember)
	if s.Noted != nil {
		server.AddReceivingMiddleware(noting(s.Noted))
	}
	if resumed {
		offers.registered, offers.told = before.Tools, before.Tools
	}
	greeted.state = before.State
	transport := &lines{in: bufio.NewReader(io.MultiReader(bytes.NewReader(before.Unread), in)), out: out, noted: s.Noted, between: offers.refresh, heard: offers.heard}
	transport.carry = func(unread []byte) carried {
		return carried{State: greeted.held(), Tools: offers.names(), Unread: unread, Kept: home.Kept}
	}
	if running, err := identify(s.Binary); s.Binary != "" && err == nil {
		transport.running = &running
	}
	session, err := server.Connect(ctx, transport, &sdk.ServerSessionOptions{State: &before.State})
	if err != nil {
		return err
	}
	stop := context.AfterFunc(ctx, func() { _ = session.Close() })
	defer stop()
	err = session.Wait()
	if clientWentAway(err) {
		return nil
	}
	return err
}

type offering struct {
	mu         sync.Mutex
	server     *sdk.Server
	access     engine.Level
	home       engine.Home
	registered []string
	told       []string
}

func (o *offering) refresh() {
	o.mu.Lock()
	defer o.mu.Unlock()
	ops := offered(o.access, engine.PagesFrom(o.home))
	var names, gone []string
	for _, op := range ops {
		names = append(names, op.Name())
	}
	if slices.Equal(names, o.registered) {
		return
	}
	for _, op := range engine.All() {
		if !slices.Contains(names, op.Name()) {
			gone = append(gone, op.Name())
		}
	}
	o.server.RemoveTools(gone...)
	for _, op := range ops {
		o.server.AddTool(toolFor(op), handlerFor(op, o.home))
	}
	o.registered = names
}

func (o *offering) heard() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.told = o.registered
}

func (o *offering) names() []string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.told
}

func clientWentAway(err error) bool {
	var wire *jsonrpc.Error
	return errors.As(err, &wire) && wire.Code == -32004
}

func noting(noted *debug.Log) sdk.Middleware {
	var once sync.Once
	return func(next sdk.MethodHandler) sdk.MethodHandler {
		return func(ctx context.Context, method string, req sdk.Request) (result sdk.Result, err error) {
			defer func() {
				if stopped := recover(); stopped != nil {
					noted.Panicked(stopped, rdebug.Stack())
					panic(stopped)
				}
			}()
			params := req.GetParams()
			if client, known := clientOf(params); known {
				once.Do(func() { noted.Say("mcp", "client "+client) })
			}
			said := method
			if call, calling := params.(*sdk.CallToolParamsRaw); calling {
				said += " " + call.Name
			}
			noted.Say("mcp", said)
			result, err = next(ctx, method, req)
			if err != nil {
				var wire *jsonrpc.Error
				if errors.As(err, &wire) {
					noted.Say("mcp", fmt.Sprintf("%s → error %d", method, wire.Code))
				} else {
					noted.Say("mcp", method+" → error "+noted.Text(err.Error()))
				}
			}
			return result, err
		}
	}
}

func clientOf(params sdk.Params) (string, bool) {
	if hello, greeting := params.(*sdk.InitializeParams); greeting && hello.ClientInfo != nil {
		return hello.ClientInfo.Name + " " + hello.ClientInfo.Version, true
	}
	if params == nil {
		return "", false
	}
	switch said := params.GetMeta()[sdk.MetaKeyClientInfo].(type) {
	case *sdk.Implementation:
		return said.Name + " " + said.Version, true
	case map[string]any:
		return fmt.Sprintf("%v %v", said["name"], said["version"]), true
	}
	return "", false
}
