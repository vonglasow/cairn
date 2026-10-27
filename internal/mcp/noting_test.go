package mcp

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/debug"
	"cairn/internal/engine"
)

func TestServeWritesDownTheClientAndEachMethodInDebugMode(t *testing.T) {
	t.Setenv("CAIRN_STORE", filepath.Join(t.TempDir(), "cairn.db"))
	noted, err := debug.Open(t.TempDir(), "mcp serve", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clientToServer, writtenByClient := io.Pipe()
	writtenByServer, serverToClient := io.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		server := Server{Name: "cairn", Version: "0", Access: engine.Full, Home: engine.Home{Dir: t.TempDir(), Lookup: os.LookupEnv}, Noted: noted}
		_ = Serve(ctx, server, clientToServer, serverToClient)
	}()
	client := sdk.NewClient(&sdk.Implementation{Name: "acceptance", Version: "7"}, nil)
	session, err := client.Connect(ctx, &sdk.IOTransport{Reader: writtenByServer, Writer: writtenByClient}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.ListTools(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "task_list"}); err != nil {
		t.Fatal(err)
	}
	_ = session.Close()
	_ = writtenByClient.Close()
	_ = clientToServer.Close()
	<-done
	_ = noted.Close()
	read, err := os.ReadFile(noted.Path())
	if err != nil {
		t.Fatal(err)
	}
	said := string(read)
	for _, want := range []string{"mcp client acceptance 7", "mcp tools/list", "mcp tools/call task_list"} {
		if !strings.Contains(said, want) {
			t.Errorf("the file does not say %q:\n%s", want, said)
		}
	}
}
