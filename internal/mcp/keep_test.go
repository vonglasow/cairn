package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/engine"
)

const showing = `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"workspace_show","arguments":{}}}`

func TestServe_handsOverTheStoreItStartedOn(t *testing.T) {
	path := onDiskAt(t, "old")
	got := handingOverTo(t, nil)
	w := serving(t, server(t, path))
	w.say(greeting, greeted, pinging)
	w.answerTo(7)
	replace(t, path, "new")
	w.say(listing)
	select {
	case c := <-got:
		if c.Kept == nil || c.Kept.Store.At != os.Getenv("CAIRN_STORE") {
			t.Errorf("handed over %+v, want the store it started on", c.Kept)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("never handed over")
	}
}

func TestServe_answersFromTheStoreHandedOver(t *testing.T) {
	kept := filepath.Join(t.TempDir(), "kept", "cairn.db")
	payload, err := json.Marshal(carried{
		State: sdk.ServerSessionState{
			InitializeParams:          &sdk.InitializeParams{ProtocolVersion: "2025-06-18", ClientInfo: &sdk.Implementation{Name: "unit", Version: "7"}},
			InitializedParams:         &sdk.InitializedParams{},
			NegotiatedProtocolVersion: "2025-06-18",
		},
		Tools:  offeredNames(t),
		Unread: []byte(showing + "\n"),
		Kept:   &engine.Started{Because: "--store, given when this server started", Store: engine.Emplacement{At: kept}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(carriedVariable, string(payload))
	w := serving(t, server(t, ""))
	said := fmt.Sprint(w.answerTo(4))
	if !strings.Contains(said, kept) || !strings.Contains(said, "not from here") {
		t.Errorf("answered %s, want the store handed over, kept", said)
	}
}
