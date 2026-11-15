package debug

import (
	"os"
	"strings"
	"testing"
	"time"
)

var moment = time.Date(2026, 10, 26, 14, 3, 4, 0, time.UTC)

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
