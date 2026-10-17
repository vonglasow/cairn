package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func readmeConfiguration(t *testing.T) string {
	t.Helper()
	said, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, after, found := strings.Cut(string(said), "## Configuration")
	if !found {
		t.Fatal("the README has no Configuration section")
	}
	_, block, found := strings.Cut(after, "```toml\n")
	if !found {
		t.Fatal("the Configuration section of the README shows no file")
	}
	block, _, _ = strings.Cut(block, "```")
	return block
}

func TestTheReadmeShowsTheConfigurationConfigInitWrites(t *testing.T) {
	if block := readmeConfiguration(t); block != Example {
		t.Errorf("the README shows:\n%s\nand config init writes:\n%s", block, Example)
	}
}

func TestTheExampleHoldsEveryKeyCairnReads(t *testing.T) {
	block := Example
	keys := acceptedKeys()
	pages := reflect.TypeOf(pages{})
	for i := range pages.NumField() {
		keys = append(keys, pages.Field(i).Tag.Get("toml"))
	}
	for _, key := range keys {
		if !strings.Contains(block, key) {
			t.Errorf("the example says nothing of %q", key)
		}
	}
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(block), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(home, func(string) (string, bool) { return "", false }); err != nil {
		t.Errorf("the example is not a configuration cairn reads: %v", err)
	}
}
