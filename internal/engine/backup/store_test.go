package backup

import (
	"testing"
)

func TestFormatSize_writesABackupsSizeTheWayAPersonReadsIt(t *testing.T) {
	for _, c := range []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{900, "900 B"},
		{48 * 1024, "48 KB"},
		{1536 * 1024, "1.5 MB"},
		{Backup{Size: 2048}.Size, "2 KB"},
	} {
		if got := FormatSize(c.bytes); got != c.want {
			t.Errorf("FormatSize(%d) = %q, want %q", c.bytes, got, c.want)
		}
	}
}
