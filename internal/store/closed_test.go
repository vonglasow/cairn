package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

func TestEveryWriteSaysWhatFailed(t *testing.T) {
	s := openStore(t)
	kept := keep(t, s, "Write the spec")
	goal, err := objective.NewObjective("Ship it")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	n := factsOf(kept).Number
	read := func(err error) error { return err }
	for what, err := range map[string]error{
		"taking up task 1":         s.AddFocus(n),
		"letting go of task 1":     s.DropFocus(n),
		"letting go of everything": s.ClearFocus(),
		"":                         s.Delete(n),
		"objective":                read(func() error { _, err := s.KeepObjective(goal); return err }()),
		"changing the objective":   s.UpdateObjective(goal),
		"deleting the objective":   s.DeleteObjective(1),
		"reading objective 1":      read(func() error { _, err := s.FindObjective(1); return err }()),
		"reading the labels":       read(func() error { _, err := s.AllObjectives(); return err }()),
		"reading the labels: ":     read(func() error { _, err := s.TasksServing(1); return err }()),
		"the task":                 read(func() error { _, err := s.Keep(kept); return err }()),
		"task":                     s.Update(kept),
		"reading":                  read(func() error { _, err := s.Find(n); return err }()),
		"reading the":              read(func() error { _, err := s.All(); return err }()),
	} {
		if err == nil {
			t.Errorf("a closed store took %q without a word", what)
			continue
		}
		if !strings.Contains(err.Error(), what) {
			t.Errorf("a closed store said %q, which does not name %q", err, what)
		}
	}
}

func TestTheEndOfAListToldChangesNothing(t *testing.T) {
	b := &boundTask{labels: []string{"x"}, blockers: []int{2}, objectives: []int{3}}
	if len(b.labels) != 1 || len(b.blockers) != 1 || len(b.objectives) != 1 {
		t.Errorf("the end of a list changed it: %+v", b)
	}
	o := &boundObjective{labels: []string{"x"}}
	if len(o.labels) != 1 {
		t.Errorf("the end of a list changed it: %+v", o)
	}
}

func TestABackupSaysWhyItCouldNotBeMade(t *testing.T) {
	v, path := vaultOn(t, "Write the spec")
	if _, err := v.Make(); err != nil {
		t.Fatal(err)
	}
	counter := filepath.Join(filepath.Dir(path), backupFolder, counterFile)
	if err := os.WriteFile(counter, []byte("not a number"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := v.Make(); err == nil || !strings.Contains(err.Error(), "reading the last backup number") {
		t.Errorf("a counter that is not a number answered %v", err)
	}
	if err := v.Delete(7); !errors.Is(err, ErrNoSuchBackup) {
		t.Errorf("deleting backup 7 answered %v", err)
	}
	if err := v.Restore(7); !errors.Is(err, ErrNoSuchBackup) {
		t.Errorf("putting back backup 7 answered %v", err)
	}
	if _, err := v.Holds(domain.Backup{Number: 9, Path: filepath.Join(t.TempDir(), "absent.db")}); err == nil || !strings.Contains(err.Error(), "reading backup 9") {
		t.Errorf("reading a backup that is not there answered %v", err)
	}
	blocked := filepath.Join(t.TempDir(), "cairn.db")
	held(t, blocked, "Ship it")
	if err := os.WriteFile(filepath.Join(filepath.Dir(blocked), backupFolder), []byte("a file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewVault(blocked).Make(); err == nil || !strings.Contains(err.Error(), "making room for the backups") {
		t.Errorf("backups where a file stands answered %v", err)
	}
	if _, err := NewVault(blocked).All(); err == nil || !strings.Contains(err.Error(), "reading the backups") {
		t.Errorf("listing backups where a file stands answered %v", err)
	}
}
