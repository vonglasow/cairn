package mcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/engine"
)

func connect(t *testing.T, s Server) *sdk.ClientSession {
	t.Helper()

	serverConn, clientConn := net.Pipe()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, s, serverConn, serverConn)
	}()

	client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, &sdk.IOTransport{Reader: clientConn, Writer: clientConn}, nil)
	if err != nil {
		cancel()
		t.Fatalf("Connect: %v", err)
	}

	t.Cleanup(func() {
		_ = session.Close()
		_ = clientConn.Close()
		<-done
		cancel()
	})

	return session
}

func TestServe_answersItsNameAndVersion(t *testing.T) {
	session := connect(t, Server{Name: "cairn", Version: "1.2.3", Access: engine.Full})

	info := session.InitializeResult().ServerInfo
	if info.Name != "cairn" {
		t.Errorf("ServerInfo.Name = %q, want %q", info.Name, "cairn")
	}
	if info.Version != "1.2.3" {
		t.Errorf("ServerInfo.Version = %q, want %q", info.Version, "1.2.3")
	}

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if got, want := len(tools.Tools), len(engine.All()); got != want {
		t.Errorf("ListTools = %d tools, want %d (one per operation, at Full access)", got, want)
	}
}

func TestServe_answersNothingWrongWhenTheClientHasGone(t *testing.T) {
	greeting := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"probe","version":"0"}}}`

	var out bytes.Buffer
	err := Serve(context.Background(), Server{Name: "cairn", Version: "1.2.3", Access: engine.Write},
		strings.NewReader(greeting+"\n"), &out)

	if err != nil {
		t.Errorf("Serve = %v, want nil", err)
	}
}

func TestServe_stillReturnsASessionErrorThatIsNotThePeerLeaving(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	never, open := io.Pipe()
	defer func() { _ = open.Close() }()

	var out bytes.Buffer
	err := Serve(ctx, Server{Name: "cairn", Version: "1.2.3", Access: engine.Write},
		never, &out)

	if err == nil {
		t.Fatal("Serve = nil, want a session error")
	}
}

func TestClientWentAway_readsTheCodeAndNothingElse(t *testing.T) {
	for _, tc := range []struct {
		what string
		err  error
		want bool
	}{
		{"the peer closed its end", &jsonrpc.Error{Code: -32004}, true},
		{"the peer closed, wrapped on the way up", fmt.Errorf("server is closing: %w", &jsonrpc.Error{Code: -32004}), true},
		{"some other protocol error", &jsonrpc.Error{Code: -32603}, false},
		{"the code with its sign lost", &jsonrpc.Error{Code: 32004}, false},
		{"a neighbouring code", &jsonrpc.Error{Code: -32005}, false},
		{"nothing went wrong at all", nil, false},
		{"an error that is not the protocol's", errors.New("the store cannot be read"), false},
	} {
		if got := clientWentAway(tc.err); got != tc.want {
			t.Errorf("clientWentAway(%s) = %v, want %v", tc.what, got, tc.want)
		}
	}
}
