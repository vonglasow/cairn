package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readThis(t *testing.T, content string) (Places, error) {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return Read(home, func(string) (string, bool) { return "", false })
}

func TestRead_aKeyUnderAnotherTableIsPassedOverSayingTheTableItBelongsUnder(t *testing.T) {
	for _, c := range []struct{ content, names, belongs string }{
		{"[pages]\nkeep_sizes = true\n", "pages.keep_sizes", "[window]"},
		{"[window]\nsession = \"1h\"\n", "window.session", "[focus]"},
		{"[time]\nstore = \"/x/cairn.db\"\n", "time.store", "before any table"},
	} {
		read, err := readThis(t, c.content)
		said := strings.Join(read.Ignored(), "\n")
		if err != nil || !strings.Contains(said, c.names) || !strings.Contains(said, c.belongs) {
			t.Errorf("%q: Read = %v, ignoring %q; want it naming %s and saying %s", c.content, err, said, c.names, c.belongs)
		}
	}
}
