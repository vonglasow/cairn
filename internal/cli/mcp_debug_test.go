package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/engine"
)

func TestMcpServeWritesItsMethodsDownInDebugMode(t *testing.T) {
	store := filepath.Join(t.TempDir(), "cairn.db")
	t.Setenv("CAIRN_STORE", store)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_DEBUG", "1")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clientToServer, writtenByClient := io.Pipe()
	writtenByServer, serverToClient := io.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		Main(ctx, []string{"mcp", "serve"}, clientToServer, serverToClient, io.Discard)
	}()
	client := sdk.NewClient(&sdk.Implementation{Name: "acceptance", Version: "0"}, nil)
	session, err := client.Connect(ctx, &sdk.IOTransport{Reader: writtenByServer, Writer: writtenByClient}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.ListTools(ctx, nil); err != nil {
		t.Fatal(err)
	}
	_ = session.Close()
	_ = writtenByClient.Close()
	_ = clientToServer.Close()
	<-done
	folder := engine.Places{Store: store}.Debug()
	entries, err := os.ReadDir(folder)
	if err != nil || len(entries) != 1 {
		t.Fatalf("the debug folder holds %v, %v", entries, err)
	}
	read, err := os.ReadFile(filepath.Join(folder, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(read), "mcp tools/list") {
		t.Errorf("the server's methods are not written down:\n%s", read)
	}
}
