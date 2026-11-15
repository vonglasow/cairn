package engine

import (
	"cairn/internal/engine/migration"
	"cairn/internal/engine/task"
	"errors"
	"strings"
	"testing"
)

func TestABehindStoreNamesWhatToTypeApart(t *testing.T) {
	told := Told(migration.ErrStoreBehind)
	said, command, ok := ToType(told)
	if !ok || command != "cairn migration apply" {
		t.Fatalf("the command to type is %q (%v), want cairn migration apply", command, ok)
	}
	if strings.Contains(said, "migration") || !strings.HasPrefix(said, migration.ErrStoreBehind.Error()) {
		t.Errorf("what is said apart from the command is %q", said)
	}
	if !strings.Contains(told.Error(), "run `cairn migration apply`") || !errors.Is(told, migration.ErrStoreBehind) {
		t.Errorf("the whole refusal reads %q, and a machine reads it whole", told)
	}
	if _, _, ok := ToType(task.ErrNoSuchTask); ok {
		t.Errorf("a refusal that names no command says one")
	}
}
