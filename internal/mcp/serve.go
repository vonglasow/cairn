package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	rdebug "runtime/debug"
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
}

const whereTheWorkGoes = "A task's description says what the task is; what was done or decided on it goes in its notes, never in its description. " +
	"Its notes are the directory task_show names as its context: a dated entry, YYYY-MM-DD-<subject>.md, for each piece of work, " +
	"and summary.md, rewritten whole, for where the task stands."

type noCloseWriter struct {
	io.Writer
}

func (noCloseWriter) Close() error { return nil }

// Serve builds an SDK server announcing s.Name and s.Version, and runs it over an IOTransport built on in and out.
func Serve(ctx context.Context, s Server, in io.Reader, out io.Writer) error {
	server := sdk.NewServer(&sdk.Implementation{Name: s.Name, Version: s.Version}, &sdk.ServerOptions{Instructions: whereTheWorkGoes})
	for _, op := range offered(s.Access, engine.PagesFrom(s.Home)) {
		server.AddTool(toolFor(op), handlerFor(op, s.Home))
	}
	if s.Noted != nil {
		server.AddReceivingMiddleware(noting(s.Noted))
	}
	transport := &sdk.IOTransport{Reader: io.NopCloser(in), Writer: noCloseWriter{out}}
	err := server.Run(ctx, transport)
	if clientWentAway(err) {
		return nil
	}
	return err
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
