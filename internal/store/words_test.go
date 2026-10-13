package store

import (
	"strings"
	"testing"
)

func TestNoBackupIsSaidWithTheCommandsWord(t *testing.T) {
	if said := ErrNoBackup.Error(); !strings.Contains(said, "restore") {
		t.Errorf("no backup says %q", said)
	}
}
