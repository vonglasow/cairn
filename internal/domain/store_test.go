package domain

import (
	"testing"
	"time"
)

func TestHeldText_writesBothCountsWhateverTheyAre(t *testing.T) {
	for _, tc := range []struct {
		of   Held
		want string
	}{
		{Held{}, "holds 0 task(s) and 0 objective(s)"},
		{Held{Tasks: 1}, "holds 1 task(s) and 0 objective(s)"},
		{Held{Tasks: 4, Objectives: 2}, "holds 4 task(s) and 2 objective(s)"},
	} {
		if got := HeldText(tc.of); got != tc.want {
			t.Errorf("HeldText(%+v) = %q, want %q", tc.of, got, tc.want)
		}
	}
}

func TestParseMoment_readsBackWhatFormatMomentWrote(t *testing.T) {
	at := time.Date(2026, 10, 1, 9, 0, 0, 0, time.Local)
	back, err := ParseMoment(FormatMoment(at))
	if err != nil || !back.Equal(at) {
		t.Errorf("ParseMoment = %v, %v, want %v", back, err, at)
	}
	if _, err := ParseMoment("yesterday"); err == nil {
		t.Error("a word was read as a moment")
	}
}

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
