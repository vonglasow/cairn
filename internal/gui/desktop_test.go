package gui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTheEntryStartsCairnByAFullPath(t *testing.T) {
	dir := t.TempDir()
	running := filepath.Join(dir, "Cellar", Known)
	linked := filepath.Join(dir, "bin", Known)
	other := filepath.Join(dir, "other", Known)
	for _, made := range []string{running, other} {
		if err := os.MkdirAll(filepath.Dir(made), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(made, nil, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(linked), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(running, linked); err != nil {
		t.Fatal(err)
	}
	leadsTo := func(to string) func(string) (string, error) {
		return func(string) (string, error) { return to, nil }
	}
	nowhere := func(string) (string, error) { return "", errors.New("not on the PATH") }
	for _, c := range []struct {
		name   string
		onPath func(string) (string, error)
		want   string
	}{
		{"the PATH leads to it", leadsTo(linked), linked},
		{"the PATH leads to another", leadsTo(other), running},
		{"the PATH leads nowhere", nowhere, running},
		{"the PATH leads to a file that is not there", leadsTo(filepath.Join(dir, "gone")), running},
	} {
		if got := Starter(running, c.onPath); got != c.want {
			t.Errorf("%s: the entry starts %q, want %q", c.name, got, c.want)
		}
	}
}

func TestAStarterFoundByARelativePathIsMadeFull(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(Known, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	full, err := filepath.Abs(Known)
	if err != nil {
		t.Fatal(err)
	}
	got := Starter(full, func(string) (string, error) { return Known, nil })
	if got != full {
		t.Errorf("the entry starts %q, want %q", got, full)
	}
}
