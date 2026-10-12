package store

import (
	"strings"
	"testing"
)

// A backup is restored, by the command named so: what the store says when
// there is none uses the same word.
func TestNoBackupIsSaidWithTheCommandsWord(t *testing.T) {
	if said := ErrNoBackup.Error(); !strings.Contains(said, "restore") {
		t.Errorf("no backup says %q", said)
	}
}
