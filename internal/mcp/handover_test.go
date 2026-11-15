package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"

	"cairn/internal/engine"
	"cairn/test"
)

const (
	greeting    = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"unit","version":"7"}}}`
	greeted     = `{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`
	listing     = `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"task_list","arguments":{}}}`
	pinging     = `{"jsonrpc":"2.0","id":7,"method":"ping","params":{}}`
	listingNext = `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"task_list","arguments":{}}}`
	listChanged = "notifications/tools/list_changed"
)

func onDiskAt(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn")
	if err := test.WriteProgram(path, content); err != nil {
		t.Fatal(err)
	}
	return path
}

func replace(t *testing.T, path, content string) {
	t.Helper()
	if err := test.WriteProgram(path+".new", content); err != nil {
		t.Fatal(err)
	}
	later := time.Now().Add(time.Second)
	if err := os.Chtimes(path+".new", later, later); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path+".new", path); err != nil {
		t.Fatal(err)
	}
}

func TestIdentify_findsABareNameOnThePath(t *testing.T) {
	path := onDiskAt(t, "old")
	t.Setenv("PATH", filepath.Dir(path))
	running, err := identify("cairn")
	if err != nil {
		t.Fatal(err)
	}
	if running.path != path {
		t.Errorf("path = %q, want %q, found on the PATH", running.path, path)
	}
}

func TestReplaced_staysPutWhileTheFileIsGone(t *testing.T) {
	path := onDiskAt(t, "old")
	running, err := identify(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, changed := running.replaced(); changed {
		t.Error("a cairn half-way through being installed was taken for a newer one")
	}
}

type wire struct {
	t       *testing.T
	in      *io.PipeWriter
	out     *bufio.Scanner
	stopped chan error
}

func serving(t *testing.T, s Server) *wire {
	t.Helper()
	t.Setenv("CAIRN_STORE", filepath.Join(t.TempDir(), "cairn.db"))
	clientToServer, writtenByClient := io.Pipe()
	writtenByServer, serverToClient := io.Pipe()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	w := &wire{t: t, in: writtenByClient, out: bufio.NewScanner(writtenByServer), stopped: make(chan error, 1)}
	w.out.Buffer(nil, 1<<20)
	go func() {
		w.stopped <- Serve(ctx, s, clientToServer, serverToClient)
		_ = serverToClient.Close()
	}()
	t.Cleanup(func() {
		_ = writtenByClient.Close()
		cancel()
	})
	return w
}

func (w *wire) say(lines ...string) {
	w.t.Helper()
	for _, line := range lines {
		if _, err := io.WriteString(w.in, line+"\n"); err != nil {
			w.t.Fatal(err)
		}
	}
}

func (w *wire) next() map[string]any {
	w.t.Helper()
	if !w.out.Scan() {
		w.t.Fatalf("the server said nothing more: %v", w.out.Err())
	}
	var said map[string]any
	if err := json.Unmarshal(w.out.Bytes(), &said); err != nil {
		w.t.Fatalf("%q: %v", w.out.Text(), err)
	}
	return said
}

func (w *wire) answerTo(id float64) map[string]any {
	w.t.Helper()
	for {
		said := w.next()
		if said["id"] == id {
			return said
		}
	}
}

func handingOverTo(t *testing.T, outcome error) chan carried {
	t.Helper()
	got := make(chan carried, 2)
	previous := handOver
	handOver = func(_ string, payload []byte) error {
		var c carried
		if err := json.Unmarshal(payload, &c); err != nil {
			t.Errorf("the session handed over cannot be read: %v", err)
		}
		got <- c
		return outcome
	}
	t.Cleanup(func() { handOver = previous })
	return got
}

func server(t *testing.T, binary string) Server {
	return Server{Name: "cairn", Version: "0", Access: engine.Read, Home: engine.Home{Dir: t.TempDir(), Lookup: os.LookupEnv}, Binary: binary}
}

func TestServe_carriesOnItselfWhenTheCairnOnDiskCannotTakeOver(t *testing.T) {
	path := onDiskAt(t, "old")
	got := handingOverTo(t, errors.New("exec format error"))
	w := serving(t, server(t, path))
	w.say(greeting, greeted)
	w.answerTo(1)
	replace(t, path, "not a program")
	w.say(listing)
	if answer := w.answerTo(2); answer["error"] != nil {
		t.Errorf("answered %v, want the call answered by the cairn still running", answer)
	}
	w.say(listingNext)
	w.answerTo(3)
	if tries := len(got); tries != 1 {
		t.Errorf("tried to hand over %d times, want once for what was installed", tries)
	}
}

func offeredNames(t *testing.T) []string {
	t.Helper()
	var names []string
	for _, op := range offered(engine.Read, engine.PagesFrom(engine.Home{Dir: t.TempDir(), Lookup: os.LookupEnv})) {
		names = append(names, toolFor(op).Name)
	}
	return names
}

func TestServe_refusesASessionHandedOverItCannotRead(t *testing.T) {
	t.Setenv(carriedVariable, "not json")
	w := serving(t, server(t, ""))
	err := <-w.stopped
	if err == nil || !strings.Contains(err.Error(), "handed over") {
		t.Errorf("stopped with %v, want it to say the session handed over could not be read", err)
	}
}

func TestLines_handsOverTheListeningItWasHoldingOpen(t *testing.T) {
	const listen = `{"jsonrpc":"2.0","id":9,"method":"subscriptions/listen","params":{"notifications":{"toolsListChanged":true}}}`
	path := onDiskAt(t, "old")
	running, err := identify(path)
	if err != nil {
		t.Fatal(err)
	}
	got := handingOverTo(t, nil)
	l := &lines{in: bufio.NewReader(strings.NewReader(listen + "\n" + listing + "\n")), out: io.Discard, running: &running,
		carry: func(unread []byte) carried { return carried{Unread: unread} }}
	if _, err := l.Read(context.Background()); err != nil {
		t.Fatal(err)
	}
	replace(t, path, "new")
	settled(t, l)
	read := make(chan error, 1)
	go func() {
		_, err := l.Read(context.Background())
		read <- err
	}()
	select {
	case err := <-read:
		if !errors.Is(err, io.EOF) {
			t.Fatalf("read %v, want the end of what this cairn reads once it handed over", err)
		}
	case <-time.After(hung):
		t.Fatal("still waiting for the listening to be answered before handing over")
	}
	if c := <-got; string(c.Unread) != listen+"\n"+listing+"\n" {
		t.Errorf("handed over %q, want the listening still open, then the call", c.Unread)
	}
}

func settled(t *testing.T, l *lines) {
	t.Helper()
	msg, err := l.Read(context.Background())
	if err != nil {
		t.Fatalf("read %v, want a call the SDK answers once what was read before it is handled", err)
	}
	check, calling := msg.(*jsonrpc.Request)
	if !calling || !check.IsCall() {
		t.Fatalf("read %v, want a call the SDK answers in turn", msg)
	}
	if err := l.Write(context.Background(), &jsonrpc.Response{ID: check.ID, Result: json.RawMessage(`{}`)}); err != nil {
		t.Fatal(err)
	}
}
