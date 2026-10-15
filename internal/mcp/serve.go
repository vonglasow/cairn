package mcp

import (
	"context"
	"errors"
	"io"

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
	for _, op := range offered(s.Access) {
		server.AddTool(toolFor(op), handlerFor(op, s.Home))
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
