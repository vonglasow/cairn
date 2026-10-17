package cli

import (
	"testing"

	"cairn/internal/engine"
	claudeplugin "cairn/plugin/claude"
)

func TestMCPServe_bringsThePluginBesideTheStoreToItsVersion(t *testing.T) {
	store := storeFor(t)
	t.Setenv("CAIRN_STORE", store)
	if err := claudeplugin.Write(engine.PluginBeside(store), "v0.0.1"); err != nil {
		t.Fatal(err)
	}
	run(t, "mcp", "serve")
	if got, _ := claudeplugin.WrittenAt(engine.PluginBeside(store)); got != stamps().Version {
		t.Errorf("after the server started, the plugin says %q, and the server is %q", got, stamps().Version)
	}
}
