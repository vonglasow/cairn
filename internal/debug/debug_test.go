package debug

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

var moment = time.Date(2026, 10, 26, 14, 3, 4, 0, time.UTC)

func names(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		out = append(out, e.Name())
	}
	slices.Sort(out)
	return out
}

func TestOpen_writesAFileOfItsOwnInTheFolderItMakes(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "debug")
	one, err := Open(dir, "task list", moment)
	if err != nil {
		t.Fatal(err)
	}
	two, err := Open(dir, "task list", moment)
	if err != nil {
		t.Fatal(err)
	}
	_ = one.Close()
	_ = two.Close()
	if one.Path() == two.Path() {
		t.Fatalf("two runs at the same moment share %s", one.Path())
	}
	if filepath.Dir(one.Path()) != dir || !strings.HasPrefix(filepath.Base(one.Path()), "20261026-140304.000") || !strings.Contains(one.Path(), "task-list") {
		t.Errorf("the file is %s", one.Path())
	}
	if got := len(names(t, dir)); got != 2 {
		t.Errorf("the folder holds %d files", got)
	}
}

func TestOpen_keepsTheLastFilesOnly(t *testing.T) {
	dir := t.TempDir()
	for i := range Kept {
		if err := os.WriteFile(filepath.Join(dir, "2000010"+string(rune('0'+i/10))+"-00000"+string(rune('0'+i%10))+"-old.log"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	oldest := names(t, dir)[0]
	l, err := Open(dir, "gui", moment)
	if err != nil {
		t.Fatal(err)
	}
	_ = l.Close()
	kept := names(t, dir)
	if len(kept) != Kept || slices.Contains(kept, oldest) {
		t.Errorf("the folder holds %d files, the oldest still there: %v", len(kept), slices.Contains(kept, oldest))
	}
}

func TestLog_isCutAtItsBoundAndSaysSo(t *testing.T) {
	l, err := Open(t.TempDir(), "gui", moment)
	if err != nil {
		t.Fatal(err)
	}
	l.most = 64
	for range 10 {
		if n, err := l.Write([]byte("a line of sixteen\n")); err != nil || n != 18 {
			t.Fatalf("a write past the bound answered %d, %v", n, err)
		}
	}
	_ = l.Close()
	written, err := os.ReadFile(l.Path())
	if err != nil {
		t.Fatal(err)
	}
	if len(written) > 64+len(cut) || !strings.HasSuffix(string(written), cut) {
		t.Errorf("the file holds %d bytes, ending %q", len(written), written[max(0, len(written)-40):])
	}
}
