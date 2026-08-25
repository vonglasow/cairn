package mcp

import (
	"context"
	"errors"
	"io"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server is what a client sees when it greets `cairn mcp serve`: a name, a
// version, and how far it is allowed to reach.
type Server struct {
	// Name is announced to the client as the server's identity.
	Name string
	// Version is announced to the client alongside Name.
	Version string
	// Access is how far the agent may reach. It is settled here, when the
	// server starts, and never per call — so nothing later has to work out
	// whether this particular request was allowed. A tool above the level is
	// left out of what is offered rather than offered and refused.
	Access Access
	// Path is where the store lives. internal/mcp does not read the
	// environment itself — the surface that starts the server resolves it
	// and hands it over, the way internal/cli's storePath does for every
	// other command.
	Path string
}

// noCloseWriter wraps an io.Writer so closing it is a no-op: the caller owns
// the stream Serve was handed and must not have it closed underneath it.
type noCloseWriter struct {
	io.Writer
}

func (noCloseWriter) Close() error { return nil }

// Serve builds an SDK server announcing s.Name and s.Version, and runs it
// over an IOTransport built on in and out. It returns when the client hangs
// up.
//
// A client owns the process it started, and ends the session by closing its
// end of the pipe. Nobody is left to read a failure at that point, so a
// request the SDK drops in flight because of it is not reported as one —
// only a session error of a different shape still is.
func Serve(ctx context.Context, s Server, in io.Reader, out io.Writer) error {
	server := sdk.NewServer(&sdk.Implementation{Name: s.Name, Version: s.Version}, nil)
	for _, op := range offered(s.Access) {
		server.AddTool(toolFor(op), handlerFor(op, s.Path))
	}
	transport := &sdk.IOTransport{Reader: io.NopCloser(in), Writer: noCloseWriter{out}}
	err := server.Run(ctx, transport)
	if clientWentAway(err) {
		return nil
	}
	return err
}

// clientWentAway reports whether the session ended because the client closed
// its end of the pipe while a request was still in flight.
//
// The number is ErrServerClosing in the SDK's own internal/jsonrpc2, which
// nothing outside the SDK may name; the public handle is the jsonrpc.Error
// alias and its code. It is written here as a literal rather than a constant
// on purpose: Go instruments statements, not declarations, so a constant would
// put the one number this depends on outside anything a test can be seen
// exercising. TestServe_answersNothingWrongWhenTheClientHasGone pins it, so a
// change on the SDK's side turns red here rather than going quiet in somebody's
// terminal.
func clientWentAway(err error) bool {
	var wire *jsonrpc.Error
	return errors.As(err, &wire) && wire.Code == -32004
}
