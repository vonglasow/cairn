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
		{"backup", "new"},
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
		if got := run(t, args...).Code; got != exitStore {
			t.Errorf("%v: exit %d, want %d when the store cannot even be located", args, got, exitStore)
		}
	}
}

// The server is not one call but a conversation, and the store is located at
// each call rather than once at the start: a store nobody can locate is what
// each call answers, and `config_show` — the one call that says why — is still
// reachable. It used to refuse to start, which left an agent nothing to ask.
func TestMCPServe_startsWhereTheStoreCannotBeLocated(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("CAIRN_STORE", "")
	if got := run(t, "mcp", "serve").Code; got != exitOK {
		t.Errorf("mcp serve: exit %d, want %d: the store is located at each call", got, exitOK)
	}
}

func TestBackupGroup_needsACommand(t *testing.T) {
	for _, group := range []string{"backup", "migration"} {
		r := run(t, group)
		if r.Code == exitOK {
			t.Errorf("%s on its own should fail", group)
		}
		if !strings.Contains(strings.ToLower(r.Stderr), "usage") {
			t.Errorf("%s should say how it is used, got %q", group, r.Stderr)
		}
	}
}

func TestBackupCommands_refuseAnArgumentThatIsNotANumber(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_STORE", storeFor(t))
	for _, args := range [][]string{
		{"backup", "show", "spec"},
		{"backup", "delete", "spec", "--yes"},
		{"backup", "restore", "spec", "--yes"},
	} {
		if got := run(t, args...).Code; got != exitUsage {
			t.Errorf("%v: exit %d, want %d", args, got, exitUsage)
		}
	}
}

func TestBackupList_answersAMachineAsAMachine(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "new", "Write the spec")
	mustRun(t, "backup", "new")

	out := mustRun(t, "backup", "list", "--json").Stdout
	var listing struct {
		Backups []map[string]any `json:"backups"`
	}
	if err := json.Unmarshal([]byte(out), &listing); err != nil {
		t.Fatalf("stdout is not JSON (%v): %s", err, out)
	}
	if len(listing.Backups) != 1 || listing.Backups[0]["number"] != float64(1) {
		t.Errorf("listed = %v, want the one backup", listing.Backups)
	}
}

func TestBackupAddAndShow_answerAMachineAsAMachine(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "new", "Write the spec")

	made := mustRun(t, "backup", "new", "--json").Stdout
	var one map[string]any
	if err := json.Unmarshal([]byte(made), &one); err != nil {
		t.Fatalf("stdout is not JSON (%v): %s", err, made)
	}

	shown := mustRun(t, "backup", "show", "1", "--json").Stdout
	if err := json.Unmarshal([]byte(shown), &one); err != nil {
		t.Fatalf("stdout is not JSON (%v): %s", err, shown)
	}
	holds, ok := one["holds"].(map[string]any)
	if !ok {
		t.Fatalf("holds = %v, want the two counts", one["holds"])
	}
	if holds["tasks"] != float64(1) || holds["objectives"] != float64(0) {
		t.Errorf("holds = %v, want 1 task and no objective", holds)
	}
}

func TestBackupDelete_answersAMachineAsAMachine(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "new", "Write the spec")
	mustRun(t, "backup", "new")
	mustRun(t, "backup", "restore", "1", "--yes")
	mustRun(t, "task", "delete", "1", "--yes", "--json")
}

func TestBackupRestore_saysWhichBackupItPutBack(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "new", "Write the spec")
	mustRun(t, "backup", "new")

	out := mustRun(t, "backup", "restore", "1", "--yes").Stdout
	if !strings.Contains(out, "backup 1") {
		t.Errorf("stdout = %q, want it to name the backup it put back", out)
	}
}

func TestBackupDelete_saysWhichBackupItDeleted(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "new", "Write the spec")
	mustRun(t, "backup", "new")

	out := mustRun(t, "backup", "delete", "1", "--yes").Stdout
	if !strings.Contains(out, "backup 1") {
		t.Errorf("stdout = %q, want it to name the backup it deleted", out)
	}
}

func TestBackupAdd_namesTheBackupItMade(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "new", "Write the spec")

	out := mustRun(t, "backup", "new").Stdout
	if !strings.Contains(out, "#1") {
		t.Errorf("stdout = %q, want it to name the backup it made", out)
	}
}
