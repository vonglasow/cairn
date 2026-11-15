package sqlite

import (
	"cairn/internal/engine/backup"
	"strings"
	"testing"
)

func TestNoBackupIsSaidWithTheCommandsWord(t *testing.T) {
	if said := backup.ErrNoBackup.Error(); !strings.Contains(said, "restore") {
		t.Errorf("no backup says %q", said)
	}
}
