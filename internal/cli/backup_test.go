package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCommands_reportAStoreTheyCannotEvenLocate(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("CAIRN_STORE", "")

	for _, args := range [][]string{
		{"backup", "add"},
		{"backup", "list"},
		{"backup", "show", "1"},
		{"backup", "delete", "1", "--yes"},
		{"backup", "restore", "--yes"},
		{"backup", "restore", "1", "--yes"},
		{"migration", "list"},
		{"migration", "apply"},
		{"migration", "show", "1"},
		{"migration", "rollback", "--yes"},
		{"task", "list"},
	} {
		if got := run(t, args...).code; got != exitStore {
			t.Errorf("%v: exit %d, want %d when the store cannot even be located", args, got, exitStore)
		}
	}
}

func TestBackupGroup_needsACommand(t *testing.T) {
	for _, group := range []string{"backup", "migration"} {
		r := run(t, group)
		if r.code == exitOK {
			t.Errorf("%s on its own should fail", group)
		}
		if !strings.Contains(strings.ToLower(r.stderr), "usage") {
			t.Errorf("%s should say how it is used, got %q", group, r.stderr)
		}
	}
}

func TestBackupCommands_refuseAnArgumentThatIsNotANumber(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	for _, args := range [][]string{
		{"backup", "show", "spec"},
		{"backup", "delete", "spec", "--yes"},
		{"backup", "restore", "spec", "--yes"},
	} {
		if got := run(t, args...).code; got != exitUsage {
			t.Errorf("%v: exit %d, want %d", args, got, exitUsage)
		}
	}
}

func TestBackupList_answersAMachineAsAMachine(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "add", "Write the spec")
	mustRun(t, "backup", "add")

	out := mustRun(t, "backup", "list", "--json").stdout
	var listed []map[string]any
	if err := json.Unmarshal([]byte(out), &listed); err != nil {
		t.Fatalf("stdout is not JSON (%v): %s", err, out)
	}
	if len(listed) != 1 || listed[0]["number"] != float64(1) {
		t.Errorf("listed = %v, want the one backup", listed)
	}
}

func TestBackupAddAndShow_answerAMachineAsAMachine(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "add", "Write the spec")

	made := mustRun(t, "backup", "add", "--json").stdout
	var one map[string]any
	if err := json.Unmarshal([]byte(made), &one); err != nil {
		t.Fatalf("stdout is not JSON (%v): %s", err, made)
	}

	shown := mustRun(t, "backup", "show", "1", "--json").stdout
	if err := json.Unmarshal([]byte(shown), &one); err != nil {
		t.Fatalf("stdout is not JSON (%v): %s", err, shown)
	}
	if one["holds"] != float64(1) {
		t.Errorf("holds = %v, want 1", one["holds"])
	}
}

func TestBackupDelete_answersAMachineAsAMachine(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "add", "Write the spec")
	mustRun(t, "backup", "add")
	mustRun(t, "backup", "restore", "1", "--yes")
	mustRun(t, "task", "delete", "1", "--yes", "--json")
}

func TestBackupRestore_saysWhichBackupItPutBack(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "add", "Write the spec")
	mustRun(t, "backup", "add")

	out := mustRun(t, "backup", "restore", "1", "--yes").stdout
	if !strings.Contains(out, "backup 1") {
		t.Errorf("stdout = %q, want it to name the backup it put back", out)
	}
}

func TestBackupDelete_saysWhichBackupItDeleted(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "add", "Write the spec")
	mustRun(t, "backup", "add")

	out := mustRun(t, "backup", "delete", "1", "--yes").stdout
	if !strings.Contains(out, "backup 1") {
		t.Errorf("stdout = %q, want it to name the backup it deleted", out)
	}
}

func TestBackupAdd_namesTheBackupItMade(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "add", "Write the spec")

	out := mustRun(t, "backup", "add").stdout
	if !strings.Contains(out, "#1") {
		t.Errorf("stdout = %q, want it to name the backup it made", out)
	}
}
