package engine

import (
	"cairn/internal/engine/backup"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func storeAt(t *testing.T, titles ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for _, title := range titles {
		if _, err := e.add(title, Changes{}); err != nil {
			t.Fatalf("Add: %v", err)
		}
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return path
}

func TestBackups_saysThereIsNoBackupToPutBack(t *testing.T) {
	if _, err := backupsAt(storeAt(t)).Latest(); !errors.Is(err, backup.ErrNoBackup) {
		t.Errorf("err = %v, want ErrNoBackup", err)
	}
}

func behindStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := e.rollBack(path); err != nil {
		t.Fatalf("RollBack: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return path
}

func TestEngine_reportsAPlaceItCannotBackupBeforeChanging(t *testing.T) {
	e := openEngine(t)
	blocked := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(blocked, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	nowhere := filepath.Join(blocked, "nested", "cairn.db")

	if _, err := e.rollBack(nowhere); err == nil {
		t.Error("RollBack should report a backup it could not make")
	}
	if _, err := e.applyAll(nowhere); err == nil {
		t.Error("ApplyAll should report a backup it could not make")
	}
	if _, err := e.apply(nowhere, 2); err == nil {
		t.Error("Apply should report a backup it could not make")
	}
}
