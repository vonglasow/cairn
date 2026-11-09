package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/debug"
	"cairn/internal/engine"
)

const (
	carriedVariable = "CAIRN_MCP_CARRIED"
	listening       = "subscriptions/listen"
	toolsChanged    = "notifications/tools/list_changed"
)

var handOver = replaceItself

var settling, _ = jsonrpc.MakeID("cairn: what was read before is handled")

type onDisk struct {
	path     string
	resolved string
	size     int64
	modified time.Time
}

func identify(path string) (onDisk, error) {
	if !strings.ContainsRune(path, filepath.Separator) {
		found, err := exec.LookPath(path)
		if err != nil {
			return onDisk{}, err
		}
		path = found
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return onDisk{}, err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return onDisk{}, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return onDisk{}, err
	}
	return onDisk{path: path, resolved: resolved, size: info.Size(), modified: info.ModTime()}, nil
}

func (o onDisk) replaced() (onDisk, bool) {
	now, err := identify(o.path)
	if err != nil {
		return o, false
	}
	return now, now != o
}

type carried struct {
	State  sdk.ServerSessionState `json:"state"`
	Tools  []string               `json:"tools"`
	Unread []byte                 `json:"unread"`
	Kept   *engine.Started        `json:"kept,omitempty"`
}

func takenOver() (carried, bool, error) {
	said, set := os.LookupEnv(carriedVariable)
	if !set {
		return carried{}, false, nil
	}
	_ = os.Unsetenv(carriedVariable)
	var c carried
	return c, true, json.Unmarshal([]byte(said), &c)
}

type remembered struct {
	mu    sync.Mutex
	state sdk.ServerSessionState
}

func (g *remembered) remember(next sdk.MethodHandler) sdk.MethodHandler {
	return func(ctx context.Context, method string, req sdk.Request) (sdk.Result, error) {
		result, err := next(ctx, method, req)
		g.mu.Lock()
		defer g.mu.Unlock()
		switch params := req.GetParams().(type) {
		case *sdk.InitializeParams:
			g.state.InitializeParams = params
			if answer, ok := result.(*sdk.InitializeResult); ok {
				g.state.NegotiatedProtocolVersion = answer.ProtocolVersion
			}
		case *sdk.InitializedParams:
			g.state.InitializedParams = params
		}
		return result, err
	}
}

func (g *remembered) held() sdk.ServerSessionState {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.state
}

type lines struct {
	in        *bufio.Reader
	out       io.Writer
	writing   sync.Mutex
	owed      sync.WaitGroup
	held      sync.Map
	spoken    atomic.Bool
	between   func()
	heard     func()
	running   *onDisk
	installed onDisk
	waiting   []byte
	settled   chan struct{}
	carry     func(unread []byte) carried
	noted     *debug.Log
}

func (l *lines) Connect(context.Context) (sdk.Connection, error) { return l, nil }

func (l *lines) Read(ctx context.Context) (jsonrpc.Message, error) {
	line := l.waiting
	if line != nil {
		l.waiting = nil
		select {
		case <-l.settled:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		if l.handOver(line) {
			return nil, io.EOF
		}
		*l.running = l.installed
	} else {
		read, err := l.in.ReadBytes('\n')
		if len(read) == 0 && err != nil {
			return nil, err
		}
		line = read
		if l.running != nil {
			if now, changed := l.running.replaced(); changed {
				l.waiting, l.installed, l.settled = line, now, make(chan struct{})
				return &jsonrpc.Request{ID: settling, Method: "ping", Params: json.RawMessage(`{}`)}, nil
			}
		}
	}
	if l.spoken.Load() {
		l.between()
	}
	msg, err := jsonrpc.DecodeMessage(line)
	if err != nil {
		return nil, err
	}
	if req, ok := msg.(*jsonrpc.Request); ok && req.IsCall() {
		if req.Method == listening {
			l.held.Store(req.ID, line)
		} else {
			l.owed.Add(1)
		}
	}
	return msg, nil
}

func (l *lines) handOver(line []byte) bool {
	l.owed.Wait()
	var unread []byte
	l.held.Range(func(_, open any) bool {
		unread = append(unread, open.([]byte)...)
		return true
	})
	ahead, _ := l.in.Peek(l.in.Buffered())
	payload, err := json.Marshal(l.carry(slices.Concat(unread, line, ahead)))
	if err == nil {
		err = handOver(l.running.path, payload)
	}
	if err != nil {
		if l.noted != nil {
			l.noted.Say("mcp", "the cairn now on disk could not take over: "+l.noted.Text(err.Error()))
		}
		return false
	}
	return true
}

func (l *lines) Write(_ context.Context, msg jsonrpc.Message) error {
	if answer, answering := msg.(*jsonrpc.Response); answering && answer.ID == settling {
		close(l.settled)
		return nil
	}
	data, err := jsonrpc.EncodeMessage(msg)
	if err != nil {
		return err
	}
	if note, noting := msg.(*jsonrpc.Request); noting && note.Method == toolsChanged {
		l.heard()
	}
	l.writing.Lock()
	_, err = l.out.Write(append(data, '\n'))
	l.writing.Unlock()
	if answer, answering := msg.(*jsonrpc.Response); answering {
		if _, was := l.held.LoadAndDelete(answer.ID); !was {
			l.owed.Done()
		}
		if !l.spoken.Swap(true) {
			l.between()
		}
	}
	return err
}

func (l *lines) Close() error { return nil }

func (l *lines) SessionID() string { return "" }
